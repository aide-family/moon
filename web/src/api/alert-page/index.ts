import { del, get, post, put } from "@/api/http"

import type {
  AlertColumnItem,
  AlertPageItem,
  ListAlertPagesParams,
  ListAlertPagesResult,
  RealtimeAlertTab,
  SaveAlertPageParams,
} from "./types"

/** GET /v1/alert-pages */
export function listAlertPages(params: ListAlertPagesParams = {}): Promise<ListAlertPagesResult> {
  return get<ListAlertPagesResult>("/alert-pages", params)
}

/** GET /v1/alert-page/{uid} */
export function getAlertPage(uid: string): Promise<{ page: AlertPageItem }> {
  return get<{ page: AlertPageItem }>(`/alert-page/${uid}`)
}

/** POST /v1/alert-page */
export function createAlertPage(body: SaveAlertPageParams): Promise<{ page: AlertPageItem }> {
  return post<{ page: AlertPageItem }>("/alert-page", body)
}

/** PUT /v1/alert-page/{uid} */
export function updateAlertPage(body: SaveAlertPageParams & { uid: string }): Promise<{ page: AlertPageItem }> {
  return put<{ page: AlertPageItem }>(`/alert-page/${body.uid}`, body)
}

/** DELETE /v1/alert-page/{uid} */
export function deleteAlertPage(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/alert-page/${uid}`)
}

/** GET /v1/alert-pages/my-tabs */
export function listMyRealtimeTabs(): Promise<{ tabs: RealtimeAlertTab[] }> {
  return get<{ tabs: RealtimeAlertTab[] }>("/alert-pages/my-tabs")
}

/** PUT /v1/alert-pages/my-tabs */
export function setMyRealtimeTabs(alertPageUids: string[]): Promise<{ tabs: RealtimeAlertTab[] }> {
  return put<{ tabs: RealtimeAlertTab[] }>("/alert-pages/my-tabs", { alertPageUids })
}

/** GET /v1/alert-pages/my-columns */
export function getMyAlertColumns(scope: string): Promise<{ columns: AlertColumnItem[] }> {
  return get<{ columns: AlertColumnItem[] }>("/alert-pages/my-columns", { scope })
}

/** PUT /v1/alert-pages/my-columns */
export function saveMyAlertColumns(
  scope: string,
  columns: AlertColumnItem[]
): Promise<{ columns: AlertColumnItem[] }> {
  return put<{ columns: AlertColumnItem[] }>("/alert-pages/my-columns", {
    scope,
    columns,
  })
}

export type {
  AlertColumnItem,
  AlertPageItem,
  ListAlertPagesParams,
  ListAlertPagesResult,
  RealtimeAlertTab,
  SaveAlertPageParams,
} from "./types"
