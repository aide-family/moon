package data

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

const (
	// nodeRoleLabelPrefix 是节点角色的标签前缀，值就是角色名。
	nodeRoleLabelPrefix = "node-role.kubernetes.io/"
	// ingressClassAnnotation 是旧版 Ingress 指定 class 的注解。
	ingressClassAnnotation = "kubernetes.io/ingress.class"
)

// unixTime 把 Kubernetes 时间转成 Unix 秒；零值返回 0，避免未设置时间的对象返回负数。
func unixTime(value metav1.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.Unix()
}

// toKubeNode 把节点和 Pod 计数转成业务对象。
// podCounts 由一次全量 Pod 列表按节点名聚合得到。
func toKubeNode(node *corev1.Node, podCounts map[string]int32) bo.KubeNode {
	item := bo.KubeNode{
		Name:           node.Name,
		Status:         nodeStatus(node.Status),
		Schedulable:    !node.Spec.Unschedulable,
		Roles:          nodeRoles(node.Labels),
		InternalIP:     nodeInternalIP(node.Status.Addresses),
		KubeletVersion: node.Status.NodeInfo.KubeletVersion,
		OSImage:        node.Status.NodeInfo.OSImage,
		PodCount:       podCounts[node.Name],
		CreatedAt:      unixTime(node.CreationTimestamp),
	}
	if quantity, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
		item.CPUCapacity = quantity.String()
	}
	if quantity, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
		item.MemoryCapacity = quantity.String()
	}
	return item
}

// nodeStatus 只按 Ready condition 判定状态，其它 condition 不影响展示。
func nodeStatus(status corev1.NodeStatus) string {
	for _, condition := range status.Conditions {
		if condition.Type != corev1.NodeReady {
			continue
		}
		switch condition.Status {
		case corev1.ConditionTrue:
			return "Ready"
		case corev1.ConditionFalse:
			return "NotReady"
		default:
			return "Unknown"
		}
	}
	return "Unknown"
}

// nodeRoles 从 node-role.kubernetes.io/* 标签取角色名，按字母序返回。
func nodeRoles(labels map[string]string) []string {
	roles := make([]string, 0, len(labels))
	for key := range labels {
		if !strings.HasPrefix(key, nodeRoleLabelPrefix) {
			continue
		}
		role := strings.TrimPrefix(key, nodeRoleLabelPrefix)
		if role == "" {
			continue
		}
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roles
}

func nodeInternalIP(addresses []corev1.NodeAddress) string {
	for _, address := range addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}
	return ""
}

func toKubeDeployment(deployment *appsv1.Deployment) bo.KubeWorkload {
	item := bo.KubeWorkload{
		Namespace: deployment.Namespace,
		Kind:      "Deployment",
		Name:      deployment.Name,
		Images:    workloadImages(deployment.Spec.Template.Spec),
		CreatedAt: unixTime(deployment.CreationTimestamp),
	}
	if deployment.Spec.Replicas != nil {
		item.DesiredReplicas = *deployment.Spec.Replicas
	}
	item.ReadyReplicas = deployment.Status.ReadyReplicas
	item.UpdatedReplicas = deployment.Status.UpdatedReplicas
	item.AvailableReplicas = deployment.Status.AvailableReplicas
	return item
}

func toKubeStatefulSet(statefulSet *appsv1.StatefulSet) bo.KubeWorkload {
	item := bo.KubeWorkload{
		Namespace: statefulSet.Namespace,
		Kind:      "StatefulSet",
		Name:      statefulSet.Name,
		Images:    workloadImages(statefulSet.Spec.Template.Spec),
		CreatedAt: unixTime(statefulSet.CreationTimestamp),
	}
	if statefulSet.Spec.Replicas != nil {
		item.DesiredReplicas = *statefulSet.Spec.Replicas
	}
	item.ReadyReplicas = statefulSet.Status.ReadyReplicas
	item.UpdatedReplicas = statefulSet.Status.UpdatedReplicas
	item.AvailableReplicas = statefulSet.Status.AvailableReplicas
	return item
}

// toKubeDaemonSet 的期望副本数取应调度的节点数，可用副本数取已经就绪的节点数。
func toKubeDaemonSet(daemonSet *appsv1.DaemonSet) bo.KubeWorkload {
	return bo.KubeWorkload{
		Namespace:         daemonSet.Namespace,
		Kind:              "DaemonSet",
		Name:              daemonSet.Name,
		DesiredReplicas:   daemonSet.Status.DesiredNumberScheduled,
		ReadyReplicas:     daemonSet.Status.NumberReady,
		UpdatedReplicas:   daemonSet.Status.UpdatedNumberScheduled,
		AvailableReplicas: daemonSet.Status.NumberAvailable,
		Images:            workloadImages(daemonSet.Spec.Template.Spec),
		CreatedAt:         unixTime(daemonSet.CreationTimestamp),
	}
}

// workloadImages 汇总普通容器和初始化容器的镜像，按出现顺序去重。
func workloadImages(spec corev1.PodSpec) []string {
	images := make([]string, 0, len(spec.Containers)+len(spec.InitContainers))
	seen := make(map[string]struct{}, len(spec.Containers)+len(spec.InitContainers))
	add := func(image string) {
		if image == "" {
			return
		}
		if _, ok := seen[image]; ok {
			return
		}
		seen[image] = struct{}{}
		images = append(images, image)
	}
	for _, container := range spec.Containers {
		add(container.Image)
	}
	for _, container := range spec.InitContainers {
		add(container.Image)
	}
	return images
}

func toKubePod(pod *corev1.Pod) bo.KubePod {
	item := bo.KubePod{
		Namespace: pod.Namespace,
		Name:      pod.Name,
		Phase:     string(pod.Status.Phase),
		NodeName:  pod.Spec.NodeName,
		PodIP:     pod.Status.PodIP,
		CreatedAt: unixTime(pod.CreationTimestamp),
	}
	if len(pod.OwnerReferences) > 0 {
		// 直接所属控制器就是 ownerReferences[0]；Deployment 的 Pod 这里会是 ReplicaSet。
		item.OwnerKind = pod.OwnerReferences[0].Kind
		item.OwnerName = pod.OwnerReferences[0].Name
	}
	statuses := make(map[string]corev1.ContainerStatus, len(pod.Status.ContainerStatuses))
	for _, status := range pod.Status.ContainerStatuses {
		statuses[status.Name] = status
		item.Restarts += status.RestartCount
	}
	ready := 0
	item.Containers = make([]bo.KubeContainer, 0, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		status, ok := statuses[container.Name]
		containerReady := ok && status.Ready
		if containerReady {
			ready++
		}
		item.Containers = append(item.Containers, bo.KubeContainer{
			Name:     container.Name,
			Ready:    containerReady,
			Restarts: status.RestartCount,
			Image:    container.Image,
		})
	}
	item.Ready = fmt.Sprintf("%d/%d", ready, len(pod.Spec.Containers))
	return item
}

func toKubeIngress(ingress *networkingv1.Ingress) bo.KubeIngress {
	return bo.KubeIngress{
		Namespace: ingress.Namespace,
		Name:      ingress.Name,
		ClassName: ingressClassName(ingress),
		Hosts:     ingressHosts(ingress),
		Paths:     ingressPaths(ingress),
		TLS:       ingressTLS(ingress),
		CreatedAt: unixTime(ingress.CreationTimestamp),
	}
}

// ingressClassName 优先取 spec.ingressClassName，旧配置回退到注解。
func ingressClassName(ingress *networkingv1.Ingress) string {
	if ingress.Spec.IngressClassName != nil && *ingress.Spec.IngressClassName != "" {
		return *ingress.Spec.IngressClassName
	}
	if ingress.Annotations == nil {
		return ""
	}
	return ingress.Annotations[ingressClassAnnotation]
}

// ingressHosts 汇总规则和 TLS 中的主机名，保持出现顺序并去重。
func ingressHosts(ingress *networkingv1.Ingress) []string {
	hosts := make([]string, 0, len(ingress.Spec.Rules)+len(ingress.Spec.TLS))
	seen := make(map[string]struct{}, len(ingress.Spec.Rules)+len(ingress.Spec.TLS))
	add := func(host string) {
		host = strings.TrimSpace(host)
		if host == "" {
			return
		}
		if _, ok := seen[host]; ok {
			return
		}
		seen[host] = struct{}{}
		hosts = append(hosts, host)
	}
	for _, rule := range ingress.Spec.Rules {
		add(rule.Host)
	}
	for _, tls := range ingress.Spec.TLS {
		for _, host := range tls.Hosts {
			add(host)
		}
	}
	return hosts
}

func ingressPaths(ingress *networkingv1.Ingress) []bo.KubeIngressPath {
	paths := make([]bo.KubeIngressPath, 0, len(ingress.Spec.Rules))
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			paths = append(paths, bo.KubeIngressPath{
				Host:        rule.Host,
				Path:        path.Path,
				PathType:    ingressPathType(path.PathType),
				ServiceName: ingressServiceName(path.Backend),
				ServicePort: ingressServicePort(path.Backend),
			})
		}
	}
	return paths
}

func ingressPathType(pathType *networkingv1.PathType) string {
	if pathType == nil {
		return ""
	}
	return string(*pathType)
}

func ingressServiceName(backend networkingv1.IngressBackend) string {
	if backend.Service == nil {
		return ""
	}
	return backend.Service.Name
}

// ingressServicePort 按端口名优先、端口号兜底转成字符串。
func ingressServicePort(backend networkingv1.IngressBackend) string {
	if backend.Service == nil {
		return ""
	}
	if backend.Service.Port.Name != "" {
		return backend.Service.Port.Name
	}
	if backend.Service.Port.Number != 0 {
		return strconv.Itoa(int(backend.Service.Port.Number))
	}
	return ""
}

// ingressTLS 返回 TLS 使用的 Secret 名称，按出现顺序去重。
func ingressTLS(ingress *networkingv1.Ingress) []string {
	secrets := make([]string, 0, len(ingress.Spec.TLS))
	seen := make(map[string]struct{}, len(ingress.Spec.TLS))
	for _, tls := range ingress.Spec.TLS {
		if tls.SecretName == "" {
			continue
		}
		if _, ok := seen[tls.SecretName]; ok {
			continue
		}
		seen[tls.SecretName] = struct{}{}
		secrets = append(secrets, tls.SecretName)
	}
	return secrets
}

// matchesName 是名称过滤的统一口径：不区分大小写的包含匹配，空过滤表示全部。
func matchesName(value, filter string) bool {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(filter))
}

func sortKubeWorkloads(items []bo.KubeWorkload) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		return items[i].Name < items[j].Name
	})
}

func sortKubePods(items []bo.KubePod) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})
}

func sortKubeIngresses(items []bo.KubeIngress) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})
}

func toKubeService(service *corev1.Service) bo.KubeService {
	return bo.KubeService{
		Namespace:      service.Namespace,
		Name:           service.Name,
		Type:           string(service.Spec.Type),
		ClusterIP:      service.Spec.ClusterIP,
		ExternalIPs:    append([]string(nil), service.Spec.ExternalIPs...),
		LoadBalancerIP: serviceLoadBalancerIP(service.Status.LoadBalancer),
		Ports:          kubeServicePorts(service.Spec.Ports),
		Selector:       serviceSelector(service.Spec.Selector),
		CreatedAt:      unixTime(service.CreationTimestamp),
	}
}

// kubeServicePorts 按声明顺序转换端口，不重排。
func kubeServicePorts(ports []corev1.ServicePort) []bo.KubeServicePort {
	out := make([]bo.KubeServicePort, 0, len(ports))
	for _, port := range ports {
		out = append(out, bo.KubeServicePort{
			Name:       port.Name,
			Protocol:   string(port.Protocol),
			Port:       port.Port,
			TargetPort: serviceTargetPort(port.TargetPort),
			NodePort:   port.NodePort,
		})
	}
	return out
}

// serviceTargetPort 把 intstr 目标端口转成名称或数字。未设置时返回空串。
func serviceTargetPort(target intstr.IntOrString) string {
	if target.Type == intstr.String {
		return target.StrVal
	}
	if target.IntVal == 0 {
		return ""
	}
	return strconv.Itoa(int(target.IntVal))
}

// serviceSelector 把选择器序列化成 JSON 对象文本。空选择器返回 {}，
// 调用方始终能按 JSON 解析。encoding/json 对 map 键排序，输出稳定。
func serviceSelector(selector map[string]string) string {
	if len(selector) == 0 {
		return "{}"
	}
	encoded, err := json.Marshal(selector)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

// serviceLoadBalancerIP 取负载均衡入口的第一个 IP，没有入口时为空。
func serviceLoadBalancerIP(status corev1.LoadBalancerStatus) string {
	if len(status.Ingress) == 0 {
		return ""
	}
	return status.Ingress[0].IP
}

func toKubeConfigMap(configMap *corev1.ConfigMap) bo.KubeConfigMap {
	return bo.KubeConfigMap{
		Namespace:      configMap.Namespace,
		Name:           configMap.Name,
		DataKeys:       sortedStringKeys(configMap.Data),
		BinaryDataKeys: sortedByteKeys(configMap.BinaryData),
		Immutable:      configMap.Immutable != nil && *configMap.Immutable,
		CreatedAt:      unixTime(configMap.CreationTimestamp),
	}
}

func toKubeSecret(secret *corev1.Secret) bo.KubeSecret {
	return bo.KubeSecret{
		Namespace: secret.Namespace,
		Name:      secret.Name,
		Type:      string(secret.Type),
		DataKeys:  sortedByteKeys(secret.Data),
		Immutable: secret.Immutable != nil && *secret.Immutable,
		CreatedAt: unixTime(secret.CreationTimestamp),
	}
}

// sortedStringKeys 只返回键名并按字母序排列，值不会进入调用方。
func sortedStringKeys(data map[string]string) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// sortedByteKeys 只返回键名并按字母序排列，值不会进入调用方。
func sortedByteKeys(data map[string][]byte) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func toKubePersistentVolume(volume *corev1.PersistentVolume) bo.KubePersistentVolume {
	return bo.KubePersistentVolume{
		Name:          volume.Name,
		Capacity:      storageCapacity(volume.Spec.Capacity),
		StorageClass:  volume.Spec.StorageClassName,
		AccessModes:   kubeAccessModes(volume.Spec.AccessModes),
		Claim:         persistentVolumeClaimRef(volume.Spec.ClaimRef),
		Phase:         string(volume.Status.Phase),
		ReclaimPolicy: string(volume.Spec.PersistentVolumeReclaimPolicy),
		CreatedAt:     unixTime(volume.CreationTimestamp),
	}
}

// persistentVolumeClaimRef 把绑定的 PVC 表示成 namespace/name；未绑定时为空。
func persistentVolumeClaimRef(claimRef *corev1.ObjectReference) string {
	if claimRef == nil || claimRef.Name == "" {
		return ""
	}
	if claimRef.Namespace == "" {
		return claimRef.Name
	}
	return claimRef.Namespace + "/" + claimRef.Name
}

func toKubePersistentVolumeClaim(claim *corev1.PersistentVolumeClaim) bo.KubePersistentVolumeClaim {
	// 未绑定的 PVC 还没有 status.capacity，展示申请容量。
	capacity := storageCapacity(claim.Status.Capacity)
	if capacity == "" {
		capacity = storageCapacity(claim.Spec.Resources.Requests)
	}
	return bo.KubePersistentVolumeClaim{
		Namespace:    claim.Namespace,
		Name:         claim.Name,
		Capacity:     capacity,
		StorageClass: storageClassOf(claim.Spec.StorageClassName),
		AccessModes:  kubeAccessModes(claim.Spec.AccessModes),
		VolumeName:   claim.Spec.VolumeName,
		Phase:        string(claim.Status.Phase),
		CreatedAt:    unixTime(claim.CreationTimestamp),
	}
}

// storageClassOf 解开 StorageClass 指针，未设置时返回空串。
func storageClassOf(storageClass *string) string {
	if storageClass == nil {
		return ""
	}
	return *storageClass
}

// storageCapacity 取 storage 资源容量，未设置时返回空串。
func storageCapacity(capacity corev1.ResourceList) string {
	quantity, ok := capacity[corev1.ResourceStorage]
	if !ok {
		return ""
	}
	return quantity.String()
}

func kubeAccessModes(modes []corev1.PersistentVolumeAccessMode) []string {
	out := make([]string, 0, len(modes))
	for _, mode := range modes {
		out = append(out, string(mode))
	}
	return out
}

func sortKubeServices(items []bo.KubeService) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})
}

func sortKubeConfigMaps(items []bo.KubeConfigMap) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})
}

func sortKubeSecrets(items []bo.KubeSecret) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})
}

func sortKubePersistentVolumes(items []bo.KubePersistentVolume) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
}

func sortKubePersistentVolumeClaims(items []bo.KubePersistentVolumeClaim) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})
}
