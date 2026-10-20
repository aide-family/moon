import { del, get, post, put } from "@/api/http"

import type {
  CreateNotificationChannelParams,
  ListNotificationChannelsParams,
  ListNotificationChannelsResult,
  NotificationChannelItem,
  UpdateNotificationChannelParams,
} from "./types"

/** GET /v1/notification-channels */
export function listNotificationChannels(params: ListNotificationChannelsParams = {}): Promise<ListNotificationChannelsResult> {
  return get<ListNotificationChannelsResult>("/notification-channels", params)
}

/** POST /v1/notification-channel */
export function createNotificationChannel(body: CreateNotificationChannelParams): Promise<{ channel: NotificationChannelItem }> {
  return post<{ channel: NotificationChannelItem }>("/notification-channel", body)
}

/** PUT /v1/notification-channel/{uid} */
export function updateNotificationChannel(body: UpdateNotificationChannelParams): Promise<{ channel: NotificationChannelItem }> {
  return put<{ channel: NotificationChannelItem }>(`/notification-channel/${body.uid}`, body)
}

/** PUT /v1/notification-channel/{uid}/status */
export function updateNotificationChannelStatus(uid: string, status: string): Promise<{ channel: NotificationChannelItem }> {
  return put<{ channel: NotificationChannelItem }>(`/notification-channel/${uid}/status`, { uid, status })
}

/** DELETE /v1/notification-channel/{uid} */
export function deleteNotificationChannel(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/notification-channel/${uid}`)
}

/** GET /v1/notification-channel/{uid} */
export function getNotificationChannel(uid: string): Promise<{ channel: NotificationChannelItem }> {
  return get<{ channel: NotificationChannelItem }>(`/notification-channel/${uid}`)
}

/** POST /v1/notification-channel/{uid}/test */
export function testNotificationChannel(uid: string, body: { templateUid?: string; receivers: string; subject?: string; body?: string; params?: string }): Promise<{ taskUid: string }> {
  return post<{ taskUid: string }>(`/notification-channel/${uid}/test`, body)
}

export type {
  CreateNotificationChannelParams,
  ListNotificationChannelsParams,
  ListNotificationChannelsResult,
  NotificationChannelItem,
  UpdateNotificationChannelParams,
} from "./types"
