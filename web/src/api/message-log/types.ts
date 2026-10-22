export interface MessageLogItem {
  uid: string
  taskUid: string
  status: string
  channelUid: string
  templateUid: string
  sender?: string
  receivers: string
  subject?: string
  content?: string
  detail: string
  createdAt: string
  groupName?: string
  subscriptionName?: string
}

export interface ListMessageLogsParams {
  page?: number
  pageSize?: number
  taskUid?: string
}

export interface ListMessageLogsResult {
  items: MessageLogItem[]
  total: number
  page: number
  pageSize: number
}

export interface NotificationTaskItem {
  uid: string
  eventUid: string
  phase: string
  status: string
  ruleUid: string
  groupUid: string
  channelUid: string
  templateUid: string
  sender?: string
  receivers: string
  subject: string
  content?: string
  retriesLeft: number
  lastError: string
  subscriptionUid?: string
  groupName?: string
  subscriptionName?: string
  channelName?: string
  templateName?: string
}

export interface ListNotificationTasksParams {
  page?: number
  pageSize?: number
  status?: string
}

export interface ListNotificationTasksResult {
  items: NotificationTaskItem[]
  total: number
  page: number
  pageSize: number
}
