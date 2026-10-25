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
export type OAuth2Provider = "GITHUB" | "GITEE" | "FEISHU"

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
