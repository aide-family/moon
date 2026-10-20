import { get, post } from "@/api/http"

import type { AcceptInvitationResult, ListSelfInvitationsResult, MemberInviteItem } from "./types"

/** GET /v1/self/invitations */
export function listSelfInvitations(): Promise<ListSelfInvitationsResult> {
  return get<ListSelfInvitationsResult>("/self/invitations")
}

/** GET /v1/invitation/{token} */
export function getInvitation(token: string): Promise<{ invitation: MemberInviteItem }> {
  return get<{ invitation: MemberInviteItem }>(`/invitation/${encodeURIComponent(token)}`)
}

/** POST /v1/invitation/accept */
export function acceptInvitation(body: { token?: string; uid?: string }): Promise<AcceptInvitationResult> {
  return post<AcceptInvitationResult>("/invitation/accept", body)
}

/** POST /v1/invitation/decline */
export function declineInvitation(body: { token?: string; uid?: string }): Promise<Record<string, never>> {
  return post<Record<string, never>>("/invitation/decline", body)
}
