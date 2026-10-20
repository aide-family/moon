import { get, put } from "@/api/http"

import type { ListSelfMessagesParams, ListSelfMessagesResult } from "./types"

/** GET /v1/self/messages */
export function listSelfMessages(params: ListSelfMessagesParams = {}): Promise<ListSelfMessagesResult> {
  return get<ListSelfMessagesResult>("/self/messages", params, undefined, { silent: true })
}

/** GET /v1/self/messages/unread-count */
export function getUnreadMessageCount(): Promise<{ count: number }> {
  return get<{ count: number }>("/self/messages/unread-count", undefined, undefined, { silent: true })
}

/** PUT /v1/self/messages/{uid}/read */
export function markMessageRead(uid: string): Promise<Record<string, never>> {
  return put<Record<string, never>>(`/self/messages/${uid}/read`)
}

/** PUT /v1/self/messages/read-all */
export function markAllMessagesRead(): Promise<Record<string, never>> {
  return put<Record<string, never>>("/self/messages/read-all")
}
