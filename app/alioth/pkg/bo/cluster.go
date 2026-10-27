package bo

import "time"

// Cluster 是租户接入的一个 Kubernetes 集群。
// kubeconfig 原文只在写入时以独立参数传给仓储，不出现在这个对象里，
// 因此把它交给接口层也不会带出凭据。
type Cluster struct {
	UID           string
	NamespaceUID  string
	Name          string
	Server        string
	Version       string
	ClusterName   string
	ContextName   string
	Remark        string
	LastCheckedAt int64
	LastError     string
	CreatedAt     time.Time
}

// ClusterProbe 是一次 kubeconfig 连通性探测的结果。
// Version 为空表示没能读到 API Server 版本；NodeCount 为 0 可能是没有节点读取权限。
type ClusterProbe struct {
	Server      string
	Version     string
	ClusterName string
	ContextName string
	NodeCount   int32
}

// KubeNode 是集群中的一个节点。
type KubeNode struct {
	Name           string
	Status         string
	Schedulable    bool
	Roles          []string
	InternalIP     string
	KubeletVersion string
	OSImage        string
	CPUCapacity    string
	MemoryCapacity string
	PodCount       int32
	CreatedAt      int64
}

// KubeWorkload 是集群中的一个工作负载。Kind 取 Deployment、StatefulSet 或 DaemonSet。
// DaemonSet 的期望副本数取应调度的节点数。
type KubeWorkload struct {
	Namespace         string
	Kind              string
	Name              string
	DesiredReplicas   int32
	ReadyReplicas     int32
	UpdatedReplicas   int32
	AvailableReplicas int32
	Images            []string
	CreatedAt         int64
}

// KubePod 是集群中的一个 Pod。OwnerKind 和 OwnerName 取直接所属控制器，
// 例如 Deployment 的 Pod 直接属于 ReplicaSet。
type KubePod struct {
	Namespace  string
	Name       string
	Phase      string
	Ready      string
	Restarts   int32
	NodeName   string
	PodIP      string
	OwnerKind  string
	OwnerName  string
	Containers []KubeContainer
	CreatedAt  int64
}

// KubeContainer 是 Pod 中的一个容器。
type KubeContainer struct {
	Name     string
	Ready    bool
	Restarts int32
	Image    string
}

// KubePodLogs 是一次 Pod 日志读取的结果。
type KubePodLogs struct {
	Content   string
	Container string
}

// KubeIngress 是集群中的一个 Ingress。
type KubeIngress struct {
	Namespace string
	Name      string
	ClassName string
	Hosts     []string
	Paths     []KubeIngressPath
	TLS       []string
	CreatedAt int64
}

// KubeIngressPath 是 Ingress 规则中的一个路径。
type KubeIngressPath struct {
	Host        string
	Path        string
	PathType    string
	ServiceName string
	ServicePort string
}

// KubeService 是集群中的一个 Service。
// Selector 是 JSON 对象文本，空选择器序列化为 {}。
type KubeService struct {
	Namespace      string
	Name           string
	Type           string
	ClusterIP      string
	ExternalIPs    []string
	LoadBalancerIP string
	Ports          []KubeServicePort
	Selector       string
	CreatedAt      int64
}

// KubeServicePort 是 Service 的一个端口。TargetPort 可能是容器端口名或数字。
type KubeServicePort struct {
	Name       string
	Protocol   string
	Port       int32
	TargetPort string
	NodePort   int32
}

// KubeConfigMap 是集群中的一个 ConfigMap。这里只保留键名，值不进入业务对象。
type KubeConfigMap struct {
	Namespace      string
	Name           string
	DataKeys       []string
	BinaryDataKeys []string
	Immutable      bool
	CreatedAt      int64
}

// KubeSecret 是集群中的一个 Secret。这里只保留类型和键名，值不进入业务对象。
type KubeSecret struct {
	Namespace string
	Name      string
	Type      string
	DataKeys  []string
	Immutable bool
	CreatedAt int64
}

// KubePersistentVolume 是集群级资源，不属于任何命名空间。
// Claim 是绑定 PVC 的 namespace/name，未绑定时为空。
type KubePersistentVolume struct {
	Name          string
	Capacity      string
	StorageClass  string
	AccessModes   []string
	Claim         string
	Phase         string
	ReclaimPolicy string
	CreatedAt     int64
}

// KubePersistentVolumeClaim 是命名空间中的 PVC。
// Capacity 优先取 status.capacity.storage，未绑定时回退到 spec.resources.requests.storage。
type KubePersistentVolumeClaim struct {
	Namespace    string
	Name         string
	Capacity     string
	StorageClass string
	AccessModes  []string
	VolumeName   string
	Phase        string
	CreatedAt    int64
}

// KubeResourceYAML 是单个资源的 YAML 预览。
// Redacted 为真表示 Secret 或 ConfigMap 的值已被替换，响应里没有敏感数据。
type KubeResourceYAML struct {
	YAML     string
	Redacted bool
}

// KubeReplicaSet 是一个 Deployment 名下的 ReplicaSet。
type KubeReplicaSet struct {
	Namespace       string
	Name            string
	DesiredReplicas int32
	ReadyReplicas   int32
	CurrentReplicas int32
	Images          []string
	CreatedAt       int64
}

// KubeWorkloadDetail 是工作负载详情：基础信息、名下的 ReplicaSet（仅 Deployment）和 Pod。
type KubeWorkloadDetail struct {
	Workload    KubeWorkload
	ReplicaSets []KubeReplicaSet
	Pods        []KubePod
}

// KubeEndpointPort 是 Endpoints 子集里的一个端口。
type KubeEndpointPort struct {
	Name     string
	Port     int32
	Protocol string
}

// KubeEndpoint 是 Service 的一个后端地址。Ready 为假表示只在 notReadyAddresses 中。
type KubeEndpoint struct {
	Address    string
	Hostname   string
	NodeName   string
	TargetKind string
	TargetName string
	Ready      bool
	Ports      []KubeEndpointPort
}

// KubeServiceDetail 是 Service 详情：基础信息和 Endpoints。
type KubeServiceDetail struct {
	Service   KubeService
	Endpoints []KubeEndpoint
}

// KubeIngressDetail 是 Ingress 详情：基础信息和规则里引用的 Service。
type KubeIngressDetail struct {
	Ingress  KubeIngress
	Services []KubeService
}

// KubeWorkloadRef 表示一个引用或挂载了配置/存储资源的工作负载。
type KubeWorkloadRef struct {
	Kind      string
	Namespace string
	Name      string
}

// KubeConfigMapDetail 是 ConfigMap 详情和引用它的工作负载。
type KubeConfigMapDetail struct {
	ConfigMap        KubeConfigMap
	RelatedWorkloads []KubeWorkloadRef
}

// KubeSecretDetail 是 Secret 详情和引用它的工作负载。Secret 只带类型和键名。
type KubeSecretDetail struct {
	Secret           KubeSecret
	RelatedWorkloads []KubeWorkloadRef
}

// KubePersistentVolumeClaimDetail 是 PVC 详情和挂载它的工作负载。
type KubePersistentVolumeClaimDetail struct {
	Claim            KubePersistentVolumeClaim
	RelatedWorkloads []KubeWorkloadRef
}
