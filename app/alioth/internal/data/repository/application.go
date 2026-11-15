package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// ApplicationRepo 管理应用、发布包版本与镜像快照。
type ApplicationRepo interface {
	CreateApplication(ctx context.Context, in bo.Application) (*bo.Application, error)
	UpdateApplication(ctx context.Context, in bo.Application) (*bo.Application, error)
	DeleteApplication(ctx context.Context, namespaceUID, uid string) error
	GetApplication(ctx context.Context, namespaceUID, uid string) (*bo.Application, error)
	ListApplications(ctx context.Context, namespaceUID string, filter bo.ApplicationListFilter, page, pageSize int) ([]bo.Application, int, error)

	CreateVersion(ctx context.Context, in bo.ApplicationVersion) (*bo.ApplicationVersion, error)
	// UpdateVersion 只允许草稿；整体替换资源与镜像快照。
	UpdateVersion(ctx context.Context, in bo.ApplicationVersion) (*bo.ApplicationVersion, error)
	DeleteVersion(ctx context.Context, namespaceUID, applicationUID, uid string) error
	GetVersion(ctx context.Context, namespaceUID, applicationUID, uid string) (*bo.ApplicationVersion, error)
	ListVersions(ctx context.Context, namespaceUID, applicationUID string, filter bo.ApplicationVersionFilter, page, pageSize int) ([]bo.ApplicationVersion, int, error)
	UpdateVersionStatus(ctx context.Context, namespaceUID, applicationUID, uid, status string) (*bo.ApplicationVersion, error)
	// ReplaceVersionImages 在事务里替换某个版本的镜像快照。
	ReplaceVersionImages(ctx context.Context, namespaceUID, versionUID string, images []bo.ApplicationVersionImage) error
	// ListVersionImages 供镜像引用反查。
	ListVersionImages(ctx context.Context, namespaceUID, versionUID string) ([]bo.ApplicationVersionImage, error)
	// GetVersionSecrets 解密版本里保存的 Secret 值，仅供执行器使用。
	GetVersionSecrets(ctx context.Context, namespaceUID, versionUID string) (map[string]string, error)
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
}
