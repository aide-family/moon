import type { UserItem } from "@/api/user/types"

export type TokenPair = {
  accessToken: string
  refreshToken: string
  user: UserItem
}

export type LoginResult = {
  session: TokenPair
}

export type RegisterParams = {
  email: string
  password: string
  captchaId: string
  captchaCode: string
  name?: string
}

export type PasswordLoginParams = {
  email: string
  password: string
}

export type EmailLoginParams = {
  email: string
  code: string
}

export type OAuth2LoginParams = {
  provider: "GITHUB" | "GITEE" | "FEISHU"
  code: string
  redirectUri?: string
}
