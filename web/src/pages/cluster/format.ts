import type { TFunction } from "i18next"

import type { BadgeTone } from "@/components/status-badge"

/** 把 Unix 秒格式化成浏览器本地日期时间；0 表示从未发生。 */
export function formatClusterTime(raw: number | string): string {
  const seconds = Number(raw)
  if (!seconds) {
    return ""
  }
  const date = new Date(seconds * 1000)
  if (Number.isNaN(date.getTime())) {
    return ""
  }
  return date.toLocaleString()
}

const MEMORY_UNITS: Record<string, number> = {
  K: 1_000,
  M: 1_000 ** 2,
  G: 1_000 ** 3,
  T: 1_000 ** 4,
  P: 1_000 ** 5,
  Ki: 1024,
  Mi: 1024 ** 2,
  Gi: 1024 ** 3,
  Ti: 1024 ** 4,
  Pi: 1024 ** 5,
}

/** 把节点的 CPU 容量（如 8、7500m、100000n）解析成毫核。无法解析时返回 0。 */
export function parseCpuMillis(raw: string): number {
  const value = (raw ?? "").trim()
  if (!value) {
    return 0
  }
  const match = /^([0-9]+(?:\.[0-9]+)?)([num])?$/.exec(value)
  if (!match) {
    return 0
  }
  const amount = Number(match[1])
  if (!Number.isFinite(amount)) {
    return 0
  }
  switch (match[2]) {
    case "n":
      return amount / 1_000_000
    case "u":
      return amount / 1_000
    case "m":
      return amount
    default:
      return amount * 1_000
  }
}

/** 把节点的内存容量（如 16384Mi、16Gi、17179869184）解析成字节。无法解析时返回 0。 */
export function parseMemoryBytes(raw: string): number {
  const value = (raw ?? "").trim()
  if (!value) {
    return 0
  }
  const match = /^([0-9]+(?:\.[0-9]+)?)([KMGTP]i?)?$/.exec(value)
  if (!match) {
    return 0
  }
  const amount = Number(match[1])
  if (!Number.isFinite(amount)) {
    return 0
  }
  const unit = match[2] ? MEMORY_UNITS[match[2]] : 1
  return unit ? amount * unit : 0
}

function formatDecimal(value: number): string {
  if (value >= 100) {
    return String(Math.round(value))
  }
  const rounded = Math.round(value * 10) / 10
  return Number.isInteger(rounded) ? String(rounded) : rounded.toFixed(1)
}

/** 把毫核汇总成可读容量，例如 16000 → 16 核。 */
export function formatCpuCapacity(millis: number, t: TFunction): string {
  if (!millis) {
    return "—"
  }
  return `${formatDecimal(millis / 1_000)} ${t("kube.unitCores")}`
}

/** 把字节汇总成 GiB/TiB，内存容量不需要更细的单位。 */
export function formatMemoryCapacity(bytes: number): string {
  if (!bytes) {
    return "—"
  }
  if (bytes >= 1024 ** 4) {
    return `${formatDecimal(bytes / 1024 ** 4)} TiB`
  }
  return `${formatDecimal(bytes / 1024 ** 3)} GiB`
}

/** 就绪副本占比，0..1；分母为 0 时按满值处理，避免出现 NaN 或 0% 的误读。 */
export function readReadyRatio(ready: number, desired: number): number {
  if (desired <= 0) {
    return 1
  }
  return Math.min(1, Math.max(0, ready / desired))
}

/** 把 0..1 的比例格式化成整数百分比。 */
export function formatRatio(value: number): string {
  return `${Math.round(Math.min(1, Math.max(0, value)) * 100)}%`
}

/** Pod 的 ready 是「就绪容器数/容器总数」摘要；字段缺失时按未就绪处理。 */
export function isPodReady(ready: string): boolean {
  const [readyCount, totalCount] = (ready ?? "").split("/")
  const readyValue = Number(readyCount)
  const totalValue = Number(totalCount)
  if (!Number.isFinite(readyValue) || !Number.isFinite(totalValue)) {
    return false
  }
  return totalValue > 0 && readyValue >= totalValue
}

/**
 * 把「创建时间」换算成相对现在的运行时长。
 * 只用于展示，不参与排序；未来时间或无效值统一回落为占位符。
 */
export function formatClusterAge(raw: number | string, t: TFunction): string {
  const seconds = Number(raw)
  if (!seconds) {
    return "—"
  }
  const elapsed = Math.max(0, Math.floor(Date.now() / 1000) - seconds)
  if (elapsed < 60) {
    return t("kube.ageSeconds", { count: elapsed })
  }
  if (elapsed < 3600) {
    return t("kube.ageMinutes", { count: Math.floor(elapsed / 60) })
  }
  if (elapsed < 86400) {
    return t("kube.ageHours", { count: Math.floor(elapsed / 3600) })
  }
  return t("kube.ageDays", { count: Math.floor(elapsed / 86400) })
}

/** Job 执行状态 → 标签颜色：完成=成功、失败=危险、运行=信息、暂停/等待=警告。 */
export function readJobStatusTone(status: string): BadgeTone {
  switch (status) {
    case "Complete":
      return "success"
    case "Failed":
      return "danger"
    case "Running":
      return "info"
    case "Suspended":
    case "Pending":
      return "warning"
    default:
      return "neutral"
  }
}

/** Pod 阶段 → 标签颜色，与 Pod / 工作负载页签的状态语义一致。 */
export function readPodPhaseTone(phase: string): BadgeTone {
  switch (phase) {
    case "Running":
      return "success"
    case "Pending":
      return "warning"
    case "Failed":
      return "danger"
    case "Succeeded":
      return "info"
    default:
      return "neutral"
  }
}

const JOB_STATUS_KEYS: Record<string, string> = {
  Suspended: "kube.jobStatusSuspended",
  Failed: "kube.jobStatusFailed",
  Complete: "kube.jobStatusComplete",
  Running: "kube.jobStatusRunning",
  Unknown: "kube.jobStatusUnknown",
  Pending: "kube.jobStatusPending",
}

/** Job 执行状态标签；未知状态回落到服务端原文。 */
export function formatJobStatus(t: TFunction, status: string): string {
  const key = JOB_STATUS_KEYS[status]
  return key ? t(key) : status
}
