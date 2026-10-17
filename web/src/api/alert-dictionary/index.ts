import { del, get, post, put } from "@/api/http"

import type {
  AlertDictionaryItem,
  ListAlertDictionariesParams,
  ListAlertDictionariesResult,
  SaveAlertDictionaryParams,
} from "./types"

/** GET /v1/alert-dictionaries */
export function listAlertDictionaries(params: ListAlertDictionariesParams = {}): Promise<ListAlertDictionariesResult> {
  return get<ListAlertDictionariesResult>("/alert-dictionaries", params)
}

/** POST /v1/alert-dictionary */
export function createAlertDictionary(body: SaveAlertDictionaryParams): Promise<{ dictionary: AlertDictionaryItem }> {
  return post<{ dictionary: AlertDictionaryItem }>("/alert-dictionary", body)
}

/** PUT /v1/alert-dictionary/{uid} */
export function updateAlertDictionary(
  body: SaveAlertDictionaryParams & { uid: string },
): Promise<{ dictionary: AlertDictionaryItem }> {
  return put<{ dictionary: AlertDictionaryItem }>(`/alert-dictionary/${body.uid}`, body)
}

/** PUT /v1/alert-dictionary/{uid}/status */
export function updateAlertDictionaryStatus(uid: string, status: string): Promise<{ dictionary: AlertDictionaryItem }> {
  return put<{ dictionary: AlertDictionaryItem }>(`/alert-dictionary/${uid}/status`, { uid, status })
}

/** DELETE /v1/alert-dictionary/{uid} */
export function deleteAlertDictionary(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/alert-dictionary/${uid}`)
}

export type {
  AlertDictionaryItem,
  ListAlertDictionariesParams,
  ListAlertDictionariesResult,
  SaveAlertDictionaryParams,
} from "./types"
