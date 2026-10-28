package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// KubernetesRepo 通过保存的 kubeconfig 访问集群 API Server。
// 调用方必须先确认集群属于当前租户，这里的 clusterUID 只用于取客户端。
type KubernetesRepo interface {
	ListNamespaces(ctx context.Context, clusterUID string) ([]string, error)

	ListNodes(ctx context.Context, clusterUID, name string) ([]bo.KubeNode, error)
	SetNodeUnschedulable(ctx context.Context, clusterUID, name string, unschedulable bool) error

	ListWorkloads(ctx context.Context, clusterUID, namespace, kind, name string) ([]bo.KubeWorkload, error)
	ScaleWorkload(ctx context.Context, clusterUID, namespace, kind, name string, replicas int32) (*bo.KubeWorkload, error)
	RestartWorkload(ctx context.Context, clusterUID, namespace, kind, name string) (*bo.KubeWorkload, error)

	// ListPods 返回按 namespace、name 升序排好的全部匹配项，分页由调用方切片。
	ListPods(ctx context.Context, clusterUID, namespace, name, ownerKind, ownerName string) ([]bo.KubePod, error)
	DeletePod(ctx context.Context, clusterUID, namespace, name string) error
	PodLogs(ctx context.Context, clusterUID, namespace, name, container string, tailLines int32, previous bool) (*bo.KubePodLogs, error)

	// ListIngresses 返回按 namespace、name 升序排好的全部匹配项，分页由调用方切片。
	ListIngresses(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeIngress, error)

	// ListServices 返回按 namespace、name 升序排好的全部匹配项，分页由调用方切片。
	ListServices(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeService, error)
	// ListConfigMaps 只返回键名，值不会离开数据层。
	ListConfigMaps(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeConfigMap, error)
	// ListSecrets 只返回类型和键名，值不会离开数据层。
	ListSecrets(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeSecret, error)
	// ListPersistentVolumes 是集群级资源，不接受命名空间过滤，按 name 升序返回。
	ListPersistentVolumes(ctx context.Context, clusterUID, name string) ([]bo.KubePersistentVolume, error)
	// ListPersistentVolumeClaims 返回按 namespace、name 升序排好的全部匹配项。
	ListPersistentVolumeClaims(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubePersistentVolumeClaim, error)

	// ResourceYAML 读取单个资源的 YAML。Node 和 PersistentVolume 忽略 namespace；
	// Secret 与 ConfigMap 的值会被替换成占位符，绝不返回明文。
	ResourceYAML(ctx context.Context, clusterUID, kind, namespace, name string) (*bo.KubeResourceYAML, error)

	// WorkloadDetail 返回工作负载、名下的 ReplicaSet（仅 Deployment）和 Pod。
	WorkloadDetail(ctx context.Context, clusterUID, kind, namespace, name string) (*bo.KubeWorkloadDetail, error)
	// IngressDetail 返回 Ingress 和规则里引用的 Service。找不到的 Service 会被跳过。
	IngressDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeIngressDetail, error)
	// ServiceDetail 返回 Service 和 Endpoints。没有 Endpoints 时返回空列表。
	ServiceDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeServiceDetail, error)
	// ConfigMapDetail 返回 ConfigMap 和引用它作为卷、环境变量来源的工作负载。
	ConfigMapDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeConfigMapDetail, error)
	// SecretDetail 返回 Secret 和引用它的工作负载。响应只含类型和键名，不含值。
	SecretDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeSecretDetail, error)
	// PersistentVolumeClaimDetail 返回 PVC 和把它作为卷挂载的工作负载。
	PersistentVolumeClaimDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubePersistentVolumeClaimDetail, error)
}
