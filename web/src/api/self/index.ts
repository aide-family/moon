import { get, put } from "@/api/http"

import type { SelfInfo, SelfNamespaces, SelfPermissions } from "./types"

/** GET /v1/self/info */
export function getSelfInfo(): Promise<SelfInfo> {
  return get<SelfInfo>("/self/info")
}

/** GET /v1/self/namespaces */
export function listSelfNamespaces(): Promise<SelfNamespaces> {
  return get<SelfNamespaces>("/self/namespaces")
}

/** GET /v1/self/permissions */
export function getSelfPermissions(): Promise<SelfPermissions> {
  return get<SelfPermissions>("/self/permissions")
}

/** PUT /v1/self/change-remark */
export function changeRemark(remark: string): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-remark", { remark })
}

/** PUT /v1/self/change-email */
export function changeEmail(email: string): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-email", { email })
}

/** PUT /v1/self/change-avatar */
export function changeAvatar(avatar: string): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-avatar", { avatar })
}
