/** Kubernetes 集群。kubeconfig 原文不会出现在任何响应里。 */
export interface ClusterItem {
  uid: string
  name: string
  /** API Server 地址。 */
  server: string
  /** 探测到的 Kubernetes 版本，例如 v1.35.5。 */
  version: string
  /** kubeconfig 中当前 context 指向的集群名。 */
  clusterName: string
  /** kubeconfig 中使用的 context 名。 */
  contextName: string
  remark: string
  /**
   * 最近一次连通性检测时间，Unix 秒；从未检测为 0。
   * protojson 把 int64 序列化成字符串，展示时用 Number() 归一。
   */
  lastCheckedAt: number | string
  /** 最近一次检测的失败原因；为空表示成功。 */
  lastError: string
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

/** 连通性探测结果。 */
export interface ClusterProbe {
  server: string
  version: string
  clusterName: string
  contextName: string
  nodeCount: number
}

/** 接入或更新集群的表单参数。 */
export interface SaveClusterParams {
  name: string
  /** 接入必填；更新时留空表示保持原值。 */
  kubeconfig?: string
  remark: string
}

export interface UpdateClusterParams extends SaveClusterParams {
  uid: string
}

export interface ListClustersParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListClustersResult {
  items: ClusterItem[]
  total: number
  page: number
  pageSize: number
}

export interface TestClusterResult {
  probe: ClusterProbe
  cluster: ClusterItem
}

export interface ProbeKubeconfigResult {
  probe: ClusterProbe
}

export interface KubernetesNodeItem {
  name: string
  /** Ready、NotReady 或 Unknown。 */
  status: string
  /** false 表示已 cordon。 */
  schedulable: boolean
  /** 节点角色，例如 control-plane、worker。 */
  roles: string[]
  internalIp: string
  kubeletVersion: string
  osImage: string
  /** CPU 容量，例如 8。 */
  cpuCapacity: string
  /** 内存容量，例如 16384Mi。 */
  memoryCapacity: string
  podCount: number
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesNodesParams {
  name?: string
}

export interface ListKubernetesNodesResult {
  items: KubernetesNodeItem[]
  total: number
}

/** 工作负载列表只覆盖三种控制器。 */
export type KubernetesWorkloadKind = "Deployment" | "StatefulSet" | "DaemonSet"

/**
 * 详情抽屉与 YAML 预览支持的对象类型。
 * 在三种工作负载之外还包括 Pod、Job、CronJob 及其余只读资源；
 * 工作负载列表的 kind 过滤仍然只用 KubernetesWorkloadKind。
 */
export type KubernetesResourceKind =
  | KubernetesWorkloadKind
  | "Pod"
  | "Job"
  | "CronJob"
  | "Ingress"
  | "Service"
  | "ConfigMap"
  | "Secret"
  | "Node"
  | "PersistentVolume"
  | "PersistentVolumeClaim"

export interface KubernetesWorkloadItem {
  namespace: string
  /** Deployment、StatefulSet 或 DaemonSet。 */
  kind: KubernetesWorkloadKind
  name: string
  desiredReplicas: number
  readyReplicas: number
  updatedReplicas: number
  availableReplicas: number
  images: string[]
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesWorkloadsParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  /** 留空表示全部三种。 */
  kind?: string
  name?: string
}

export interface ListKubernetesWorkloadsResult {
  items: KubernetesWorkloadItem[]
  total: number
}

export interface ScaleKubernetesWorkloadParams {
  namespace: string
  kind: string
  name: string
  /** 目标副本数，0..10000。 */
  replicas: number
}

export interface RestartKubernetesWorkloadParams {
  namespace: string
  kind: string
  name: string
}

/** 某个容器镜像的历史版本，来自 ReplicaSet 或 ControllerRevision。 */
export interface KubernetesImageVersionItem {
  image: string
  /** 版本号（proto 为 int32）。 */
  revision: number
  /** 历史对象创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface KubernetesWorkloadImageItem {
  /** 容器名。 */
  container: string
  /** 是否为 init 容器；设置镜像时必须与之一致。 */
  init: boolean
  image: string
  repository: string
  tag: string
  /** 该容器的历史版本，新版本在前；没有历史对象时为空。 */
  versions: KubernetesImageVersionItem[]
}

export interface GetKubernetesWorkloadImagesParams {
  kind: KubernetesWorkloadKind
  namespace: string
  name: string
}

export interface GetKubernetesWorkloadImagesResult {
  images: KubernetesWorkloadImageItem[]
}

export interface SetKubernetesWorkloadImageParams {
  namespace: string
  kind: KubernetesWorkloadKind
  name: string
  container: string
  /** 目标容器是否为 init 容器。 */
  init: boolean
  image: string
}

/** 工作负载的一个历史版本。 */
export interface KubernetesWorkloadRevisionItem {
  /** 版本号（proto 为 int32）。 */
  revision: number
  /** 承载该版本的历史对象名（ReplicaSet 或 ControllerRevision）。 */
  name: string
  /** 是否为当前版本；当前版本不能回滚。 */
  current: boolean
  replicas: number
  images: string[]
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesWorkloadRevisionsParams {
  kind: KubernetesWorkloadKind
  namespace: string
  name: string
}

export interface ListKubernetesWorkloadRevisionsResult {
  /** 新版本在前。 */
  revisions: KubernetesWorkloadRevisionItem[]
}

export interface RollbackKubernetesWorkloadParams {
  namespace: string
  kind: KubernetesWorkloadKind
  name: string
  /** 目标版本，必须是该工作负载已存在的历史版本且不是当前版本。 */
  revision: number
}

/** Job 的执行状态；proto 字段是 string，取值由服务端判定。 */
export type KubernetesJobStatus =
  | "Suspended"
  | "Failed"
  | "Complete"
  | "Running"
  | "Unknown"
  | "Pending"

export interface KubernetesJobItem {
  namespace: string
  name: string
  /** 执行状态；服务端按 suspend / failed / succeeded / active 的顺序判定。 */
  status: KubernetesJobStatus
  /** 期望完成数；未设置时为 0。 */
  completions: number
  /** 并行度；未设置时为 0。 */
  parallelism: number
  active: number
  succeeded: number
  failed: number
  /** 直接控制器的类型，通常是 CronJob。 */
  ownerKind: string
  /** 直接控制器的名称。 */
  ownerName: string
  /** 开始时间，Unix 秒；未开始为 0。protojson 把 int64 序列化成字符串。 */
  startTime: number | string
  /** 完成时间，Unix 秒；未完成或未成功为 0。 */
  completionTime: number | string
  /** 创建时间，Unix 秒。 */
  createdAt: number | string
  /** 按容器顺序的镜像。 */
  images: string[]
}

export interface ListKubernetesJobsParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesJobsResult {
  items: KubernetesJobItem[]
  total: number
  page: number
  pageSize: number
}

export interface GetKubernetesJobDetailParams {
  namespace: string
  name: string
}

export interface GetKubernetesJobDetailResult {
  job: KubernetesJobItem
  /** 该 Job 创建的 Pod，按名称升序。 */
  pods: KubernetesPodItem[]
}

export interface KubernetesCronJobItem {
  namespace: string
  name: string
  /** Cron 调度表达式。 */
  schedule: string
  /** 调度时区；未设置时为空。 */
  timeZone: string
  /** 是否已暂停。 */
  suspend: boolean
  /** 并发策略，例如 Allow、Forbid、Replace。 */
  concurrencyPolicy: string
  /** 最近一次调度时间，Unix 秒；从未调度为 0。protojson 把 int64 序列化成字符串。 */
  lastScheduleTime: number | string
  /** 当前活跃的 Job 数。 */
  activeCount: number
  images: string[]
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesCronJobsParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesCronJobsResult {
  items: KubernetesCronJobItem[]
  total: number
  page: number
  pageSize: number
}

export interface GetKubernetesCronJobDetailParams {
  namespace: string
  name: string
}

export interface GetKubernetesCronJobDetailResult {
  cronjob: KubernetesCronJobItem
  /** 最近的 Job，最多 10 条，按创建时间倒序。 */
  jobs: KubernetesJobItem[]
}

export interface SetKubernetesCronJobSuspendParams {
  namespace: string
  name: string
  suspend: boolean
}

export interface TriggerKubernetesCronJobParams {
  namespace: string
  name: string
}

export interface KubernetesPodContainerItem {
  name: string
  ready: boolean
  restarts: number
  image: string
}

export interface KubernetesPodItem {
  namespace: string
  name: string
  /** 阶段，例如 Running、Pending、Succeeded、Failed。 */
  phase: string
  /** 就绪摘要，例如 1/2。 */
  ready: string
  /** 所有容器的重启次数合计。 */
  restarts: number
  nodeName: string
  podIp: string
  /** 直接所属控制器的类型，例如 Deployment、Job。 */
  ownerKind: string
  /** 直接所属控制器的名称。 */
  ownerName: string
  containers: KubernetesPodContainerItem[]
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesPodsParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  name?: string
  ownerKind?: string
  ownerName?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesPodsResult {
  items: KubernetesPodItem[]
  total: number
  page: number
  pageSize: number
}

export interface KubernetesPodLogsParams {
  namespace: string
  name: string
  /** 留空表示第一个容器。 */
  container?: string
  /** 0 表示服务端默认 200 行，最大 2000。 */
  tailLines?: number
  previous?: boolean
}

export interface KubernetesPodLogsResult {
  content: string
  /** 实际读取的容器名。 */
  container: string
}

export interface KubernetesIngressPathItem {
  /** 为空表示不限制主机。 */
  host: string
  path: string
  pathType: string
  serviceName: string
  servicePort: string
}

export interface KubernetesIngressItem {
  namespace: string
  name: string
  className: string
  hosts: string[]
  paths: KubernetesIngressPathItem[]
  /** TLS Secret 名称。 */
  tls: string[]
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesIngressesParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesIngressesResult {
  items: KubernetesIngressItem[]
  total: number
  page: number
  pageSize: number
}

export interface ListKubernetesNamespacesResult {
  items: string[]
}

/** Service 的一个端口。 */
export interface KubernetesServicePortItem {
  /** 端口名。 */
  name: string
  /** 协议，例如 TCP、UDP。 */
  protocol: string
  /** Service 端口。 */
  port: number
  /** 容器端口，可能是名称或数字。 */
  targetPort: string
  /** NodePort；ClusterIP 类型时为 0。 */
  nodePort: number
}

export interface KubernetesServiceItem {
  namespace: string
  name: string
  /** 类型，例如 ClusterIP、NodePort、LoadBalancer、ExternalName。 */
  type: string
  clusterIp: string
  externalIps: string[]
  /** LoadBalancer 的入口 IP。 */
  loadBalancerIp: string
  ports: KubernetesServicePortItem[]
  /** 选择器，JSON 对象字符串。 */
  selector: string
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesServicesParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesServicesResult {
  items: KubernetesServiceItem[]
  total: number
  page: number
  pageSize: number
}

export interface KubernetesConfigMapItem {
  namespace: string
  name: string
  /** data 的键名，按字母序；值不返回。 */
  dataKeys: string[]
  /** binaryData 的键名，按字母序；值不返回。 */
  binaryDataKeys: string[]
  immutable: boolean
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesConfigMapsParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesConfigMapsResult {
  items: KubernetesConfigMapItem[]
  total: number
  page: number
  pageSize: number
}

export interface KubernetesSecretItem {
  namespace: string
  name: string
  /** 类型，例如 Opaque、kubernetes.io/tls。 */
  type: string
  /** data 的键名，按字母序；值永远不返回。 */
  dataKeys: string[]
  immutable: boolean
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesSecretsParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesSecretsResult {
  items: KubernetesSecretItem[]
  total: number
  page: number
  pageSize: number
}

/** PersistentVolume 是集群级资源，没有 namespace 字段。 */
export interface KubernetesPersistentVolumeItem {
  name: string
  /** 容量，例如 10Gi。 */
  capacity: string
  storageClass: string
  /** 访问模式，例如 ReadWriteOnce。 */
  accessModes: string[]
  /** 绑定的 PVC，格式为 namespace/name；未绑定时为空。 */
  claim: string
  /** 阶段，例如 Available、Bound、Released。 */
  phase: string
  /** 回收策略，例如 Retain、Delete。 */
  reclaimPolicy: string
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesPersistentVolumesParams {
  name?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesPersistentVolumesResult {
  items: KubernetesPersistentVolumeItem[]
  total: number
  page: number
  pageSize: number
}

export interface KubernetesPersistentVolumeClaimItem {
  namespace: string
  name: string
  /** 容量，例如 10Gi。 */
  capacity: string
  storageClass: string
  /** 访问模式，例如 ReadWriteOnce。 */
  accessModes: string[]
  /** 绑定的 PV 名称；未绑定时为空。 */
  volumeName: string
  /** 阶段，例如 Pending、Bound、Lost。 */
  phase: string
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface ListKubernetesPersistentVolumeClaimsParams {
  /** 留空表示全部命名空间。 */
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListKubernetesPersistentVolumeClaimsResult {
  items: KubernetesPersistentVolumeClaimItem[]
  total: number
  page: number
  pageSize: number
}

/** 读取单个资源 YAML 的参数。kind 取 proto 允许的资源类型，含 Job 与 CronJob。 */
export interface GetKubernetesResourceYamlParams {
  kind: KubernetesResourceKind
  /** Node 与 PersistentVolume 忽略该字段，其余资源必填。 */
  namespace?: string
  name: string
}

export interface GetKubernetesResourceYamlResult {
  /** YAML 文本；Secret / ConfigMap 的敏感值已被替换。 */
  yaml: string
  /** 是否隐藏了敏感值。 */
  redacted: boolean
}

// ---------------------------------------------------------------------------
// 资源详情：工作负载、Ingress、Service、ConfigMap、Secret、PVC 的关联资源。
// ---------------------------------------------------------------------------

/** 工作负载详情里的副本集。 */
export interface KubernetesReplicaSetItem {
  namespace: string
  name: string
  desiredReplicas: number
  readyReplicas: number
  currentReplicas: number
  images: string[]
  /** 创建时间，Unix 秒。protojson 把 int64 序列化成字符串。 */
  createdAt: number | string
}

export interface GetKubernetesWorkloadDetailParams {
  /** Deployment、StatefulSet 或 DaemonSet。 */
  kind: string
  namespace: string
  name: string
}

export interface GetKubernetesWorkloadDetailResult {
  workload: KubernetesWorkloadItem
  /** StatefulSet 和 DaemonSet 通常为空。 */
  replicasets: KubernetesReplicaSetItem[]
  pods: KubernetesPodItem[]
}

export interface GetKubernetesIngressDetailParams {
  namespace: string
  name: string
}

export interface GetKubernetesIngressDetailResult {
  ingress: KubernetesIngressItem
  /** 规则中引用的 Service，按名称去重。 */
  services: KubernetesServiceItem[]
}

/** Endpoint 端口。 */
export interface KubernetesEndpointPortItem {
  name: string
  port: number
  protocol: string
}

/** Service 的一个 Endpoint。 */
export interface KubernetesEndpointItem {
  address: string
  hostname: string
  nodeName: string
  /** 目标对象类型，例如 Pod。 */
  targetKind: string
  targetName: string
  ready: boolean
  ports: KubernetesEndpointPortItem[]
}

export interface GetKubernetesServiceDetailParams {
  namespace: string
  name: string
}

export interface GetKubernetesServiceDetailResult {
  service: KubernetesServiceItem
  endpoints: KubernetesEndpointItem[]
}

/** 引用 ConfigMap / Secret / PVC 的工作负载。 */
export interface KubernetesWorkloadRefItem {
  kind: string
  namespace: string
  name: string
}

export interface GetKubernetesConfigMapDetailParams {
  namespace: string
  name: string
}

export interface GetKubernetesConfigMapDetailResult {
  configmap: KubernetesConfigMapItem
  relatedWorkloads: KubernetesWorkloadRefItem[]
}

export interface GetKubernetesSecretDetailParams {
  namespace: string
  name: string
}

export interface GetKubernetesSecretDetailResult {
  secret: KubernetesSecretItem
  relatedWorkloads: KubernetesWorkloadRefItem[]
}

export interface GetKubernetesPersistentVolumeClaimDetailParams {
  namespace: string
  name: string
}

export interface GetKubernetesPersistentVolumeClaimDetailResult {
  claim: KubernetesPersistentVolumeClaimItem
  relatedWorkloads: KubernetesWorkloadRefItem[]
}
