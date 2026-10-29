import { get } from "@/api/http"

import type {
  ListSystemNamespacesParams,
  ListSystemNamespacesResult,
} from "./types"

/** GET /v1/system/namespaces。需要系统权限 system:namespace:read。 */
export function listSystemNamespaces(
  params: ListSystemNamespacesParams = {}
): Promise<ListSystemNamespacesResult> {
  return get<ListSystemNamespacesResult>("/system/namespaces", params)
}
