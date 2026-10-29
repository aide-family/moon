import type { NamespaceItem } from "@/api/namespace/types"

/**
 * 系统管理员视角的租户条目 = 租户本体 + 后端聚合出来的成员数与创建时间。
 * createdAt 是 Unix 秒（int64），与仓库其它 created_at 字段一致，展示时 *1000。
 */
export type SystemNamespaceItem = NamespaceItem & {
  memberCount: number
  createdAt: number
}

export type ListSystemNamespacesParams = {
  page?: number
  pageSize?: number
  name?: string
  /** "" 表示全部；后端只接受 ACTIVE / DISABLED。 */
  status?: string
}

/** PUT /v1/system/namespace/{uid}/status 的请求体。 */
export type UpdateNamespaceStatusPayload = {
  status: "ACTIVE" | "DISABLED"
  /** DISABLED 必填，1-500 字；后端会邮件通知租户负责人。 */
  reason?: string
}

export type ListSystemNamespacesResult = {
  items?: SystemNamespaceItem[]
  total?: number
  page?: number
  pageSize?: number
}
