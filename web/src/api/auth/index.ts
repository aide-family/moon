import { post, send } from "@/api/http"

import type {
  EmailLoginParams,
  ListOAuth2ProvidersParams,
  ListOAuth2ProvidersResult,
  LoginResult,
  OAuth2LoginParams,
  RegisterParams,
  SendEmailLoginCodeParams,
  SendRegisterEmailCodeParams,
} from "./types"

/** POST /v1/auth/register/code */
export function sendRegisterEmailCode(
  body: SendRegisterEmailCodeParams
): Promise<Record<string, never>> {
  return post<Record<string, never>>("/auth/register/code", body)
}

/** POST /v1/auth/register */
export function register(body: RegisterParams): Promise<LoginResult> {
  return post<LoginResult>("/auth/register", body)
}

/** POST /v1/auth/email/login/code */
export function sendEmailLoginCode(
  body: SendEmailLoginCodeParams
): Promise<Record<string, never>> {
  return post<Record<string, never>>("/auth/email/login/code", body)
}

/** POST /v1/auth/email/login */
export function emailLogin(body: EmailLoginParams): Promise<LoginResult> {
  return post<LoginResult>("/auth/email/login", body)
}

/** POST /v1/auth/oauth2/providers。未配置/不可用时登录页不弹全局错误，只隐藏第三方入口。 */
export function listOAuth2Providers(
  body: ListOAuth2ProvidersParams
): Promise<ListOAuth2ProvidersResult> {
  return send<ListOAuth2ProvidersResult>(
    "/auth/oauth2/providers",
    { method: "POST", body: JSON.stringify(body) },
    { silent: true }
  )
}

/** POST /v1/auth/oauth2/login */
export function oauth2Login(body: OAuth2LoginParams): Promise<LoginResult> {
  return post<LoginResult>("/auth/oauth2/login", body)
}
