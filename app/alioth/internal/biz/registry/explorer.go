package registry

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// Explorer 读取远端镜像仓库、Tag 与平台信息，不写远端。
type Explorer struct {
	registries repository.RegistryRepo
	images     repository.ImageRegistryRepo
	access     access.Checker
	log        *slog.Logger
}

func NewExplorer(registries repository.RegistryRepo, images repository.ImageRegistryRepo, checker access.Checker, logger *slog.Logger) *Explorer {
	return &Explorer{registries: registries, images: images, access: checker, log: logger}
}

func (explorer *Explorer) ListProjects(ctx context.Context, registryUID, name string) ([]bo.RegistryProject, error) {
	item, err := explorer.registry(ctx, registryUID)
	if err != nil {
		return nil, err
	}
	projects, err := explorer.images.ListProjects(ctx, *item, strings.TrimSpace(name))
	return projects, explorer.access.Wrap(err)
}

func (explorer *Explorer) ListRepositories(ctx context.Context, registryUID, project, name string, pageNo, pageSize int32) ([]bo.RegistryRepository, int, int, int, error) {
	item, err := explorer.registry(ctx, registryUID)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := explorer.images.ListRepositories(ctx, *item, strings.TrimSpace(project), strings.TrimSpace(name), p, s)
	return items, total, p, s, explorer.access.Wrap(err)
}

func (explorer *Explorer) ListTags(ctx context.Context, registryUID, project, repository, architecture, name string, pageNo, pageSize int32) ([]bo.RegistryTag, int, int, int, error) {
	item, err := explorer.registry(ctx, registryUID)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	repository = strings.TrimSpace(repository)
	if repository == "" {
		return nil, 0, 0, 0, merr.Invalid("repository is required")
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := explorer.images.ListTags(ctx, *item, strings.TrimSpace(project), repository, strings.TrimSpace(architecture), strings.TrimSpace(name), p, s)
	return items, total, p, s, explorer.access.Wrap(err)
}

func (explorer *Explorer) GetTag(ctx context.Context, registryUID, project, repository, tag string) (*bo.RegistryTagDetail, error) {
	item, err := explorer.registry(ctx, registryUID)
	if err != nil {
		return nil, err
	}
	repository, tag = strings.TrimSpace(repository), strings.TrimSpace(tag)
	if repository == "" || tag == "" {
		return nil, merr.Invalid("repository and tag are required")
	}
	detail, err := explorer.images.GetTag(ctx, *item, strings.TrimSpace(project), repository, tag)
	return detail, explorer.access.Wrap(err)
}

func (explorer *Explorer) CopyTag(ctx context.Context, registryUID, project, repository, tag, digest, targetRepository, targetTag string) (*bo.RegistryTag, error) {
	ns, err := explorer.access.Allow(ctx, perm.RegistryImageWrite)
	if err != nil {
		return nil, err
	}
	item, err := explorer.registries.GetRegistry(ctx, ns, registryUID)
	if err != nil {
		return nil, explorer.access.Wrap(err)
	}
	targetRepository = strings.TrimSpace(targetRepository)
	targetTag = strings.TrimSpace(targetTag)
	if targetRepository == "" || targetTag == "" {
		return nil, merr.Invalid("target repository and tag are required")
	}
	if len([]rune(targetRepository)) > maxRepositoryRunes || len([]rune(targetTag)) > maxTagRunes {
		return nil, merr.Invalid("target repository or tag is too long")
	}
	copied, err := explorer.images.CopyTag(ctx, *item, strings.TrimSpace(project), strings.TrimSpace(repository), strings.TrimSpace(tag), strings.TrimSpace(digest), targetRepository, targetTag)
	return copied, explorer.access.Wrap(err)
}

func (explorer *Explorer) DeleteTag(ctx context.Context, registryUID, project, repository, tag, digest string) error {
	ns, err := explorer.access.Allow(ctx, perm.RegistryImageDelete)
	if err != nil {
		return err
	}
	item, err := explorer.registries.GetRegistry(ctx, ns, registryUID)
	if err != nil {
		return explorer.access.Wrap(err)
	}
	repository, tag = strings.TrimSpace(repository), strings.TrimSpace(tag)
	if repository == "" || tag == "" {
		return merr.Invalid("repository and tag are required")
	}
	return explorer.access.Wrap(explorer.images.DeleteTag(ctx, *item, strings.TrimSpace(project), repository, tag, strings.TrimSpace(digest)))
}

func (explorer *Explorer) registry(ctx context.Context, uid string) (*bo.Registry, error) {
	ns, err := explorer.access.Allow(ctx, perm.RegistryRead)
	if err != nil {
		return nil, err
	}
	item, err := explorer.registries.GetRegistry(ctx, ns, uid)
	if err != nil {
		return nil, explorer.access.Wrap(err)
	}
	return item, nil
}
