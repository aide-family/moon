package data

import (
	"context"
	"log/slog"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversioncomponent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversionimage"
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

// 版本允许 0 个服务/0 个资源；服务可以先建出来再加资源，删除唯一服务后回到 0 个组件。
// 历史空组件清理只删除没有任何资源的组件行。
func TestApplicationCompositionAllowsEmptyVersionAgainstRealDatabase(t *testing.T) {
	dsn := testDatabaseURL(t)
	db := openTestDB(t, dsn)
	ctx := context.Background()
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	if err := ApplySchema(ctx, client, db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	box, err := newSecretBox("application-composition-empty-test")
	if err != nil {
		t.Fatal(err)
	}
	repo := NewApplicationRepo(&Dependencies{db: db, client: client, box: box}, slog.Default())
	namespaceUID, applicationUID := id.New(), id.New()
	if _, err := client.Application.Create().
		SetUID(applicationUID).SetNamespaceUID(namespaceUID).
		SetName("appstore-empty-" + applicationUID[:8]).Save(ctx); err != nil {
		t.Fatalf("create application: %v", err)
	}

	empty, err := repo.CreateVersion(ctx, bo.ApplicationVersion{
		NamespaceUID: namespaceUID, ApplicationUID: applicationUID, Version: "1.0.0",
	})
	if err != nil {
		t.Fatalf("create empty version: %v", err)
	}
	if empty.CompositionVersion != normalizedComposition || len(empty.Components) != 0 ||
		empty.ComponentCount != 0 || empty.ResourceCount != 0 {
		t.Fatalf("empty version = %+v", empty)
	}

	// 只有一个没有资源的服务，备注以 Markdown 原样保存。
	withComponent := *empty
	withComponent.Components = []bo.ApplicationVersionComponent{{
		Name: "worker", DisplayName: "worker", Description: "## worker\n负责后台任务",
	}}
	saved, err := repo.SaveVersionContent(ctx, withComponent, empty.Revision)
	if err != nil {
		t.Fatalf("save empty component: %v", err)
	}
	if len(saved.Components) != 1 || saved.Components[0].Name != "worker" ||
		saved.Components[0].UID == "" || saved.Components[0].Description != "## worker\n负责后台任务" {
		t.Fatalf("saved component = %+v", saved.Components)
	}

	// 删除唯一服务：组件行与资源行一起清空，版本回到 0 个组件。
	withoutComponent := *saved
	withoutComponent.Components = nil
	cleared, err := repo.SaveVersionContent(ctx, withoutComponent, saved.Revision)
	if err != nil {
		t.Fatalf("clear components: %v", err)
	}
	if len(cleared.Components) != 0 || cleared.ComponentCount != 0 || cleared.ResourceCount != 0 {
		t.Fatalf("cleared version = %+v", cleared)
	}
	componentRows, err := client.ApplicationVersionComponent.Query().
		Where(applicationversioncomponent.VersionUIDEQ(cleared.UID)).Count(ctx)
	if err != nil || componentRows != 0 {
		t.Fatalf("component rows = %d err = %v", componentRows, err)
	}

	// 历史空组件：直接插入一行没有资源的组件，清理后应删除；带资源的组件保留。
	orphan, err := client.ApplicationVersionComponent.Create().
		SetUID(id.New()).SetNamespaceUID(namespaceUID).SetApplicationUID(applicationUID).
		SetVersionUID(cleared.UID).SetName("legacy-default").SetDisplayName("legacy-default").Save(ctx)
	if err != nil {
		t.Fatalf("create orphan component: %v", err)
	}
	keptVersion, err := repo.CreateVersion(ctx, bo.ApplicationVersion{
		NamespaceUID: namespaceUID, ApplicationUID: applicationUID, Version: "1.0.1",
		Components: []bo.ApplicationVersionComponent{{
			Name: "gateway", DisplayName: "gateway",
			Resources: []bo.ApplicationResource{{
				Item: "api", Component: "gateway", Manifest: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: api\n",
				APIVersion: "v1", Kind: "ConfigMap", Name: "api", Scope: "NAMESPACED",
			}},
		}},
	})
	if err != nil {
		t.Fatalf("create kept version: %v", err)
	}
	if err := cleanupEmptyVersionComponents(ctx, db); err != nil {
		t.Fatalf("cleanup empty components: %v", err)
	}
	if ok, err := client.ApplicationVersionComponent.Query().Where(
		applicationversioncomponent.UIDEQ(orphan.UID),
	).Exist(ctx); err != nil || ok {
		t.Fatalf("orphan component still exists: ok = %v err = %v", ok, err)
	}
	keptComponents, err := client.ApplicationVersionComponent.Query().
		Where(applicationversioncomponent.VersionUIDEQ(keptVersion.UID)).Count(ctx)
	if err != nil || keptComponents != 1 {
		t.Fatalf("kept component rows = %d err = %v", keptComponents, err)
	}
}

// 旧库没有 description 列时，迁移只补一次列并清理历史空组件；后续启动不再清理，
// 否则会把用户刚建好、还没加资源的服务删掉。
func TestApplicationCompositionUpgradeCleansLegacyEmptyComponentsOnce(t *testing.T) {
	dsn := testDatabaseURL(t)
	db := openTestDB(t, dsn)
	ctx := context.Background()
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	if err := ApplySchema(ctx, client, db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	namespaceUID, applicationUID := id.New(), id.New()
	if _, err := client.Application.Create().
		SetUID(applicationUID).SetNamespaceUID(namespaceUID).
		SetName("appstore-upgrade-" + applicationUID[:8]).Save(ctx); err != nil {
		t.Fatalf("create application: %v", err)
	}
	version, err := client.ApplicationVersion.Create().
		SetUID(id.New()).SetNamespaceUID(namespaceUID).SetApplicationUID(applicationUID).
		SetVersion("1.0.0").SetCompositionVersion(normalizedComposition).Save(ctx)
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	// 模拟旧库结构：先造一行历史空组件，再去掉 description 列。
	legacy, err := client.ApplicationVersionComponent.Create().
		SetUID(id.New()).SetNamespaceUID(namespaceUID).SetApplicationUID(applicationUID).
		SetVersionUID(version.UID).SetName(defaultComponentName).SetDisplayName(defaultComponentName).Save(ctx)
	if err != nil {
		t.Fatalf("create legacy empty component: %v", err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE application_version_components DROP COLUMN IF EXISTS description`); err != nil {
		t.Fatalf("drop description: %v", err)
	}

	upgraded, err := prepareApplicationCompositionColumns(ctx, db)
	if err != nil {
		t.Fatalf("prepare columns: %v", err)
	}
	if !upgraded {
		t.Fatal("first upgrade should be detected")
	}
	if err := cleanupEmptyVersionComponents(ctx, db); err != nil {
		t.Fatalf("cleanup legacy empty components: %v", err)
	}
	if ok, err := client.ApplicationVersionComponent.Query().Where(
		applicationversioncomponent.UIDEQ(legacy.UID),
	).Exist(ctx); err != nil || ok {
		t.Fatalf("legacy empty component still exists: ok = %v err = %v", ok, err)
	}

	// 升级完成后再建的空服务不会被后续启动清理。
	draft, err := client.ApplicationVersionComponent.Create().
		SetUID(id.New()).SetNamespaceUID(namespaceUID).SetApplicationUID(applicationUID).
		SetVersionUID(version.UID).SetName("worker").SetDisplayName("worker").Save(ctx)
	if err != nil {
		t.Fatalf("create draft component: %v", err)
	}
	upgraded, err = prepareApplicationCompositionColumns(ctx, db)
	if err != nil || upgraded {
		t.Fatalf("second prepare upgraded = %v err = %v", upgraded, err)
	}
	if ok, err := client.ApplicationVersionComponent.Query().Where(
		applicationversioncomponent.UIDEQ(draft.UID),
	).Exist(ctx); err != nil || !ok {
		t.Fatalf("draft component was removed: ok = %v err = %v", ok, err)
	}
}

// 组件改名后，资源行与镜像行的 component 都要切到新名。
func TestApplicationCompositionRenamesComponentImagesAgainstRealDatabase(t *testing.T) {
	dsn := testDatabaseURL(t)
	db := openTestDB(t, dsn)
	ctx := context.Background()
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	if err := ApplySchema(ctx, client, db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	box, err := newSecretBox("application-composition-rename-test")
	if err != nil {
		t.Fatal(err)
	}
	repo := NewApplicationRepo(&Dependencies{db: db, client: client, box: box}, slog.Default())
	namespaceUID, applicationUID := id.New(), id.New()
	if _, err := client.Application.Create().
		SetUID(applicationUID).SetNamespaceUID(namespaceUID).
		SetName("appstore-rename-" + applicationUID[:8]).Save(ctx); err != nil {
		t.Fatalf("create application: %v", err)
	}

	manifest := "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: api\nspec:\n  template:\n    spec:\n      containers:\n        - name: app\n          image: harbor.local/library/api:1.0.0\n"
	created, err := repo.CreateVersion(ctx, bo.ApplicationVersion{
		NamespaceUID: namespaceUID, ApplicationUID: applicationUID, Version: "1.0.0",
		Components: []bo.ApplicationVersionComponent{{
			Name: "gateway", DisplayName: "gateway",
			Resources: []bo.ApplicationResource{{
				Item: "api", Component: "gateway", Manifest: manifest,
				APIVersion: "apps/v1", Kind: "Deployment", Name: "api", Scope: "NAMESPACED",
			}},
		}},
		Images: []bo.ApplicationVersionImage{{
			ResourceItem: "api", Component: "gateway", Container: "app",
			Image: "harbor.local/library/api:1.0.0", Source: "EXTERNAL", External: true,
		}},
	})
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if len(created.Images) != 1 || created.Images[0].Component != "gateway" {
		t.Fatalf("created images = %+v", created.Images)
	}

	// 改名时沿用组件 UID；镜像仍带着旧 component，数据层按资源视图改写成新名。
	renamed := *created
	renamed.Components[0].Name = "edge"
	renamed.Components[0].DisplayName = "edge"
	renamed.Components[0].Description = "入口"
	renamed.Components[0].Resources[0].Component = "edge"
	renamed.Resources = []bo.ApplicationResource{renamed.Components[0].Resources[0]}
	updated, err := repo.SaveVersionContent(ctx, renamed, created.Revision)
	if err != nil {
		t.Fatalf("rename component: %v", err)
	}
	if len(updated.Components) != 1 || updated.Components[0].Name != "edge" ||
		updated.Components[0].Description != "入口" {
		t.Fatalf("renamed component = %+v", updated.Components)
	}
	if updated.Components[0].Resources[0].Component != "edge" {
		t.Fatalf("resource component = %q", updated.Components[0].Resources[0].Component)
	}
	if len(updated.Images) != 1 || updated.Images[0].Component != "edge" {
		t.Fatalf("images after rename = %+v", updated.Images)
	}
	imageRows, err := client.ApplicationVersionImage.Query().
		Where(applicationversionimage.VersionUIDEQ(updated.UID)).All(ctx)
	if err != nil || len(imageRows) != 1 || imageRows[0].Component != "edge" {
		t.Fatalf("image rows = %+v err = %v", imageRows, err)
	}
}
