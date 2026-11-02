import { del, get, post, put } from "@/api/http"

import type {
  ListStatusPagesParams,
  ListStatusPagesResult,
  SaveStatusPageParams,
  StatusPageItem,
  StatusPageStatus,
} from "./types"

/** GET /v1/status-pages */
export function listStatusPages(
  params: ListStatusPagesParams = {}
): Promise<ListStatusPagesResult> {
  return get<ListStatusPagesResult>("/status-pages", params)
}

/** GET /v1/status-page/{uid} */
export function getStatusPage(
  uid: string
): Promise<{ statusPage: StatusPageItem }> {
  return get<{ statusPage: StatusPageItem }>(`/status-page/${uid}`)
}

/** POST /v1/status-page */
export function createStatusPage(
  body: SaveStatusPageParams
): Promise<{ statusPage: StatusPageItem }> {
  return post<{ statusPage: StatusPageItem }>("/status-page", body)
}

/** PUT /v1/status-page/{uid} */
export function updateStatusPage(
  body: SaveStatusPageParams & { uid: string }
): Promise<{ statusPage: StatusPageItem }> {
  return put<{ statusPage: StatusPageItem }>(`/status-page/${body.uid}`, body)
}

/** PUT /v1/status-page/{uid}/status：停用后公开链接立即失效。 */
export function updateStatusPageStatus(
  uid: string,
  status: StatusPageStatus
): Promise<{ statusPage: StatusPageItem }> {
  return put<{ statusPage: StatusPageItem }>(`/status-page/${uid}/status`, {
    uid,
    status,
  })
}

/** DELETE /v1/status-page/{uid} */
export function deleteStatusPage(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/status-page/${uid}`)
}

/** POST /v1/status-page/{uid}/token：重新生成公开链接，旧链接立即失效。 */
export function regenerateStatusPageToken(
  uid: string
): Promise<{ statusPage: StatusPageItem }> {
  return post<{ statusPage: StatusPageItem }>(`/status-page/${uid}/token`, {})
}

export type {
  ListStatusPagesParams,
  ListStatusPagesResult,
  PublicComponentStatus,
  PublicEventStatus,
  PublicStatusComponent,
  PublicStatusEvent,
  PublicStatusPage,
  SaveStatusPageParams,
  StatusPageItem,
  StatusPageStatus,
} from "./types"
