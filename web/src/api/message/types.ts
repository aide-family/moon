export type SystemMessageItem = {
  uid: string
  kind: string
  actorName: string
  namespaceUid: string
  namespaceName: string
  roleName: string
  read: boolean
  createdAt: number
}

export type ListSelfMessagesParams = {
  page?: number
  pageSize?: number
  unreadOnly?: boolean
}

export type ListSelfMessagesResult = {
  items?: SystemMessageItem[]
  total?: number
  page?: number
  pageSize?: number
  unread?: number
}
