import { get, post } from "@/api/http"

import type {
  ListMessageLogsParams,
  ListMessageLogsResult,
  ListNotificationTasksParams,
  ListNotificationTasksResult,
  MessageLogItem,
  NotificationTaskItem,
} from "./types"

/** GET /v1/message-logs */
export function listMessageLogs(params: ListMessageLogsParams = {}): Promise<ListMessageLogsResult> {
  return get<ListMessageLogsResult>("/message-logs", params)
}

/** GET /v1/message-log/{uid} */
export function getMessageLog(uid: string): Promise<{ log: MessageLogItem }> {
  return get<{ log: MessageLogItem }>(`/message-log/${uid}`)
}

/** GET /v1/notification-tasks */
export function listNotificationTasks(params: ListNotificationTasksParams = {}): Promise<ListNotificationTasksResult> {
  return get<ListNotificationTasksResult>("/notification-tasks", params)
}

/** POST /v1/notification-task/{uid}/retry */
export function retryNotificationTask(uid: string): Promise<{ task: NotificationTaskItem }> {
  return post<{ task: NotificationTaskItem }>(`/notification-task/${uid}/retry`, {})
}

/** POST /v1/notification-task/{uid}/cancel */
export function cancelNotificationTask(uid: string): Promise<{ task: NotificationTaskItem }> {
  return post<{ task: NotificationTaskItem }>(`/notification-task/${uid}/cancel`, {})
}

export type {
  ListMessageLogsParams,
  ListMessageLogsResult,
  ListNotificationTasksParams,
  ListNotificationTasksResult,
  MessageLogItem,
  NotificationTaskItem,
} from "./types"
