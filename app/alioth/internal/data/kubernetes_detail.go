package data

import (
	"context"
	"sort"

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
		replicaSets, err := client.AppsV1().ReplicaSets(namespace).List(callCtx, metav1.ListOptions{})
		if err != nil {
			return nil, kubeError(r.log, err)
		}
		owned := make(map[string]struct{}, len(replicaSets.Items))
		for index := range replicaSets.Items {
			replicaSet := &replicaSets.Items[index]
			if !ownedBy(replicaSet.OwnerReferences, workloadKindDeployment, deployment.Name) {
				continue
			}
			owned[replicaSet.Name] = struct{}{}
			detail.ReplicaSets = append(detail.ReplicaSets, toKubeReplicaSet(replicaSet))
		}
		sortKubeReplicaSets(detail.ReplicaSets)
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
