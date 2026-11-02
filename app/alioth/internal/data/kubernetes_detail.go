package data

import (
	"context"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// resourceRefKind 区分工作负载模板引用的是 ConfigMap、Secret 还是 PVC。
type resourceRefKind int

const (
	refConfigMap resourceRefKind = iota
	refSecret
	refPersistentVolumeClaim
)

// WorkloadDetail 返回工作负载、名下的 ReplicaSet（仅 Deployment）和 Pod。
func (r *kubernetesRepo) WorkloadDetail(ctx context.Context, clusterUID, kind, namespace, name string) (*bo.KubeWorkloadDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	detail := &bo.KubeWorkloadDetail{}
	switch kind {
	case workloadKindDeployment:
		deployment, err := client.AppsV1().Deployments(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		detail.Workload = toKubeDeployment(deployment)
		detail.Conditions = deploymentConditions(deployment.Status.Conditions)
		detail.ReplicaStatus = deploymentReplicaStatus(deployment)
		replicaSets, err := client.AppsV1().ReplicaSets(namespace).List(callCtx, metav1.ListOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		// ReplicaSet 的历史版本号来自注解，Deployment 自身也带同一注解；
		// 两者相等的就是当前版本，供详情抽屉标出「当前」。
		currentRevision, _ := revisionOf(deployment.Annotations)
		owned := make(map[string]struct{}, len(replicaSets.Items))
		for index := range replicaSets.Items {
			replicaSet := &replicaSets.Items[index]
			if !ownedBy(replicaSet.OwnerReferences, workloadKindDeployment, deployment.Name) {
				continue
			}
			owned[replicaSet.Name] = struct{}{}
			item := toKubeReplicaSet(replicaSet)
			item.Revision, _ = revisionOf(replicaSet.Annotations)
			item.Current = item.Revision > 0 && item.Revision == currentRevision
			detail.ReplicaSets = append(detail.ReplicaSets, item)
		}
		sortKubeReplicaSets(detail.ReplicaSets)
		detail.ReplicaSets = trimReplicaSetsByRevision(detail.ReplicaSets)
		pods, err := client.CoreV1().Pods(namespace).List(callCtx, metav1.ListOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		for index := range pods.Items {
			pod := &pods.Items[index]
			// Deployment 的 Pod 直接属于 ReplicaSet，需要再按名称落到本 Deployment 的 RS 上。
			if len(pod.OwnerReferences) == 0 || pod.OwnerReferences[0].Kind != "ReplicaSet" {
				continue
			}
			if _, ok := owned[pod.OwnerReferences[0].Name]; !ok {
				continue
			}
			detail.Pods = append(detail.Pods, toKubePod(pod))
		}
		sortKubePods(detail.Pods)
	case workloadKindStatefulSet:
		statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		detail.Workload = toKubeStatefulSet(statefulSet)
		detail.Conditions = statefulSetConditions(statefulSet.Status.Conditions)
		detail.ReplicaStatus = statefulSetReplicaStatus(statefulSet)
		detail.Pods, err = r.ownedPods(callCtx, client, namespace, workloadKindStatefulSet, name)
		if err != nil {
			return nil, err
		}
	case workloadKindDaemonSet:
		daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(callCtx, name, metav1.GetOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		detail.Workload = toKubeDaemonSet(daemonSet)
		detail.Conditions = daemonSetConditions(daemonSet.Status.Conditions)
		detail.ReplicaStatus = daemonSetReplicaStatus(daemonSet)
		detail.Pods, err = r.ownedPods(callCtx, client, namespace, workloadKindDaemonSet, name)
		if err != nil {
			return nil, err
		}
	default:
		return nil, merr.Invalid("invalid resource kind")
	}
	return detail, nil
}

// ownedPods 返回直接归属于指定工作负载的 Pod。
func (r *kubernetesRepo) ownedPods(ctx context.Context, client kubernetes.Interface, namespace, kind, name string) ([]bo.KubePod, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	items := make([]bo.KubePod, 0, len(pods.Items))
	for index := range pods.Items {
		pod := &pods.Items[index]
		if len(pod.OwnerReferences) == 0 {
			continue
		}
		if pod.OwnerReferences[0].Kind != kind || pod.OwnerReferences[0].Name != name {
			continue
		}
		items = append(items, toKubePod(pod))
	}
	sortKubePods(items)
	return items, nil
}

// IngressDetail 返回 Ingress 和规则里引用的 Service。找不到的 Service 会被跳过。
func (r *kubernetesRepo) IngressDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeIngressDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	detail := &bo.KubeIngressDetail{Ingress: toKubeIngress(ingress)}
	for _, serviceName := range ingressServiceNames(ingress) {
		service, err := client.CoreV1().Services(namespace).Get(callCtx, serviceName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// 规则可能指向尚未创建或已删除的 Service，详情仍然可用。
				continue
			}
			return nil, kubeError(r.log, err)
		}
		detail.Services = append(detail.Services, toKubeService(service))
	}
	sortKubeServices(detail.Services)
	return detail, nil
}

// ServiceDetail 返回 Service 和 Endpoints。没有 Endpoints 时返回空列表。
func (r *kubernetesRepo) ServiceDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeServiceDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	service, err := client.CoreV1().Services(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	detail := &bo.KubeServiceDetail{Service: toKubeService(service)}
	endpoints, err := client.CoreV1().Endpoints(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return detail, nil
		}
		return nil, kubeError(r.log, err)
	}
	detail.Endpoints = toKubeEndpoints(endpoints)
	return detail, nil
}

// ConfigMapDetail 返回 ConfigMap 和引用它的工作负载。
func (r *kubernetesRepo) ConfigMapDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeConfigMapDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	related, err := r.relatedWorkloads(callCtx, client, namespace, refConfigMap, name)
	if err != nil {
		return nil, err
	}
	return &bo.KubeConfigMapDetail{ConfigMap: toKubeConfigMap(configMap), RelatedWorkloads: related}, nil
}

// SecretDetail 返回 Secret 和引用它的工作负载。Secret 只有类型和键名，没有值。
func (r *kubernetesRepo) SecretDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeSecretDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	secret, err := client.CoreV1().Secrets(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	related, err := r.relatedWorkloads(callCtx, client, namespace, refSecret, name)
	if err != nil {
		return nil, err
	}
	return &bo.KubeSecretDetail{Secret: toKubeSecret(secret), RelatedWorkloads: related}, nil
}

// PersistentVolumeClaimDetail 返回 PVC 和挂载它的工作负载。
func (r *kubernetesRepo) PersistentVolumeClaimDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubePersistentVolumeClaimDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	claim, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	related, err := r.relatedWorkloads(callCtx, client, namespace, refPersistentVolumeClaim, name)
	if err != nil {
		return nil, err
	}
	return &bo.KubePersistentVolumeClaimDetail{Claim: toKubePersistentVolumeClaim(claim), RelatedWorkloads: related}, nil
}

// relatedWorkloads 扫描命名空间中的 Deployment、StatefulSet 和 DaemonSet 的 Pod 模板，
// 返回引用了目标 ConfigMap / Secret / PVC 的工作负载，按类型和名称排序去重。
func (r *kubernetesRepo) relatedWorkloads(ctx context.Context, client kubernetes.Interface, namespace string, refKind resourceRefKind, name string) ([]bo.KubeWorkloadRef, error) {
	references := make([]bo.KubeWorkloadRef, 0)
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	for index := range deployments.Items {
		deployment := &deployments.Items[index]
		if podTemplateReferences(deployment.Spec.Template.Spec, refKind, name) {
			references = append(references, bo.KubeWorkloadRef{Kind: workloadKindDeployment, Namespace: namespace, Name: deployment.Name})
		}
	}
	statefulSets, err := client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	for index := range statefulSets.Items {
		statefulSet := &statefulSets.Items[index]
		if podTemplateReferences(statefulSet.Spec.Template.Spec, refKind, name) {
			references = append(references, bo.KubeWorkloadRef{Kind: workloadKindStatefulSet, Namespace: namespace, Name: statefulSet.Name})
		}
	}
	daemonSets, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	for index := range daemonSets.Items {
		daemonSet := &daemonSets.Items[index]
		if podTemplateReferences(daemonSet.Spec.Template.Spec, refKind, name) {
			references = append(references, bo.KubeWorkloadRef{Kind: workloadKindDaemonSet, Namespace: namespace, Name: daemonSet.Name})
		}
	}
	sort.Slice(references, func(i, j int) bool {
		if references[i].Kind != references[j].Kind {
			return references[i].Kind < references[j].Kind
		}
		return references[i].Name < references[j].Name
	})
	return references, nil
}

// podTemplateReferences 判断 Pod 模板是否引用目标资源：卷（含 projected）、
// envFrom、env.valueFrom；Secret 另含 imagePullSecrets。
func podTemplateReferences(spec corev1.PodSpec, refKind resourceRefKind, name string) bool {
	for _, volume := range spec.Volumes {
		switch refKind {
		case refConfigMap:
			if volume.ConfigMap != nil && volume.ConfigMap.Name == name {
				return true
			}
		case refSecret:
			if volume.Secret != nil && volume.Secret.SecretName == name {
				return true
			}
		case refPersistentVolumeClaim:
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == name {
				return true
			}
		}
		if volume.Projected == nil || (refKind != refConfigMap && refKind != refSecret) {
			continue
		}
		for _, source := range volume.Projected.Sources {
			if refKind == refConfigMap && source.ConfigMap != nil && source.ConfigMap.Name == name {
				return true
			}
			if refKind == refSecret && source.Secret != nil && source.Secret.Name == name {
				return true
			}
		}
	}
	// envFrom 和 env.valueFrom 只出现在普通容器与初始化容器上。
	containers := make([]corev1.Container, 0, len(spec.InitContainers)+len(spec.Containers))
	containers = append(containers, spec.InitContainers...)
	containers = append(containers, spec.Containers...)
	for _, container := range containers {
		for _, envFrom := range container.EnvFrom {
			if refKind == refConfigMap && envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == name {
				return true
			}
			if refKind == refSecret && envFrom.SecretRef != nil && envFrom.SecretRef.Name == name {
				return true
			}
		}
		for _, env := range container.Env {
			if env.ValueFrom == nil {
				continue
			}
			if refKind == refConfigMap && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == name {
				return true
			}
			if refKind == refSecret && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == name {
				return true
			}
		}
	}
	if refKind != refSecret {
		return false
	}
	for _, imagePullSecret := range spec.ImagePullSecrets {
		if imagePullSecret.Name == name {
			return true
		}
	}
	return false
}

// toKubeReplicaSet 只保留详情抽屉需要的副本数、镜像和时间。
func toKubeReplicaSet(replicaSet *appsv1.ReplicaSet) bo.KubeReplicaSet {
	item := bo.KubeReplicaSet{
		Namespace: replicaSet.Namespace,
		Name:      replicaSet.Name,
		Images:    workloadImages(replicaSet.Spec.Template.Spec),
		CreatedAt: unixTime(replicaSet.CreationTimestamp),
	}
	if replicaSet.Spec.Replicas != nil {
		item.DesiredReplicas = *replicaSet.Spec.Replicas
	}
	item.ReadyReplicas = replicaSet.Status.ReadyReplicas
	item.CurrentReplicas = replicaSet.Status.Replicas
	return item
}

func sortKubeReplicaSets(items []bo.KubeReplicaSet) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
}

// ingressServiceNames 汇总 defaultBackend 和规则中引用的 Service 名称，保持首次出现顺序并去重。
func ingressServiceNames(ingress *networkingv1.Ingress) []string {
	names := make([]string, 0, len(ingress.Spec.Rules)+1)
	seen := make(map[string]struct{}, len(ingress.Spec.Rules)+1)
	add := func(name string) {
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	if ingress.Spec.DefaultBackend != nil && ingress.Spec.DefaultBackend.Service != nil {
		add(ingress.Spec.DefaultBackend.Service.Name)
	}
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service != nil {
				add(path.Backend.Service.Name)
			}
		}
	}
	return names
}

// toKubeEndpoints 把 Endpoints 的各个子集摊平成地址列表，就绪和未就绪分开标记。
func toKubeEndpoints(endpoints *corev1.Endpoints) []bo.KubeEndpoint {
	items := make([]bo.KubeEndpoint, 0)
	for _, subset := range endpoints.Subsets {
		ports := make([]bo.KubeEndpointPort, 0, len(subset.Ports))
		for _, port := range subset.Ports {
			ports = append(ports, bo.KubeEndpointPort{Name: port.Name, Port: port.Port, Protocol: string(port.Protocol)})
		}
		for _, address := range subset.Addresses {
			items = append(items, toKubeEndpoint(address, ports, true))
		}
		for _, address := range subset.NotReadyAddresses {
			items = append(items, toKubeEndpoint(address, ports, false))
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Address != items[j].Address {
			return items[i].Address < items[j].Address
		}
		// 同一地址同时出现在就绪和未就绪列表时，就绪的排在前面。
		return items[i].Ready && !items[j].Ready
	})
	return items
}

func toKubeEndpoint(address corev1.EndpointAddress, ports []bo.KubeEndpointPort, ready bool) bo.KubeEndpoint {
	item := bo.KubeEndpoint{
		Address:  address.IP,
		Hostname: address.Hostname,
		Ready:    ready,
		Ports:    append([]bo.KubeEndpointPort(nil), ports...),
	}
	if address.NodeName != nil {
		item.NodeName = *address.NodeName
	}
	if address.TargetRef != nil {
		item.TargetKind = address.TargetRef.Kind
		item.TargetName = address.TargetRef.Name
	}
	return item
}

// ownedBy 判断对象的直接属主是否是指定类型和名称。
func ownedBy(ownerReferences []metav1.OwnerReference, kind, name string) bool {
	for _, reference := range ownerReferences {
		if reference.Kind == kind && reference.Name == name {
			return true
		}
	}
	return false
}

// maxWorkloadReplicaSets 是 Deployment 详情返回的 ReplicaSet 历史条数上限。
const maxWorkloadReplicaSets = 10

// trimReplicaSetsByRevision 按版本号倒序排列并截断到上限。没有版本号的历史对象
// 排在最后，仍按名称升序，保持详情可读。
func trimReplicaSetsByRevision(items []bo.KubeReplicaSet) []bo.KubeReplicaSet {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Revision != items[j].Revision {
			return items[i].Revision > items[j].Revision
		}
		return items[i].Name < items[j].Name
	})
	if len(items) > maxWorkloadReplicaSets {
		items = items[:maxWorkloadReplicaSets]
	}
	return items
}

// ServerVersion 读取 API Server 的 gitVersion，用于概况展示实时版本。
func (r *kubernetesRepo) ServerVersion(ctx context.Context, clusterUID string) (string, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return "", err
	}
	// Discovery 不接受上下文；客户端自身的请求超时负责兜底。
	info, err := client.Discovery().ServerVersion()
	if err != nil {
		return "", kubeError(r.log, err)
	}
	return info.GitVersion, nil
}

// NodeDetail 返回节点基础信息、全部条件、污点与容量。Pod 数依赖全量 Pod 列表，
// 缺少 Pod 读取权限时按 0 计，不影响节点详情。
func (r *kubernetesRepo) NodeDetail(ctx context.Context, clusterUID, name string) (*bo.KubeNodeDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, merr.Invalid("node name is required")
	}
	node, err := client.CoreV1().Nodes().Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	podCounts, err := r.podCountsByNode(callCtx, client)
	if err != nil {
		return nil, err
	}
	return &bo.KubeNodeDetail{
		Node:        toKubeNode(node, podCounts),
		Conditions:  nodeConditions(node.Status.Conditions),
		Taints:      nodeTaints(node.Spec.Taints),
		Capacity:    resourceListMap(node.Status.Capacity),
		Allocatable: resourceListMap(node.Status.Allocatable),
	}, nil
}

// PodDetail 返回 Pod 基础信息、容器状态、条件、卷与归属，不含事件。
func (r *kubernetesRepo) PodDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubePodDetail, error) {
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
	defer cancel()
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, merr.Invalid("namespace and name are required")
	}
	pod, err := client.CoreV1().Pods(namespace).Get(callCtx, name, metav1.GetOptions{})
	if err != nil {
		return nil, kubeError(r.log, err)
	}
	return &bo.KubePodDetail{
		Pod:             toKubePod(pod),
		Containers:      containerStatuses(pod, pod.Status.ContainerStatuses),
		InitContainers:  containerStatuses(pod, pod.Status.InitContainerStatuses),
		Conditions:      podConditions(pod.Status.Conditions),
		Volumes:         podVolumes(pod.Spec.Volumes),
		OwnerReferences: ownerReferenceItems(pod.OwnerReferences),
	}, nil
}

// nodeConditions 转换节点条件。
func nodeConditions(conditions []corev1.NodeCondition) []bo.KubeCondition {
	if len(conditions) == 0 {
		return nil
	}
	items := make([]bo.KubeCondition, 0, len(conditions))
	for _, condition := range conditions {
		items = append(items, bo.KubeCondition{
			Type:             string(condition.Type),
			Status:           string(condition.Status),
			Reason:           condition.Reason,
			Message:          condition.Message,
			LastTransitionAt: unixTime(condition.LastTransitionTime),
		})
	}
	return items
}

// podConditions 转换 Pod 条件。
func podConditions(conditions []corev1.PodCondition) []bo.KubeCondition {
	if len(conditions) == 0 {
		return nil
	}
	items := make([]bo.KubeCondition, 0, len(conditions))
	for _, condition := range conditions {
		items = append(items, bo.KubeCondition{
			Type:             string(condition.Type),
			Status:           string(condition.Status),
			Reason:           condition.Reason,
			Message:          condition.Message,
			LastTransitionAt: unixTime(condition.LastTransitionTime),
		})
	}
	return items
}

// deploymentConditions 转换 Deployment 条件。
func deploymentConditions(conditions []appsv1.DeploymentCondition) []bo.KubeCondition {
	if len(conditions) == 0 {
		return nil
	}
	items := make([]bo.KubeCondition, 0, len(conditions))
	for _, condition := range conditions {
		items = append(items, bo.KubeCondition{
			Type:             string(condition.Type),
			Status:           string(condition.Status),
			Reason:           condition.Reason,
			Message:          condition.Message,
			LastTransitionAt: unixTime(condition.LastTransitionTime),
		})
	}
	return items
}

// statefulSetConditions 转换 StatefulSet 条件。
func statefulSetConditions(conditions []appsv1.StatefulSetCondition) []bo.KubeCondition {
	if len(conditions) == 0 {
		return nil
	}
	items := make([]bo.KubeCondition, 0, len(conditions))
	for _, condition := range conditions {
		items = append(items, bo.KubeCondition{
			Type:             string(condition.Type),
			Status:           string(condition.Status),
			Reason:           condition.Reason,
			Message:          condition.Message,
			LastTransitionAt: unixTime(condition.LastTransitionTime),
		})
	}
	return items
}

// daemonSetConditions 转换 DaemonSet 条件。
func daemonSetConditions(conditions []appsv1.DaemonSetCondition) []bo.KubeCondition {
	if len(conditions) == 0 {
		return nil
	}
	items := make([]bo.KubeCondition, 0, len(conditions))
	for _, condition := range conditions {
		items = append(items, bo.KubeCondition{
			Type:             string(condition.Type),
			Status:           string(condition.Status),
			Reason:           condition.Reason,
			Message:          condition.Message,
			LastTransitionAt: unixTime(condition.LastTransitionTime),
		})
	}
	return items
}

// nodeTaints 转换节点污点。
func nodeTaints(taints []corev1.Taint) []bo.KubeTaint {
	if len(taints) == 0 {
		return nil
	}
	items := make([]bo.KubeTaint, 0, len(taints))
	for _, taint := range taints {
		items = append(items, bo.KubeTaint{Key: taint.Key, Value: taint.Value, Effect: string(taint.Effect)})
	}
	return items
}

// resourceListMap 把容量 / 可分配量转成字符串 map，键是资源名（含扩展资源）。
func resourceListMap(resources corev1.ResourceList) map[string]string {
	items := make(map[string]string, len(resources))
	for name, quantity := range resources {
		items[string(name)] = quantity.String()
	}
	return items
}

// deploymentReplicaStatus 汇总 Deployment 副本状态。不可用副本数取 status 的
// unavailableReplicas，缺失时按期望减可用计算。
func deploymentReplicaStatus(deployment *appsv1.Deployment) bo.KubeReplicaStatus {
	desired := int32(0)
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}
	status := bo.KubeReplicaStatus{
		Desired:   desired,
		Current:   deployment.Status.Replicas,
		Ready:     deployment.Status.ReadyReplicas,
		Updated:   deployment.Status.UpdatedReplicas,
		Available: deployment.Status.AvailableReplicas,
	}
	status.Unavailable = deployment.Status.UnavailableReplicas
	if status.Unavailable == 0 {
		status.Unavailable = missingReplicas(status.Desired, status.Available)
	}
	return status
}

// statefulSetReplicaStatus 汇总 StatefulSet 副本状态。
func statefulSetReplicaStatus(statefulSet *appsv1.StatefulSet) bo.KubeReplicaStatus {
	desired := int32(0)
	if statefulSet.Spec.Replicas != nil {
		desired = *statefulSet.Spec.Replicas
	}
	status := bo.KubeReplicaStatus{
		Desired:   desired,
		Current:   statefulSet.Status.CurrentReplicas,
		Ready:     statefulSet.Status.ReadyReplicas,
		Updated:   statefulSet.Status.UpdatedReplicas,
		Available: statefulSet.Status.AvailableReplicas,
	}
	status.Unavailable = missingReplicas(status.Desired, status.Available)
	return status
}

// daemonSetReplicaStatus 汇总 DaemonSet 副本状态，口径与列表里的期望 / 就绪一致。
func daemonSetReplicaStatus(daemonSet *appsv1.DaemonSet) bo.KubeReplicaStatus {
	status := bo.KubeReplicaStatus{
		Desired:   daemonSet.Status.DesiredNumberScheduled,
		Current:   daemonSet.Status.CurrentNumberScheduled,
		Ready:     daemonSet.Status.NumberReady,
		Updated:   daemonSet.Status.UpdatedNumberScheduled,
		Available: daemonSet.Status.NumberAvailable,
	}
	status.Unavailable = missingReplicas(status.Desired, status.Available)
	return status
}

// missingReplicas 返回期望减可用的差额，负数按 0 计。
func missingReplicas(desired, available int32) int32 {
	if desired <= available {
		return 0
	}
	return desired - available
}

// containerStatuses 按 Pod 里的容器顺序转换容器状态，并补上声明端口。
// 只看运行中的状态；已删除的容器不会出现在详情里。
func containerStatuses(pod *corev1.Pod, statuses []corev1.ContainerStatus) []bo.KubeContainerStatus {
	if len(statuses) == 0 {
		return nil
	}
	ports := podContainerPorts(pod)
	items := make([]bo.KubeContainerStatus, 0, len(statuses))
	for _, status := range statuses {
		item := bo.KubeContainerStatus{
			Name:     status.Name,
			Image:    status.Image,
			Ready:    status.Ready,
			Restarts: status.RestartCount,
			Ports:    ports[status.Name],
		}
		switch {
		case status.State.Running != nil:
			item.State = "RUNNING"
			item.StartedAt = unixTime(status.State.Running.StartedAt)
		case status.State.Waiting != nil:
			item.State = "WAITING"
			item.StateReason = status.State.Waiting.Reason
			item.StateMessage = status.State.Waiting.Message
		case status.State.Terminated != nil:
			item.State = "TERMINATED"
			item.StateReason = status.State.Terminated.Reason
			item.StateMessage = status.State.Terminated.Message
			item.StartedAt = unixTime(status.State.Terminated.StartedAt)
		}
		items = append(items, item)
	}
	return items
}

// podContainerPorts 汇总每个容器声明的端口，格式为 8080/TCP。
func podContainerPorts(pod *corev1.Pod) map[string][]string {
	containers := make([]corev1.Container, 0, len(pod.Spec.Containers)+len(pod.Spec.InitContainers))
	containers = append(containers, pod.Spec.Containers...)
	containers = append(containers, pod.Spec.InitContainers...)
	ports := make(map[string][]string, len(containers))
	for _, container := range containers {
		if len(container.Ports) == 0 {
			continue
		}
		values := make([]string, 0, len(container.Ports))
		for _, port := range container.Ports {
			protocol := string(port.Protocol)
			if protocol == "" {
				protocol = string(corev1.ProtocolTCP)
			}
			values = append(values, strconv.Itoa(int(port.ContainerPort))+"/"+protocol)
		}
		ports[container.Name] = values
	}
	return ports
}

// podVolumes 转换 Pod 卷。Secret / ConfigMap 只返回名称，不读取任何值。
func podVolumes(volumes []corev1.Volume) []bo.KubeVolume {
	if len(volumes) == 0 {
		return nil
	}
	items := make([]bo.KubeVolume, 0, len(volumes))
	for _, volume := range volumes {
		items = append(items, toKubeVolume(volume))
	}
	return items
}

// toKubeVolume 按卷来源给出类型和摘要。只读标记只取 PVC 的 readOnly，
// 其它卷类型没有稳定的只读语义。
func toKubeVolume(volume corev1.Volume) bo.KubeVolume {
	item := bo.KubeVolume{Name: volume.Name}
	switch {
	case volume.PersistentVolumeClaim != nil:
		item.Type = "PersistentVolumeClaim"
		item.Summary = volume.PersistentVolumeClaim.ClaimName
		item.ReadOnly = volume.PersistentVolumeClaim.ReadOnly
	case volume.Secret != nil:
		item.Type = "Secret"
		item.Summary = volume.Secret.SecretName
	case volume.ConfigMap != nil:
		item.Type = "ConfigMap"
		item.Summary = volume.ConfigMap.Name
	case volume.EmptyDir != nil:
		item.Type = "EmptyDir"
	case volume.HostPath != nil:
		item.Type = "HostPath"
		item.Summary = volume.HostPath.Path
	case volume.Projected != nil:
		item.Type = "Projected"
	case volume.DownwardAPI != nil:
		item.Type = "DownwardAPI"
	case volume.NFS != nil:
		item.Type = "NFS"
		item.Summary = volume.NFS.Server + ":" + volume.NFS.Path
	default:
		item.Type = "Other"
	}
	return item
}

// ownerReferenceItems 转换 ownerReferences。
func ownerReferenceItems(references []metav1.OwnerReference) []bo.KubeOwnerReference {
	if len(references) == 0 {
		return nil
	}
	items := make([]bo.KubeOwnerReference, 0, len(references))
	for _, reference := range references {
		controller := reference.Controller != nil && *reference.Controller
		items = append(items, bo.KubeOwnerReference{
			Kind: reference.Kind, Name: reference.Name, UID: string(reference.UID), Controller: controller,
		})
	}
	return items
}
