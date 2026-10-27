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

export type KubernetesWorkloadKind = "Deployment" | "StatefulSet" | "DaemonSet"

export interface KubernetesWorkloadItem {
  namespace: string
  /** Deployment、StatefulSet 或 DaemonSet。 */
  kind: string
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

/** 读取单个资源 YAML 的参数。kind 取 proto 允许的 11 种资源类型。 */
export interface GetKubernetesResourceYamlParams {
  kind: string
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
