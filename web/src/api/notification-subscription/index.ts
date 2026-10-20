import { del, get, post, put } from "@/api/http"

import type {
  CreateNotificationSubscriptionParams,
  ListNotificationSubscriptionsParams,
  ListNotificationSubscriptionsResult,
  NotificationSubscriptionItem,
  UpdateNotificationSubscriptionParams,
} from "./types"

/** GET /v1/notification-subscriptions */
export function listNotificationSubscriptions(
  params: ListNotificationSubscriptionsParams = {},
): Promise<ListNotificationSubscriptionsResult> {
  return get<ListNotificationSubscriptionsResult>("/notification-subscriptions", params)
}

/** POST /v1/notification-subscription */
export function createNotificationSubscription(
  body: CreateNotificationSubscriptionParams,
): Promise<{ subscription: NotificationSubscriptionItem }> {
  return post<{ subscription: NotificationSubscriptionItem }>("/notification-subscription", body)
}

/** PUT /v1/notification-subscription/{uid} */
export function updateNotificationSubscription(
  body: UpdateNotificationSubscriptionParams,
): Promise<{ subscription: NotificationSubscriptionItem }> {
  return put<{ subscription: NotificationSubscriptionItem }>(`/notification-subscription/${body.uid}`, body)
}

/** PUT /v1/notification-subscription/{uid}/status */
export function updateNotificationSubscriptionStatus(
  uid: string,
  status: string,
): Promise<{ subscription: NotificationSubscriptionItem }> {
  return put<{ subscription: NotificationSubscriptionItem }>(`/notification-subscription/${uid}/status`, { uid, status })
}

/** DELETE /v1/notification-subscription/{uid} */
export function deleteNotificationSubscription(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/notification-subscription/${uid}`)
}

export type {
  CreateNotificationSubscriptionParams,
  ListNotificationSubscriptionsParams,
  ListNotificationSubscriptionsResult,
  NotificationSubscriptionItem,
  UpdateNotificationSubscriptionParams,
} from "./types"
