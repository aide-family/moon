package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"sigs.k8s.io/yaml"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversion"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversioncomponent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversionimage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversionresource"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

// prepareApplicationCompositionColumns 给已经存在的应用表补上归一化模型的新列。
// ent 在已有数据上直接 ADD COLUMN NOT NULL 不会回填默认值，会报 23502；这里用
// IF NOT EXISTS + DEFAULT 预建列，新库由 ent 自己建表时跳过。
func prepareApplicationCompositionColumns(ctx context.Context, db *sql.DB) error {
	tables := map[string][]string{
		"application_versions": {
			`ALTER TABLE application_versions ADD COLUMN IF NOT EXISTS revision bigint NOT NULL DEFAULT 1`,
			`ALTER TABLE application_versions ADD COLUMN IF NOT EXISTS composition_version bigint NOT NULL DEFAULT 0`,
			`ALTER TABLE application_versions ADD COLUMN IF NOT EXISTS status_changed_at timestamptz NULL`,
			`ALTER TABLE application_versions ADD COLUMN IF NOT EXISTS deprecated_at timestamptz NULL`,
			`ALTER TABLE application_versions ADD COLUMN IF NOT EXISTS copied_from_version_uid character varying NOT NULL DEFAULT ''`,
			`UPDATE application_versions SET status_changed_at = COALESCE(published_at, created_at) WHERE status_changed_at IS NULL`,
		},
		"application_version_images": {
			`ALTER TABLE application_version_images ADD COLUMN IF NOT EXISTS component character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE application_version_images ADD COLUMN IF NOT EXISTS resource_uid character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE application_version_images ADD COLUMN IF NOT EXISTS container_path character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE application_version_images ADD COLUMN IF NOT EXISTS project character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE application_version_images ADD COLUMN IF NOT EXISTS source character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE application_version_images ADD COLUMN IF NOT EXISTS resolved_image character varying NOT NULL DEFAULT ''`,
		},
		"application_release_operations": {
			`ALTER TABLE application_release_operations ADD COLUMN IF NOT EXISTS application_uid character varying NOT NULL DEFAULT ''`,
			`ALTER TABLE application_release_operations ADD COLUMN IF NOT EXISTS cluster_uid character varying NOT NULL DEFAULT ''`,
		},
	}
	for table, statements := range tables {
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			continue
		}
		for _, statement := range statements {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				slog.Default().Error(fmt.Sprintf("schema: %v", err))
				return merr.Server("request failed")
			}
		}
	}
	return nil
}

// backfillApplicationComposition 把尚未归一化的版本回填成「default 组件 + 资源行」。
// Secret 值与 spec 保持原样，不需要解密。回填失败时保留 composition_version=0，
// 读取路径继续回退 spec，不影响已有版本。
func backfillApplicationComposition(ctx context.Context, client *ent.Client) error {
	rows, err := client.ApplicationVersion.Query().Where(
		applicationversion.CompositionVersionNEQ(normalizedComposition),
	).All(ctx)
	if err != nil {
		return databaseErr(slog.Default(), err)
	}
	for _, row := range rows {
		if err := backfillVersionComposition(ctx, client, row); err != nil {
			return err
		}
	}
	return nil
}

func backfillVersionComposition(ctx context.Context, client *ent.Client, row *ent.ApplicationVersion) error {
	var spec versionSpec
	_ = json.Unmarshal([]byte(row.Spec), &spec)
	tx, err := client.Tx(ctx)
	if err != nil {
		return databaseErr(slog.Default(), err)
	}
	defer func() { _ = tx.Rollback() }()
	// 重复执行或上次中途失败时先清掉不完整的内容行，再按 spec 重建。
	if _, err := tx.ApplicationVersionResource.Delete().Where(
		applicationversionresource.VersionUIDEQ(row.UID),
	).Exec(ctx); err != nil {
		return databaseErr(slog.Default(), err)
	}
	if _, err := tx.ApplicationVersionComponent.Delete().Where(
		applicationversioncomponent.VersionUIDEQ(row.UID),
	).Exec(ctx); err != nil {
		return databaseErr(slog.Default(), err)
	}
	component, err := tx.ApplicationVersionComponent.Create().
		SetUID(id.New()).SetNamespaceUID(row.NamespaceUID).SetApplicationUID(row.ApplicationUID).
		SetVersionUID(row.UID).SetName(defaultComponentName).
		SetDisplayName(defaultComponentName).SetSortOrder(0).Save(ctx)
	if err != nil {
		return databaseErr(slog.Default(), err)
	}
	seen := make(map[string]struct{}, len(spec.Resources))
	resourceUIDs := make(map[string]string, len(spec.Resources))
	for position, resource := range spec.Resources {
		item := resource.Item
		if item == "" {
			item = fmt.Sprintf("r-%d", position+1)
		}
		base := item
		for suffix := 1; ; suffix++ {
			if _, ok := seen[item]; !ok {
				break
			}
			item = fmt.Sprintf("%s-%d", base, suffix+1)
		}
		seen[item] = struct{}{}
		manifest := ""
		if len(resource.Manifest) > 0 {
			var value any
			if err := json.Unmarshal(resource.Manifest, &value); err == nil {
				if raw, marshalErr := yaml.Marshal(value); marshalErr == nil {
					manifest = string(raw)
				}
			}
		}
		images, _ := json.Marshal(resource.Images)
		secretKeys, _ := json.Marshal(resource.SecretKeys)
		scope := resource.Scope
		if scope == "" {
			scope = "UNKNOWN"
		}
		created, err := tx.ApplicationVersionResource.Create().
			SetUID(id.New()).SetNamespaceUID(row.NamespaceUID).SetApplicationUID(row.ApplicationUID).
			SetVersionUID(row.UID).SetComponentUID(component.UID).SetComponent(defaultComponentName).
			SetItem(item).SetManifest(manifest).SetAPIVersion(resource.APIVersion).
			SetKind(resource.Kind).SetName(resource.Name).SetNamespace(resource.Namespace).
			SetScope(scope).SetClusterScoped(resource.ClusterScoped).SetImages(string(images)).
			SetSecretKeys(string(secretKeys)).SetContentHash(bo.ApplicationManifestHash(manifest)).
			SetSortOrder(position).Save(ctx)
		if err != nil {
			return databaseErr(slog.Default(), err)
		}
		resourceUIDs[item] = created.UID
	}
	// 旧镜像行补上组件、资源关联与来源，读取时不再依赖空值回退。
	for item, resourceUID := range resourceUIDs {
		if err := tx.ApplicationVersionImage.Update().
			Where(applicationversionimage.VersionUIDEQ(row.UID), applicationversionimage.ResourceItemEQ(item)).
			SetResourceUID(resourceUID).SetComponent(defaultComponentName).Exec(ctx); err != nil {
			return databaseErr(slog.Default(), err)
		}
	}
	if err := tx.ApplicationVersionImage.Update().
		Where(applicationversionimage.VersionUIDEQ(row.UID), applicationversionimage.SourceEQ(""), applicationversionimage.External(true)).
		SetSource("EXTERNAL").Exec(ctx); err != nil {
		return databaseErr(slog.Default(), err)
	}
	if err := tx.ApplicationVersionImage.Update().
		Where(applicationversionimage.VersionUIDEQ(row.UID), applicationversionimage.SourceEQ(""), applicationversionimage.External(false)).
		SetSource("REGISTRY").Exec(ctx); err != nil {
		return databaseErr(slog.Default(), err)
	}
	revision := row.Revision
	if revision < 1 {
		revision = 1
	}
	if _, err := tx.ApplicationVersion.UpdateOne(row).
		SetCompositionVersion(normalizedComposition).SetRevision(revision).Save(ctx); err != nil {
		return databaseErr(slog.Default(), err)
	}
	if err := tx.Commit(); err != nil {
		return databaseErr(slog.Default(), err)
	}
	return nil
}
