import type { UserItem } from "@/api/user/types"

export type TokenPair = {
  accessToken: string
  refreshToken: string
  user: UserItem
}

export type LoginResult = {
  session: TokenPair
}

/** 已配置客户端的第三方登录来源，取值与 OAuth2LoginRequest 一致。 */
export type OAuth2Provider = "GITHUB" | "GITEE" | "FEISHU" | "DINGTALK"

export type SendRegisterEmailCodeParams = {
  email: string
  captchaId: string
  captchaCode: string
}

export type RegisterParams = {
  email: string
  code: string
  name?: string
  captchaId: string
  captchaCode: string
}

export type SendEmailLoginCodeParams = {
  email: string
  captchaId: string
  captchaCode: string
}

export type EmailLoginParams = {
  email: string
  code: string
  captchaId: string
  captchaCode: string
}

export type ListOAuth2ProvidersParams = {
  redirectUri: string
  state: string
}

export type OAuth2ProviderItem = {
  provider: OAuth2Provider
  name: string
  authorizeUrl: string
}

export type ListOAuth2ProvidersResult = {
  providers: OAuth2ProviderItem[]
}

export type OAuth2LoginParams = {
  provider: OAuth2Provider
  code: string
  redirectUri?: string
}

/** 第三方登录结果：有可用邮箱或已绑定时返回会话，否则返回待绑定凭证。 */
export type OAuth2LoginResult = {
  /** 登录会话；需要补充邮箱绑定（emailBindingRequired）时缺省。 */
  session?: TokenPair
  /** 第三方没有可用邮箱，需要用户补充并验证邮箱。 */
  emailBindingRequired?: boolean
  /** 待绑定凭证，只用于发送绑定验证码和完成绑定，过期后需要重新发起第三方登录。 */
  bindToken?: string
  /** 第三方返回的未验证邮箱，用于预填绑定表单；平台没有返回时为空。 */
  suggestedEmail?: string
}

export type OAuthBindEmailCodeParams = {
  bindToken: string
  email: string
  captchaId: string
  captchaCode: string
}

export type BindOAuthEmailParams = {
  bindToken: string
  email: string
  code: string
  captchaId: string
  captchaCode: string
}
