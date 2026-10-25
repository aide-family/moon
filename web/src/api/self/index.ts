import { get, put, type RequestOptions } from "@/api/http"

import type { SelfInfo, SelfNamespaces, SelfPermissions } from "./types"

/** GET /v1/self/info */
export function getSelfInfo(options?: RequestOptions): Promise<SelfInfo> {
  return get<SelfInfo>("/self/info", undefined, undefined, options)
}

/** GET /v1/self/namespaces */
export function listSelfNamespaces(
  options?: RequestOptions
): Promise<SelfNamespaces> {
  return get<SelfNamespaces>("/self/namespaces", undefined, undefined, options)
}

/** GET /v1/self/permissions */
export function getSelfPermissions(
  namespace?: string,
  options?: RequestOptions
): Promise<SelfPermissions> {
  return get<SelfPermissions>(
    "/self/permissions",
    undefined,
    namespace ? { headers: { "X-Namespace": namespace } } : undefined,
    options
  )
}

/** PUT /v1/self/change-remark */
export function changeRemark(remark: string): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-remark", { remark })
}

/** PUT /v1/self/change-phone */
export function changePhone(phone: string): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-phone", { phone })
}

/** PUT /v1/self/change-email */
export function changeEmail(email: string): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-email", { email })
}

/** PUT /v1/self/change-avatar */
export function changeAvatar(avatar: string): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-avatar", { avatar })
}
