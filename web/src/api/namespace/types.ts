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
