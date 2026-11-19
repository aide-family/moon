package data

import (
	"context"
	"log/slog"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

// 版本归一化模型依赖 SELECT ... FOR UPDATE、组件/资源 upsert 和 spec 回填，
// 只有真库能验证这些 SQL。测试在独立库上跑，先建表再回填旧版本。
//
//	ALIOTH_TEST_DATABASE_URL='postgres://alioth:alioth@127.0.0.1:5432/alioth_appstore_test?sslmode=disable' \
//	  go test ./internal/data -run TestApplicationComposition -v
func TestApplicationCompositionNormalizationAgainstRealDatabase(t *testing.T) {
	dsn := testDatabaseURL(t)
	db := openTestDB(t, dsn)
	ctx := context.Background()
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	if err := ApplySchema(ctx, client, db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	box, err := newSecretBox("application-composition-test")
	if err != nil {
		t.Fatal(err)
	}
	repo := NewApplicationRepo(&Dependencies{db: db, client: client, box: box}, slog.Default())
	namespaceUID, applicationUID := id.New(), id.New()
	if _, err := client.Application.Create().
		SetUID(applicationUID).SetNamespaceUID(namespaceUID).
		SetName("appstore-test-" + applicationUID[:8]).Save(ctx); err != nil {
		t.Fatalf("create application: %v", err)
	}

	// 旧版本只有 spec：回填后应有一个 default 组件和资源行。
	legacyUID := id.New()
	if _, err := client.ApplicationVersion.Create().
		SetUID(legacyUID).SetNamespaceUID(namespaceUID).SetApplicationUID(applicationUID).
		SetVersion("1.0.0").SetSpec(`{"resources":[{"item":"api","manifest":{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"api"}}}]}`).
		Save(ctx); err != nil {
		t.Fatalf("create legacy version: %v", err)
	}
	if err := backfillApplicationComposition(ctx, client); err != nil {
		t.Fatalf("backfill: %v", err)
	}
	legacy, err := repo.GetVersion(ctx, namespaceUID, applicationUID, legacyUID)
	if err != nil {
		t.Fatalf("get legacy version: %v", err)
	}
	if legacy.CompositionVersion != normalizedComposition || len(legacy.Components) != 1 ||
		legacy.Components[0].Name != defaultComponentName || len(legacy.Components[0].Resources) != 1 {
		t.Fatalf("legacy backfill = %+v", legacy)
	}
	if legacy.Components[0].Resources[0].UID == "" || legacy.Components[0].Resources[0].ContentHash == "" {
		t.Fatalf("legacy resource = %+v", legacy.Components[0].Resources[0])
	}

	manifest := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: api\ndata:\n  key: value\n"
	created, err := repo.CreateVersion(ctx, bo.ApplicationVersion{
		NamespaceUID: namespaceUID, ApplicationUID: applicationUID, Version: "2.0.0",
		Components: []bo.ApplicationVersionComponent{{
			Name: "gateway", DisplayName: "gateway", SortOrder: 0,
			Resources: []bo.ApplicationResource{{
				Item: "api", Component: "gateway", Manifest: manifest,
				APIVersion: "v1", Kind: "ConfigMap", Name: "api", Scope: "NAMESPACED",
				ContentHash: bo.ApplicationManifestHash(manifest),
			}},
		}},
		Resources: []bo.ApplicationResource{{
			Item: "api", Component: "gateway", Manifest: manifest,
			APIVersion: "v1", Kind: "ConfigMap", Name: "api", Scope: "NAMESPACED",
			ContentHash: bo.ApplicationManifestHash(manifest),
		}},
	})
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if created.Revision != 1 || created.ComponentCount != 1 || len(created.Components[0].Resources) != 1 {
		t.Fatalf("created version = %+v", created)
	}

	// 资源级写入合并版本级密文，并把 revision +1。
	resourceUID := created.Components[0].Resources[0].UID
	secretManifest := "apiVersion: v1\nkind: Secret\nmetadata:\n  name: api\n"
	change := bo.ApplicationResourceChange{
		NamespaceUID: namespaceUID, ApplicationUID: applicationUID, VersionUID: created.UID,
		ResourceUID: resourceUID, Resource: bo.ApplicationResource{
			UID: resourceUID, Item: "api", Component: "gateway", Manifest: secretManifest,
			APIVersion: "v1", Kind: "Secret", Name: "api", Scope: "NAMESPACED",
			ContentHash: bo.ApplicationManifestHash(secretManifest),
		},
		SecretValues:     map[string]string{"api/password": "s3cret"},
		ExpectedRevision: 1,
	}
	edited := *created
	edited.Components[0].Resources[0] = change.Resource
	edited.Resources = []bo.ApplicationResource{change.Resource}
	updated, err := repo.SaveVersionResource(ctx, edited, change)
	if err != nil {
		t.Fatalf("save resource: %v", err)
	}
	if updated.Revision != 2 {
		t.Fatalf("revision after save = %d", updated.Revision)
	}
	secrets, err := repo.GetVersionSecrets(ctx, namespaceUID, updated.UID)
	if err != nil || secrets["api/password"] != "s3cret" {
		t.Fatalf("secrets = %v err = %v", secrets, err)
	}
	if _, err := repo.SaveVersionResource(ctx, edited, change); !merr.IsConflict(err) {
		t.Fatalf("stale revision error = %v", err)
	}

	// 发布前的整包写入把镜像和内容固定下来，并再 +1 revision。
	publish := *updated
	publish.Images = nil
	frozen, err := repo.SaveVersionContent(ctx, publish, updated.Revision)
	if err != nil {
		t.Fatalf("save content: %v", err)
	}
	if frozen.Revision != 3 {
		t.Fatalf("revision after publish content = %d", frozen.Revision)
	}
	published, err := repo.UpdateVersionStatus(ctx, namespaceUID, applicationUID, updated.UID, "PUBLISHED")
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.Status != "PUBLISHED" || published.PublishedAt == nil || published.StatusChangedAt == nil {
		t.Fatalf("published version = %+v", published)
	}
}
