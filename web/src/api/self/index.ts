import { del, get, post, put, type RequestOptions } from "@/api/http"

import type {
  BindSelfOAuthParams,
  SelfInfo,
  SelfNamespaces,
  SelfOAuthBindingResult,
  SelfOAuthLinks,
  SelfPermissions,
} from "./types"

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

/**
 * GET /v1/self/namespaces，可要求包含已停用的租户。
 * includeDisabled 只用于切换失败时解释原因；切换器本身不展示停用租户。
 */
export function listSelfNamespacesIncludingDisabled(
  options?: RequestOptions
): Promise<SelfNamespaces> {
  return get<SelfNamespaces>(
    "/self/namespaces",
    { includeDisabled: true },
    undefined,
    options
  )
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

/** POST /v1/self/change-email/code */
export function sendSelfChangeEmailCode(params: {
  email: string
  captchaId: string
  captchaCode: string
}): Promise<void> {
  return post<void>("/self/change-email/code", params)
}

/** PUT /v1/self/change-email */
export function changeEmail(params: {
  email: string
  code: string
  captchaId: string
  captchaCode: string
}): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-email", params)
}

/** PUT /v1/self/change-avatar */
export function changeAvatar(avatar: string): Promise<SelfInfo> {
  return put<SelfInfo>("/self/change-avatar", { avatar })
}

/**
 * GET /v1/self/oauth/links
 * 后端未部署时返回 404，页面据此整块隐藏关联卡片，因此固定 silent，避免弹全局错误 toast。
 */
export function listSelfOAuthLinks(
  options?: RequestOptions
): Promise<SelfOAuthLinks> {
  return get<SelfOAuthLinks>("/self/oauth/links", undefined, undefined, {
    ...options,
    silent: true,
  })
}

/** POST /v1/self/oauth/bind。用授权码把第三方身份绑到当前账号，不签发会话。 */
export function bindSelfOAuth(
  body: BindSelfOAuthParams
): Promise<SelfOAuthBindingResult> {
  return post<SelfOAuthBindingResult>("/self/oauth/bind", body)
}

/** DELETE /v1/self/oauth/links/{uid}。404 表示已经解绑，调用方按成功处理。 */
export function unbindSelfOAuth(
  uid: string,
  options?: RequestOptions
): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/self/oauth/links/${uid}`, options)
}
