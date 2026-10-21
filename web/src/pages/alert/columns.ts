import type { AlertColumnItem } from "@/api/alert-page/types"

// 未配置列时实时告警表格使用的默认列，保持旧页面的信息密度。
export const DEFAULT_HISTORY_COLUMNS: AlertColumnItem[] = [
  { key: "summary", type: "FIELD", label: "" },
  { key: "datasource", type: "FIELD", label: "" },
  { key: "status", type: "FIELD", label: "" },
  { key: "sample_value", type: "FIELD", label: "" },
  { key: "starts_at", type: "FIELD", label: "" },
]

export const DEFAULT_REALTIME_COLUMNS: AlertColumnItem[] = [
  { key: "summary", type: "FIELD", label: "" },
  { key: "datasource", type: "FIELD", label: "" },
  { key: "level", type: "FIELD", label: "" },
  { key: "sample_value", type: "FIELD", label: "" },
  { key: "starts_at", type: "FIELD", label: "" },
  { key: "duration", type: "FIELD", label: "" },
  { key: "intervened", type: "FIELD", label: "" },
]

// 内置字段列允许用户选择，顺序就是可选下拉的顺序。
export const FIELD_COLUMNS: Array<{ key: string; labelKey: string }> = [
  { key: "summary", labelKey: "alertEvent.summary" },
  { key: "datasource", labelKey: "alertEvent.datasource" },
  { key: "level", labelKey: "alertEvent.level" },
  { key: "sample_value", labelKey: "alertEvent.sample" },
  { key: "starts_at", labelKey: "alertEvent.startsAt" },
  { key: "duration", labelKey: "alertEvent.duration" },
  { key: "intervened", labelKey: "alertEvent.intervened" },
  { key: "status", labelKey: "alertEvent.status" },
  { key: "description", labelKey: "alertEvent.descriptionField" },
  { key: "expr", labelKey: "alertEvent.expr" },
  { key: "fingerprint", labelKey: "alertEvent.fingerprint" },
  { key: "rule", labelKey: "alertEvent.rule" },
]

const fieldLabels = new Map(FIELD_COLUMNS.map((item) => [item.key, item.labelKey]))

// fieldLabel 返回内置字段在表格里的默认列名；标签列没有别名时直接用标签键。
export function fieldLabel(t: (key: string) => string, key: string): string {
  const textKey = fieldLabels.get(key)
  return textKey ? t(textKey) : key
}

export function isFieldMetricCount(key: string): boolean {
  return key === "duration" || key === "intervened" || key === "level"
}

// parseEventLabels 把事件的标签 JSON 转成对象；解析失败时返回空对象。
export function parseEventLabels(raw: string): Record<string, string> {
  try {
    const parsed = JSON.parse(raw || "{}") as Record<string, unknown>
    const out: Record<string, string> = {}
    for (const [key, value] of Object.entries(parsed ?? {})) {
      out[key] = value == null ? "" : String(value)
    }
    return out
  } catch {
    return {}
  }
}
