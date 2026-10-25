import type { AlertEventStats } from "@/api/alert-event/types"
import { get } from "@/api/http"
import type { ListAlertEventsResult } from "@/api/alert-event/types"
import type { ListMessageLogsResult } from "@/api/message-log/types"
import type { ListWorkerInstancesResult } from "@/api/worker-instances/types"
import {
  addDays,
  dayStartSeconds,
  formatDateValue,
  nextDayStartSeconds,
} from "@/lib/date-range"

import type {
  DashboardAlertSnapshot,
  DashboardDeliverySnapshot,
  DashboardScope,
  DashboardSnapshot,
  DeliveryTrend,
  DeliveryTrendPoint,
} from "./types"

// 概览是自动刷新的落地页：任何单个区块失败都只在卡片内降级，
// 不弹全局 toast、也不整页报错。
const quiet = { silent: true }

async function settle<T>(task: Promise<T>): Promise<T | null> {
  try {
    return await task
  } catch {
    return null
  }
}

function countTotal(path: string): Promise<number> {
  return get<{ total: number }>(path, { pageSize: 1 }, undefined, quiet).then(
    (result) => result.total
  )
}

/** 今天的本地自然日起止（Unix 秒，上界排除），与消息日志接口的口径一致。 */
function todayBounds(): { startTime: number; endTime: number } | null {
  const today = formatDateValue(new Date())
  const startTime = dayStartSeconds(today)
  const endTime = nextDayStartSeconds(today)
  if (startTime === undefined || endTime === undefined) {
    return null
  }
  return { startTime, endTime }
}

function countLogs(params: {
  status: string
  startTime: number
  endTime: number
}): Promise<number> {
  return get<ListMessageLogsResult>(
    "/message-logs",
    { ...params, pageSize: 1 },
    undefined,
    quiet
  ).then((result) => result.total)
}

async function loadAlerts(): Promise<DashboardAlertSnapshot> {
  const [stats, list] = await Promise.all([
    settle(
      get<AlertEventStats>("/alert-event-stats", undefined, undefined, quiet)
    ),
    settle(
      get<ListAlertEventsResult>(
        "/alert-events",
        { view: "REALTIME", pageSize: 6 },
        undefined,
        quiet
      )
    ),
  ])
  return {
    stats,
    items: list?.items ?? null,
    total: list?.total ?? 0,
  }
}

async function loadDelivery(): Promise<DashboardDeliverySnapshot | null> {
  const bounds = todayBounds()
  if (!bounds) {
    return null
  }
  const [sent, failed] = await Promise.all([
    settle(countLogs({ status: "SENT", ...bounds })),
    settle(countLogs({ status: "FAILED", ...bounds })),
  ])
  if (sent === null || failed === null) {
    return null
  }
  // 成功率的分母是「已尝试」的投递：取消和积压没有尝试过，不能算失败。
  const attempts = sent + failed
  return {
    sent,
    failed,
    attempts,
    successRate: attempts === 0 ? null : sent / attempts,
  }
}

/**
 * 概览快照。所有数字都来自已有接口，没有新增后端聚合端点：
 * 具备精确计数的项（实时数量、当日投递、资产总数）直接展示，
 * 需要跨周表聚合的项（趋势、P90 闭环时长）不在这里伪造。
 */
export async function fetchDashboard(
  scope: DashboardScope
): Promise<DashboardSnapshot> {
  const [alerts, delivery, workers, datasources, rules] = await Promise.all([
    scope.alerts ? loadAlerts() : Promise.resolve(null),
    scope.delivery ? loadDelivery() : Promise.resolve(null),
    scope.workers
      ? settle(
          get<ListWorkerInstancesResult>(
            "/worker-instances",
            undefined,
            undefined,
            quiet
          )
        )
      : Promise.resolve(null),
    scope.datasources
      ? settle(countTotal("/datasources"))
      : Promise.resolve(null),
    scope.rules ? settle(countTotal("/alert-rules")) : Promise.resolve(null),
  ])

  return {
    observedAt: new Date().toISOString(),
    alerts,
    delivery,
    workers,
    inventory:
      scope.datasources || scope.rules ? { datasources, rules } : null,
  }
}

/**
 * 近 N 天投递趋势。每天两个精确计数（SENT / FAILED），失败的那天不出点，
 * 由页面标注「部分日期统计失败」，而不是用 0 冒充实测值。
 */
export async function fetchDeliveryTrend(days = 7): Promise<DeliveryTrend> {
  const today = new Date()
  const dates = Array.from({ length: days }, (_, index) =>
    addDays(today, -(days - 1 - index))
  )
  const results = await Promise.all(
    dates.map(async (date): Promise<DeliveryTrendPoint | null> => {
      const key = formatDateValue(date)
      const startTime = dayStartSeconds(key)
      const endTime = nextDayStartSeconds(key)
      if (startTime === undefined || endTime === undefined) {
        return null
      }
      const [sent, failed] = await Promise.all([
        settle(countLogs({ status: "SENT", startTime, endTime })),
        settle(countLogs({ status: "FAILED", startTime, endTime })),
      ])
      if (sent === null || failed === null) {
        return null
      }
      return { date: key, sent, failed }
    })
  )
  const points = results.filter(
    (point): point is DeliveryTrendPoint => point !== null
  )
  return { points, complete: points.length === results.length }
}

export type {
  DashboardAlertSnapshot,
  DashboardDeliverySnapshot,
  DashboardInventory,
  DashboardScope,
  DashboardSnapshot,
  DeliveryTrend,
  DeliveryTrendPoint,
} from "./types"
