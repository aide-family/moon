import type { MemberItem } from "@/api/member/types"

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

export type ListSelfInvitationsResult = {
  items?: MemberInviteItem[]
}

export type AcceptInvitationResult = {
  member: MemberItem
}
