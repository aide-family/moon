import type { AlertColumnItem } from "@/api/alert-page/types"

// 历史明细是事件台账：必须给出终态、结束时间和持续时长，
// 否则“已恢复”的记录看不出什么时候结束、持续了多久。
export const DEFAULT_HISTORY_COLUMNS: AlertColumnItem[] = [
  { key: "summary", type: "FIELD", label: "" },
  { key: "status", type: "FIELD", label: "" },
  { key: "level", type: "FIELD", label: "" },
  { key: "datasource", type: "FIELD", label: "" },
  { key: "sample_value", type: "FIELD", label: "" },
  { key: "starts_at", type: "FIELD", label: "" },
  { key: "ends_at", type: "FIELD", label: "" },
  { key: "duration", type: "FIELD", label: "" },
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

// 只在历史明细里出现的字段：实时列表只含仍在触发的事件，ends_at 必然为空。
// 历史明细允许把它加回列配置，所以单独列出一份供选择器使用。
const HISTORY_ONLY_FIELD_COLUMNS: Array<{ key: string; labelKey: string }> = [
  { key: "ends_at", labelKey: "alertEvent.endsAt" },
]

// 历史明细列选择器可选的内置字段：内置字段 + 仅历史字段。
export const HISTORY_FIELD_COLUMNS: Array<{ key: string; labelKey: string }> = [
  ...FIELD_COLUMNS,
  ...HISTORY_ONLY_FIELD_COLUMNS,
]

// 历史明细的列配置视图标识，与后端 alertpage 包的 tabHistory 一致。
// 和实时告警的 mine 一样，配置按用户存在后端，随账号走而不是跟着浏览器。
export const HISTORY_COLUMNS_SCOPE = "history"

const fieldLabels = new Map(
  [...FIELD_COLUMNS, ...HISTORY_ONLY_FIELD_COLUMNS].map((item) => [
    item.key,
    item.labelKey,
  ])
)

// fieldLabel 返回内置字段在表格里的默认列名，未知字段回退到键。
export function fieldLabel(t: (key: string) => string, key: string): string {
  const textKey = fieldLabels.get(key)
  return textKey ? t(textKey) : key
}

// columnTitle 返回表格列头：内置字段优先用别名，没有别名时走 i18n 名称；
// 标签列的别名通常是中文，所以英文界面统一显示标签键，其他语言显示别名。
export function columnTitle(
  t: (key: string) => string,
  lang: string,
  column: AlertColumnItem
): string {
  const alias = column.label.trim()
  if (column.type === "LABEL") {
    return lang.startsWith("en") ? column.key : alias || column.key
  }
  return alias || fieldLabel(t, column.key)
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
