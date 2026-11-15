package registry

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewRegistry, NewExplorer, NewSynchronizer)

const (
	maxNameRunes        = 64
	maxEndpointBytes    = 512
	maxRemarkBytes      = 512
	maxUsernameBytes    = 256
	maxPasswordBytes    = 1024
	maxRepositoryRunes  = 256
	maxTagRunes         = 128
	maxSourceImageBytes = 1024
)

// Registry 管理镜像仓库连接的增删改查与连通性测试。
type Registry struct {
	registries repository.RegistryRepo
	images     repository.ImageRegistryRepo
	apps       repository.ApplicationRepo
	syncs      repository.ImageSyncRepo
	access     access.Checker
	log        *slog.Logger
}

func NewRegistry(registries repository.RegistryRepo, images repository.ImageRegistryRepo, apps repository.ApplicationRepo, syncs repository.ImageSyncRepo, checker access.Checker, logger *slog.Logger) *Registry {
	return &Registry{registries: registries, images: images, apps: apps, syncs: syncs, access: checker, log: logger}
}

func (registry *Registry) Create(ctx context.Context, in bo.Registry, password string) (*bo.Registry, error) {
	ns, err := registry.access.Allow(ctx, perm.RegistryWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := validateRegistry(&in, password); err != nil {
		return nil, err
	}
	item, err := registry.registries.CreateRegistry(ctx, in, password)
	return item, registry.access.Wrap(err)
}

func (registry *Registry) Update(ctx context.Context, in bo.Registry, password string, clearPassword bool) (*bo.Registry, error) {
	ns, err := registry.access.Allow(ctx, perm.RegistryWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := validateRegistry(&in, password); err != nil {
		return nil, err
	}
	item, err := registry.registries.UpdateRegistry(ctx, in, password, clearPassword)
	return item, registry.access.Wrap(err)
}

func (registry *Registry) Delete(ctx context.Context, uid string) error {
	ns, err := registry.access.Allow(ctx, perm.RegistryDelete)
	if err != nil {
		return err
	}
	if active, err := registry.syncs.HasActiveForRegistry(ctx, ns, uid); err != nil {
		return registry.access.Wrap(err)
	} else if active {
		return merr.Conflict("registry has active image sync jobs")
	}
	if referenced, err := registry.apps.HasImageReferencesForRegistry(ctx, ns, uid); err != nil {
		return registry.access.Wrap(err)
	} else if referenced {
		return merr.Conflict("registry is referenced by application versions")
	}
	return registry.access.Wrap(registry.registries.DeleteRegistry(ctx, ns, uid))
}

func (registry *Registry) Get(ctx context.Context, uid string) (*bo.Registry, error) {
	ns, err := registry.access.Allow(ctx, perm.RegistryRead)
	if err != nil {
		return nil, err
	}
	item, err := registry.registries.GetRegistry(ctx, ns, uid)
	return item, registry.access.Wrap(err)
}

func (registry *Registry) List(ctx context.Context, pageNo, pageSize int32, name, driver string) ([]bo.Registry, int, int, int, error) {
	ns, err := registry.access.Allow(ctx, perm.RegistryRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := registry.registries.ListRegistries(ctx, ns, bo.RegistryListFilter{Name: strings.TrimSpace(name), Driver: strings.TrimSpace(driver)}, p, s)
	return items, total, p, s, registry.access.Wrap(err)
}

// Test 用已保存的凭据测试连接，并刷新最近检测时间与错误。
func (registry *Registry) Test(ctx context.Context, uid string) (*bo.RegistryProbe, *bo.Registry, error) {
	ns, err := registry.access.Allow(ctx, perm.RegistryRead)
	if err != nil {
		return nil, nil, err
	}
	if _, err := registry.registries.GetRegistry(ctx, ns, uid); err != nil {
		return nil, nil, registry.access.Wrap(err)
	}
	probe, probeErr := registry.images.ProbeSaved(ctx, uid)
	lastError := ""
	if probeErr != nil {
		lastError = probeErr.Error()
	}
	item, err := registry.registries.TouchRegistry(ctx, ns, uid, time.Now().Unix(), lastError)
	if err != nil {
		return nil, nil, registry.access.Wrap(err)
	}
	if probeErr != nil {
		return nil, item, probeErr
	}
	return probe, item, nil
}

// Probe 用尚未保存的凭据测试连接。
func (registry *Registry) Probe(ctx context.Context, driver, endpoint, username, password string, allowInsecureTLS bool) (*bo.RegistryProbe, error) {
	if _, err := registry.access.Allow(ctx, perm.RegistryWrite); err != nil {
		return nil, err
	}
	in := bo.Registry{Driver: driver, Endpoint: endpoint, Username: username, AllowInsecureTLS: allowInsecureTLS}
	if err := validateRegistry(&in, password); err != nil {
		return nil, err
	}
	probe, err := registry.images.Probe(ctx, driver, strings.TrimSuffix(strings.TrimSpace(endpoint), "/"), strings.TrimSpace(username), password, allowInsecureTLS)
	return probe, registry.access.Wrap(err)
}

func validateRegistry(in *bo.Registry, password string) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len([]rune(in.Name)) > maxNameRunes {
		return merr.Invalid("registry name is required and must not exceed 64 characters")
	}
	if in.Driver != "HARBOR" && in.Driver != "DOCKER_V2" {
		return merr.Invalid("invalid registry driver")
	}
	in.Endpoint = strings.TrimSuffix(strings.TrimSpace(in.Endpoint), "/")
	parsed, err := url.Parse(in.Endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return merr.Invalid("registry endpoint must be an http or https URL")
	}
	if len(in.Endpoint) > maxEndpointBytes {
		return merr.Invalid("registry endpoint is too long")
	}
	in.Username = strings.TrimSpace(in.Username)
	if len(in.Username) > maxUsernameBytes {
		return merr.Invalid("registry username is too long")
	}
	if len(password) > maxPasswordBytes {
		return merr.Invalid("registry password is too long")
	}
	in.Remark = strings.TrimSpace(in.Remark)
	if len(in.Remark) > maxRemarkBytes {
		return merr.Invalid("registry remark is too long")
	}
	return nil
}
