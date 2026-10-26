import type { DatasourceItem } from "@/api/datasource/types"
import { labelMatchersMatch, type LabelPair } from "@/lib/labels"

/** 解析数据源标签 JSON；无法解析时返回 null，调用方按不匹配处理。 */
export function parseDatasourceLabels(
  raw: string
): Record<string, string> | null {
  try {
    return JSON.parse(raw || "{}") as Record<string, string>
  } catch {
    return null
  }
}

// datasourceMatchLabels 与后端保持一致：数据源自身标签之外，
// datasource_level 和 datasource 也可以被规则匹配。
export function datasourceMatchLabels(
  item: DatasourceItem
): Record<string, string> | null {
  const labels = parseDatasourceLabels(item.labels)
  if (labels === null) {
    return null
  }
  const merged: Record<string, string> = { ...labels, datasource: item.name }
  if (item.levelName) {
    merged.datasource_level = item.levelName
  } else {
    // 与后端一致：数据源没有等级时不并入 datasource_level，
    // 即使数据源自己的标签里带了这个键（后端也会在评测前删掉）。
    delete merged.datasource_level
  }
  return merged
}

/** 数据源标签匹配的系统键：数据源名称和等级名称由后端在评测时并入标签。 */
export const DATASOURCE_MATCH_SYSTEM_KEYS = [
  "datasource",
  "datasource_level",
] as const

// 匹配键和值的候选集合只来自同一驱动、启用状态的数据源，
// 和 matchedDatasources/matchedDatasourceCount 使用的集合保持一致。
function matchableDatasources(
  sources: DatasourceItem[],
  driver: string
): DatasourceItem[] {
  return sources.filter(
    (item) => item.driver === driver && item.status === "ACTIVE"
  )
}

function compareText(left: string, right: string): number {
  return left < right ? -1 : left > right ? 1 : 0
}

/**
 * 规则可选的匹配键：系统键固定在前，其后是启用数据源里出现的自定义标签键。
 * 自定义键去重并排序；与系统键同名的标签键不重复出现。
 */
export function datasourceMatchKeys(
  sources: DatasourceItem[],
  driver: string
): string[] {
  const customKeys = new Set<string>()
  for (const item of matchableDatasources(sources, driver)) {
    const labels = parseDatasourceLabels(item.labels)
    if (labels === null) {
      continue
    }
    for (const rawKey of Object.keys(labels)) {
      const key = rawKey.trim()
      if (
        key !== "" &&
        !(DATASOURCE_MATCH_SYSTEM_KEYS as readonly string[]).includes(key)
      ) {
        customKeys.add(key)
      }
    }
  }
  return [...DATASOURCE_MATCH_SYSTEM_KEYS, ...[...customKeys].sort(compareText)]
}

/**
 * 选中匹配键后可以带出的值：
 * datasource -> 数据源名称，datasource_level -> 等级名称，其余键 -> 数据源同名标签值。
 * 值保留原始内容（匹配按完整正则比较，去掉空白会改变语义），只去掉空值并去重排序。
 */
export function datasourceMatchValues(
  sources: DatasourceItem[],
  driver: string,
  key: string
): string[] {
  const matchKey = key.trim()
  if (matchKey === "") {
    return []
  }
  const values = new Set<string>()
  for (const item of matchableDatasources(sources, driver)) {
    if (matchKey === "datasource") {
      if (item.name.trim() !== "") {
        values.add(item.name)
      }
      continue
    }
    if (matchKey === "datasource_level") {
      if (item.levelName.trim() !== "") {
        values.add(item.levelName)
      }
      continue
    }
    const labels = parseDatasourceLabels(item.labels)
    const value = labels?.[matchKey]
    if (typeof value === "string" && value.trim() !== "") {
      values.add(value)
    }
  }
  return [...values].sort(compareText)
}

/** 规则会命中哪些启用数据源：驱动一致且标签匹配器全部命中。 */
export function matchedDatasources(
  sources: DatasourceItem[],
  driver: string,
  matchers: LabelPair[]
): DatasourceItem[] {
  return sources.filter((item) => {
    if (item.driver !== driver || item.status !== "ACTIVE") {
      return false
    }
    const labels = datasourceMatchLabels(item)
    return labels !== null && labelMatchersMatch(matchers, labels)
  })
}

export function matchDatasourceCount(
  sources: DatasourceItem[],
  driver: string,
  matchers: LabelPair[]
): number {
  return matchedDatasources(sources, driver, matchers).length
}

/** 表达式编辑器用来选数据源的启发式：取第一个匹配的启用数据源。 */
export function matchedDatasourceUid(
  sources: DatasourceItem[],
  driver: string,
  matchers: LabelPair[]
): string {
  if (
    !driver.startsWith("METRICS_") &&
    driver !== "LOGS_LOKI" &&
    driver !== "LOGS_VLOGS" &&
    driver !== "LOGS_ELASTICSEARCH"
  ) {
    return ""
  }
  return matchedDatasources(sources, driver, matchers)[0]?.uid ?? ""
}

/** 规则标签 JSON 的键名，评测时会并入事件标签，用于模板补全。 */
export function ruleLabelKeys(labels: string): string[] {
  try {
    const parsed: unknown = JSON.parse(labels || "{}")
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return []
    }
    return Object.keys(parsed as Record<string, unknown>).filter(
      (key) => key !== ""
    )
  } catch {
    return []
  }
}
