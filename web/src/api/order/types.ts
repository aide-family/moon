import type { NamespaceItem } from "@/api/namespace/types"

/** 套餐版本：PUBLIC 可被用户创建租户时选择，BETA 只由平台管理员分配。 */
export type PackageVisibility = "PUBLIC" | "BETA"

/** 订单状态：待支付、已支付、已取消。 */
export type OrderStatus = "PENDING" | "PAID" | "CANCELLED"

/** 订单类型：新建租户、套餐升级、套餐续期。 */
export type OrderKind = "CREATE" | "UPGRADE" | "RENEW"

/**
 * 创建租户时可选的功能套餐。字段与 FeaturePackageItem 对齐，
 * 价格单位为人民币分，0 表示免费；maxMembers 为 0 表示不限制人数。
 */
export type PublicPackageItem = {
  uid: string
  name: string
  remark?: string
  features?: string[]
  isBase: boolean
  status: string
  visibility: string
  /** 授权时长（天），0 表示永久。 */
  durationDays?: number
  /** 可容纳的成员数上限，0 表示不限制。 */
  maxMembers?: number
  /** 价格，单位为该币种的最小单位（分）。 */
  priceCents?: number
  /** 定价币种：CNY、USD、EUR 或 JPY。 */
  currency?: string
}

/** 一笔租户创建订单。价格、时长和人数上限是下单时的套餐快照。 */
export type OrderItem = {
  uid: string
  /** 支付后创建的租户标识；待支付或已取消时为空。 */
  namespaceUid?: string
  namespaceName: string
  namespaceRemark?: string
  packageUid: string
  packageName: string
  /** 订单金额，单位为该币种的最小单位（分），0 表示免费。 */
  priceCents: number
  /** 下单时的币种快照：CNY、USD、EUR 或 JPY。 */
  currency: string
  /** 授权时长（天）快照，0 表示永久。 */
  durationDays: number
  /** 成员数上限快照，0 表示不限制。 */
  maxMembers: number
  status: OrderStatus
  kind: OrderKind
  /** Unix 秒（int64），展示时 *1000。 */
  createdAt?: number
  /** Unix 秒；未支付时为 0。 */
  paidAt?: number
}

export type CreateOrderParams = {
  name: string
  remark?: string
  packageUid: string
}

export type ListOrdersParams = {
  page?: number
  pageSize?: number
  status?: string
}

export type ListOrdersResult = {
  items?: OrderItem[]
  total?: number
  page?: number
  pageSize?: number
}

export type PayOrderResult = {
  order: OrderItem
  namespace: NamespaceItem
}

/** 租户当前套餐。升级 / 续期入口用它展示当前价格、时长和到期时间。 */
export type NamespacePackageInfo = {
  packageUid?: string
  packageName?: string
  status?: string
  visibility?: string
  durationDays?: number
  maxMembers?: number
  priceCents?: number
  /** 定价币种：CNY、USD、EUR 或 JPY。 */
  currency?: string
  /** Unix 秒；0 表示永久。 */
  expiresAt?: number
  expired?: boolean
}
