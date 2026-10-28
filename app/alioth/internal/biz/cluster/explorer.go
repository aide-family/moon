package cluster

import (
	"context"
	"strings"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// YAML 预览支持的资源类型，和 proto 的 kind 枚举一致。
const (
	resourceKindNode                  = "Node"
	resourceKindPod                   = "Pod"
	resourceKindDeployment            = "Deployment"
	resourceKindStatefulSet           = "StatefulSet"
	resourceKindDaemonSet             = "DaemonSet"
	resourceKindJob                   = "Job"
	resourceKindCronJob               = "CronJob"
	resourceKindIngress               = "Ingress"
	resourceKindService               = "Service"
	resourceKindConfigMap             = "ConfigMap"
	resourceKindSecret                = "Secret"
	resourceKindPersistentVolume      = "PersistentVolume"
	resourceKindPersistentVolumeClaim = "PersistentVolumeClaim"
)

// Explorer 读取集群资源。每个方法都先确认调用方有 cluster:read，
// 再确认集群属于当前租户，最后才访问集群 API Server。
type Explorer struct {
	clusters  repository.ClusterRepo
	resources repository.KubernetesRepo
	access    access.Checker
}

func NewExplorer(clusters repository.ClusterRepo, resources repository.KubernetesRepo, checker access.Checker) *Explorer {
	return &Explorer{clusters: clusters, resources: resources, access: checker}
}

func (explorer *Explorer) ListNamespaces(ctx context.Context, clusterUID string) ([]string, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	names, err := explorer.resources.ListNamespaces(ctx, clusterUID)
	return names, explorer.access.Wrap(err)
}

// ListNodes 返回全部匹配的节点，节点数量有限，不分页。
func (explorer *Explorer) ListNodes(ctx context.Context, clusterUID, name string) ([]bo.KubeNode, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, err
	}
	items, err := explorer.resources.ListNodes(ctx, clusterUID, strings.TrimSpace(name))
	return items, len(items), explorer.access.Wrap(err)
}

// ListWorkloads 返回全部匹配的工作负载，不分页。
func (explorer *Explorer) ListWorkloads(ctx context.Context, clusterUID, namespace, kind, name string) ([]bo.KubeWorkload, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, err
	}
	items, err := explorer.resources.ListWorkloads(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(kind), strings.TrimSpace(name))
	return items, len(items), explorer.access.Wrap(err)
}

// ListPods 在内存里分页：集群 API 的列表不分页，仓储返回按 namespace、name 排好的完整结果。
func (explorer *Explorer) ListPods(ctx context.Context, clusterUID, namespace, name, ownerKind, ownerName string, pageNo, pageSize int32) ([]bo.KubePod, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListPods(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name), strings.TrimSpace(ownerKind), strings.TrimSpace(ownerName))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// ListIngresses 和 Pod 一样在内存里分页。
func (explorer *Explorer) ListIngresses(ctx context.Context, clusterUID, namespace, name string, pageNo, pageSize int32) ([]bo.KubeIngress, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListIngresses(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// ListServices 和 Pod 一样在内存里分页。
func (explorer *Explorer) ListServices(ctx context.Context, clusterUID, namespace, name string, pageNo, pageSize int32) ([]bo.KubeService, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListServices(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// ListConfigMaps 只返回键名，值不会离开数据层。
func (explorer *Explorer) ListConfigMaps(ctx context.Context, clusterUID, namespace, name string, pageNo, pageSize int32) ([]bo.KubeConfigMap, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListConfigMaps(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// ListSecrets 只返回类型和键名，值不会离开数据层。
func (explorer *Explorer) ListSecrets(ctx context.Context, clusterUID, namespace, name string, pageNo, pageSize int32) ([]bo.KubeSecret, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListSecrets(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// ListPersistentVolumes 是集群级资源，没有命名空间过滤，只在名称上过滤。
func (explorer *Explorer) ListPersistentVolumes(ctx context.Context, clusterUID, name string, pageNo, pageSize int32) ([]bo.KubePersistentVolume, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListPersistentVolumes(ctx, clusterUID, strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// ListPersistentVolumeClaims 在内存里分页。
func (explorer *Explorer) ListPersistentVolumeClaims(ctx context.Context, clusterUID, namespace, name string, pageNo, pageSize int32) ([]bo.KubePersistentVolumeClaim, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListPersistentVolumeClaims(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// PodLogs 读取容器日志尾部。容器名留空时由仓储取第一个容器。
func (explorer *Explorer) PodLogs(ctx context.Context, clusterUID, namespace, name, container string, tailLines int32, previous bool) (*bo.KubePodLogs, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	logs, err := explorer.resources.PodLogs(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name), strings.TrimSpace(container), tailLines, previous)
	return logs, explorer.access.Wrap(err)
}

// WorkloadDetail 返回工作负载、名下 ReplicaSet 和 Pod。
func (explorer *Explorer) WorkloadDetail(ctx context.Context, clusterUID, kind, namespace, name string) (*bo.KubeWorkloadDetail, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	kind = strings.TrimSpace(kind)
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return nil, err
	}
	if !validWorkloadKind(kind) {
		return nil, merr.Invalid("invalid resource kind")
	}
	item, err := explorer.resources.WorkloadDetail(ctx, clusterUID, kind, namespace, name)
	return item, explorer.access.Wrap(err)
}

// IngressDetail 返回 Ingress 和规则里引用的 Service。
func (explorer *Explorer) IngressDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeIngressDetail, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return nil, err
	}
	item, err := explorer.resources.IngressDetail(ctx, clusterUID, namespace, name)
	return item, explorer.access.Wrap(err)
}

// ServiceDetail 返回 Service 和 Endpoints。
func (explorer *Explorer) ServiceDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeServiceDetail, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return nil, err
	}
	item, err := explorer.resources.ServiceDetail(ctx, clusterUID, namespace, name)
	return item, explorer.access.Wrap(err)
}

// ConfigMapDetail 返回 ConfigMap 和引用它的工作负载。
func (explorer *Explorer) ConfigMapDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeConfigMapDetail, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return nil, err
	}
	item, err := explorer.resources.ConfigMapDetail(ctx, clusterUID, namespace, name)
	return item, explorer.access.Wrap(err)
}

// SecretDetail 返回 Secret 和引用它的工作负载。响应只有类型和键名。
func (explorer *Explorer) SecretDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeSecretDetail, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return nil, err
	}
	item, err := explorer.resources.SecretDetail(ctx, clusterUID, namespace, name)
	return item, explorer.access.Wrap(err)
}

// PersistentVolumeClaimDetail 返回 PVC 和挂载它的工作负载。
func (explorer *Explorer) PersistentVolumeClaimDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubePersistentVolumeClaimDetail, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return nil, err
	}
	item, err := explorer.resources.PersistentVolumeClaimDetail(ctx, clusterUID, namespace, name)
	return item, explorer.access.Wrap(err)
}

// WorkloadImages 返回每个容器的当前镜像与历史版本。历史版本来自集群里已存在的历史对象。
func (explorer *Explorer) WorkloadImages(ctx context.Context, clusterUID, namespace, kind, name string) ([]bo.KubeWorkloadImage, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, kind, name, err := normalizeWorkloadTarget(namespace, kind, name)
	if err != nil {
		return nil, err
	}
	items, err := explorer.resources.WorkloadImages(ctx, clusterUID, namespace, kind, name)
	return items, explorer.access.Wrap(err)
}

// WorkloadRevisions 返回工作负载的历史版本，新版本在前。
func (explorer *Explorer) WorkloadRevisions(ctx context.Context, clusterUID, namespace, kind, name string) ([]bo.KubeWorkloadRevision, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, kind, name, err := normalizeWorkloadTarget(namespace, kind, name)
	if err != nil {
		return nil, err
	}
	items, err := explorer.resources.WorkloadRevisions(ctx, clusterUID, namespace, kind, name)
	return items, explorer.access.Wrap(err)
}

// ListJobs 在内存里分页：仓储返回按 namespace、name 排好的完整结果。
func (explorer *Explorer) ListJobs(ctx context.Context, clusterUID, namespace, name string, pageNo, pageSize int32) ([]bo.KubeJob, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListJobs(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// JobDetail 返回 Job 和它的 Pod。
func (explorer *Explorer) JobDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeJobDetail, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return nil, err
	}
	detail, err := explorer.resources.JobDetail(ctx, clusterUID, namespace, name)
	return detail, explorer.access.Wrap(err)
}

// ListCronJobs 在内存里分页：仓储返回按 namespace、name 排好的完整结果。
func (explorer *Explorer) ListCronJobs(ctx context.Context, clusterUID, namespace, name string, pageNo, pageSize int32) ([]bo.KubeCronJob, int, int, int, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, 0, 0, 0, err
	}
	items, err := explorer.resources.ListCronJobs(ctx, clusterUID, strings.TrimSpace(namespace), strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, explorer.access.Wrap(err)
	}
	paged, total, p, s := slicePage(items, pageNo, pageSize)
	return paged, total, p, s, nil
}

// CronJobDetail 返回 CronJob 和它最近创建的 Job。
func (explorer *Explorer) CronJobDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeCronJobDetail, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return nil, err
	}
	detail, err := explorer.resources.CronJobDetail(ctx, clusterUID, namespace, name)
	return detail, explorer.access.Wrap(err)
}

// normalizeWorkloadTarget 清理并校验工作负载类型、命名空间和名称。工作负载只接受
// Deployment、StatefulSet 和 DaemonSet，空类型不表示全部。
func normalizeWorkloadTarget(namespace, kind, name string) (string, string, string, error) {
	kind = strings.TrimSpace(kind)
	namespace, name, err := normalizeNamespacedTarget(namespace, name)
	if err != nil {
		return "", "", "", err
	}
	if kind == "" || !validWorkloadKind(kind) {
		return "", "", "", merr.Invalid("invalid resource kind")
	}
	return namespace, kind, name, nil
}

// normalizeNamespacedTarget 清理并校验命名空间和名称。
func normalizeNamespacedTarget(namespace, name string) (string, string, error) {
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return "", "", merr.Invalid("namespace and name are required")
	}
	return namespace, name, nil
}

// ResourceYAML 读取单个资源的 YAML 预览。Secret / ConfigMap 的值由仓储替换，
// 这里只负责权限、租户范围和资源定位参数的校验。
func (explorer *Explorer) ResourceYAML(ctx context.Context, clusterUID, kind, namespace, name string) (*bo.KubeResourceYAML, error) {
	if _, err := explorer.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	kind, namespace, name, err := normalizeYAMLResource(kind, namespace, name)
	if err != nil {
		return nil, err
	}
	item, err := explorer.resources.ResourceYAML(ctx, clusterUID, kind, namespace, name)
	return item, explorer.access.Wrap(err)
}

// normalizeYAMLResource 清理并校验资源类型、命名空间和名称，返回实际使用的定位参数。
// Node 和 PersistentVolume 是集群级资源，命名空间参数会被忽略。
func normalizeYAMLResource(kind, namespace, name string) (string, string, string, error) {
	kind = strings.TrimSpace(kind)
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", "", merr.Invalid("resource name is required")
	}
	if !yamlResourceKindSupported(kind) {
		return "", "", "", merr.Invalid("invalid resource kind")
	}
	if yamlResourceClusterScoped(kind) {
		return kind, "", name, nil
	}
	if namespace == "" {
		return "", "", "", merr.Invalid("namespace is required")
	}
	return kind, namespace, name, nil
}

// yamlResourceKindSupported 判断资源类型是否在 YAML 预览范围内。
func yamlResourceKindSupported(kind string) bool {
	switch kind {
	case resourceKindNode, resourceKindPod, resourceKindDeployment, resourceKindStatefulSet,
		resourceKindDaemonSet, resourceKindJob, resourceKindCronJob, resourceKindIngress,
		resourceKindService, resourceKindConfigMap, resourceKindSecret, resourceKindPersistentVolume,
		resourceKindPersistentVolumeClaim:
		return true
	default:
		return false
	}
}

// yamlResourceClusterScoped 判断资源是否属于集群级。
func yamlResourceClusterScoped(kind string) bool {
	return kind == resourceKindNode || kind == resourceKindPersistentVolume
}

// scope 校验读权限并确认集群属于当前租户。不属于当前租户时仓储返回 NotFound，
// 不区分「集群不存在」和「属于别的租户」。
func (explorer *Explorer) scope(ctx context.Context, clusterUID string) (string, error) {
	ns, err := explorer.access.Allow(ctx, perm.ClusterRead)
	if err != nil {
		return "", err
	}
	if _, err := explorer.clusters.GetCluster(ctx, ns, clusterUID); err != nil {
		return "", explorer.access.Wrap(err)
	}
	return ns, nil
}

// slicePage 对已经排好序的完整列表做内存分页，返回当前页、总数和归一化后的页码。
func slicePage[T any](items []T, pageNo, pageSize int32) ([]T, int, int, int) {
	p, s := page.Normalize(pageNo, pageSize)
	total := len(items)
	start := (p - 1) * s
	if start >= total {
		return []T{}, total, p, s
	}
	end := start + s
	if end > total {
		end = total
	}
	return items[start:end], total, p, s
}
