export type NamespaceItem = {
  uid: string
  name: string
  ownerUid: string
  logo: string
  remark: string
  status: string
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
