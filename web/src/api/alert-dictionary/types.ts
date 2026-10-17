export type AlertDictionaryType = "ALERT" | "DATASOURCE"

export interface AlertDictionaryItem {
  uid: string
  name: string
  color: string
  type: AlertDictionaryType | string
  status: string
}

export interface ListAlertDictionariesParams {
  page?: number
  pageSize?: number
  type?: string
  name?: string
}

export interface ListAlertDictionariesResult {
  items: AlertDictionaryItem[]
  total: number
  page: number
  pageSize: number
}

export interface SaveAlertDictionaryParams {
  name: string
  color: string
  type?: AlertDictionaryType
}
