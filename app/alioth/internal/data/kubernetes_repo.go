package data

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// defaultPodLogTailLines 是未指定行数时读取的日志尾部行数。
	defaultPodLogTailLines = 200
	// maxPodLogTailLines 与 proto 校验一致，避免绕过 HTTP 的调用拉取超长日志。
	maxPodLogTailLines = 2000
	// maxPodLogBytes 是单次日志读取的字节上限。
	maxPodLogBytes = 512 * 1024
	// workloadKindDeployment 等是工作负载类型的固定取值，和 proto 枚举一致。
	workloadKindDeployment  = "Deployment"
	workloadKindStatefulSet = "StatefulSet"
	workloadKindDaemonSet   = "DaemonSet"
)

var _ repository.KubernetesRepo = (*kubernetesRepo)(nil)

func NewKubernetesRepo(deps *Dependencies, logger *slog.Logger) repository.KubernetesRepo {
	return &kubernetesRepo{clients: deps.clients, log: logger}
}

type kubernetesRepo struct {
	clients *clusterClients
	log     *slog.Logger
}

func (r *kubernetesRepo) ListNamespaces(ctx context.Context, clusterUID string) ([]string, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.CoreV1().Namespaces().List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		names = append(names, item.Name)
	}
	sort.Strings(names)
	return names, nil
}

func (r *kubernetesRepo) ListNodes(ctx context.Context, clusterUID, name string) ([]bo.KubeNode, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	nodes, err := client.CoreV1().Nodes().List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	// 节点上的 Pod 数需要一次全量 Pod 列表，按 spec.nodeName 聚合。
	// 受限凭据可能没有全量 Pod 读取权限：这时仍然展示节点，Pod 数按 0 处理，
	// 不因为一个辅助字段让整个节点页失败。
	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(callCtx, metav1.ListOptions{})
	if err != nil {
		if !apierrors.IsForbidden(err) {
			return nil, kubeError(r.log, err)
		}
		pods = &corev1.PodList{}
	}
	podCounts := make(map[string]int32, len(nodes.Items))
	for _, pod := range pods.Items {
		if pod.Spec.NodeName == "" {
			continue
		}
		// 已结束的 Pod 不再占用节点，计入会让「Pod 数」长期偏高。
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		podCounts[pod.Spec.NodeName]++
	}
	items := make([]bo.KubeNode, 0, len(nodes.Items))
	for index := range nodes.Items {
		node := &nodes.Items[index]
		if !matchesName(node.Name, name) {
			continue
		}
		items = append(items, toKubeNode(node, podCounts))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

func (r *kubernetesRepo) SetNodeUnschedulable(ctx context.Context, clusterUID, name string, unschedulable bool) error {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	// 用 merge patch 只改 spec.unschedulable，不覆盖节点上的其它字段。
	patch := []byte(fmt.Sprintf(`{"spec":{"unschedulable":%t}}`, unschedulable))
	if _, err := client.CoreV1().Nodes().Patch(callCtx, name, types.MergePatchType, patch, metav1.PatchOptions{}); err != nil {
		return kubeError(r.log, err)
	}
	return nil
}

func (r *kubernetesRepo) ListWorkloads(ctx context.Context, clusterUID, namespace, kind, name string) ([]bo.KubeWorkload, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace = strings.TrimSpace(namespace)
	kind = strings.TrimSpace(kind)
	if !validWorkloadKind(kind) {
		return nil, merr.Invalid("invalid request")
	}
	items := make([]bo.KubeWorkload, 0)
	if kind == "" || kind == workloadKindDeployment {
		list, err := client.AppsV1().Deployments(namespace).List(callCtx, metav1.ListOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		for index := range list.Items {
			items = append(items, toKubeDeployment(&list.Items[index]))
		}
	}
	if kind == "" || kind == workloadKindStatefulSet {
		list, err := client.AppsV1().StatefulSets(namespace).List(callCtx, metav1.ListOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		for index := range list.Items {
			items = append(items, toKubeStatefulSet(&list.Items[index]))
		}
	}
	if kind == "" || kind == workloadKindDaemonSet {
		list, err := client.AppsV1().DaemonSets(namespace).List(callCtx, metav1.ListOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		for index := range list.Items {
			items = append(items, toKubeDaemonSet(&list.Items[index]))
		}
	}
	filtered := items[:0]
	for _, item := range items {
		if matchesName(item.Name, name) {
			filtered = append(filtered, item)
		}
	}
	sortKubeWorkloads(filtered)
	return filtered, nil
}

func (r *kubernetesRepo) ScaleWorkload(ctx context.Context, clusterUID, namespace, kind, name string, replicas int32) (*bo.KubeWorkload, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	switch kind {
	case workloadKindDeployment:
		scale, err := client.AppsV1().Deployments(namespace).GetScale(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		scale.Spec.Replicas = replicas
		if _, err := client.AppsV1().Deployments(namespace).UpdateScale(callCtx, name, scale, metav1.UpdateOptions{}); err != nil {
			return nil, kubeError(r.log, err)
		}
		deployment, err := client.AppsV1().Deployments(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeDeployment(deployment)
		return &item, nil
	case workloadKindStatefulSet:
		scale, err := client.AppsV1().StatefulSets(namespace).GetScale(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		scale.Spec.Replicas = replicas
		if _, err := client.AppsV1().StatefulSets(namespace).UpdateScale(callCtx, name, scale, metav1.UpdateOptions{}); err != nil {
			return nil, kubeError(r.log, err)
		}
		statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeStatefulSet(statefulSet)
		return &item, nil
	default:
		return nil, merr.Invalid("invalid request")
	}
}

// RestartWorkload 给 Pod 模板打重启时间注解，由控制器滚动重建 Pod。
func (r *kubernetesRepo) RestartWorkload(ctx context.Context, clusterUID, namespace, kind, name string) (*bo.KubeWorkload, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	restartedAt := time.Now().Format(time.RFC3339)
	patch := []byte(fmt.Sprintf(
		`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`,
		restartedAt,
	))
	// 用 strategic merge patch 保留模板上已有的其它注解。
	switch kind {
	case workloadKindDeployment:
		if _, err := client.AppsV1().Deployments(namespace).Patch(callCtx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}); err != nil {
			return nil, kubeError(r.log, err)
		}
		deployment, err := client.AppsV1().Deployments(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeDeployment(deployment)
		return &item, nil
	case workloadKindStatefulSet:
		if _, err := client.AppsV1().StatefulSets(namespace).Patch(callCtx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}); err != nil {
			return nil, kubeError(r.log, err)
		}
		statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeStatefulSet(statefulSet)
		return &item, nil
	case workloadKindDaemonSet:
		if _, err := client.AppsV1().DaemonSets(namespace).Patch(callCtx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}); err != nil {
			return nil, kubeError(r.log, err)
		}
		daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		item := toKubeDaemonSet(daemonSet)
		return &item, nil
	default:
		return nil, merr.Invalid("invalid request")
	}
}

func (r *kubernetesRepo) ListPods(ctx context.Context, clusterUID, namespace, name, ownerKind, ownerName string) ([]bo.KubePod, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.CoreV1().Pods(strings.TrimSpace(namespace)).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubePod, 0, len(list.Items))
	for index := range list.Items {
		item := toKubePod(&list.Items[index])
		if !matchesName(item.Name, name) {
			continue
		}
		if !matchesPodOwner(item, ownerKind, ownerName) {
			continue
		}
		items = append(items, item)
	}
	sortKubePods(items)
	return items, nil
}

func (r *kubernetesRepo) DeletePod(ctx context.Context, clusterUID, namespace, name string) error {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	// 不设 Force 和 GracePeriodSeconds：交给 API Server 按 Pod 自身的优雅退出时间删除。
	if err := client.CoreV1().Pods(namespace).Delete(callCtx, name, metav1.DeleteOptions{}); err != nil {
		return kubeError(r.log, err)
	}
	return nil
}

func (r *kubernetesRepo) PodLogs(ctx context.Context, clusterUID, namespace, name, container string, tailLines int32, previous bool) (*bo.KubePodLogs, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	container = strings.TrimSpace(container)
	if container == "" {
		pod, err := client.CoreV1().Pods(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		if len(pod.Spec.Containers) == 0 {
			return nil, merr.Invalid("pod has no containers")
		}
		container = pod.Spec.Containers[0].Name
	}
	tail := int64(tailLines)
	if tail <= 0 {
		tail = defaultPodLogTailLines
	}
	if tail > maxPodLogTailLines {
		tail = maxPodLogTailLines
	}
	limitBytes := int64(maxPodLogBytes)
	request := client.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
		Container:  container,
		TailLines:  &tail,
		LimitBytes: &limitBytes,
		Previous:   previous,
	})
	stream, err := request.Stream(callCtx)
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	defer func() { _ = stream.Close() }()
	content, err := io.ReadAll(io.LimitReader(stream, maxPodLogBytes))
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	return &bo.KubePodLogs{Content: string(content), Container: container}, nil
}

func (r *kubernetesRepo) ListIngresses(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeIngress, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.NetworkingV1().Ingresses(strings.TrimSpace(namespace)).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubeIngress, 0, len(list.Items))
	for index := range list.Items {
		ingress := &list.Items[index]
		if !matchesName(ingress.Name, name) {
			continue
		}
		items = append(items, toKubeIngress(ingress))
	}
	sortKubeIngresses(items)
	return items, nil
}

// validWorkloadKind 判断过滤用的工作负载类型是否受支持，空表示全部。
func validWorkloadKind(kind string) bool {
	switch kind {
	case "", workloadKindDeployment, workloadKindStatefulSet, workloadKindDaemonSet:
		return true
	default:
		return false
	}
}

// matchesPodOwner 按 Pod 的直接控制器过滤。按 Deployment 过滤时把 ReplicaSet 也算上，
// 因为 Deployment 的 Pod 直接属于 ReplicaSet；名称同时接受控制器全名和工作负载名加
// ReplicaSet 哈希前缀。
func matchesPodOwner(pod bo.KubePod, ownerKind, ownerName string) bool {
	ownerKind = strings.TrimSpace(ownerKind)
	ownerName = strings.TrimSpace(ownerName)
	if ownerKind != "" && pod.OwnerKind != ownerKind {
		if !(ownerKind == workloadKindDeployment && pod.OwnerKind == "ReplicaSet") {
			return false
		}
	}
	if ownerName != "" && pod.OwnerName != ownerName && !hasReplicaSetHash(pod.OwnerName, ownerName) {
		return false
	}
	return true
}

// hasReplicaSetHash 只把 <工作负载名>-<ReplicaSet 哈希> 当成同一个工作负载。
// 单纯的前缀匹配会把 "web" 也命中 "web-other-<hash>"。
func hasReplicaSetHash(ownerName, workloadName string) bool {
	hash, ok := strings.CutPrefix(ownerName, workloadName+"-")
	if !ok || len(hash) < 5 || len(hash) > 10 {
		return false
	}
	for _, char := range hash {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}

func (r *kubernetesRepo) ListServices(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeService, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.CoreV1().Services(strings.TrimSpace(namespace)).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubeService, 0, len(list.Items))
	for index := range list.Items {
		service := &list.Items[index]
		if !matchesName(service.Name, name) {
			continue
		}
		items = append(items, toKubeService(service))
	}
	sortKubeServices(items)
	return items, nil
}

func (r *kubernetesRepo) ListConfigMaps(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeConfigMap, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.CoreV1().ConfigMaps(strings.TrimSpace(namespace)).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubeConfigMap, 0, len(list.Items))
	for index := range list.Items {
		configMap := &list.Items[index]
		if !matchesName(configMap.Name, name) {
			continue
		}
		items = append(items, toKubeConfigMap(configMap))
	}
	sortKubeConfigMaps(items)
	return items, nil
}

func (r *kubernetesRepo) ListSecrets(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubeSecret, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.CoreV1().Secrets(strings.TrimSpace(namespace)).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubeSecret, 0, len(list.Items))
	for index := range list.Items {
		secret := &list.Items[index]
		if !matchesName(secret.Name, name) {
			continue
		}
		items = append(items, toKubeSecret(secret))
	}
	sortKubeSecrets(items)
	return items, nil
}

// ListPersistentVolumes 是集群级资源，不接受命名空间过滤。
func (r *kubernetesRepo) ListPersistentVolumes(ctx context.Context, clusterUID, name string) ([]bo.KubePersistentVolume, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.CoreV1().PersistentVolumes().List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubePersistentVolume, 0, len(list.Items))
	for index := range list.Items {
		volume := &list.Items[index]
		if !matchesName(volume.Name, name) {
			continue
		}
		items = append(items, toKubePersistentVolume(volume))
	}
	sortKubePersistentVolumes(items)
	return items, nil
}

func (r *kubernetesRepo) ListPersistentVolumeClaims(ctx context.Context, clusterUID, namespace, name string) ([]bo.KubePersistentVolumeClaim, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	list, err := client.CoreV1().PersistentVolumeClaims(strings.TrimSpace(namespace)).List(callCtx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubePersistentVolumeClaim, 0, len(list.Items))
	for index := range list.Items {
		claim := &list.Items[index]
		if !matchesName(claim.Name, name) {
			continue
		}
		items = append(items, toKubePersistentVolumeClaim(claim))
	}
	sortKubePersistentVolumeClaims(items)
	return items, nil
}
