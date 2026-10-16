export interface AlertPageItem {
  uid: string
  name: string
  ownerUid: string
  labels: string
  public: boolean
  remark: string
}

export interface RealtimeAlertTab {
  id: string
  name: string
  kind: "mine" | "page" | string
}

export interface ListAlertPagesParams {
  page?: number
  pageSize?: number
  name?: string
  scope?: "manage" | "selectable"
}

export interface ListAlertPagesResult {
  items: AlertPageItem[]
  total: number
  page: number
  pageSize: number
}

export interface SaveAlertPageParams {
  name: string
  labels: string
  public: boolean
  remark: string
}
