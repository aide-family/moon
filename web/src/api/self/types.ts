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
  /**
   * 当前租户已开通的功能 code（GET /v1/self/permissions 的 features 字段）。
   * 旧版本后端不返回该字段时保持 undefined，前端按「功能未启用过滤」处理，
   * 避免升级过程中整个租户菜单突然消失。
   */
  features?: string[]
  /**
   * 系统管理员只读视角：调用者不是该租户成员，但仍可查看租户数据。
   * 为真时 permissions 只包含租户的 :read 权限；写接口仍由服务端 403 拒绝。
   */
  readOnly?: boolean
  /** 带 X-Namespace 时返回当前租户；readOnly 为真时一定存在。 */
  namespace?: NamespaceItem
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
