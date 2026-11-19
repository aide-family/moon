package application

import (
	"context"
	"log/slog"
	"net/url"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewCatalog, NewComposer, NewDeployer, NewExecutor, NewInspector)

const (
	maxVersionRunes     = 32
	maxChangelog        = 4096
	maxComponentRunes   = 63
	draftVersionStatus  = "DRAFT"
	publishVersionState = "PUBLISHED"
	deprecateState      = "DEPRECATED"
)

// Catalog 管理应用与发布包版本：版本状态机、资源级写入、解析任意资源与镜像。
// 版本内的服务/组件与资源由 Composer 维护归一化视图。
type Catalog struct {
	apps       repository.ApplicationRepo
	releases   repository.ApplicationReleaseRepo
	registries repository.RegistryRepo
	images     repository.ImageRegistryRepo
	composer   *Composer
	access     access.Checker
	log        *slog.Logger
}

func NewCatalog(apps repository.ApplicationRepo, releases repository.ApplicationReleaseRepo, registries repository.RegistryRepo, images repository.ImageRegistryRepo, composer *Composer, checker access.Checker, logger *slog.Logger) *Catalog {
	return &Catalog{apps: apps, releases: releases, registries: registries, images: images, composer: composer, access: checker, log: logger}
}

func (catalog *Catalog) CreateApplication(ctx context.Context, in bo.Application) (*bo.Application, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizeApplication(&in); err != nil {
		return nil, err
	}
	item, err := catalog.apps.CreateApplication(ctx, in)
	return item, catalog.access.Wrap(err)
}

func (catalog *Catalog) UpdateApplication(ctx context.Context, in bo.Application) (*bo.Application, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizeApplication(&in); err != nil {
		return nil, err
	}
	item, err := catalog.apps.UpdateApplication(ctx, in)
	return item, catalog.access.Wrap(err)
}

func (catalog *Catalog) DeleteApplication(ctx context.Context, uid string) error {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationDelete)
	if err != nil {
		return err
	}
	// 已发布版本或未卸载部署组存在时不能删除应用，避免历史发布记录失去归属。
	if published, err := catalog.apps.HasPublishedVersions(ctx, ns, uid); err != nil {
		return catalog.access.Wrap(err)
	} else if published {
		return merr.Conflict("application has published versions")
	}
	if active, err := catalog.releases.HasActiveDeploymentsForApplication(ctx, ns, uid); err != nil {
		return catalog.access.Wrap(err)
	} else if active {
		return merr.Conflict("application has active deployments")
	}
	return catalog.access.Wrap(catalog.apps.DeleteApplication(ctx, ns, uid))
}

func (catalog *Catalog) GetApplication(ctx context.Context, uid string) (*bo.Application, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationRead)
	if err != nil {
		return nil, err
	}
	item, err := catalog.apps.GetApplication(ctx, ns, uid)
	return item, catalog.access.Wrap(err)
}

func (catalog *Catalog) ListApplications(ctx context.Context, pageNo, pageSize int32, filter bo.ApplicationListFilter) ([]bo.Application, int, int, int, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := catalog.apps.ListApplications(ctx, ns, filter, p, s)
	return items, total, p, s, catalog.access.Wrap(err)
}

func (catalog *Catalog) CreateVersion(ctx context.Context, applicationUID string, version bo.ApplicationVersion) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	if _, err := catalog.apps.GetApplication(ctx, ns, applicationUID); err != nil {
		return nil, catalog.access.Wrap(err)
	}
	version.NamespaceUID = ns
	version.ApplicationUID = applicationUID
	if err := normalizeVersion(&version); err != nil {
		return nil, err
	}
	if len(version.Resources) == 0 {
		version.Components = catalog.composer.Compose(nil)
		version.Resources = nil
		version.Validation = bo.ApplicationValidation{Warnings: []string{"add at least one resource before publishing"}}
		item, err := catalog.apps.CreateVersion(ctx, version)
		if err != nil {
			return nil, catalog.access.Wrap(err)
		}
		catalog.decorate(item)
		return item, nil
	}
	parsed, secrets, errs, warnings, err := ParseResources(version.Resources)
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, merr.Invalid(strings.Join(errs, "; "))
	}
	if err := validateComponentNames(parsed); err != nil {
		return nil, err
	}
	version.SecretValues = secrets
	components := catalog.composer.Compose(parsed)
	version.Components = components
	version.Resources = catalog.composer.Flatten(components)
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, components)
	warnings = append(warnings, imageWarnings...)
	version.Images = images
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	item, err := catalog.apps.CreateVersion(ctx, version)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	catalog.decorate(item)
	return item, nil
}

// UpdateVersion 兼容旧接口：只允许改 changelog；归一化版本的 resources 必须走
// 资源级接口，避免旧客户端按整包覆盖丢弃并发编辑。
func (catalog *Catalog) UpdateVersion(ctx context.Context, applicationUID string, version bo.ApplicationVersion) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	version.NamespaceUID = ns
	version.ApplicationUID = applicationUID
	if err := normalizeChangelog(&version); err != nil {
		return nil, err
	}
	existing, err := catalog.apps.GetVersion(ctx, ns, applicationUID, version.UID)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	if existing.Status != draftVersionStatus {
		return nil, merr.Conflict("published version is immutable")
	}
	if existing.CompositionVersion == 1 {
		if len(version.Resources) > 0 {
			return nil, merr.Conflict("normalized version resources must be edited through the resource API")
		}
		item, err := catalog.apps.UpdateVersion(ctx, bo.ApplicationVersion{
			UID: version.UID, NamespaceUID: ns, ApplicationUID: applicationUID, Changelog: version.Changelog,
		})
		if err != nil {
			return nil, catalog.access.Wrap(err)
		}
		catalog.decorate(item)
		return item, nil
	}
	if len(version.Resources) == 0 {
		item, err := catalog.apps.UpdateVersion(ctx, bo.ApplicationVersion{
			UID: version.UID, NamespaceUID: ns, ApplicationUID: applicationUID, Changelog: version.Changelog,
		})
		if err != nil {
			return nil, catalog.access.Wrap(err)
		}
		catalog.decorate(item)
		return item, nil
	}
	parsed, secrets, errs, warnings, err := ParseResources(version.Resources)
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, merr.Invalid(strings.Join(errs, "; "))
	}
	version.SecretValues = secrets
	version.Resources = parsed
	version.Components = catalog.composer.Compose(parsed)
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, version.Components)
	warnings = append(warnings, imageWarnings...)
	version.Images = images
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	item, err := catalog.apps.UpdateVersion(ctx, version)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	catalog.decorate(item)
	return item, nil
}

func (catalog *Catalog) DeleteVersion(ctx context.Context, applicationUID, uid string) error {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return catalog.access.Wrap(err)
	}
	// 已发布版本必须先后下架；已下架版本仍可能被未卸载的集群实例运行。
	if version.Status == publishVersionState {
		return merr.Conflict("published version cannot be deleted")
	}
	if version.Status == deprecateState {
		if active, err := catalog.releases.HasDeploymentsForVersion(ctx, ns, uid); err != nil {
			return catalog.access.Wrap(err)
		} else if active {
			return merr.Conflict("version is referenced by active deployments")
		}
	}
	return catalog.access.Wrap(catalog.apps.DeleteVersion(ctx, ns, applicationUID, uid))
}

func (catalog *Catalog) GetVersion(ctx context.Context, applicationUID, uid string) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationRead)
	if err != nil {
		return nil, err
	}
	item, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	catalog.decorate(item)
	return item, nil
}

func (catalog *Catalog) ListVersions(ctx context.Context, applicationUID string, pageNo, pageSize int32, filter bo.ApplicationVersionFilter) ([]bo.ApplicationVersion, int, int, int, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := catalog.apps.ListVersions(ctx, ns, applicationUID, filter, p, s)
	if err != nil {
		return nil, 0, 0, 0, catalog.access.Wrap(err)
	}
	for index := range items {
		catalog.decorate(&items[index])
	}
	return items, total, p, s, nil
}

// Validate 重新解析资源与镜像，返回错误与警告；不改变版本内容或状态。
func (catalog *Catalog) Validate(ctx context.Context, applicationUID, uid string) (*bo.ApplicationVersion, []string, []string, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, nil, nil, err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, nil, nil, catalog.access.Wrap(err)
	}
	catalog.decorate(version)
	components := version.Components
	if len(components) == 0 {
		components = catalog.composer.Compose(version.Resources)
	}
	warnings := make([]string, 0)
	if err := catalog.composer.ValidateForPublish(components); err != nil {
		return version, []string{err.Error()}, warnings, nil
	}
	parsed, _, errs, parseWarnings, err := ParseResources(catalog.composer.Flatten(components))
	if err != nil {
		return nil, nil, nil, err
	}
	warnings = append(warnings, parseWarnings...)
	if len(errs) > 0 {
		return version, errs, warnings, nil
	}
	_, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, catalog.composer.Compose(parsed))
	warnings = append(warnings, imageWarnings...)
	return version, nil, warnings, nil
}

// Publish 把草稿固定为已发布：要求每个组件至少一个资源，并把镜像 digest 固定下来。
func (catalog *Catalog) Publish(ctx context.Context, applicationUID, uid string) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	if version.Status != draftVersionStatus {
		return nil, merr.Conflict("only a draft version can be published")
	}
	components := version.Components
	if len(components) == 0 {
		components = catalog.composer.Compose(version.Resources)
	}
	if err := catalog.composer.ValidateForPublish(components); err != nil {
		return nil, err
	}
	parsed, _, errs, warnings, err := ParseResources(catalog.composer.Flatten(components))
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, merr.Invalid(strings.Join(errs, "; "))
	}
	components = catalog.composer.Compose(parsed)
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, components)
	warnings = append(warnings, imageWarnings...)
	version.Components = components
	version.Resources = catalog.composer.Flatten(components)
	version.Images = images
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	// 发布时把归一化内容、镜像 digest 和版本级密文一起写入，并 revision +1。
	if _, err := catalog.apps.SaveVersionContent(ctx, *version, version.Revision); err != nil {
		return nil, catalog.access.Wrap(err)
	}
	item, err := catalog.apps.UpdateVersionStatus(ctx, ns, applicationUID, uid, publishVersionState)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	catalog.decorate(item)
	return item, nil
}

// Deprecate 只把 PUBLISHED 版本下架；DRAFT 需要先发布，DEPRECATED 幂等返回。
func (catalog *Catalog) Deprecate(ctx context.Context, applicationUID, uid string) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	switch version.Status {
	case publishVersionState:
		item, err := catalog.apps.UpdateVersionStatus(ctx, ns, applicationUID, uid, deprecateState)
		if err != nil {
			return nil, catalog.access.Wrap(err)
		}
		catalog.decorate(item)
		return item, nil
	case deprecateState:
		catalog.decorate(version)
		return version, nil
	default:
		return nil, merr.Conflict("only a published version can be deprecated")
	}
}

// DuplicateVersion 复制任意状态的版本为新草稿，复制组件、资源、镜像与版本级密文。
func (catalog *Catalog) DuplicateVersion(ctx context.Context, applicationUID, uid, version, changelog string) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	if _, err := catalog.apps.GetApplication(ctx, ns, applicationUID); err != nil {
		return nil, catalog.access.Wrap(err)
	}
	source, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	target := bo.ApplicationVersion{Version: version, Changelog: changelog}
	if err := normalizeVersion(&target); err != nil {
		return nil, err
	}
	if strings.TrimSpace(target.Changelog) == "" {
		target.Changelog = source.Changelog
	}
	item, err := catalog.apps.DuplicateVersion(ctx, bo.ApplicationVersionCopy{
		NamespaceUID: ns, ApplicationUID: applicationUID, SourceVersionUID: uid,
		Version: target.Version, Changelog: target.Changelog,
	})
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	catalog.decorate(item)
	return item, nil
}

// CreateVersionResource 在草稿版本里新增一条资源；expected_revision 是乐观锁。
func (catalog *Catalog) CreateVersionResource(ctx context.Context, applicationUID, uid string, input bo.ApplicationResource, expectedRevision int32) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	if version.Status != draftVersionStatus {
		return nil, merr.Conflict("only draft version resources can be changed")
	}
	parsed, secrets, errs, warnings, err := ParseResources([]bo.ApplicationResource{input})
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, merr.Invalid(strings.Join(errs, "; "))
	}
	resource := parsed[0]
	var conflict error
	for _, component := range version.Components {
		for _, existing := range component.Resources {
			if existing.Item == resource.Item {
				conflict = merr.Conflict("resource item already exists in this version")
			}
		}
	}
	if conflict != nil {
		return nil, conflict
	}
	components, err := catalog.composer.CreateResource(version.Components, resource)
	if err != nil {
		return nil, err
	}
	version.Components = components
	version.Resources = catalog.composer.Flatten(components)
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, components)
	version.Images = images
	warnings = append(warnings, imageWarnings...)
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	item, err := catalog.apps.SaveVersionResource(ctx, *version, bo.ApplicationResourceChange{
		NamespaceUID: ns, ApplicationUID: applicationUID, VersionUID: uid,
		Resource: parsed[0], SecretValues: secrets, ExpectedRevision: expectedRevision,
	})
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	catalog.decorate(item)
	return item, nil
}

// SaveVersionResource 编辑草稿版本里的单条资源，可迁移组件并合并版本级密文。
func (catalog *Catalog) SaveVersionResource(ctx context.Context, applicationUID, uid, resourceUID, component, manifest string, secretValues map[string]string, clearSecretKeys []string, expectedRevision int32) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	if version.Status != draftVersionStatus {
		return nil, merr.Conflict("only draft version resources can be changed")
	}
	existing, err := catalog.composer.FindResource(version.Components, resourceUID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(component) == "" {
		component = existing.Component
	}
	parsed, secrets, errs, warnings, err := ParseResources([]bo.ApplicationResource{{
		UID: existing.UID, Item: existing.Item, Component: component, Manifest: manifest,
	}})
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, merr.Invalid(strings.Join(errs, "; "))
	}
	// 资源级 Secret 值按 key 传入；版本级密文统一用 "item/key" 存储。
	for key, value := range secretValues {
		secrets[existing.Item+"/"+key] = value
	}
	parsed[0].SecretKeys = mergeResourceSecretKeys(existing.SecretKeys, parsed[0].SecretKeys, secretValues, clearSecretKeys)
	components, err := catalog.composer.SaveResource(version.Components, resourceUID, parsed[0])
	if err != nil {
		return nil, err
	}
	version.Components = components
	version.Resources = catalog.composer.Flatten(components)
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, components)
	version.Images = images
	warnings = append(warnings, imageWarnings...)
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	item, err := catalog.apps.SaveVersionResource(ctx, *version, bo.ApplicationResourceChange{
		NamespaceUID: ns, ApplicationUID: applicationUID, VersionUID: uid,
		ResourceUID: resourceUID, Resource: parsed[0], SecretValues: secrets,
		ClearSecretKeys: clearSecretKeys, ExpectedRevision: expectedRevision,
	})
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	catalog.decorate(item)
	return item, nil
}

// DeleteVersionResource 删除草稿版本里的单条资源；组件变空时删除，但保留至少一个。
func (catalog *Catalog) DeleteVersionResource(ctx context.Context, applicationUID, uid, resourceUID string, expectedRevision int32) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	if version.Status != draftVersionStatus {
		return nil, merr.Conflict("only draft version resources can be changed")
	}
	components, err := catalog.composer.DeleteResource(version.Components, resourceUID)
	if err != nil {
		return nil, err
	}
	version.Components = components
	version.Resources = catalog.composer.Flatten(components)
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, components)
	version.Images = images
	version.Validation = bo.ApplicationValidation{Warnings: imageWarnings}
	item, err := catalog.apps.DeleteVersionResource(ctx, *version, resourceUID, expectedRevision)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	catalog.decorate(item)
	return item, nil
}

func (catalog *Catalog) GetVersionResource(ctx context.Context, applicationUID, uid, resourceUID string) (*bo.ApplicationResource, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationRead)
	if err != nil {
		return nil, err
	}
	resource, err := catalog.apps.GetVersionResource(ctx, ns, applicationUID, uid, resourceUID)
	return resource, catalog.access.Wrap(err)
}

// ValidateVersionResource 只解析校验一段资源 YAML，不写入版本。
func (catalog *Catalog) ValidateVersionResource(ctx context.Context, applicationUID, uid, component, manifest string) ([]string, []string, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, nil, err
	}
	if _, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid); err != nil {
		return nil, nil, catalog.access.Wrap(err)
	}
	_, _, errs, warnings, err := ParseResources([]bo.ApplicationResource{{
		Component: component, Manifest: manifest,
	}})
	if err != nil {
		return nil, nil, err
	}
	return errs, warnings, nil
}

func (catalog *Catalog) decorate(version *bo.ApplicationVersion) {
	if version == nil {
		return
	}
	version.ComponentCount = int32(len(version.Components))
	resourceCount := int32(0)
	for _, component := range version.Components {
		resourceCount += int32(len(component.Resources))
	}
	version.ResourceCount = resourceCount
	version.AllowedActions = catalog.composer.AllowedActions(version.Status)
	if version.Revision == 0 {
		version.Revision = 1
	}
}

func normalizeApplication(in *bo.Application) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len([]rune(in.Name)) > 64 {
		return merr.Invalid("application name is required and must not exceed 64 characters")
	}
	switch in.Category {
	case "DATABASE", "MIDDLEWARE", "OBSERVABILITY", "TOOLING", "OTHER":
	default:
		return merr.Invalid("invalid application category")
	}
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	in.Description = strings.TrimSpace(in.Description)
	if len([]rune(in.Description)) > 1024 {
		return merr.Invalid("description is too long")
	}
	return nil
}

func normalizeChangelog(in *bo.ApplicationVersion) error {
	in.Changelog = strings.TrimSpace(in.Changelog)
	if len([]rune(in.Changelog)) > maxChangelog {
		return merr.Invalid("changelog is too long")
	}
	return nil
}

func normalizeVersion(in *bo.ApplicationVersion) error {
	in.Version = strings.TrimSpace(strings.TrimPrefix(in.Version, "v"))
	if in.Version == "" || len([]rune(in.Version)) > maxVersionRunes {
		return merr.Invalid("version is required and must not exceed 32 characters")
	}
	in.Changelog = strings.TrimSpace(in.Changelog)
	if len([]rune(in.Changelog)) > maxChangelog {
		return merr.Invalid("changelog is too long")
	}
	return nil
}

func validateComponentNames(resources []bo.ApplicationResource) error {
	for _, resource := range resources {
		if len([]rune(strings.TrimSpace(resource.Component))) > maxComponentRunes {
			return merr.Invalid("component name must not exceed 63 characters")
		}
	}
	return nil
}

// resolveImages 把资源里的镜像解析成版本快照。能走仓库连接的镜像固定 digest 并
// 标记为 REGISTRY，其余按 EXTERNAL 保留原引用。
func (catalog *Catalog) resolveImages(ctx context.Context, namespaceUID, applicationUID string, components []bo.ApplicationVersionComponent) ([]bo.ApplicationVersionImage, []string) {
	registries, _, err := catalog.registries.ListRegistries(ctx, namespaceUID, bo.RegistryListFilter{}, 1, 100)
	if err != nil {
		return nil, []string{"image registry list is unavailable; images are treated as external"}
	}
	byHost := map[string]bo.Registry{}
	for _, registry := range registries {
		parsed, err := url.Parse(registry.Endpoint)
		if err == nil && parsed.Host != "" {
			byHost[strings.ToLower(parsed.Host)] = registry
		}
	}
	images := make([]bo.ApplicationVersionImage, 0)
	warnings := make([]string, 0)
	for _, component := range components {
		for _, resource := range component.Resources {
			refs := resource.ImageRefs
			if len(refs) == 0 {
				for _, image := range resource.Images {
					refs = append(refs, bo.ApplicationImageRef{Image: image})
				}
			}
			for _, ref := range refs {
				host, repository, tag, digest := splitImage(ref.Image)
				snapshot := bo.ApplicationVersionImage{
					NamespaceUID: namespaceUID, ApplicationUID: applicationUID,
					ResourceItem: resource.Item, ResourceUID: resource.UID, Component: component.Name,
					Container: ref.Container, ContainerPath: ref.ContainerPath, Image: ref.Image,
					Repository: repository, Tag: tag, Digest: digest,
					Project: projectOf(repository), Source: "EXTERNAL", ResolvedImage: ref.Image,
					External: true,
				}
				if registry, ok := byHost[strings.ToLower(host)]; ok {
					preview, err := catalog.images.ResolveSource(ctx, bo.ImageSyncSource{
						Kind: "CONNECTION", RegistryUID: registry.UID, Image: ref.Image,
					}, nil)
					if err == nil {
						snapshot.RegistryUID = registry.UID
						snapshot.RegistryName = registry.Name
						snapshot.Digest = preview.SourceDigest
						snapshot.Platforms = preview.SourcePlatforms
						snapshot.External = false
						snapshot.Source = "REGISTRY"
						snapshot.ResolvedImage = resolvedImage(host, repository, preview.SourceDigest, ref.Image)
					} else {
						warnings = append(warnings, "image "+ref.Image+" could not be resolved and is treated as external")
					}
				} else {
					warnings = append(warnings, "image "+ref.Image+" is external and is not digest-pinned")
				}
				images = append(images, snapshot)
			}
		}
	}
	return images, warnings
}

func splitImage(image string) (host, repository, tag, digest string) {
	image = strings.TrimSpace(image)
	if index := strings.LastIndex(image, "@"); index > 0 {
		digest = image[index+1:]
		image = image[:index]
	}
	first := strings.Index(image, "/")
	if first > 0 && (strings.Contains(image[:first], ".") || strings.Contains(image[:first], ":") || image[:first] == "localhost") {
		host = image[:first]
		image = image[first+1:]
	} else {
		host = "docker.io"
	}
	slash := strings.LastIndex(image, "/")
	if index := strings.LastIndex(image, ":"); index > slash {
		tag = image[index+1:]
		image = image[:index]
	}
	repository = image
	return host, repository, tag, digest
}

// mergeResourceSecretKeys 让资源上的 SecretKeys 与版本级密文的实际键保持一致：
// 保留已保存的键，加入本次明文与清单里出现的键，去掉显式清除的键。
func mergeResourceSecretKeys(existing, parsed []string, added map[string]string, cleared []string) []string {
	keys := make([]string, 0, len(existing)+len(parsed)+len(added))
	seen := map[string]struct{}{}
	removed := make(map[string]struct{}, len(cleared))
	for _, key := range cleared {
		removed[key] = struct{}{}
	}
	appendKey := func(key string) {
		if _, ok := removed[key]; ok {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	for _, key := range existing {
		appendKey(key)
	}
	for _, key := range parsed {
		appendKey(key)
	}
	for key := range added {
		appendKey(key)
	}
	return keys
}

// projectOf 取仓库路径的第一段作为项目名（Harbor 等按项目隔离）。
func projectOf(repository string) string {
	index := strings.Index(repository, "/")
	if index <= 0 {
		return ""
	}
	return repository[:index]
}

// resolvedImage 用解析到的 digest 固定实际部署引用。
func resolvedImage(host, repository, digest, fallback string) string {
	if digest == "" {
		return fallback
	}
	reference := repository
	if host != "" && !strings.EqualFold(host, "docker.io") {
		reference = host + "/" + repository
	}
	return reference + "@" + digest
}
