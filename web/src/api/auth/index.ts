import { post, send } from "@/api/http"

import type {
  BindOAuthEmailParams,
  EmailLoginParams,
  ListOAuth2ProvidersParams,
  ListOAuth2ProvidersResult,
  LoginResult,
  OAuth2LoginParams,
  OAuth2LoginResult,
  OAuthBindEmailCodeParams,
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

/** POST /v1/auth/oauth2/login。第三方没有可用邮箱时返回待绑定凭证，不签发会话。 */
export function oauth2Login(
  body: OAuth2LoginParams
): Promise<OAuth2LoginResult> {
  return post<OAuth2LoginResult>("/auth/oauth2/login", body)
}

/** POST /v1/auth/oauth2/bind/code。给待绑定的第三方账号发送邮箱验证码。 */
export function sendOAuthBindEmailCode(
  body: OAuthBindEmailCodeParams
): Promise<Record<string, never>> {
  return post<Record<string, never>>("/auth/oauth2/bind/code", body)
}

/** POST /v1/auth/oauth2/bind。校验邮箱验证码，完成绑定并返回登录会话。 */
export function bindOAuthEmail(
  body: BindOAuthEmailParams
): Promise<LoginResult> {
  return post<LoginResult>("/auth/oauth2/bind", body)
}
