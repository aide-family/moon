package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"sigs.k8s.io/yaml"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/application"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversion"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversioncomponent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversionimage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversionresource"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

var _ repository.ApplicationRepo = (*applicationRepo)(nil)

// draftStatus 是唯一允许修改资源的版本状态。
const draftStatus = "DRAFT"

// normalizedComposition 是归一化内容的版本标记。
const normalizedComposition = 1

func NewApplicationRepo(deps *Dependencies, logger *slog.Logger) repository.ApplicationRepo {
	return &applicationRepo{repoBase: newBase(deps, logger)}
}

type applicationRepo struct{ *repoBase }

type versionSpec struct {
	Resources []versionSpecResource `json:"resources"`
}
type versionSpecResource struct {
	Item          string          `json:"item"`
	Component     string          `json:"component,omitempty"`
	Manifest      json.RawMessage `json:"manifest"`
	APIVersion    string          `json:"apiVersion,omitempty"`
	Kind          string          `json:"kind,omitempty"`
	Name          string          `json:"name,omitempty"`
	Namespace     string          `json:"namespace,omitempty"`
	Scope         string          `json:"scope,omitempty"`
	ClusterScoped bool            `json:"clusterScoped,omitempty"`
	Images        []string        `json:"images,omitempty"`
	SecretKeys    []string        `json:"secretKeys,omitempty"`
}

func (r *applicationRepo) CreateApplication(ctx context.Context, in bo.Application) (*bo.Application, error) {
	row, err := r.client.Application.Create().
		SetUID(id.New()).SetNamespaceUID(in.NamespaceUID).SetName(in.Name).
		SetDisplayName(in.DisplayName).SetCategory(in.Category).
		SetDescription(in.Description).SetReadme(in.Readme).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetApplication(ctx, row.NamespaceUID, row.UID)
}

func (r *applicationRepo) UpdateApplication(ctx context.Context, in bo.Application) (*bo.Application, error) {
	n, err := r.client.Application.Update().
		Where(application.UIDEQ(in.UID), application.NamespaceUIDEQ(in.NamespaceUID), application.DeletedAtIsNil()).
		SetDisplayName(in.DisplayName).SetCategory(in.Category).
		SetDescription(in.Description).SetReadme(in.Readme).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("application not found")
	}
	return r.GetApplication(ctx, in.NamespaceUID, in.UID)
}

func (r *applicationRepo) DeleteApplication(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.Application.Update().
		Where(application.UIDEQ(uid), application.NamespaceUIDEQ(namespaceUID), application.DeletedAtIsNil()).
		SetDeletedAt(time.Now().UTC()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("application not found")
	}
	return nil
}

func (r *applicationRepo) GetApplication(ctx context.Context, namespaceUID, uid string) (*bo.Application, error) {
	row, err := r.client.Application.Query().Where(
		application.UIDEQ(uid), application.NamespaceUIDEQ(namespaceUID), application.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toApplication(row)
	versions, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.NamespaceUIDEQ(namespaceUID),
		applicationversion.ApplicationUIDEQ(uid),
		applicationversion.DeletedAtIsNil(),
	).All(ctx)
	if err == nil {
		item.VersionCount = int32(len(versions))
		var latest *ent.ApplicationVersion
		for _, version := range versions {
			if version.Status == applicationversion.StatusPUBLISHED {
				if latest == nil || version.CreatedAt.After(latest.CreatedAt) {
					latest = version
				}
			}
		}
		if latest != nil {
			item.LatestVersion = latest.Version
		}
	}
	return &item, nil
}

func (r *applicationRepo) ListApplications(ctx context.Context, namespaceUID string, filter bo.ApplicationListFilter, page, pageSize int) ([]bo.Application, int, error) {
	query := r.client.Application.Query().Where(application.NamespaceUIDEQ(namespaceUID), application.DeletedAtIsNil())
	if name := strings.TrimSpace(filter.Name); name != "" {
		query.Where(application.NameContainsFold(name))
	}
	if category := strings.TrimSpace(filter.Category); category != "" {
		query.Where(application.CategoryEQ(category))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(application.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.Application, 0, len(rows))
	for _, row := range rows {
		item := toApplication(row)
		hydrated, err := r.GetApplication(ctx, namespaceUID, row.UID)
		if err == nil {
			item = *hydrated
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *applicationRepo) CreateVersion(ctx context.Context, in bo.ApplicationVersion) (*bo.ApplicationVersion, error) {
	if _, err := r.getApplicationRow(ctx, in.NamespaceUID, in.ApplicationUID); err != nil {
		return nil, err
	}
	spec, cipher, validation, err := r.encodeVersion(in)
	if err != nil {
		return nil, err
	}
	sqlTx, client, err := r.beginVersionTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sqlTx.Rollback() }()
	versionUID := id.New()
	if _, err := client.ApplicationVersion.Create().
		SetUID(versionUID).SetNamespaceUID(in.NamespaceUID).SetApplicationUID(in.ApplicationUID).
		SetVersion(in.Version).SetStatus(applicationversion.StatusDRAFT).SetChangelog(in.Changelog).
		SetSpec(spec).SetSecretsCipher(cipher).SetValidation(validation).
		SetRevision(1).SetCompositionVersion(normalizedComposition).
		SetCopiedFromVersionUID(in.CopiedFromVersionUID).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	resourceUIDs, err := r.writeComposition(ctx, client, in.NamespaceUID, in.ApplicationUID, versionUID, in.Components)
	if err != nil {
		return nil, err
	}
	if err := r.replaceVersionImages(ctx, client, in, versionUID, resourceUIDs); err != nil {
		return nil, err
	}
	if err := sqlTx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, versionUID)
}

func (r *applicationRepo) UpdateVersionChangelog(ctx context.Context, namespaceUID, applicationUID, uid, changelog string) (*bo.ApplicationVersion, error) {
	n, err := r.client.ApplicationVersion.Update().
		Where(
			applicationversion.UIDEQ(uid), applicationversion.NamespaceUIDEQ(namespaceUID),
			applicationversion.ApplicationUIDEQ(applicationUID), applicationversion.DeletedAtIsNil(),
		).SetChangelog(changelog).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("version not found")
	}
	return r.GetVersion(ctx, namespaceUID, applicationUID, uid)
}

func (r *applicationRepo) UpdateVersion(ctx context.Context, in bo.ApplicationVersion) (*bo.ApplicationVersion, error) {
	row, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.UIDEQ(in.UID), applicationversion.NamespaceUIDEQ(in.NamespaceUID),
		applicationversion.ApplicationUIDEQ(in.ApplicationUID), applicationversion.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if row.Status != applicationversion.StatusDRAFT {
		return nil, merr.Conflict("published version is immutable")
	}
	if row.CompositionVersion == normalizedComposition {
		// 归一化版本的资源只能通过资源级接口按 revision 修改，防止旧客户端整包覆盖。
		if len(in.Resources) > 0 {
			return nil, merr.Conflict("normalized version resources must be edited through the resource API")
		}
		if _, err := r.client.ApplicationVersion.UpdateOne(row).SetChangelog(in.Changelog).Save(ctx); err != nil {
			return nil, r.fail(err)
		}
		return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, in.UID)
	}
	spec, cipher, validation, err := r.encodeVersion(in)
	if err != nil {
		return nil, err
	}
	// 版本详情不会回读 Secret 值；更新时未携带新值则保留旧密文，
	// 避免只改其它资源就把已保存的 Secret 清空。
	if cipher == "" && row.SecretsCipher != "" {
		cipher = row.SecretsCipher
	}
	if _, err := r.client.ApplicationVersion.UpdateOne(row).
		SetChangelog(in.Changelog).SetSpec(spec).SetSecretsCipher(cipher).
		SetValidation(validation).AddRevision(1).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := r.ReplaceVersionImages(ctx, in.NamespaceUID, in.UID, in.Images); err != nil {
		return nil, err
	}
	return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, in.UID)
}

func (r *applicationRepo) SaveVersionContent(ctx context.Context, in bo.ApplicationVersion, expectedRevision int32) (*bo.ApplicationVersion, error) {
	sqlTx, client, err := r.beginVersionTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sqlTx.Rollback() }()
	revision, cipher, err := r.lockDraftVersion(ctx, sqlTx, in.NamespaceUID, in.ApplicationUID, in.UID, expectedRevision)
	if err != nil {
		return nil, err
	}
	resourceUIDs, err := r.writeComposition(ctx, client, in.NamespaceUID, in.ApplicationUID, in.UID, in.Components)
	if err != nil {
		return nil, err
	}
	if err := r.replaceVersionImages(ctx, client, in, in.UID, resourceUIDs); err != nil {
		return nil, err
	}
	spec, _, validation, err := r.encodeVersion(in)
	if err != nil {
		return nil, err
	}
	cipher, err = r.mergeVersionCipher(cipher, in.SecretValues)
	if err != nil {
		return nil, err
	}
	if _, err := client.ApplicationVersion.Update().Where(
		applicationversion.UIDEQ(in.UID), applicationversion.NamespaceUIDEQ(in.NamespaceUID),
	).SetChangelog(in.Changelog).SetSpec(spec).SetSecretsCipher(cipher).SetValidation(validation).
		SetRevision(int(revision) + 1).SetCompositionVersion(normalizedComposition).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := sqlTx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, in.UID)
}

func (r *applicationRepo) SaveVersionResource(ctx context.Context, in bo.ApplicationVersion, change bo.ApplicationResourceChange) (*bo.ApplicationVersion, error) {
	sqlTx, client, err := r.beginVersionTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sqlTx.Rollback() }()
	revision, cipher, err := r.lockDraftVersion(ctx, sqlTx, in.NamespaceUID, in.ApplicationUID, in.UID, change.ExpectedRevision)
	if err != nil {
		return nil, err
	}
	resourceUIDs, err := r.writeComposition(ctx, client, in.NamespaceUID, in.ApplicationUID, in.UID, in.Components)
	if err != nil {
		return nil, err
	}
	if err := r.replaceVersionImages(ctx, client, in, in.UID, resourceUIDs); err != nil {
		return nil, err
	}
	cipher, err = r.patchVersionCipher(cipher, change)
	if err != nil {
		return nil, err
	}
	spec, _, validation, err := r.encodeVersion(in)
	if err != nil {
		return nil, err
	}
	if _, err := client.ApplicationVersion.Update().Where(
		applicationversion.UIDEQ(in.UID), applicationversion.NamespaceUIDEQ(in.NamespaceUID),
	).SetSpec(spec).SetSecretsCipher(cipher).SetValidation(validation).
		SetRevision(int(revision) + 1).SetCompositionVersion(normalizedComposition).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := sqlTx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, in.UID)
}

func (r *applicationRepo) DeleteVersionResource(ctx context.Context, in bo.ApplicationVersion, resourceUID string, expectedRevision int32) (*bo.ApplicationVersion, error) {
	sqlTx, client, err := r.beginVersionTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sqlTx.Rollback() }()
	revision, cipher, err := r.lockDraftVersion(ctx, sqlTx, in.NamespaceUID, in.ApplicationUID, in.UID, expectedRevision)
	if err != nil {
		return nil, err
	}
	row, err := client.ApplicationVersionResource.Query().Where(
		applicationversionresource.UIDEQ(resourceUID),
		applicationversionresource.NamespaceUIDEQ(in.NamespaceUID),
		applicationversionresource.VersionUIDEQ(in.UID),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if _, err := r.writeComposition(ctx, client, in.NamespaceUID, in.ApplicationUID, in.UID, in.Components); err != nil {
		return nil, err
	}
	if _, err := client.ApplicationVersionImage.Delete().Where(
		applicationversionimage.VersionUIDEQ(in.UID),
		applicationversionimage.ResourceItemEQ(row.Item),
	).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	cipher, err = r.removeVersionSecretItem(cipher, row.Item)
	if err != nil {
		return nil, err
	}
	spec, _, validation, err := r.encodeVersion(in)
	if err != nil {
		return nil, err
	}
	if _, err := client.ApplicationVersion.Update().Where(
		applicationversion.UIDEQ(in.UID), applicationversion.NamespaceUIDEQ(in.NamespaceUID),
	).SetSpec(spec).SetSecretsCipher(cipher).SetValidation(validation).
		SetRevision(int(revision) + 1).SetCompositionVersion(normalizedComposition).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := sqlTx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, in.UID)
}

func (r *applicationRepo) GetVersionResource(ctx context.Context, namespaceUID, applicationUID, versionUID, resourceUID string) (*bo.ApplicationResource, error) {
	row, err := r.client.ApplicationVersionResource.Query().Where(
		applicationversionresource.UIDEQ(resourceUID),
		applicationversionresource.NamespaceUIDEQ(namespaceUID),
		applicationversionresource.ApplicationUIDEQ(applicationUID),
		applicationversionresource.VersionUIDEQ(versionUID),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toResource(row)
	return &item, nil
}

func (r *applicationRepo) DuplicateVersion(ctx context.Context, in bo.ApplicationVersionCopy) (*bo.ApplicationVersion, error) {
	source, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.UIDEQ(in.SourceVersionUID),
		applicationversion.NamespaceUIDEQ(in.NamespaceUID),
		applicationversion.ApplicationUIDEQ(in.ApplicationUID),
		applicationversion.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	sqlTx, client, err := r.beginVersionTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sqlTx.Rollback() }()
	versionUID := id.New()
	if _, err := client.ApplicationVersion.Create().
		SetUID(versionUID).SetNamespaceUID(in.NamespaceUID).SetApplicationUID(in.ApplicationUID).
		SetVersion(in.Version).SetStatus(applicationversion.StatusDRAFT).SetChangelog(in.Changelog).
		SetSpec(source.Spec).SetSecretsCipher(source.SecretsCipher).SetValidation(source.Validation).
		SetRevision(1).SetCompositionVersion(source.CompositionVersion).
		SetCopiedFromVersionUID(source.UID).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	components := []bo.ApplicationVersionComponent(nil)
	if source.CompositionVersion == normalizedComposition {
		if components, err = r.loadComposition(ctx, client, in.NamespaceUID, source.UID); err != nil {
			return nil, err
		}
		for index := range components {
			components[index].UID = ""
			for position := range components[index].Resources {
				components[index].Resources[position].UID = ""
			}
		}
		if _, err := r.writeComposition(ctx, client, in.NamespaceUID, in.ApplicationUID, versionUID, components); err != nil {
			return nil, err
		}
	}
	sourceImages, err := r.listVersionImages(ctx, client, in.NamespaceUID, source.UID)
	if err != nil {
		return nil, err
	}
	if len(sourceImages) > 0 {
		copied := bo.ApplicationVersion{
			NamespaceUID: in.NamespaceUID, ApplicationUID: in.ApplicationUID,
			Images: sourceImages,
		}
		for index := range copied.Images {
			copied.Images[index].UID = ""
			copied.Images[index].VersionUID = ""
			copied.Images[index].ResourceUID = ""
		}
		if err := r.replaceVersionImages(ctx, client, copied, versionUID, nil); err != nil {
			return nil, err
		}
	}
	if err := sqlTx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, versionUID)
}

func (r *applicationRepo) DeleteVersion(ctx context.Context, namespaceUID, applicationUID, uid string) error {
	row, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.UIDEQ(uid), applicationversion.NamespaceUIDEQ(namespaceUID),
		applicationversion.ApplicationUIDEQ(applicationUID), applicationversion.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	if row.Status == applicationversion.StatusPUBLISHED {
		return merr.Conflict("published version cannot be deleted")
	}
	sqlTx, client, err := r.beginVersionTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = sqlTx.Rollback() }()
	if _, err := client.ApplicationVersionImage.Delete().Where(applicationversionimage.VersionUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := client.ApplicationVersionResource.Delete().Where(applicationversionresource.VersionUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := client.ApplicationVersionComponent.Delete().Where(applicationversioncomponent.VersionUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := client.ApplicationVersion.Update().Where(
		applicationversion.UIDEQ(uid), applicationversion.NamespaceUIDEQ(namespaceUID),
	).SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := sqlTx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *applicationRepo) GetVersion(ctx context.Context, namespaceUID, applicationUID, uid string) (*bo.ApplicationVersion, error) {
	row, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.UIDEQ(uid), applicationversion.NamespaceUIDEQ(namespaceUID),
		applicationversion.ApplicationUIDEQ(applicationUID), applicationversion.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := r.versionToBO(row)
	components := make([]bo.ApplicationVersionComponent, 0)
	if row.CompositionVersion == normalizedComposition {
		if components, err = r.loadComposition(ctx, r.client, namespaceUID, uid); err != nil {
			return nil, err
		}
	} else {
		components = legacyComponents(item.Resources)
	}
	if images, err := r.ListVersionImages(ctx, namespaceUID, uid); err == nil {
		item.Images = images
		attachVersionImageRefs(components, images)
	}
	applyComponents(&item, components)
	return &item, nil
}

func (r *applicationRepo) ListVersions(ctx context.Context, namespaceUID, applicationUID string, filter bo.ApplicationVersionFilter, page, pageSize int) ([]bo.ApplicationVersion, int, error) {
	query := r.client.ApplicationVersion.Query().Where(
		applicationversion.NamespaceUIDEQ(namespaceUID),
		applicationversion.ApplicationUIDEQ(applicationUID),
		applicationversion.DeletedAtIsNil(),
	)
	if filter.Status != "" {
		query.Where(applicationversion.StatusEQ(applicationversion.Status(filter.Status)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(applicationversion.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.ApplicationVersion, 0, len(rows))
	for _, row := range rows {
		item := r.versionToBO(row)
		components := make([]bo.ApplicationVersionComponent, 0)
		if row.CompositionVersion == normalizedComposition {
			if components, err = r.loadComposition(ctx, r.client, namespaceUID, row.UID); err != nil {
				return nil, 0, err
			}
		} else {
			components = legacyComponents(item.Resources)
		}
		if images, err := r.ListVersionImages(ctx, namespaceUID, row.UID); err == nil {
			item.Images = images
			attachVersionImageRefs(components, images)
		}
		applyComponents(&item, components)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *applicationRepo) UpdateVersionStatus(ctx context.Context, namespaceUID, applicationUID, uid, status string) (*bo.ApplicationVersion, error) {
	now := time.Now().UTC()
	update := r.client.ApplicationVersion.Update().
		Where(applicationversion.UIDEQ(uid), applicationversion.NamespaceUIDEQ(namespaceUID),
			applicationversion.ApplicationUIDEQ(applicationUID), applicationversion.DeletedAtIsNil()).
		SetStatus(applicationversion.Status(status)).SetStatusChangedAt(now)
	switch status {
	case "PUBLISHED":
		update.SetPublishedAt(now).ClearDeprecatedAt()
	case "DEPRECATED":
		update.SetDeprecatedAt(now)
	}
	n, err := update.Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("version not found")
	}
	return r.GetVersion(ctx, namespaceUID, applicationUID, uid)
}

func (r *applicationRepo) ReplaceVersionImages(ctx context.Context, namespaceUID, versionUID string, images []bo.ApplicationVersionImage) error {
	sqlTx, client, err := r.beginVersionTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = sqlTx.Rollback() }()
	if err := r.replaceVersionImages(ctx, client, bo.ApplicationVersion{NamespaceUID: namespaceUID, Images: images}, versionUID, nil); err != nil {
		return err
	}
	if err := sqlTx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *applicationRepo) ListVersionImages(ctx context.Context, namespaceUID, versionUID string) ([]bo.ApplicationVersionImage, error) {
	return r.listVersionImages(ctx, r.client, namespaceUID, versionUID)
}

func (r *applicationRepo) listVersionImages(ctx context.Context, client *ent.Client, namespaceUID, versionUID string) ([]bo.ApplicationVersionImage, error) {
	rows, err := client.ApplicationVersionImage.Query().Where(
		applicationversionimage.NamespaceUIDEQ(namespaceUID),
		applicationversionimage.VersionUIDEQ(versionUID),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.ApplicationVersionImage, 0, len(rows))
	for _, row := range rows {
		var platforms []string
		_ = json.Unmarshal([]byte(row.Platforms), &platforms)
		source := row.Source
		if source == "" {
			if row.External {
				source = "EXTERNAL"
			} else {
				source = "REGISTRY"
			}
		}
		items = append(items, bo.ApplicationVersionImage{
			UID: row.UID, NamespaceUID: row.NamespaceUID, ApplicationUID: row.ApplicationUID,
			VersionUID: row.VersionUID, ResourceItem: row.ResourceItem, ResourceUID: row.ResourceUID,
			Component: row.Component, Container: row.Container, ContainerPath: row.ContainerPath,
			Image: row.Image, RegistryUID: row.RegistryUID, RegistryName: row.RegistryName,
			Repository: row.Repository, Tag: row.Tag, Digest: row.Digest, Project: row.Project,
			Source: source, ResolvedImage: row.ResolvedImage, Platforms: platforms,
			External: row.External, ResolvedAt: row.ResolvedAt,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].ResourceItem == items[j].ResourceItem {
			return items[i].Container < items[j].Container
		}
		return items[i].ResourceItem < items[j].ResourceItem
	})
	return items, nil
}

func (r *applicationRepo) GetVersionSecrets(ctx context.Context, namespaceUID, versionUID string) (map[string]string, error) {
	row, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.UIDEQ(versionUID), applicationversion.NamespaceUIDEQ(namespaceUID),
		applicationversion.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.decodeSecrets(row.SecretsCipher)
}

func (r *applicationRepo) getApplicationRow(ctx context.Context, namespaceUID, uid string) (*ent.Application, error) {
	row, err := r.client.Application.Query().Where(
		application.UIDEQ(uid), application.NamespaceUIDEQ(namespaceUID), application.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return row, nil
}

// transactionDriver 把已有的 *sql.Tx 包成 ent 的 dialect.Driver。ent 的某些写操作
// （例如 UpdateOne 的图更新）会自己调用 Driver.Tx；这里返回自身，Commit/Rollback
// 保持空操作，由外层 sql.Tx 统一提交或回滚。
type transactionDriver struct {
	conn entsql.Conn
}

func (driver *transactionDriver) Exec(ctx context.Context, query string, args, v any) error {
	return driver.conn.Exec(ctx, query, args, v)
}

func (driver *transactionDriver) Query(ctx context.Context, query string, args, v any) error {
	return driver.conn.Query(ctx, query, args, v)
}

func (driver *transactionDriver) Tx(context.Context) (dialect.Tx, error) { return driver, nil }

func (driver *transactionDriver) Close() error { return nil }

func (driver *transactionDriver) Dialect() string { return dialect.Postgres }

func (driver *transactionDriver) Commit() error { return nil }

func (driver *transactionDriver) Rollback() error { return nil }

// beginVersionTx 打开一个绑定 ent client 的 SQL 事务。版本内容写入需要在同一连接上
// 先 SELECT ... FOR UPDATE 再改组件/资源/密文，因此这里用 sql.Tx 构造 ent 客户端。
func (r *applicationRepo) beginVersionTx(ctx context.Context) (*sql.Tx, *ent.Client, error) {
	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, r.fail(err)
	}
	driver := &transactionDriver{conn: entsql.Conn{ExecQuerier: sqlTx}}
	return sqlTx, ent.NewClient(ent.Driver(driver)), nil
}

// lockDraftVersion 锁定版本行并校验状态与内容版本号，返回当前 revision 与密文。
func (r *applicationRepo) lockDraftVersion(ctx context.Context, sqlTx *sql.Tx, namespaceUID, applicationUID, versionUID string, expectedRevision int32) (int32, string, error) {
	var (
		revision       int32
		status, cipher string
	)
	err := sqlTx.QueryRowContext(ctx, `
SELECT revision, status::text, secrets_cipher
FROM application_versions
WHERE uid = $1 AND namespace_uid = $2 AND application_uid = $3 AND deleted_at IS NULL
FOR UPDATE`, versionUID, namespaceUID, applicationUID).Scan(&revision, &status, &cipher)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", merr.NotFound("version not found")
		}
		return 0, "", r.fail(err)
	}
	if status != draftStatus {
		return 0, "", merr.Conflict("only draft version resources can be changed")
	}
	if expectedRevision != 0 && revision != expectedRevision {
		return 0, "", merr.Conflict("version content changed; reload and retry")
	}
	return revision, cipher, nil
}

// writeComposition 用期望的组件视图覆盖版本内容：按 uid/名称匹配并 upsert 组件与
// 资源，删除不再存在的行。版本允许 0 个组件；此时清空该版本的全部组件与资源。
// 调用方负责事务与状态校验。
func (r *applicationRepo) writeComposition(ctx context.Context, client *ent.Client, namespaceUID, applicationUID, versionUID string, components []bo.ApplicationVersionComponent) (map[string]string, error) {
	if len(components) == 0 {
		if _, err := client.ApplicationVersionResource.Delete().Where(
			applicationversionresource.VersionUIDEQ(versionUID),
		).Exec(ctx); err != nil {
			return nil, r.fail(err)
		}
		if _, err := client.ApplicationVersionComponent.Delete().Where(
			applicationversioncomponent.VersionUIDEQ(versionUID),
		).Exec(ctx); err != nil {
			return nil, r.fail(err)
		}
		return map[string]string{}, nil
	}
	componentRows, err := client.ApplicationVersionComponent.Query().Where(
		applicationversioncomponent.VersionUIDEQ(versionUID),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	componentByUID := make(map[string]*ent.ApplicationVersionComponent, len(componentRows))
	componentByName := make(map[string]*ent.ApplicationVersionComponent, len(componentRows))
	for _, row := range componentRows {
		componentByUID[row.UID] = row
		componentByName[row.Name] = row
	}
	resourceRows, err := client.ApplicationVersionResource.Query().Where(
		applicationversionresource.VersionUIDEQ(versionUID),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	resourceByUID := make(map[string]*ent.ApplicationVersionResource, len(resourceRows))
	resourceByItem := make(map[string]*ent.ApplicationVersionResource, len(resourceRows))
	for _, row := range resourceRows {
		resourceByUID[row.UID] = row
		resourceByItem[row.Item] = row
	}
	keepComponents := make(map[string]struct{}, len(components))
	keepComponentUIDs := make(map[string]struct{}, len(components))
	keepResources := make(map[string]struct{})
	resourceUIDByItem := make(map[string]string, len(resourceRows))
	for position := range components {
		component := components[position]
		name := strings.TrimSpace(component.Name)
		if name == "" {
			name = defaultComponentName
		}
		if _, ok := keepComponents[name]; ok {
			return nil, merr.Invalid("duplicate component name: " + name)
		}
		keepComponents[name] = struct{}{}
		displayName := strings.TrimSpace(component.DisplayName)
		if displayName == "" {
			displayName = name
		}
		row := componentByUID[component.UID]
		if row == nil {
			row = componentByName[name]
		}
		if row != nil {
			if _, err := client.ApplicationVersionComponent.UpdateOne(row).
				SetName(name).SetDisplayName(displayName).
				SetDescription(strings.TrimSpace(component.Description)).
				SetSortOrder(int(component.SortOrder)).Save(ctx); err != nil {
				return nil, r.fail(err)
			}
		} else {
			row, err = client.ApplicationVersionComponent.Create().
				SetUID(id.New()).SetNamespaceUID(namespaceUID).SetApplicationUID(applicationUID).
				SetVersionUID(versionUID).SetName(name).SetDisplayName(displayName).
				SetDescription(strings.TrimSpace(component.Description)).
				SetSortOrder(int(component.SortOrder)).Save(ctx)
			if err != nil {
				return nil, r.fail(err)
			}
		}
		keepComponentUIDs[row.UID] = struct{}{}
		for resourcePosition := range component.Resources {
			resource := component.Resources[resourcePosition]
			item := strings.TrimSpace(resource.Item)
			if item == "" {
				item = fmt.Sprintf("r-%d", resourcePosition+1)
			}
			if _, ok := keepResources[item]; ok {
				return nil, merr.Invalid("duplicate resource item: " + item)
			}
			keepResources[item] = struct{}{}
			resource.Component = name
			existing := resourceByUID[resource.UID]
			if existing == nil {
				existing = resourceByItem[item]
			}
			images, _ := json.Marshal(resource.Images)
			secretKeys, _ := json.Marshal(resource.SecretKeys)
			contentHash := resource.ContentHash
			if contentHash == "" {
				contentHash = bo.ApplicationManifestHash(resource.Manifest)
			}
			if existing != nil {
				if _, err := client.ApplicationVersionResource.UpdateOne(existing).
					SetComponentUID(row.UID).SetComponent(name).SetItem(item).SetManifest(resource.Manifest).
					SetAPIVersion(resource.APIVersion).SetKind(resource.Kind).SetName(resource.Name).
					SetNamespace(resource.Namespace).SetScope(resource.Scope).
					SetClusterScoped(resource.ClusterScoped).SetImages(string(images)).
					SetSecretKeys(string(secretKeys)).SetContentHash(contentHash).
					SetSortOrder(resourcePosition).Save(ctx); err != nil {
					return nil, r.fail(err)
				}
				keepResources[existing.UID] = struct{}{}
				resourceUIDByItem[item] = existing.UID
				continue
			}
			created, err := client.ApplicationVersionResource.Create().
				SetUID(id.New()).SetNamespaceUID(namespaceUID).SetApplicationUID(applicationUID).
				SetVersionUID(versionUID).SetComponentUID(row.UID).SetComponent(name).SetItem(item).
				SetManifest(resource.Manifest).SetAPIVersion(resource.APIVersion).SetKind(resource.Kind).
				SetName(resource.Name).SetNamespace(resource.Namespace).SetScope(resource.Scope).
				SetClusterScoped(resource.ClusterScoped).SetImages(string(images)).
				SetSecretKeys(string(secretKeys)).SetContentHash(contentHash).
				SetSortOrder(resourcePosition).Save(ctx)
			if err != nil {
				return nil, r.fail(err)
			}
			keepResources[created.UID] = struct{}{}
			resourceUIDByItem[item] = created.UID
		}
	}
	staleResources := make([]string, 0)
	for uid := range resourceByUID {
		if _, ok := keepResources[uid]; !ok {
			staleResources = append(staleResources, uid)
		}
	}
	if len(staleResources) > 0 {
		if _, err := client.ApplicationVersionResource.Delete().Where(
			applicationversionresource.UIDIn(staleResources...),
		).Exec(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	staleComponents := make([]string, 0)
	for _, row := range componentRows {
		if _, ok := keepComponentUIDs[row.UID]; !ok {
			staleComponents = append(staleComponents, row.UID)
		}
	}
	if len(staleComponents) > 0 {
		if _, err := client.ApplicationVersionComponent.Delete().Where(
			applicationversioncomponent.UIDIn(staleComponents...),
		).Exec(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	return resourceUIDByItem, nil
}

// replaceVersionImages 替换版本的全部镜像快照，并按 item 关联资源 uid。
func (r *applicationRepo) replaceVersionImages(ctx context.Context, client *ent.Client, in bo.ApplicationVersion, versionUID string, resourceUIDByItem map[string]string) error {
	if _, err := client.ApplicationVersionImage.Delete().Where(
		applicationversionimage.VersionUIDEQ(versionUID),
	).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if resourceUIDByItem == nil {
		resourceUIDByItem = make(map[string]string, len(in.Resources))
	}
	for _, resource := range in.Resources {
		if _, ok := resourceUIDByItem[resource.Item]; !ok {
			resourceUIDByItem[resource.Item] = resource.UID
		}
	}
	if len(resourceUIDByItem) == 0 && len(in.Components) > 0 {
		for _, component := range in.Components {
			for _, resource := range component.Resources {
				resourceUIDByItem[resource.Item] = resource.UID
			}
		}
	}
	if len(resourceUIDByItem) == 0 {
		rows, err := client.ApplicationVersionResource.Query().Where(
			applicationversionresource.VersionUIDEQ(versionUID),
		).Select(applicationversionresource.FieldItem, applicationversionresource.FieldUID).All(ctx)
		if err == nil {
			for _, row := range rows {
				resourceUIDByItem[row.Item] = row.UID
			}
		}
	}
	// 组件改名后资源与镜像的 component 要一起切换到新名；以资源视图为准。
	componentByItem := make(map[string]string, len(in.Components)+len(in.Resources))
	for _, component := range in.Components {
		name := strings.TrimSpace(component.Name)
		if name == "" {
			name = defaultComponentName
		}
		for _, resource := range component.Resources {
			componentByItem[resource.Item] = name
		}
	}
	for _, resource := range in.Resources {
		if _, ok := componentByItem[resource.Item]; ok {
			continue
		}
		name := strings.TrimSpace(resource.Component)
		if name == "" {
			name = defaultComponentName
		}
		componentByItem[resource.Item] = name
	}
	applicationUID := in.ApplicationUID
	for _, image := range in.Images {
		platforms, _ := json.Marshal(image.Platforms)
		resourceUID := image.ResourceUID
		if resourceUID == "" {
			resourceUID = resourceUIDByItem[image.ResourceItem]
		}
		component := image.Component
		if name, ok := componentByItem[image.ResourceItem]; ok {
			component = name
		}
		if applicationUID == "" {
			applicationUID = image.ApplicationUID
		}
		builder := client.ApplicationVersionImage.Create().
			SetUID(id.New()).SetNamespaceUID(in.NamespaceUID).SetApplicationUID(applicationUID).
			SetVersionUID(versionUID).SetResourceItem(image.ResourceItem).SetResourceUID(resourceUID).
			SetComponent(component).SetContainer(image.Container).
			SetContainerPath(image.ContainerPath).SetImage(image.Image).
			SetRegistryUID(image.RegistryUID).SetRegistryName(image.RegistryName).
			SetRepository(image.Repository).SetTag(image.Tag).SetDigest(image.Digest).
			SetProject(image.Project).SetSource(image.Source).SetResolvedImage(image.ResolvedImage).
			SetPlatforms(string(platforms)).SetExternal(image.External)
		if image.ResolvedAt != nil {
			builder.SetResolvedAt(*image.ResolvedAt)
		}
		if _, err := builder.Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	return nil
}

func (r *applicationRepo) loadComposition(ctx context.Context, client *ent.Client, namespaceUID, versionUID string) ([]bo.ApplicationVersionComponent, error) {
	componentRows, err := client.ApplicationVersionComponent.Query().Where(
		applicationversioncomponent.NamespaceUIDEQ(namespaceUID),
		applicationversioncomponent.VersionUIDEQ(versionUID),
	).Order(ent.Asc(applicationversioncomponent.FieldSortOrder)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	resourceRows, err := client.ApplicationVersionResource.Query().Where(
		applicationversionresource.NamespaceUIDEQ(namespaceUID),
		applicationversionresource.VersionUIDEQ(versionUID),
	).Order(ent.Asc(applicationversionresource.FieldSortOrder)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	componentIndex := make(map[string]int, len(componentRows))
	components := make([]bo.ApplicationVersionComponent, 0, len(componentRows))
	for _, row := range componentRows {
		componentIndex[row.UID] = len(components)
		components = append(components, bo.ApplicationVersionComponent{
			UID: row.UID, Name: row.Name, DisplayName: row.DisplayName,
			Description: row.Description, SortOrder: int32(row.SortOrder),
			Resources: make([]bo.ApplicationResource, 0),
		})
	}
	for _, row := range resourceRows {
		position, ok := componentIndex[row.ComponentUID]
		if !ok {
			continue
		}
		resource := toResource(row)
		components[position].Resources = append(components[position].Resources, resource)
		components[position].ResourceCount = int32(len(components[position].Resources))
	}
	return components, nil
}

func toResource(row *ent.ApplicationVersionResource) bo.ApplicationResource {
	var images, secretKeys []string
	_ = json.Unmarshal([]byte(row.Images), &images)
	_ = json.Unmarshal([]byte(row.SecretKeys), &secretKeys)
	item := bo.ApplicationResource{
		UID: row.UID, Item: row.Item, Component: row.Component, Manifest: row.Manifest,
		APIVersion: row.APIVersion, Kind: row.Kind, Name: row.Name, Namespace: row.Namespace,
		Scope: row.Scope, ClusterScoped: row.ClusterScoped, Images: images,
		SecretKeys: secretKeys, ContentHash: row.ContentHash,
	}
	return item
}

// applyComponents 把组件视图写回版本对象，同时生成扁平资源列表和计数。
func applyComponents(item *bo.ApplicationVersion, components []bo.ApplicationVersionComponent) {
	item.Components = components
	resources := make([]bo.ApplicationResource, 0)
	resourceCount := int32(0)
	for _, component := range components {
		resources = append(resources, component.Resources...)
		resourceCount += int32(len(component.Resources))
	}
	item.Resources = resources
	item.ComponentCount = int32(len(components))
	item.ResourceCount = resourceCount
}

// attachVersionImageRefs 把镜像快照还原成资源上的容器引用，保证改名/整版保存
// 重新解析镜像时不会丢掉 container 与 containerPath。
func attachVersionImageRefs(components []bo.ApplicationVersionComponent, images []bo.ApplicationVersionImage) {
	refsByResource := make(map[string][]bo.ApplicationImageRef, len(images))
	for _, image := range images {
		if image.ResourceUID == "" || image.ContainerPath == "" {
			continue
		}
		refsByResource[image.ResourceUID] = append(refsByResource[image.ResourceUID], bo.ApplicationImageRef{
			Image: image.Image, Container: image.Container, ContainerPath: image.ContainerPath,
		})
	}
	for component := range components {
		for position := range components[component].Resources {
			resource := &components[component].Resources[position]
			if refs, ok := refsByResource[resource.UID]; ok {
				resource.ImageRefs = refs
			}
		}
	}
}

// legacyComponents 把 spec 里的扁平资源包成一个默认组件，供尚未回填的版本读取。
// 版本没有资源时返回空组件列表，避免旧版本继续出现空 default 服务。
func legacyComponents(resources []bo.ApplicationResource) []bo.ApplicationVersionComponent {
	if len(resources) == 0 {
		return []bo.ApplicationVersionComponent{}
	}
	for index := range resources {
		if resources[index].Component == "" {
			resources[index].Component = defaultComponentName
		}
	}
	return []bo.ApplicationVersionComponent{{
		Name: defaultComponentName, DisplayName: defaultComponentName, SortOrder: 0,
		Resources: resources, ResourceCount: int32(len(resources)),
	}}
}

const defaultComponentName = "default"

// flattenComponents 按组件顺序展开资源，组件名写回资源。
func flattenComponents(components []bo.ApplicationVersionComponent) []bo.ApplicationResource {
	resources := make([]bo.ApplicationResource, 0)
	for _, component := range components {
		name := component.Name
		if name == "" {
			name = defaultComponentName
		}
		for _, resource := range component.Resources {
			resource.Component = name
			resources = append(resources, resource)
		}
	}
	return resources
}

func (r *applicationRepo) encodeVersion(in bo.ApplicationVersion) (string, string, string, error) {
	resources := in.Resources
	if len(in.Components) > 0 {
		resources = flattenComponents(in.Components)
	}
	spec := versionSpec{Resources: make([]versionSpecResource, 0, len(resources))}
	for _, resource := range resources {
		var manifest map[string]any
		if err := yaml.Unmarshal([]byte(resource.Manifest), &manifest); err != nil {
			return "", "", "", merr.Invalid("invalid resource yaml")
		}
		raw, err := json.Marshal(manifest)
		if err != nil {
			return "", "", "", merr.Invalid("invalid resource yaml")
		}
		component := resource.Component
		if component == "" {
			component = defaultComponentName
		}
		spec.Resources = append(spec.Resources, versionSpecResource{
			Item: resource.Item, Component: component, Manifest: raw,
			APIVersion: resource.APIVersion, Kind: resource.Kind, Name: resource.Name,
			Namespace: resource.Namespace, Scope: resource.Scope,
			ClusterScoped: resource.ClusterScoped, Images: resource.Images, SecretKeys: resource.SecretKeys,
		})
	}
	rawSpec, err := json.Marshal(spec)
	if err != nil {
		return "", "", "", merr.Server("request failed")
	}
	cipher := ""
	if len(in.SecretValues) > 0 {
		raw, err := json.Marshal(in.SecretValues)
		if err != nil {
			return "", "", "", merr.Server("request failed")
		}
		cipher, err = r.box.encrypt(string(raw))
		if err != nil {
			return "", "", "", err
		}
	}
	rawValidation, err := json.Marshal(in.Validation)
	if err != nil {
		return "", "", "", merr.Server("request failed")
	}
	return string(rawSpec), cipher, string(rawValidation), nil
}

func (r *applicationRepo) mergeVersionCipher(current string, values map[string]string) (string, error) {
	if values == nil {
		return current, nil
	}
	if len(values) == 0 {
		return "", nil
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "", merr.Server("request failed")
	}
	return r.box.encrypt(string(raw))
}

func (r *applicationRepo) patchVersionCipher(current string, change bo.ApplicationResourceChange) (string, error) {
	secrets, err := r.decodeSecrets(current)
	if err != nil {
		return "", err
	}
	item := change.Resource.Item
	for _, key := range change.ClearSecretKeys {
		delete(secrets, item+"/"+key)
	}
	for key, value := range change.SecretValues {
		secrets[key] = value
	}
	if len(secrets) == 0 {
		return "", nil
	}
	raw, err := json.Marshal(secrets)
	if err != nil {
		return "", merr.Server("request failed")
	}
	return r.box.encrypt(string(raw))
}

func (r *applicationRepo) removeVersionSecretItem(current, item string) (string, error) {
	secrets, err := r.decodeSecrets(current)
	if err != nil {
		return "", err
	}
	prefix := item + "/"
	for key := range secrets {
		if strings.HasPrefix(key, prefix) {
			delete(secrets, key)
		}
	}
	if len(secrets) == 0 {
		return "", nil
	}
	raw, err := json.Marshal(secrets)
	if err != nil {
		return "", merr.Server("request failed")
	}
	return r.box.encrypt(string(raw))
}

// versionToBO 只映射版本行本身；组件与资源由调用方按 composition_version 补充。
func (r *applicationRepo) versionToBO(row *ent.ApplicationVersion) bo.ApplicationVersion {
	item := bo.ApplicationVersion{
		UID: row.UID, NamespaceUID: row.NamespaceUID, ApplicationUID: row.ApplicationUID,
		Version: row.Version, Status: string(row.Status), Changelog: row.Changelog,
		Revision: int32(row.Revision), CompositionVersion: int32(row.CompositionVersion),
		CopiedFromVersionUID: row.CopiedFromVersionUID,
		PublishedAt:          row.PublishedAt, StatusChangedAt: row.StatusChangedAt,
		DeprecatedAt: row.DeprecatedAt, CreatedAt: row.CreatedAt,
	}
	var spec versionSpec
	if err := json.Unmarshal([]byte(row.Spec), &spec); err == nil {
		for _, resource := range spec.Resources {
			manifest := ""
			if len(resource.Manifest) > 0 {
				var value any
				if err := json.Unmarshal(resource.Manifest, &value); err == nil {
					if raw, err := yaml.Marshal(value); err == nil {
						manifest = string(raw)
					}
				}
			}
			component := resource.Component
			if component == "" {
				component = defaultComponentName
			}
			item.Resources = append(item.Resources, bo.ApplicationResource{
				Item: resource.Item, Component: component, Manifest: manifest,
				APIVersion: resource.APIVersion, Kind: resource.Kind, Name: resource.Name,
				Namespace: resource.Namespace, Scope: resource.Scope,
				ClusterScoped: resource.ClusterScoped, Images: resource.Images,
				SecretKeys: resource.SecretKeys, ContentHash: bo.ApplicationManifestHash(manifest),
			})
		}
	}
	_ = json.Unmarshal([]byte(row.Validation), &item.Validation)
	return item
}

func (r *applicationRepo) decodeSecrets(cipher string) (map[string]string, error) {
	if strings.TrimSpace(cipher) == "" {
		return map[string]string{}, nil
	}
	plain, err := r.box.decrypt(cipher)
	if err != nil {
		r.log.Error(fmt.Sprintf("application secrets: %v", err))
		return nil, merr.Server("request failed")
	}
	values := map[string]string{}
	if err := json.Unmarshal([]byte(plain), &values); err != nil {
		return nil, merr.Server("request failed")
	}
	return values, nil
}

func toApplication(row *ent.Application) bo.Application {
	return bo.Application{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, DisplayName: row.DisplayName,
		Category: row.Category, Description: row.Description, Readme: row.Readme, CreatedAt: row.CreatedAt,
	}
}

func (r *applicationRepo) HasPublishedVersions(ctx context.Context, namespaceUID, applicationUID string) (bool, error) {
	ok, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.NamespaceUIDEQ(namespaceUID),
		applicationversion.ApplicationUIDEQ(applicationUID),
		applicationversion.DeletedAtIsNil(),
		applicationversion.StatusEQ(applicationversion.StatusPUBLISHED),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}

func (r *applicationRepo) FindImageReferences(ctx context.Context, namespaceUID, registryUID, repository, tag string) ([]bo.ApplicationImageReference, error) {
	images, err := r.client.ApplicationVersionImage.Query().Where(
		applicationversionimage.NamespaceUIDEQ(namespaceUID),
		applicationversionimage.RegistryUIDEQ(registryUID),
		applicationversionimage.RepositoryEQ(repository),
		applicationversionimage.TagEQ(tag),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	seen := map[string]struct{}{}
	items := make([]bo.ApplicationImageReference, 0)
	for _, image := range images {
		version, err := r.client.ApplicationVersion.Query().Where(
			applicationversion.UIDEQ(image.VersionUID), applicationversion.DeletedAtIsNil(),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, r.fail(err)
		}
		if _, ok := seen[version.UID]; ok {
			continue
		}
		seen[version.UID] = struct{}{}
		app, err := r.client.Application.Query().Where(
			application.UIDEQ(version.ApplicationUID), application.DeletedAtIsNil(),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, r.fail(err)
		}
		items = append(items, bo.ApplicationImageReference{
			ApplicationUID: app.UID, ApplicationName: app.Name,
			VersionUID: version.UID, Version: version.Version, Status: string(version.Status),
		})
	}
	return items, nil
}

func (r *applicationRepo) HasImageReferencesForRegistry(ctx context.Context, namespaceUID, registryUID string) (bool, error) {
	ok, err := r.client.ApplicationVersionImage.Query().Where(
		applicationversionimage.NamespaceUIDEQ(namespaceUID),
		applicationversionimage.RegistryUIDEQ(registryUID),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}
