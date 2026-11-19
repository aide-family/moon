/**
 * 应用分类是自由字符串（proto 只限制长度 ≤32）。界面用内置分类做提示，
 * 自定义值原样展示；空值按 OTHER 处理。
 */
export type ApplicationCategory = string
export type ApplicationVersionStatus = "DRAFT" | "PUBLISHED" | "DEPRECATED"

/** 服务端按版本状态计算的可执行动作，前端不自行推断。 */
export type ApplicationVersionAction =
  | "PUBLISH"
  | "DEPRECATE"
  | "COPY"
  | "DELETE"
  | "RELEASE"

export interface ApplicationItem {
  uid: string
  name: string
  displayName: string
  category: ApplicationCategory
  description: string
  readme: string
  versionCount: number
  latestVersion: string
  createdAt: number | string
}

export interface SaveApplicationParams {
  name: string
  displayName: string
  category: ApplicationCategory
  description: string
  readme: string
}

export interface UpdateApplicationParams extends SaveApplicationParams {
  uid: string
}

export interface ListApplicationsParams {
  page?: number
  pageSize?: number
  name?: string
  category?: ApplicationCategory | ""
}
export interface ListApplicationsResult {
  items: ApplicationItem[]
  total: number
  page: number
  pageSize: number
}

export interface ApplicationResource {
  /** 资源在版本内的展示标识（历史字段，向后兼容）。 */
  item: string
  manifest: string
  apiVersion: string
  kind: string
  name: string
  namespace: string
  scope: "NAMESPACED" | "CLUSTER" | "UNKNOWN"
  clusterScoped: boolean
  images: string[]
  secretKeys: string[]
  /** 资源在版本内的稳定标识，编辑与删除使用。 */
  uid: string
  /** 资源所属的服务（组件）名。 */
  component: string
  /** 资源内容的 sha256，用于并发编辑提示。 */
  contentHash: string
}

export interface ApplicationResourceInput {
  item: string
  manifest: string
  /** 资源所属的服务（组件）名；为空时归入默认服务。 */
  component?: string
}

export interface ApplicationVersionImage {
  resourceItem: string
  container: string
  image: string
  registryUid: string
  registryName: string
  repository: string
  tag: string
  digest: string
  platforms: string[]
  external: boolean
  component: string
  resourceUid: string
  /** 容器在清单里的 JSON 路径。 */
  containerPath: string
  project: string
  /** REGISTRY / EXTERNAL。 */
  source: string
  /** 上架时固定的实际部署引用。 */
  resolvedImage: string
}

/** 版本内的一个服务，由若干 Kubernetes 资源组成。 */
export interface ApplicationVersionComponent {
  uid: string
  name: string
  displayName: string
  sortOrder: number
  resources: ApplicationResource[]
  resourceCount: number
  /** 服务备注，支持 Markdown。 */
  description: string
}

export interface ApplicationVersionItem {
  uid: string
  applicationUid: string
  version: string
  status: ApplicationVersionStatus
  changelog: string
  /** 扁平视图，保留给旧客户端；新界面使用 components。 */
  resources: ApplicationResource[]
  images: ApplicationVersionImage[]
  publishedAt: number | string
  createdAt: number | string
  validation: string
  components: ApplicationVersionComponent[]
  /** 版本内容的乐观锁版本号，每次内容变更 +1。 */
  revision: number
  /** 服务端计算的当前可执行动作。 */
  allowedActions: ApplicationVersionAction[]
  componentCount: number
  resourceCount: number
  statusChangedAt: number | string
  deprecatedAt: number | string
  copiedFromVersionUid: string
}

export interface ListApplicationVersionsParams {
  page?: number
  pageSize?: number
  status?: ApplicationVersionStatus | ""
}
export interface ListApplicationVersionsResult {
  items: ApplicationVersionItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateApplicationVersionParams {
  applicationUid: string
  version: string
  changelog: string
  resources: ApplicationResourceInput[]
}

export interface UpdateApplicationVersionParams {
  applicationUid: string
  uid: string
  changelog: string
  resources: ApplicationResourceInput[]
  /** 客户端读取到的版本内容版本号；不一致时返回冲突。 */
  revision: number
}

export interface DuplicateApplicationVersionParams {
  applicationUid: string
  uid: string
  version: string
  changelog: string
}

export interface CreateApplicationVersionResourceParams {
  applicationUid: string
  uid: string
  resource: ApplicationResourceInput
  expectedRevision: number
}

export interface SaveApplicationVersionResourceParams {
  applicationUid: string
  uid: string
  resourceUid: string
  component: string
  manifest: string
  secretValues?: Record<string, string>
  clearSecretKeys?: string[]
  expectedRevision: number
}

export interface DeleteApplicationVersionResourceParams {
  applicationUid: string
  uid: string
  resourceUid: string
  expectedRevision: number
}

export interface SaveApplicationVersionComponentParams {
  applicationUid: string
  uid: string
  /** 为空表示新增服务，否则更新对应服务。 */
  componentUid: string
  name: string
  /** 服务备注，支持 Markdown，最多 2048 字符。 */
  description: string
  expectedRevision: number
}

export interface DeleteApplicationVersionComponentParams {
  applicationUid: string
  uid: string
  componentUid: string
  expectedRevision: number
}

export interface ValidateApplicationVersionResourceParams {
  applicationUid: string
  uid: string
  component: string
  manifest: string
}

export interface ValidateApplicationVersionResult {
  version?: ApplicationVersionItem
  errors: string[]
  warnings: string[]
}

export interface ValidateApplicationVersionResourceResult {
  errors: string[]
  warnings: string[]
}

export type DeploymentStatus =
  | "PENDING"
  | "DEPLOYING"
  | "RUNNING"
  | "PARTIAL"
  | "FAILED"
  | "UNINSTALLING"
  | "UNINSTALLED"

export interface DeploymentCluster {
  uid: string
  name: string
  status: string
  releaseUid: string
  /** 部署组期望版本。 */
  version: string
  message: string
  /** 集群当前实际运行的版本（最后一次成功的发布）。 */
  currentVersion: string
  currentVersionUid: string
  /** 集群当前版本与部署组期望版本一致。 */
  inSync: boolean
}

export interface ApplicationReleaseItem {
  uid: string
  deploymentUid: string
  clusterUid: string
  clusterName: string
  targetNamespace: string
  name: string
  status: string
  versionUid: string
  version: string
  message: string
  lastOperationUid: string
  deployedAt: number | string
  applicationUid: string
  applicationName: string
}

export interface ApplicationDeploymentItem {
  uid: string
  name: string
  applicationUid: string
  applicationName: string
  versionUid: string
  version: string
  targetNamespace: string
  clusters: DeploymentCluster[]
  status: DeploymentStatus
  message: string
  releases: ApplicationReleaseItem[]
  createdAt: number | string
  deployedAt: number | string
}

export interface ApplicationDeployValues {
  replicas: Record<string, number>
  imageTags: Record<string, string>
  ingressHosts: Record<string, string>
  secretValues: Record<string, string>
}

export interface ListDeploymentsParams {
  page?: number
  pageSize?: number
  name?: string
  status?: DeploymentStatus | ""
  applicationUid?: string
  clusterUid?: string
}
export interface ListDeploymentsResult {
  items: ApplicationDeploymentItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateDeploymentParams {
  applicationUid: string
  versionUid: string
  name: string
  targetNamespace: string
  clusterUids: string[]
  values?: ApplicationDeployValues
}

export interface ListApplicationReleasesParams {
  page?: number
  pageSize?: number
  applicationUid?: string
  clusterUid?: string
  status?: DeploymentStatus | ""
}
export interface ListApplicationReleasesResult {
  items: ApplicationReleaseItem[]
  total: number
  page: number
  pageSize: number
}

export interface ApplicationOperationItem {
  uid: string
  deploymentUid: string
  releaseUid: string
  clusterUid: string
  clusterName: string
  action: string
  status: string
  fromVersionUid: string
  toVersionUid: string
  message: string
  operatorName: string
  startedAt: number | string
  finishedAt: number | string
  createdAt: number | string
  resources: { apiVersion: string; kind: string; namespace: string; name: string }[]
}
export interface ListOperationsResult {
  items: ApplicationOperationItem[]
  total: number
  page: number
  pageSize: number
}
