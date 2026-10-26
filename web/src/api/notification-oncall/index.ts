import { del, get, post, put } from "@/api/http"

import type {
  CreateNotificationOnCallParams,
  ListNotificationOnCallsParams,
  ListNotificationOnCallsResult,
  NotificationOnCallItem,
  UpdateNotificationOnCallParams,
} from "./types"

/** GET /v1/notification-oncalls */
export function listNotificationOnCalls(
  params: ListNotificationOnCallsParams = {}
): Promise<ListNotificationOnCallsResult> {
  return get<ListNotificationOnCallsResult>("/notification-oncalls", params)
}

/** POST /v1/notification-oncall */
export function createNotificationOnCall(
  body: CreateNotificationOnCallParams
): Promise<{ onCall: NotificationOnCallItem }> {
  return post<{ onCall: NotificationOnCallItem }>("/notification-oncall", body)
}

/** PUT /v1/notification-oncall/{uid} */
export function updateNotificationOnCall(
  body: UpdateNotificationOnCallParams
): Promise<{ onCall: NotificationOnCallItem }> {
  return put<{ onCall: NotificationOnCallItem }>(
    `/notification-oncall/${body.uid}`,
    body
  )
}

/** PUT /v1/notification-oncall/{uid}/status */
export function updateNotificationOnCallStatus(
  uid: string,
  status: string
): Promise<{ onCall: NotificationOnCallItem }> {
  return put<{ onCall: NotificationOnCallItem }>(
    `/notification-oncall/${uid}/status`,
    { uid, status }
  )
}

/** DELETE /v1/notification-oncall/{uid} */
export function deleteNotificationOnCall(
  uid: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/notification-oncall/${uid}`)
}

/** GET /v1/notification-oncall/{uid} */
export function getNotificationOnCall(
  uid: string
): Promise<{ onCall: NotificationOnCallItem }> {
  return get<{ onCall: NotificationOnCallItem }>(`/notification-oncall/${uid}`)
}

export type {
  CreateNotificationOnCallParams,
  ListNotificationOnCallsParams,
  ListNotificationOnCallsResult,
  NotificationOnCallItem,
  OnCallGroupInput,
  OnCallGroupItem,
  OnCallWeekday,
  OnCallWindow,
  UpdateNotificationOnCallParams,
} from "./types"
