import { del, get, post, put } from "@/api/http"

import type {
  InviteMemberParams,
  InviteMemberResult,
  ListMemberInvitesResult,
  ListMembersParams,
  ListMembersResult,
  MemberItem,
} from "./types"

/** GET /v1/members */
export function listMembers(params: ListMembersParams = {}): Promise<ListMembersResult> {
  return get<ListMembersResult>("/members", params)
}

/** GET /v1/member/invites */
export function listMemberInvites(): Promise<ListMemberInvitesResult> {
  return get<ListMemberInvitesResult>("/member-invites")
}

/** POST /v1/member/invite */
export function inviteMember(body: InviteMemberParams): Promise<InviteMemberResult> {
  return post<InviteMemberResult>("/member/invite", body)
}

/** DELETE /v1/member/invite/{uid} */
export function revokeMemberInvite(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/member/invite/${uid}`)
}

/** DELETE /v1/member/{uid} */
export function deleteMember(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/member/${uid}`)
}

/** PUT /v1/member/{uid}/status */
export function updateMemberStatus(uid: string, status: string): Promise<{ member: MemberItem }> {
  return put<{ member: MemberItem }>(`/member/${uid}/status`, { uid, status })
}

/** PUT /v1/member/{uid}/roles */
export function updateMemberRoles(uid: string, roleUids: string[]): Promise<{ member: MemberItem }> {
  return put<{ member: MemberItem }>(`/member/${uid}/roles`, { uid, roleUids })
}
