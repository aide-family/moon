package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// ApplicationRepo 管理应用、发布包版本与镜像快照。
//
// 版本内容有两套存储：composition_version 为 1 的版本读 application_version_components
// / application_version_resources；旧版本回退 application_versions.spec。写入归一化
// 内容的接口都要在事务里锁版本行并校验状态与 revision。
type ApplicationRepo interface {
	CreateApplication(ctx context.Context, in bo.Application) (*bo.Application, error)
	UpdateApplication(ctx context.Context, in bo.Application) (*bo.Application, error)
	DeleteApplication(ctx context.Context, namespaceUID, uid string) error
	GetApplication(ctx context.Context, namespaceUID, uid string) (*bo.Application, error)
	ListApplications(ctx context.Context, namespaceUID string, filter bo.ApplicationListFilter, page, pageSize int) ([]bo.Application, int, error)

	CreateVersion(ctx context.Context, in bo.ApplicationVersion) (*bo.ApplicationVersion, error)
	// UpdateVersion 只允许改 changelog；未归一化的旧版本仍允许整体替换 resources。
	UpdateVersion(ctx context.Context, in bo.ApplicationVersion) (*bo.ApplicationVersion, error)
	// SaveVersionContent 在事务里锁定版本、校验 DRAFT 与 revision，整体替换组件/
	// 资源/镜像，合并或保留版本级密文，并把 revision +1。
	SaveVersionContent(ctx context.Context, in bo.ApplicationVersion, expectedRevision int32) (*bo.ApplicationVersion, error)
	// SaveVersionResource 在事务里写入单个资源的完整组件视图，并按 change 合并密文。
	SaveVersionResource(ctx context.Context, in bo.ApplicationVersion, change bo.ApplicationResourceChange) (*bo.ApplicationVersion, error)
	// DeleteVersionResource 在事务里按期望组件视图删除单个资源及其镜像，并清理版本级密文。
	DeleteVersionResource(ctx context.Context, in bo.ApplicationVersion, resourceUID string, expectedRevision int32) (*bo.ApplicationVersion, error)
	GetVersionResource(ctx context.Context, namespaceUID, applicationUID, versionUID, resourceUID string) (*bo.ApplicationResource, error)
	// DuplicateVersion 复制一个版本为新草稿，复制组件/资源/镜像/密文。
	DuplicateVersion(ctx context.Context, in bo.ApplicationVersionCopy) (*bo.ApplicationVersion, error)
	DeleteVersion(ctx context.Context, namespaceUID, applicationUID, uid string) error
	GetVersion(ctx context.Context, namespaceUID, applicationUID, uid string) (*bo.ApplicationVersion, error)
	ListVersions(ctx context.Context, namespaceUID, applicationUID string, filter bo.ApplicationVersionFilter, page, pageSize int) ([]bo.ApplicationVersion, int, error)
	UpdateVersionStatus(ctx context.Context, namespaceUID, applicationUID, uid, status string) (*bo.ApplicationVersion, error)
	// GetVersionSecrets 解密版本里保存的 Secret 值，仅供执行器使用。
	GetVersionSecrets(ctx context.Context, namespaceUID, versionUID string) (map[string]string, error)
	// HasPublishedVersions 判断应用是否有已发布版本（删除应用时保护）。
	HasPublishedVersions(ctx context.Context, namespaceUID, applicationUID string) (bool, error)
	// FindImageReferences 查询引用指定镜像的所有发布包版本。
	FindImageReferences(ctx context.Context, namespaceUID, registryUID, repository, tag string) ([]bo.ApplicationImageReference, error)
	// HasImageReferencesForRegistry 判断是否有任何版本镜像引用该仓库连接。
	HasImageReferencesForRegistry(ctx context.Context, namespaceUID, registryUID string) (bool, error)
}

// ApplicationApplier 把发布包资源应用到目标集群。
type ApplicationApplier interface {
	NamespaceExists(ctx context.Context, clusterUID, namespace string) (bool, error)
	// Preflight 检查命名空间、资源类型与 scope；不写集群。
	Preflight(ctx context.Context, clusterUID, namespace string, resources []bo.ApplicationResource) ([]bo.ResourceRef, error)
	// Apply 用 Server-Side Apply 创建/更新资源，并等待标准工作负载就绪。
	Apply(ctx context.Context, clusterUID, namespace, deploymentUID, releaseUID, name string, resources []bo.ApplicationResource, secrets map[string]string) ([]bo.ResourceRef, error)
	// Delete 删除指定资源；默认保留 PVC 与集群级资源。
	Delete(ctx context.Context, clusterUID, releaseUID string, refs []bo.ResourceRef, keepPVC, deleteClusterScoped bool) error
}

// ApplicationReleaseRepo 管理部署组、集群实例与操作记录。
type ApplicationReleaseRepo interface {
	// CreateDeployment 在一个事务里创建部署组、N 个实例与 N 条 PENDING 操作。
	CreateDeployment(ctx context.Context, deployment bo.ApplicationDeployment, releases []bo.ApplicationRelease, operations []bo.ApplicationOperation) (*bo.ApplicationDeployment, error)
	GetDeployment(ctx context.Context, namespaceUID, uid string) (*bo.ApplicationDeployment, error)
	ListDeployments(ctx context.Context, namespaceUID string, filter bo.ApplicationReleaseFilter, page, pageSize int) ([]bo.ApplicationDeployment, int, error)
	UpdateDeployment(ctx context.Context, in bo.ApplicationDeployment) error

	GetRelease(ctx context.Context, namespaceUID, uid string) (*bo.ApplicationRelease, error)
	ListReleases(ctx context.Context, namespaceUID string, filter bo.ApplicationReleaseFilter, page, pageSize int) ([]bo.ApplicationRelease, int, error)
	UpdateRelease(ctx context.Context, in bo.ApplicationRelease) error

	CreateOperation(ctx context.Context, in bo.ApplicationOperation) (*bo.ApplicationOperation, error)
	GetOperation(ctx context.Context, namespaceUID, uid string) (*bo.ApplicationOperation, error)
	ListOperations(ctx context.Context, namespaceUID, deploymentUID, releaseUID string, page, pageSize int) ([]bo.ApplicationOperation, int, error)
	ClaimOperation(ctx context.Context, owner string, now, leaseUntil time.Time) (*bo.ApplicationOperation, error)
	UpdateOperation(ctx context.Context, in bo.ApplicationOperation) error
	// HasActiveOperation 判断实例是否已有 PENDING/RUNNING 操作。
	HasActiveOperation(ctx context.Context, releaseUID string) (bool, error)
	// LastSuccessfulOperation 返回实例最近一次成功的操作，用于升级裁剪。
	LastSuccessfulOperation(ctx context.Context, releaseUID string) (*bo.ApplicationOperation, error)
	// HasActiveDeploymentsForApplication 判断应用是否有未卸载的部署组。
	HasActiveDeploymentsForApplication(ctx context.Context, namespaceUID, applicationUID string) (bool, error)
	// HasDeploymentsForVersion 判断版本是否仍被未卸载的集群实例引用。
	HasDeploymentsForVersion(ctx context.Context, namespaceUID, versionUID string) (bool, error)
	// ReleaseUsedVersion 判断版本是否出现在该集群实例的历史（当前版本或任一次操作）。
	ReleaseUsedVersion(ctx context.Context, namespaceUID, releaseUID, versionUID string) (bool, error)
}
