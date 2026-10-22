export interface NotificationTemplateItem {
  uid: string
  name: string
  type: string
  app: string
  status: string
  subject: string
  body: string
  remark: string
}

export interface ListNotificationTemplatesParams {
  page?: number
  pageSize?: number
  name?: string
  type?: string
  app?: string
}

export interface ListNotificationTemplatesResult {
  items: NotificationTemplateItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateNotificationTemplateParams {
  name: string
  type: string
  app?: string
  subject?: string
  body?: string
  remark?: string
}

export interface UpdateNotificationTemplateParams extends CreateNotificationTemplateParams {
  uid: string
}

export interface NotificationPresetItem {
  uid: string
  name: string
  type: string
  app: string
  subject: string
  body: string
  description: string
}

export interface ListNotificationPresetsResult {
  items: NotificationPresetItem[]
}

export interface PreviewNotificationTemplateParams {
  type: string
  app?: string
  subject?: string
  body?: string
}

export interface PreviewNotificationTemplateResult {
  subject: string
  body: string
}
