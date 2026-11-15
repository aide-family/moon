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

var ProviderSet = wire.NewSet(NewCatalog, NewDeployer, NewExecutor, NewInspector)

const (
	maxVersionRunes = 32
	maxChangelog    = 4096
)

// Catalog 管理应用与发布包版本：解析任意资源、解析镜像、发布校验。
type Catalog struct {
	apps       repository.ApplicationRepo
	releases   repository.ApplicationReleaseRepo
	registries repository.RegistryRepo
	images     repository.ImageRegistryRepo
	access     access.Checker
	log        *slog.Logger
}

func NewCatalog(apps repository.ApplicationRepo, releases repository.ApplicationReleaseRepo, registries repository.RegistryRepo, images repository.ImageRegistryRepo, checker access.Checker, logger *slog.Logger) *Catalog {
	return &Catalog{apps: apps, releases: releases, registries: registries, images: images, access: checker, log: logger}
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
	if published, err := catalog.apps.HasNonDraftVersions(ctx, ns, uid); err != nil {
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
		version.Validation = bo.ApplicationValidation{Warnings: []string{"add at least one resource before publishing"}}
		item, err := catalog.apps.CreateVersion(ctx, version)
		return item, catalog.access.Wrap(err)
	}
	parsed, secrets, errs, warnings, err := ParseResources(version.Resources)
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, merr.Invalid(strings.Join(errs, "; "))
	}
	version.SecretValues = secrets
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, parsed)
	warnings = append(warnings, imageWarnings...)
	version.Resources = parsed
	version.Images = images
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	item, err := catalog.apps.CreateVersion(ctx, version)
	return item, catalog.access.Wrap(err)
}

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
	if len(version.Resources) == 0 {
		version.Validation = bo.ApplicationValidation{Warnings: []string{"add at least one resource before publishing"}}
		item, err := catalog.apps.UpdateVersion(ctx, version)
		return item, catalog.access.Wrap(err)
	}
	parsed, secrets, errs, warnings, err := ParseResources(version.Resources)
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, merr.Invalid(strings.Join(errs, "; "))
	}
	version.SecretValues = secrets
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, parsed)
	warnings = append(warnings, imageWarnings...)
	version.Resources = parsed
	version.Images = images
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	item, err := catalog.apps.UpdateVersion(ctx, version)
	return item, catalog.access.Wrap(err)
}

func (catalog *Catalog) DeleteVersion(ctx context.Context, applicationUID, uid string) error {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return err
	}
	if active, err := catalog.releases.HasDeploymentsForVersion(ctx, ns, uid); err != nil {
		return catalog.access.Wrap(err)
	} else if active {
		return merr.Conflict("version is referenced by active deployments")
	}
	return catalog.access.Wrap(catalog.apps.DeleteVersion(ctx, ns, applicationUID, uid))
}

func (catalog *Catalog) GetVersion(ctx context.Context, applicationUID, uid string) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationRead)
	if err != nil {
		return nil, err
	}
	item, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	return item, catalog.access.Wrap(err)
}

func (catalog *Catalog) ListVersions(ctx context.Context, applicationUID string, pageNo, pageSize int32, filter bo.ApplicationVersionFilter) ([]bo.ApplicationVersion, int, int, int, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := catalog.apps.ListVersions(ctx, ns, applicationUID, filter, p, s)
	return items, total, p, s, catalog.access.Wrap(err)
}

// Validate 重新解析资源与镜像，返回错误与警告；不改变版本状态。
func (catalog *Catalog) Validate(ctx context.Context, applicationUID, uid string) (*bo.ApplicationVersion, []string, []string, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, nil, nil, err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, nil, nil, catalog.access.Wrap(err)
	}
	if version.Status != "DRAFT" {
		return version, nil, nil, nil
	}
	parsed, secrets, errs, warnings, err := ParseResources(version.Resources)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(errs) > 0 {
		return version, errs, warnings, nil
	}
	version.SecretValues = secrets
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, parsed)
	warnings = append(warnings, imageWarnings...)
	version.Resources = parsed
	version.Images = images
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	updated, err := catalog.apps.UpdateVersion(ctx, *version)
	if err != nil {
		return nil, nil, nil, catalog.access.Wrap(err)
	}
	return updated, nil, warnings, nil
}

func (catalog *Catalog) Publish(ctx context.Context, applicationUID, uid string) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	version, err := catalog.apps.GetVersion(ctx, ns, applicationUID, uid)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	if version.Status == "PUBLISHED" {
		return version, nil
	}
	if version.Status == "DEPRECATED" {
		return catalog.apps.UpdateVersionStatus(ctx, ns, applicationUID, uid, "PUBLISHED")
	}
	parsed, secrets, errs, warnings, err := ParseResources(version.Resources)
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, merr.Invalid(strings.Join(errs, "; "))
	}
	version.SecretValues = secrets
	images, imageWarnings := catalog.resolveImages(ctx, ns, applicationUID, parsed)
	warnings = append(warnings, imageWarnings...)
	version.Resources = parsed
	version.Images = images
	version.Validation = bo.ApplicationValidation{Warnings: warnings}
	updated, err := catalog.apps.UpdateVersion(ctx, *version)
	if err != nil {
		return nil, catalog.access.Wrap(err)
	}
	item, err := catalog.apps.UpdateVersionStatus(ctx, ns, applicationUID, updated.UID, "PUBLISHED")
	return item, catalog.access.Wrap(err)
}

func (catalog *Catalog) Deprecate(ctx context.Context, applicationUID, uid string) (*bo.ApplicationVersion, error) {
	ns, err := catalog.access.Allow(ctx, perm.ApplicationWrite)
	if err != nil {
		return nil, err
	}
	item, err := catalog.apps.UpdateVersionStatus(ctx, ns, applicationUID, uid, "DEPRECATED")
	return item, catalog.access.Wrap(err)
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

func (catalog *Catalog) resolveImages(ctx context.Context, namespaceUID, applicationUID string, resources []bo.ApplicationResource) ([]bo.ApplicationVersionImage, []string) {
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
	for _, resource := range resources {
		for _, image := range resource.Images {
			host, repository, tag, digest := splitImage(image)
			registry, ok := byHost[strings.ToLower(host)]
			snapshot := bo.ApplicationVersionImage{
				NamespaceUID: namespaceUID, ApplicationUID: applicationUID,
				ResourceItem: resource.Item, Container: image, Image: image,
				Repository: repository, Tag: tag, Digest: digest, External: true,
			}
			if ok {
				preview, err := catalog.images.ResolveSource(ctx, bo.ImageSyncSource{
					Kind: "CONNECTION", RegistryUID: registry.UID, Image: image,
				}, nil)
				if err == nil {
					snapshot.RegistryUID = registry.UID
					snapshot.RegistryName = registry.Name
					snapshot.Digest = preview.SourceDigest
					snapshot.Platforms = preview.SourcePlatforms
					snapshot.External = false
				} else {
					warnings = append(warnings, "image "+image+" could not be resolved and is treated as external")
				}
			} else {
				warnings = append(warnings, "image "+image+" is external and is not digest-pinned")
			}
			images = append(images, snapshot)
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
