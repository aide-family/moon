export type AlertLevelType = "ALERT" | "DATASOURCE"

export interface AlertLevelItem {
  uid: string
  name: string
  color: string
  type: AlertLevelType | string
  status: string
}

export interface ListAlertLevelsParams {
  page?: number
  pageSize?: number
  type?: string
  name?: string
}

export interface ListAlertLevelsResult {
  items: AlertLevelItem[]
  total: number
  page: number
  pageSize: number
}

export interface SaveAlertLevelParams {
  uid?: string
  name: string
  color?: string
  type?: string
}
