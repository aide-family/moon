package repository

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// ClusterRepo 读写租户接入的 Kubernetes 集群，并负责 kubeconfig 的解析与探测。
// kubeconfig 原文只作为写入参数出现，读取接口不返回它。
type ClusterRepo interface {
	// CreateCluster 先解析并探测 kubeconfig，成功后加密保存。
	CreateCluster(ctx context.Context, in bo.Cluster, kubeconfig string) (*bo.Cluster, error)
	// UpdateCluster 在 kubeconfig 非空时先探测再替换，空值表示保持原凭据和元数据。
	UpdateCluster(ctx context.Context, in bo.Cluster, kubeconfig string) (*bo.Cluster, error)
	DeleteCluster(ctx context.Context, namespaceUID, uid string) error
	GetCluster(ctx context.Context, namespaceUID, uid string) (*bo.Cluster, error)
	ListClusters(ctx context.Context, namespaceUID, name string, page, pageSize int) ([]bo.Cluster, int, error)
	// ProbeKubeconfig 解析并探测尚未保存的 kubeconfig，不写数据库。
	ProbeKubeconfig(ctx context.Context, kubeconfig string) (*bo.ClusterProbe, error)
	// TestCluster 用已保存的 kubeconfig 探测，并刷新最近检测时间、最近错误和版本元数据。
	TestCluster(ctx context.Context, namespaceUID, uid string) (*bo.ClusterProbe, *bo.Cluster, error)
}
