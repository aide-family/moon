package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// RegistryRepo 管理镜像仓库连接的持久化。远端读取由 ImageRegistryRepo 负责。
type RegistryRepo interface {
	CreateRegistry(ctx context.Context, in bo.Registry, password string) (*bo.Registry, error)
	// UpdateRegistry 的 password 为空表示保持原值；clearPassword 为 true 时显式清空。
	UpdateRegistry(ctx context.Context, in bo.Registry, password string, clearPassword bool) (*bo.Registry, error)
	DeleteRegistry(ctx context.Context, namespaceUID, uid string) error
	GetRegistry(ctx context.Context, namespaceUID, uid string) (*bo.Registry, error)
	ListRegistries(ctx context.Context, namespaceUID string, filter bo.RegistryListFilter, page, pageSize int) ([]bo.Registry, int, error)
	// TouchRegistry 刷新最近检测时间与错误，返回最新连接。
	TouchRegistry(ctx context.Context, namespaceUID, uid string, checkedAt int64, lastError string) (*bo.Registry, error)
}

// ImageRegistryRepo 读取远端仓库并执行镜像复制。实现按 driver 选择协议。
type ImageRegistryRepo interface {
	Probe(ctx context.Context, driver, endpoint, username, password string, allowInsecureTLS bool) (*bo.RegistryProbe, error)
	// ProbeSaved 用已保存的凭据测试连接。
	ProbeSaved(ctx context.Context, registryUID string) (*bo.RegistryProbe, error)
	ListProjects(ctx context.Context, registry bo.Registry, name string) ([]bo.RegistryProject, error)
	ListRepositories(ctx context.Context, registry bo.Registry, project, name string, page, pageSize int) ([]bo.RegistryRepository, int, error)
	ListTags(ctx context.Context, registry bo.Registry, project, repository, architecture, name string, page, pageSize int) ([]bo.RegistryTag, int, error)
	GetTag(ctx context.Context, registry bo.Registry, project, repository, tag string) (*bo.RegistryTagDetail, error)
	CopyTag(ctx context.Context, registry bo.Registry, project, repository, tag, digest, targetRepository, targetTag string) (*bo.RegistryTag, error)
	DeleteTag(ctx context.Context, registry bo.Registry, project, repository, tag, digest string) error
	// ResolveSource 解析已接入或外部来源，返回完整引用、digest 与平台。
	ResolveSource(ctx context.Context, source bo.ImageSyncSource, registry *bo.Registry) (*bo.ImageSyncPreview, error)
	// TagExists 检查目标仓库是否已有 Tag，返回 digest。
	TagExists(ctx context.Context, registry bo.Registry, repository, tag string) (bool, string, error)
	// CopyImage 执行一次镜像复制，返回目标 digest。
	CopyImage(ctx context.Context, in bo.ImageSync) (string, error)
}

// ImageSyncRepo 持久化同步任务，并提供调度器认领与状态更新。
type ImageSyncRepo interface {
	CreateImageSync(ctx context.Context, in bo.ImageSync, sourcePassword string) (*bo.ImageSync, error)
	GetImageSync(ctx context.Context, namespaceUID, uid string) (*bo.ImageSync, error)
	ListImageSyncs(ctx context.Context, namespaceUID string, filter bo.ImageSyncFilter, page, pageSize int) ([]bo.ImageSync, int, error)
	RetryImageSync(ctx context.Context, namespaceUID, uid string) (*bo.ImageSync, error)
	CancelImageSync(ctx context.Context, namespaceUID, uid string) (*bo.ImageSync, error)
	// ClaimImageSync 认领一条 PENDING 任务；没有可认领任务时返回 nil。
	ClaimImageSync(ctx context.Context, owner string, now, leaseUntil time.Time) (*bo.ImageSync, error)
	UpdateImageSync(ctx context.Context, in bo.ImageSync) error
	// ClearExpiredCredential 清空终态超过 before 的外部临时凭据。
	ClearExpiredCredential(ctx context.Context, before time.Time) (int, error)
}
