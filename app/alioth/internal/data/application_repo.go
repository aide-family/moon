package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/application"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversion"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationversionimage"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
	"sigs.k8s.io/yaml"
)

var _ repository.ApplicationRepo = (*applicationRepo)(nil)

func NewApplicationRepo(deps *Dependencies, logger *slog.Logger) repository.ApplicationRepo {
	return &applicationRepo{repoBase: newBase(deps, logger)}
}

type applicationRepo struct{ *repoBase }

type versionSpec struct {
	Resources []versionSpecResource `json:"resources"`
}
type versionSpecResource struct {
	Item          string          `json:"item"`
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
		SetDisplayName(in.DisplayName).SetCategory(application.Category(in.Category)).
		SetDescription(in.Description).SetReadme(in.Readme).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetApplication(ctx, row.NamespaceUID, row.UID)
}

func (r *applicationRepo) UpdateApplication(ctx context.Context, in bo.Application) (*bo.Application, error) {
	n, err := r.client.Application.Update().
		Where(application.UIDEQ(in.UID), application.NamespaceUIDEQ(in.NamespaceUID), application.DeletedAtIsNil()).
		SetDisplayName(in.DisplayName).SetCategory(application.Category(in.Category)).
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
		query.Where(application.CategoryEQ(application.Category(category)))
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
	row, err := r.client.ApplicationVersion.Create().
		SetUID(id.New()).SetNamespaceUID(in.NamespaceUID).SetApplicationUID(in.ApplicationUID).
		SetVersion(in.Version).SetStatus(applicationversion.StatusDRAFT).SetChangelog(in.Changelog).
		SetSpec(spec).SetSecretsCipher(cipher).SetValidation(validation).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := r.ReplaceVersionImages(ctx, in.NamespaceUID, row.UID, in.Images); err != nil {
		return nil, err
	}
	return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, row.UID)
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
		SetChangelog(in.Changelog).SetSpec(spec).SetSecretsCipher(cipher).SetValidation(validation).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := r.ReplaceVersionImages(ctx, in.NamespaceUID, in.UID, in.Images); err != nil {
		return nil, err
	}
	return r.GetVersion(ctx, in.NamespaceUID, in.ApplicationUID, in.UID)
}

func (r *applicationRepo) DeleteVersion(ctx context.Context, namespaceUID, applicationUID, uid string) error {
	row, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.UIDEQ(uid), applicationversion.NamespaceUIDEQ(namespaceUID),
		applicationversion.ApplicationUIDEQ(applicationUID), applicationversion.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	if row.Status != applicationversion.StatusDRAFT {
		return merr.Conflict("only draft version can be deleted")
	}
	if _, err := r.client.ApplicationVersionImage.Delete().Where(applicationversionimage.VersionUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := r.client.ApplicationVersion.UpdateOne(row).SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
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
	images, err := r.ListVersionImages(ctx, namespaceUID, uid)
	if err == nil {
		item.Images = images
	}
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
		if images, err := r.ListVersionImages(ctx, namespaceUID, row.UID); err == nil {
			item.Images = images
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *applicationRepo) UpdateVersionStatus(ctx context.Context, namespaceUID, applicationUID, uid, status string) (*bo.ApplicationVersion, error) {
	update := r.client.ApplicationVersion.Update().
		Where(applicationversion.UIDEQ(uid), applicationversion.NamespaceUIDEQ(namespaceUID),
			applicationversion.ApplicationUIDEQ(applicationUID), applicationversion.DeletedAtIsNil()).
		SetStatus(applicationversion.Status(status))
	if status == "PUBLISHED" {
		update.SetPublishedAt(time.Now().UTC())
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
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ApplicationVersionImage.Delete().Where(applicationversionimage.VersionUIDEQ(versionUID)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	for _, image := range images {
		platforms, _ := json.Marshal(image.Platforms)
		builder := tx.ApplicationVersionImage.Create().
			SetUID(id.New()).SetNamespaceUID(namespaceUID).SetApplicationUID(image.ApplicationUID).
			SetVersionUID(versionUID).SetResourceItem(image.ResourceItem).SetContainer(image.Container).
			SetImage(image.Image).SetRegistryUID(image.RegistryUID).SetRegistryName(image.RegistryName).
			SetRepository(image.Repository).SetTag(image.Tag).SetDigest(image.Digest).
			SetPlatforms(string(platforms)).SetExternal(image.External)
		if image.ResolvedAt != nil {
			builder.SetResolvedAt(*image.ResolvedAt)
		}
		if _, err := builder.Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	return r.fail(tx.Commit())
}

func (r *applicationRepo) ListVersionImages(ctx context.Context, namespaceUID, versionUID string) ([]bo.ApplicationVersionImage, error) {
	rows, err := r.client.ApplicationVersionImage.Query().Where(
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
		items = append(items, bo.ApplicationVersionImage{
			UID: row.UID, NamespaceUID: row.NamespaceUID, ApplicationUID: row.ApplicationUID,
			VersionUID: row.VersionUID, ResourceItem: row.ResourceItem, Container: row.Container,
			Image: row.Image, RegistryUID: row.RegistryUID, RegistryName: row.RegistryName,
			Repository: row.Repository, Tag: row.Tag, Digest: row.Digest,
			Platforms: platforms, External: row.External, ResolvedAt: row.ResolvedAt,
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

func (r *applicationRepo) encodeVersion(in bo.ApplicationVersion) (string, string, string, error) {
	spec := versionSpec{Resources: make([]versionSpecResource, 0, len(in.Resources))}
	for _, resource := range in.Resources {
		var manifest map[string]any
		if err := yaml.Unmarshal([]byte(resource.Manifest), &manifest); err != nil {
			return "", "", "", merr.Invalid("invalid resource yaml")
		}
		raw, err := json.Marshal(manifest)
		if err != nil {
			return "", "", "", merr.Invalid("invalid resource yaml")
		}
		spec.Resources = append(spec.Resources, versionSpecResource{
			Item: resource.Item, Manifest: raw, APIVersion: resource.APIVersion, Kind: resource.Kind,
			Name: resource.Name, Namespace: resource.Namespace, Scope: resource.Scope,
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

func (r *applicationRepo) versionToBO(row *ent.ApplicationVersion) bo.ApplicationVersion {
	item := bo.ApplicationVersion{
		UID: row.UID, NamespaceUID: row.NamespaceUID, ApplicationUID: row.ApplicationUID,
		Version: row.Version, Status: string(row.Status), Changelog: row.Changelog,
		PublishedAt: row.PublishedAt, CreatedAt: row.CreatedAt,
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
			item.Resources = append(item.Resources, bo.ApplicationResource{
				Item: resource.Item, Manifest: manifest, APIVersion: resource.APIVersion,
				Kind: resource.Kind, Name: resource.Name, Namespace: resource.Namespace,
				Scope: resource.Scope, ClusterScoped: resource.ClusterScoped,
				Images: resource.Images, SecretKeys: resource.SecretKeys,
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
		Category: string(row.Category), Description: row.Description, Readme: row.Readme, CreatedAt: row.CreatedAt,
	}
}

func (r *applicationRepo) HasNonDraftVersions(ctx context.Context, namespaceUID, applicationUID string) (bool, error) {
	ok, err := r.client.ApplicationVersion.Query().Where(
		applicationversion.NamespaceUIDEQ(namespaceUID),
		applicationversion.ApplicationUIDEQ(applicationUID),
		applicationversion.DeletedAtIsNil(),
		applicationversion.StatusIn(applicationversion.StatusPUBLISHED, applicationversion.StatusDEPRECATED),
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
