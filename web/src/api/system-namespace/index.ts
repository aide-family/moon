import { get, put } from "@/api/http"

import type {
  ListSystemNamespacesParams,
  ListSystemNamespacesResult,
  UpdateNamespaceStatusPayload,
} from "./types"

/** GET /v1/system/namespaces。需要系统权限 system:namespace:read。 */
export function listSystemNamespaces(
  params: ListSystemNamespacesParams = {}
): Promise<ListSystemNamespacesResult> {
  return get<ListSystemNamespacesResult>("/system/namespaces", params)
}

/** PUT /v1/system/namespace/{uid}/status。停用必须带原因（1-500 字）。 */
export function updateSystemNamespaceStatus(
  uid: string,
  body: UpdateNamespaceStatusPayload
): Promise<unknown> {
  return put<unknown>(`/system/namespace/${uid}/status`, body)
}

/** PUT /v1/system/namespace/{uid}/archive。软删除租户及全部数据，不可恢复。 */
export function archiveSystemNamespace(uid: string): Promise<unknown> {
  return put<unknown>(`/system/namespace/${uid}/archive`)
}
