/** 维护窗口状态：启用 / 停用。 */
export type AlertMaintenanceWindowStatus = "ACTIVE" | "DISABLED"

/**
 * 维护窗口。生效区间左闭右开；匹配条件之间是 AND，空条件表示不限。
 * labels 是 JSON 字符串（标签键 → 完整匹配正则），与后端存储格式一致。
 */
export interface AlertMaintenanceWindowItem {
  uid: string
  name: string
  description: string
  labels: string
  levelUids: string[]
  datasourceUids: string[]
  ruleUids: string[]
  /** 生效开始时间，RFC3339。 */
  startsAt: string
  /** 生效结束时间，RFC3339，必须晚于 startsAt。 */
  endsAt: string
  status: AlertMaintenanceWindowStatus | string
  createdAt: string
  updatedAt: string
  /** 创建人标识。 */
  createdBy: string
}

export interface ListAlertMaintenanceWindowsParams {
  page?: number
  pageSize?: number
  name?: string
  /** ACTIVE / DISABLED，空值表示不限制。 */
  status?: string
}

export interface ListAlertMaintenanceWindowsResult {
  items: AlertMaintenanceWindowItem[]
  total: number
  page: number
  pageSize: number
}

/** 创建与更新共用的载荷；更新时由接口函数补上 uid。 */
export interface SaveAlertMaintenanceWindowParams {
  name: string
  description?: string
  labels: string
  levelUids?: string[]
  datasourceUids?: string[]
  ruleUids?: string[]
  /** RFC3339。 */
  startsAt: string
  /** RFC3339。 */
  endsAt: string
}
