export type NamespaceItem = {
  uid: string
  name: string
  ownerUid: string
  ownerName?: string
  ownerEmail?: string
  logo: string
  remark: string
  status: string
  /** 后端聚合的成员数；成员列表等接口不返回时缺省。 */
  memberCount?: number
  /** Unix 秒（int64），展示时 *1000。 */
  createdAt?: number
  /** 当前套餐名称；系统租户列表返回，成员视角可能缺省。 */
  packageName?: string
  /** 套餐到期时间（Unix 秒，0 表示永久）。 */
  packageExpiresAt?: number | string
  /** 停用原因；仅 DISABLED 租户有值，后端会据此发邮件通知负责人。 */
  disableReason?: string
  /** 当前套餐的成员数上限，0 表示不限制。 */
  maxMembers?: number
}

export type ListNamespacesParams = {
  page?: number
  pageSize?: number
  name?: string
}

export type ListNamespacesResult = {
  items?: NamespaceItem[]
  total?: number
  page?: number
  pageSize?: number
}

export type CreateNamespaceParams = {
  name: string
  remark?: string
}

export type UpdateNamespaceParams = {
  uid: string
  name: string
  remark?: string
  logo?: string
}
