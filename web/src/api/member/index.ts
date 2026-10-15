import { del, get, post, put } from "@/api/http"

import type { InviteMemberParams, ListMembersParams, ListMembersResult, MemberItem } from "./types"

/** GET /v1/members */
export function listMembers(params: ListMembersParams = {}): Promise<ListMembersResult> {
  return get<ListMembersResult>("/members", params)
}

/** POST /v1/member/invite */
export function inviteMember(body: InviteMemberParams): Promise<{ member: MemberItem }> {
  return post<{ member: MemberItem }>("/member/invite", body)
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
