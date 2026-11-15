export type ApplicationCategory = "DATABASE" | "MIDDLEWARE" | "OBSERVABILITY" | "TOOLING" | "OTHER"
export type ApplicationVersionStatus = "DRAFT" | "PUBLISHED" | "DEPRECATED"

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
}

export interface ApplicationResourceInput {
  item: string
  manifest: string
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
}

export interface ApplicationVersionItem {
  uid: string
  applicationUid: string
  version: string
  status: ApplicationVersionStatus
  changelog: string
  resources: ApplicationResource[]
  images: ApplicationVersionImage[]
  publishedAt: number | string
  createdAt: number | string
  validation: string
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

export interface SaveApplicationVersionParams {
  applicationUid: string
  version?: string
  changelog: string
  resources: ApplicationResourceInput[]
}

export interface ValidateApplicationVersionResult {
  version?: ApplicationVersionItem
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
  version: string
  message: string
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

export interface ApplicationOperationItem {
  uid: string
  deploymentUid: string
  releaseUid: string
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
