export interface NotificationTemplateItem {
  uid: string
  name: string
  type: string
  status: string
  subject: string
  body: string
  remark: string
}

export interface ListNotificationTemplatesParams {
  page?: number
  pageSize?: number
  name?: string
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
  subject?: string
  body?: string
  remark?: string
}

export interface UpdateNotificationTemplateParams extends CreateNotificationTemplateParams {
  uid: string
}

export interface SelectItem {
  uid: string
  name: string
}
