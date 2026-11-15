export type RegistryDriver = "HARBOR" | "DOCKER_V2"

export interface RegistryItem {
  uid: string
  name: string
  driver: RegistryDriver
  endpoint: string
  username: string
  allowInsecureTls: boolean
  remark: string
  lastCheckedAt: number | string
  lastError: string
  createdAt: number | string
}

export interface RegistryProbe {
  endpoint: string
  driver: RegistryDriver
  repositoryCount: number
}

export interface SaveRegistryParams {
  name: string
  driver: RegistryDriver
  endpoint: string
  username: string
  password?: string
  allowInsecureTls: boolean
  remark: string
}

export interface UpdateRegistryParams extends SaveRegistryParams {
  uid: string
  clearPassword?: boolean
}

export interface ListRegistriesParams {
  page?: number
  pageSize?: number
  name?: string
  driver?: RegistryDriver | ""
}

export interface ListRegistriesResult {
  items: RegistryItem[]
  total: number
  page: number
  pageSize: number
}

export interface ProbeRegistryParams {
  driver: RegistryDriver
  endpoint: string
  username: string
  password?: string
  allowInsecureTls: boolean
}

export interface RegistryProjectItem {
  name: string
  repositoryCount: number
}

export interface RegistryRepositoryItem {
  name: string
  project: string
  tagCount: number
  updatedAt: number | string
}

export interface ListRegistryRepositoriesParams {
  project?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListRegistryRepositoriesResult {
  items: RegistryRepositoryItem[]
  total: number
  page: number
  pageSize: number
}

export interface RegistryTagItem {
  tag: string
  digest: string
  size: number | string
  platforms: string[]
  pushedAt: number | string
  referenceCount: number
}

export interface ListRegistryTagsParams {
  project?: string
  repository: string
  architecture?: string
  name?: string
  page?: number
  pageSize?: number
}

export interface ListRegistryTagsResult {
  items: RegistryTagItem[]
  total: number
  page: number
  pageSize: number
}

export interface RegistryPlatform {
  os: string
  architecture: string
  variant: string
  digest: string
  size: number | string
}

export interface RegistryTagReference {
  applicationUid: string
  applicationName: string
  versionUid: string
  version: string
  status: string
}

export interface RegistryTagDetail {
  tag: string
  digest: string
  size: number | string
  mediaType: string
  platforms: RegistryPlatform[]
  configDigest: string
  labels: string[]
  createdAt: number | string
  references: RegistryTagReference[]
}

export interface GetRegistryTagParams {
  project?: string
  repository: string
  tag: string
}

export interface CopyRegistryTagParams {
  project?: string
  repository: string
  tag: string
  digest?: string
  targetRepository: string
  targetTag: string
}

export interface DeleteRegistryTagParams {
  project?: string
  repository: string
  tag: string
  digest?: string
}

export type ImageSyncSourceKind = "CONNECTION" | "EXTERNAL"
export type ImageSyncStatus =
  | "PENDING"
  | "RUNNING"
  | "SUCCEEDED"
  | "FAILED"
  | "CANCELED"

export interface ImageSyncItem {
  uid: string
  sourceKind: ImageSyncSourceKind
  sourceRegistryUid: string
  sourceRegistryName: string
  sourceImage: string
  sourceDigest: string
  sourcePlatforms: string[]
  targetRegistryUid: string
  targetRegistryName: string
  targetRepository: string
  targetTag: string
  targetDigest: string
  mode: "ALL_PLATFORMS" | "SELECTED_PLATFORMS"
  overwrite: boolean
  status: ImageSyncStatus
  stage: string
  copiedBlobs: number
  totalBlobs: number
  message: string
  attempt: number
  operatorName: string
  createdAt: number | string
  startedAt: number | string
  finishedAt: number | string
}

export interface ListImageSyncsParams {
  page?: number
  pageSize?: number
  status?: ImageSyncStatus | ""
  sourceKind?: ImageSyncSourceKind | ""
  registryUid?: string
}

export interface ListImageSyncsResult {
  items: ImageSyncItem[]
  total: number
  page: number
  pageSize: number
}

export interface ImageSyncInput {
  sourceKind: ImageSyncSourceKind
  sourceRegistryUid?: string
  sourceImage?: string
  sourceUsername?: string
  sourcePassword?: string
  sourceAllowInsecureTls?: boolean
  targetRegistryUid: string
  targetRepository: string
  targetTag: string
  platforms?: string[]
  overwrite?: boolean
}

export interface PreviewImageSyncParams {
  source: ImageSyncInput
  targetRegistryUid: string
  targetRepository: string
  targetTag: string
}

export interface PreviewImageSyncResult {
  sourceImage: string
  sourceDigest: string
  sourcePlatforms: string[]
  targetExists: boolean
  targetDigest: string
}

export interface CreateImageSyncsResult {
  created: ImageSyncItem[]
  failed: { index: number; message: string }[]
}
