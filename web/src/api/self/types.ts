import type { OAuth2Provider } from "@/api/auth/types"
import type { NamespaceItem } from "@/api/namespace/types"
import type { UserItem } from "@/api/user/types"

export type SelfInfo = {
  user: UserItem
  systemRole?: string
  systemPermissions?: string[]
}

export type SelfNamespaces = {
  items?: NamespaceItem[]
}

export type SelfPermissions = {
  permissions?: string[]
}

/** 当前账号已绑定的一条第三方身份。 */
export type SelfOAuthLink = {
  /** user_oauth.uid，解绑按它对位。 */
  uid: string
  provider: OAuth2Provider
  /** 只展示尾 4 位，由前端掩码。 */
  openId: string
  /** 绑定时从第三方拿到的邮箱，可能为空。 */
  email: string
  /** RFC3339，来自 user_oauth.created_at。 */
  boundAt?: string
}

export type SelfOAuthLinks = {
  items: SelfOAuthLink[]
}

export type BindSelfOAuthParams = {
  provider: OAuth2Provider
  code: string
  redirectUri: string
}

export type SelfOAuthBindingResult = {
  item: SelfOAuthLink
}
