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
  return {
    ...labels,
    datasource: item.name,
    ...(item.levelName ? { datasource_level: item.levelName } : {}),
  }
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
