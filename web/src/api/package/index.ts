import { del, get, post, put, type RequestOptions } from "@/api/http"

import type {
  ListFeaturesResult,
  ListPackagesParams,
  ListPackagesResult,
  NamespaceFeatures,
  PackagePayload,
  UpdateNamespaceFeaturesPayload,
  UpdatePackagePayload,
} from "./types"

/** GET /v1/system/features。需要系统权限 system:package:read。 */
export function listFeatures(
  options?: RequestOptions
): Promise<ListFeaturesResult> {
  return get<ListFeaturesResult>(
    "/system/features",
    undefined,
    undefined,
    options
  )
}

/** GET /v1/system/packages。需要系统权限 system:package:read。 */
export function listPackages(
  params: ListPackagesParams = {},
  options?: RequestOptions
): Promise<ListPackagesResult> {
  return get<ListPackagesResult>("/system/packages", params, undefined, options)
}

/** POST /v1/system/package。需要系统权限 system:package:write（仅超级管理员）。 */
export function createPackage(body: PackagePayload): Promise<unknown> {
  return post<unknown>("/system/package", body)
}

/** PUT /v1/system/package/{uid}。基础套餐不可改名，但可改功能与备注。 */
export function updatePackage(body: UpdatePackagePayload): Promise<unknown> {
  const { uid, ...payload } = body
  return put<unknown>(`/system/package/${uid}`, payload)
}

/** DELETE /v1/system/package/{uid}。基础套餐、被租户使用的套餐会被后端拒绝。 */
export function deletePackage(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/system/package/${uid}`)
}

/** GET /v1/system/namespace/{uid}/features。读取租户的套餐与有效功能。 */
export function getNamespaceFeatures(
  uid: string,
  options?: RequestOptions
): Promise<NamespaceFeatures> {
  return get<NamespaceFeatures>(
    `/system/namespace/${uid}/features`,
    undefined,
    undefined,
    options
  )
}

/** PUT /v1/system/namespace/{uid}/features。保存套餐与目标有效功能。 */
export function updateNamespaceFeatures(
  uid: string,
  body: UpdateNamespaceFeaturesPayload
): Promise<unknown> {
  return put<unknown>(`/system/namespace/${uid}/features`, body)
}
