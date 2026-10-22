export type MemberItem = {
  uid: string
  namespaceUid: string
  userUid: string
  email: string
  phone: string
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

export type MemberInviteItem = {
  uid: string
  namespaceUid: string
  namespaceName: string
  email: string
  roleUid: string
  roleName: string
  inviterName: string
  status: string
  expiresAt: number
  createdAt: number
  registered: boolean
}

export type InviteMemberResult = {
  invitation: MemberInviteItem
  emailSent: boolean
}

export type ListMemberInvitesResult = {
  items?: MemberInviteItem[]
}
