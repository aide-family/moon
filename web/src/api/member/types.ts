export type MemberItem = {
  uid: string
  namespaceUid: string
  userUid: string
  email: string
  name: string
  status: string
  roleUids?: string[]
  roleNames?: string[]
}

export type ListMembersParams = {
  page?: number
  pageSize?: number
  name?: string
}

export type ListMembersResult = {
  items?: MemberItem[]
  total?: number
  page?: number
  pageSize?: number
}

export type InviteMemberParams = {
  email: string
  roleUid: string
}
