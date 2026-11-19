import { del, get, post, put } from "@/api/http"

import type {
  ApplicationDeploymentItem,
  ApplicationItem,
  ApplicationOperationItem,
  ApplicationReleaseItem,
  ApplicationResource,
  ApplicationVersionItem,
  ApplicationDeployValues,
  CreateApplicationVersionParams,
  CreateApplicationVersionResourceParams,
  CreateDeploymentParams,
  DeleteApplicationVersionResourceParams,
  DuplicateApplicationVersionParams,
  ListApplicationReleasesParams,
  ListApplicationReleasesResult,
  ListApplicationsParams,
  ListApplicationsResult,
  ListApplicationVersionsParams,
  ListApplicationVersionsResult,
  ListDeploymentsParams,
  ListDeploymentsResult,
  ListOperationsResult,
  SaveApplicationParams,
  SaveApplicationVersionResourceParams,
  UpdateApplicationParams,
  UpdateApplicationVersionParams,
  ValidateApplicationVersionResourceParams,
  ValidateApplicationVersionResourceResult,
  ValidateApplicationVersionResult,
} from "./types"

function segment(value: string): string {
  return encodeURIComponent(value)
}

/** GET /v1/applications */
export function listApplications(params: ListApplicationsParams = {}): Promise<ListApplicationsResult> {
  return get<ListApplicationsResult>("/applications", params)
}
/** POST /v1/application */
export function createApplication(body: SaveApplicationParams): Promise<{ application: ApplicationItem }> {
  return post<{ application: ApplicationItem }>("/application", body)
}
/** PUT /v1/application/{uid} */
export function updateApplication(body: UpdateApplicationParams): Promise<{ application: ApplicationItem }> {
  return put<{ application: ApplicationItem }>(`/application/${segment(body.uid)}`, body)
}
/** GET /v1/application/{uid} */
export function getApplication(uid: string): Promise<{ application: ApplicationItem }> {
  return get<{ application: ApplicationItem }>(`/application/${segment(uid)}`)
}
/** DELETE /v1/application/{uid} */
export function deleteApplication(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/application/${segment(uid)}`)
}

/** GET /v1/application/{uid}/versions */
export function listApplicationVersions(uid: string, params: ListApplicationVersionsParams = {}): Promise<ListApplicationVersionsResult> {
  return get<ListApplicationVersionsResult>(`/application/${segment(uid)}/versions`, params)
}
/** GET /v1/application/{uid}/version/{versionUid} */
export function getApplicationVersion(uid: string, versionUid: string): Promise<{ version: ApplicationVersionItem }> {
  return get<{ version: ApplicationVersionItem }>(`/application/${segment(uid)}/version/${segment(versionUid)}`)
}
/** POST /v1/application/{uid}/version */
export function createApplicationVersion(body: CreateApplicationVersionParams): Promise<{ version: ApplicationVersionItem }> {
  return post<{ version: ApplicationVersionItem }>(`/application/${segment(body.applicationUid)}/version`, body)
}
/** PUT /v1/application/{uid}/version/{versionUid} */
export function updateApplicationVersion(body: UpdateApplicationVersionParams): Promise<{ version: ApplicationVersionItem }> {
  return put<{ version: ApplicationVersionItem }>(`/application/${segment(body.applicationUid)}/version/${segment(body.uid)}`, body)
}
/** DELETE /v1/application/{uid}/version/{versionUid} */
export function deleteApplicationVersion(uid: string, versionUid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/application/${segment(uid)}/version/${segment(versionUid)}`)
}
/** POST /v1/application/{uid}/version/{versionUid}/validate */
export function validateApplicationVersion(uid: string, versionUid: string): Promise<ValidateApplicationVersionResult> {
  return post<ValidateApplicationVersionResult>(`/application/${segment(uid)}/version/${segment(versionUid)}/validate`)
}
/** POST /v1/application/{uid}/version/{versionUid}/publish */
export function publishApplicationVersion(uid: string, versionUid: string): Promise<{ version: ApplicationVersionItem }> {
  return post<{ version: ApplicationVersionItem }>(`/application/${segment(uid)}/version/${segment(versionUid)}/publish`)
}
/** POST /v1/application/{uid}/version/{versionUid}/deprecate */
export function deprecateApplicationVersion(uid: string, versionUid: string): Promise<{ version: ApplicationVersionItem }> {
  return post<{ version: ApplicationVersionItem }>(`/application/${segment(uid)}/version/${segment(versionUid)}/deprecate`)
}
/** POST /v1/application/{uid}/version/{versionUid}/copy */
export function duplicateApplicationVersion(body: DuplicateApplicationVersionParams): Promise<{ version: ApplicationVersionItem }> {
  return post<{ version: ApplicationVersionItem }>(`/application/${segment(body.applicationUid)}/version/${segment(body.uid)}/copy`, body)
}

/** POST /v1/application/{uid}/version/{versionUid}/resource */
export function createApplicationVersionResource(body: CreateApplicationVersionResourceParams): Promise<{ version: ApplicationVersionItem }> {
  return post<{ version: ApplicationVersionItem }>(`/application/${segment(body.applicationUid)}/version/${segment(body.uid)}/resource`, body)
}
/** GET /v1/application/{uid}/version/{versionUid}/resource/{resourceUid} */
export function getApplicationVersionResource(uid: string, versionUid: string, resourceUid: string): Promise<{ resource: ApplicationResource }> {
  return get<{ resource: ApplicationResource }>(`/application/${segment(uid)}/version/${segment(versionUid)}/resource/${segment(resourceUid)}`)
}
/** PUT /v1/application/{uid}/version/{versionUid}/resource/{resourceUid} */
export function saveApplicationVersionResource(body: SaveApplicationVersionResourceParams): Promise<{ version: ApplicationVersionItem }> {
  return put<{ version: ApplicationVersionItem }>(
    `/application/${segment(body.applicationUid)}/version/${segment(body.uid)}/resource/${segment(body.resourceUid)}`,
    body,
  )
}
/** DELETE /v1/application/{uid}/version/{versionUid}/resource/{resourceUid} */
export function deleteApplicationVersionResource(body: DeleteApplicationVersionResourceParams): Promise<{ version: ApplicationVersionItem }> {
  // DELETE 没有请求体：expectedRevision 按查询参数绑定（与 proto 字段名一致）。
  return del<{ version: ApplicationVersionItem }>(
    `/application/${segment(body.applicationUid)}/version/${segment(body.uid)}/resource/${segment(body.resourceUid)}?expectedRevision=${body.expectedRevision}`,
  )
}
/** POST /v1/application/{uid}/version/{versionUid}/resource/validate */
export function validateApplicationVersionResource(body: ValidateApplicationVersionResourceParams): Promise<ValidateApplicationVersionResourceResult> {
  return post<ValidateApplicationVersionResourceResult>(
    `/application/${segment(body.applicationUid)}/version/${segment(body.uid)}/resource/validate`,
    body,
  )
}

/** GET /v1/application-deployments */
export function listApplicationDeployments(params: ListDeploymentsParams = {}): Promise<ListDeploymentsResult> {
  return get<ListDeploymentsResult>("/application-deployments", params)
}
/** POST /v1/application-deployments */
export function createApplicationDeployment(body: CreateDeploymentParams): Promise<{ deployment: ApplicationDeploymentItem }> {
  return post<{ deployment: ApplicationDeploymentItem }>("/application-deployments", body)
}
/** GET /v1/application-deployments/{uid} */
export function getApplicationDeployment(uid: string): Promise<{ deployment: ApplicationDeploymentItem }> {
  return get<{ deployment: ApplicationDeploymentItem }>(`/application-deployments/${segment(uid)}`)
}
export function upgradeApplicationDeployment(uid: string, versionUid: string, clusterUids: string[] = [], values?: ApplicationDeployValues): Promise<{ deployment: ApplicationDeploymentItem }> {
  return post<{ deployment: ApplicationDeploymentItem }>(`/application-deployments/${segment(uid)}/upgrade`, { versionUid, clusterUids, values })
}
export function rollbackApplicationDeployment(uid: string, targetVersionUid: string, clusterUids: string[] = []): Promise<{ deployment: ApplicationDeploymentItem }> {
  return post<{ deployment: ApplicationDeploymentItem }>(`/application-deployments/${segment(uid)}/rollback`, { targetVersionUid, clusterUids })
}
export function redeployApplicationDeployment(uid: string, clusterUids: string[] = []): Promise<{ deployment: ApplicationDeploymentItem }> {
  return post<{ deployment: ApplicationDeploymentItem }>(`/application-deployments/${segment(uid)}/redeploy`, { clusterUids })
}
export function uninstallApplicationDeployment(uid: string, clusterUids: string[] = [], keepPersistentVolumeClaims = true, deleteClusterScopedResources = false): Promise<{ deployment: ApplicationDeploymentItem }> {
  return post<{ deployment: ApplicationDeploymentItem }>(`/application-deployments/${segment(uid)}/uninstall`, { clusterUids, keepPersistentVolumeClaims, deleteClusterScopedResources })
}
export function listApplicationDeploymentOperations(uid: string, params: { page?: number; pageSize?: number } = {}): Promise<ListOperationsResult> {
  return get<ListOperationsResult>(`/application-deployments/${segment(uid)}/operations`, params)
}
/** GET /v1/application-releases */
export function listApplicationReleases(params: ListApplicationReleasesParams = {}): Promise<ListApplicationReleasesResult> {
  return get<ListApplicationReleasesResult>("/application-releases", params)
}
export function getApplicationRelease(uid: string): Promise<{ release: ApplicationReleaseItem }> {
  return get<{ release: ApplicationReleaseItem }>(`/application-releases/${segment(uid)}`)
}
export function listApplicationReleaseOperations(uid: string, params: { page?: number; pageSize?: number } = {}): Promise<ListOperationsResult> {
  return get<ListOperationsResult>(`/application-releases/${segment(uid)}/operations`, params)
}
export function getApplicationOperation(uid: string): Promise<{ operation: ApplicationOperationItem }> {
  return get<{ operation: ApplicationOperationItem }>(`/application-release-operations/${segment(uid)}`)
}

export type * from "./types"
