import type { AlertEventItem, AlertEventStats } from "@/api/alert-event/types"
import type { ListWorkerInstancesResult } from "@/api/worker-instances/types"

/**
 * 概览页一个区块的权限可见性。
 * 概览只聚合已有接口，没有独立的后端端点，因此每个区块独立判断权限：
 * 缺权限的区块不渲染，而不是让整页 403。
 */
export interface DashboardScope {
  alerts: boolean
  delivery: boolean
  workers: boolean
  datasources: boolean
  rules: boolean
}

/** 告警快照：实时数量 + 触发中清单。任一子项失败时为 null，卡片各自降级。 */
export interface DashboardAlertSnapshot {
  stats: AlertEventStats | null
  items: AlertEventItem[] | null
  total: number
}

/**
 * 今日投递。successRate 为 null 表示当天没有任何投递尝试（不是 0，也不是 1）。
 * attempts 是成功率的真实分母，页面必须展示，避免把取消/积压算进失败率。
 */
export interface DashboardDeliverySnapshot {
  sent: number
  failed: number
  attempts: number
  successRate: number | null
}

export interface DashboardInventory {
  datasources: number | null
  rules: number | null
}

export interface DashboardSnapshot {
  observedAt: string
  alerts: DashboardAlertSnapshot | null
  delivery: DashboardDeliverySnapshot | null
  workers: ListWorkerInstancesResult | null
  inventory: DashboardInventory | null
}

/** 投递趋势的一天。date 为本地自然日 YYYY-MM-DD。 */
export interface DeliveryTrendPoint {
  date: string
  sent: number
  failed: number
}

export interface DeliveryTrend {
  points: DeliveryTrendPoint[]
  /** false 表示有日期统计失败，图上数字会偏低，页面必须提示。 */
  complete: boolean
}
