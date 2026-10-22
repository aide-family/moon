import { del, get, post, put } from "@/api/http"

import type {
  CreateNotificationTemplateParams,
  ListNotificationTemplatesParams,
  ListNotificationTemplatesResult,
  NotificationPresetItem,
  NotificationTemplateItem,
  PreviewNotificationTemplateParams,
  PreviewNotificationTemplateResult,
  UpdateNotificationTemplateParams,
} from "./types"

/** GET /v1/notification-templates */
export function listNotificationTemplates(params: ListNotificationTemplatesParams = {}): Promise<ListNotificationTemplatesResult> {
  return get<ListNotificationTemplatesResult>("/notification-templates", params)
}

/** GET /v1/notification-presets */
export function listNotificationPresets(): Promise<{ items: NotificationPresetItem[] }> {
  return get<{ items: NotificationPresetItem[] }>("/notification-presets")
}

/** POST /v1/notification-template */
export function createNotificationTemplate(body: CreateNotificationTemplateParams): Promise<{ template: NotificationTemplateItem }> {
  return post<{ template: NotificationTemplateItem }>("/notification-template", body)
}

/** PUT /v1/notification-template/{uid} */
export function updateNotificationTemplate(body: UpdateNotificationTemplateParams): Promise<{ template: NotificationTemplateItem }> {
  return put<{ template: NotificationTemplateItem }>(`/notification-template/${body.uid}`, body)
}

/** PUT /v1/notification-template/{uid}/status */
export function updateNotificationTemplateStatus(uid: string, status: string): Promise<{ template: NotificationTemplateItem }> {
  return put<{ template: NotificationTemplateItem }>(`/notification-template/${uid}/status`, { uid, status })
}

/** DELETE /v1/notification-template/{uid} */
export function deleteNotificationTemplate(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/notification-template/${uid}`)
}

/** GET /v1/notification-template/{uid} */
export function getNotificationTemplate(uid: string): Promise<{ template: NotificationTemplateItem }> {
  return get<{ template: NotificationTemplateItem }>(`/notification-template/${uid}`)
}

/** POST /v1/notification-template/preview */
export function previewNotificationTemplate(
  body: PreviewNotificationTemplateParams,
): Promise<PreviewNotificationTemplateResult> {
  return post<PreviewNotificationTemplateResult>("/notification-template/preview", body)
}

/** POST /v1/notification-template/{uid}/test */
export function testNotificationTemplate(uid: string, body: { channelUid: string; receivers: string; params?: string }): Promise<{ taskUid: string }> {
  return post<{ taskUid: string }>(`/notification-template/${uid}/test`, body)
}

export type {
  CreateNotificationTemplateParams,
  ListNotificationTemplatesParams,
  ListNotificationTemplatesResult,
  NotificationPresetItem,
  NotificationTemplateItem,
  PreviewNotificationTemplateParams,
  PreviewNotificationTemplateResult,
  UpdateNotificationTemplateParams,
} from "./types"
