import { del, get, post, put } from "@/api/http"

import type {
  CreateNotificationGroupParams,
  ListNotificationGroupsParams,
  ListNotificationGroupsResult,
  NotificationGroupItem,
  SelectItem,
  UpdateNotificationGroupParams,
} from "./types"

/** GET /v1/notification-groups */
export function listNotificationGroups(params: ListNotificationGroupsParams = {}): Promise<ListNotificationGroupsResult> {
  return get<ListNotificationGroupsResult>("/notification-groups", params)
}

/** GET /v1/notification-groups/select */
export function selectNotificationGroups(params: { limit?: number; lastUid?: string; activeOnly?: boolean } = {}): Promise<{ items: SelectItem[]; lastUid: string }> {
  return get<{ items: SelectItem[]; lastUid: string }>("/notification-groups/select", params)
}

/** POST /v1/notification-group */
export function createNotificationGroup(body: CreateNotificationGroupParams): Promise<{ group: NotificationGroupItem }> {
  return post<{ group: NotificationGroupItem }>("/notification-group", body)
}

/** PUT /v1/notification-group/{uid} */
export function updateNotificationGroup(body: UpdateNotificationGroupParams): Promise<{ group: NotificationGroupItem }> {
  return put<{ group: NotificationGroupItem }>(`/notification-group/${body.uid}`, body)
}

/** PUT /v1/notification-group/{uid}/status */
export function updateNotificationGroupStatus(uid: string, status: string): Promise<{ group: NotificationGroupItem }> {
  return put<{ group: NotificationGroupItem }>(`/notification-group/${uid}/status`, { uid, status })
}

/** DELETE /v1/notification-group/{uid} */
export function deleteNotificationGroup(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/notification-group/${uid}`)
}

/** GET /v1/notification-group/{uid} */
export function getNotificationGroup(uid: string): Promise<{ group: NotificationGroupItem }> {
  return get<{ group: NotificationGroupItem }>(`/notification-group/${uid}`)
}

export type {
  CreateNotificationGroupParams,
  ListNotificationGroupsParams,
  ListNotificationGroupsResult,
  NotificationGroupItem,
  UpdateNotificationGroupParams,
} from "./types"
