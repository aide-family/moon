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
	// MetricsDatasourceUID 是绑定用来读取节点与 Pod 用量的指标数据源。
	// 空值表示未绑定；数据源已删除时 MetricsDatasourceName 为空。
	MetricsDatasourceUID  string
	MetricsDatasourceName string
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
	// UID 是节点对象标识，详情按它过滤事件。
	UID            string
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
	// AllocatableCPUMillis 和 AllocatableMemoryBytes 是节点可分配量，解析失败按 0 计。
	AllocatableCPUMillis   int64
	AllocatableMemoryBytes int64
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
	// UID 是 Pod 对象标识，详情按它过滤事件。
	UID        string
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

// KubePodShell 是一条 Pod 终端会话。Data 是终端输出，Closed 在会话结束时收到关闭原因；
// Write / Resize / Stop 由数据层实现，调用方不关心 SPDY 细节。
type KubePodShell struct {
	Container string
	Data      <-chan []byte
	Closed    <-chan string
	Write     func([]byte) error
	Resize    func(cols, rows int32) error
	Stop      func()
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
	// Revision 取自 deployment.kubernetes.io/revision 注解；没有注解时为 0。
	Revision int32
	// Current 表示该 ReplicaSet 是 Deployment 的当前版本。
	Current bool
}

// KubeWorkloadDetail 是工作负载详情：基础信息、名下的 ReplicaSet（仅 Deployment）和 Pod。
// Conditions 和 ReplicaStatus 来自工作负载的 status，用于详情抽屉的副本与条件区块。
type KubeWorkloadDetail struct {
	Workload      KubeWorkload
	ReplicaSets   []KubeReplicaSet
	Pods          []KubePod
	Conditions    []KubeCondition
	ReplicaStatus KubeReplicaStatus
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

// KubeWorkloadImage 是工作负载中一个容器的当前镜像与历史版本。
// Versions 只包含历史对象里的镜像，不含当前镜像，按版本从新到旧、按镜像去重。
type KubeWorkloadImage struct {
	Container  string
	Init       bool
	Image      string
	Repository string
	Tag        string
	Versions   []KubeImageVersion
}

// KubeImageVersion 是某个容器在一个历史版本里的镜像。
type KubeImageVersion struct {
	Image     string
	Revision  int32
	CreatedAt int64
}

// KubeWorkloadRevision 是工作负载的一个历史版本。
// ControllerRevision 不记录副本数，StatefulSet / DaemonSet 的 Replicas 取当前期望副本数。
type KubeWorkloadRevision struct {
	Revision  int32
	Name      string
	Current   bool
	Replicas  int32
	Images    []string
	CreatedAt int64
}

// KubeJob 是集群中的一个 Job。Status 取 Suspended、Failed、Complete、Running、Unknown
// 或 Pending；OwnerKind 和 OwnerName 取直接所属控制器，通常是 CronJob。
type KubeJob struct {
	Namespace      string
	Name           string
	Status         string
	Completions    int32
	Parallelism    int32
	Active         int32
	Succeeded      int32
	Failed         int32
	OwnerKind      string
	OwnerName      string
	StartTime      int64
	CompletionTime int64
	CreatedAt      int64
	Images         []string
}

// KubeJobDetail 是 Job 详情：基础信息和它的 Pod。
type KubeJobDetail struct {
	Job  KubeJob
	Pods []KubePod
}

// KubeCronJob 是集群中的一个 CronJob。LastScheduleTime 未调度过时为 0，
// ActiveCount 是当前活跃 Job 的数量。
type KubeCronJob struct {
	Namespace         string
	Name              string
	Schedule          string
	TimeZone          string
	Suspend           bool
	ConcurrencyPolicy string
	LastScheduleTime  int64
	ActiveCount       int32
	Images            []string
	CreatedAt         int64
}

// KubeCronJobDetail 是 CronJob 详情：基础信息和它最近创建的 Job。
type KubeCronJobDetail struct {
	CronJob KubeCronJob
	Jobs    []KubeJob
}

// KubeCondition 是资源 status.conditions 里的一条条件。
type KubeCondition struct {
	Type             string
	Status           string
	Reason           string
	Message          string
	LastTransitionAt int64
}

// KubeTaint 是节点上的一个污点。
type KubeTaint struct {
	Key    string
	Value  string
	Effect string
}

// KubeEvent 是一条与资源相关的 Kubernetes 事件。
type KubeEvent struct {
	Type        string
	Reason      string
	Message     string
	Count       int32
	Source      string
	FirstSeenAt int64
	LastSeenAt  int64
}

// KubeContainerStatus 是 Pod 中一个容器的运行状态。
// State 取 RUNNING、WAITING 或 TERMINATED；StateReason 和 StateMessage 取自对应状态。
type KubeContainerStatus struct {
	Name         string
	Image        string
	Ready        bool
	Restarts     int32
	State        string
	StateReason  string
	StateMessage string
	StartedAt    int64
	Ports        []string
}

// KubeVolume 是 Pod 的一个存储卷。Secret 与 ConfigMap 只带名称，不含任何值。
type KubeVolume struct {
	Name     string
	Type     string
	ReadOnly bool
	Summary  string
}

// KubeOwnerReference 是资源的一个 ownerReference。
type KubeOwnerReference struct {
	Kind       string
	Name       string
	UID        string
	Controller bool
}

// KubeNodeDetail 是节点详情：基础信息、条件、污点与容量。Events 由调用方按需补。
type KubeNodeDetail struct {
	Node        KubeNode
	Conditions  []KubeCondition
	Taints      []KubeTaint
	Capacity    map[string]string
	Allocatable map[string]string
	Events      []KubeEvent
	Warnings    []string
}

// KubePodDetail 是 Pod 详情：基础信息、容器状态、条件、卷与归属。Events 由调用方按需补。
type KubePodDetail struct {
	Pod             KubePod
	Containers      []KubeContainerStatus
	InitContainers  []KubeContainerStatus
	Conditions      []KubeCondition
	Volumes         []KubeVolume
	OwnerReferences []KubeOwnerReference
	Events          []KubeEvent
	Warnings        []string
}

// KubeReplicaStatus 是工作负载的副本状态汇总。
type KubeReplicaStatus struct {
	Desired     int32
	Current     int32
	Ready       int32
	Updated     int32
	Available   int32
	Unavailable int32
}

// KubePodPhaseCount 是概况里一个 Pod 阶段的数量。
type KubePodPhaseCount struct {
	Phase string
	Count int32
}

// KubeWorkloadSummary 是概况里一类工作负载的统计。
type KubeWorkloadSummary struct {
	Kind    string
	Total   int32
	Desired int32
	Ready   int32
}

// KubeOverview 是集群概况的一次聚合结果。DeniedSections 和 FailedSections 记录
// 没有统计成功的区块：前者缺读权限，后者查询失败。
type KubeOverview struct {
	APIServerVersion       string
	NodeTotal              int32
	NodeReady              int32
	NodeNotReady           int32
	AllocatableCPUMillis   int64
	AllocatableMemoryBytes int64
	PodTotal               int32
	PodPhases              []KubePodPhaseCount
	Workloads              []KubeWorkloadSummary
	JobTotal               int32
	JobActive              int32
	JobFailed              int32
	ServiceTotal           int32
	IngressTotal           int32
	DeniedSections         []string
	FailedSections         []string
	GeneratedAt            int64
}

// ClusterMetricsSummary 是集群绑定的指标数据源汇总。Status 取 UNBOUND、OK、
// PARTIAL 或 FAILED；未取到的指标为 nil，不能用 0 冒充。
type ClusterMetricsSummary struct {
	Status                 string
	NodeCPUUsagePercent    *float64
	NodeMemoryUsagePercent *float64
	PodCPUUsageCores       *float64
	PodMemoryUsageBytes    *float64
	UnavailableMetrics     []string
	Message                string
}
