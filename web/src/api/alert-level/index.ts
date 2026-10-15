import { del, get, post, put } from "@/api/http"

import type { AlertLevelItem, ListAlertLevelsParams, ListAlertLevelsResult, SaveAlertLevelParams } from "./types"

/** GET /v1/alert-levels */
export function listAlertLevels(params: ListAlertLevelsParams = {}): Promise<ListAlertLevelsResult> {
  return get<ListAlertLevelsResult>("/alert-levels", params)
}

/** POST /v1/alert-level */
export function createAlertLevel(body: SaveAlertLevelParams): Promise<{ level: AlertLevelItem }> {
  return post<{ level: AlertLevelItem }>("/alert-level", body)
}

/** PUT /v1/alert-level/{uid} */
export function updateAlertLevel(body: SaveAlertLevelParams & { uid: string }): Promise<{ level: AlertLevelItem }> {
  return put<{ level: AlertLevelItem }>(`/alert-level/${body.uid}`, body)
}

/** PUT /v1/alert-level/{uid}/status */
export function updateAlertLevelStatus(uid: string, status: string): Promise<{ level: AlertLevelItem }> {
  return put<{ level: AlertLevelItem }>(`/alert-level/${uid}/status`, { uid, status })
}

/** DELETE /v1/alert-level/{uid} */
export function deleteAlertLevel(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/alert-level/${uid}`)
}

export type { AlertLevelItem, ListAlertLevelsParams, ListAlertLevelsResult, SaveAlertLevelParams } from "./types"
