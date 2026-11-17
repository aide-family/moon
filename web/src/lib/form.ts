import i18n from "@/i18n"

/** Return an error message when value is blank after trim; otherwise null. */
export function requireTrimmed(value: string, message: string): string | null {
  if (value.trim() === "") {
    return message
  }
  return null
}

/** Return an error message when the trimmed value is longer than max characters; otherwise null. */
export function maxLength(value: string, max: number, message: string): string | null {
  return Array.from(value.trim()).length > max ? message : null
}

/** Return an error message when value is empty or the sentinel NONE; otherwise null. */
export function requireSelected(value: string, message: string): string | null {
  if (value.trim() === "" || value === "NONE") {
    return message
  }
  return null
}

/** 与后端一致的标签键正则：1–63 位字母、数字、点、下划线和连字符。 */
export const LABEL_KEY_PATTERN = /^[A-Za-z0-9._-]{1,63}$/

/** 与后端一致的标签值正则：0–63 位字母、数字、点、下划线和连字符。 */
export const LABEL_VALUE_PATTERN = /^[A-Za-z0-9._-]{0,63}$/

/** First non-null message from checks, or null if all pass. */
export function firstError(...checks: Array<string | null>): string | null {
  for (const check of checks) {
    if (check) {
      return check
    }
  }
  return null
}

/**
 * 正则格式校验：值非空且不符合 pattern 时返回格式化后的错误，空值交给必填规则处理。
 * 调用方传入的 message 必须已经说明字段名与期望格式（或用 formatError 生成）。
 */
export function patternError(
  value: string,
  pattern: RegExp,
  message: string
): string | null {
  const trimmed = value.trim()
  if (trimmed === "" || pattern.test(trimmed)) {
    return null
  }
  return message
}

/** 必填 + 正则：空值返回 requiredMessage，非空但不匹配返回 patternMessage。 */
export function requirePattern(
  value: string,
  pattern: RegExp,
  requiredMessage: string,
  patternMessage: string
): string | null {
  const trimmed = value.trim()
  if (trimmed === "") {
    return requiredMessage
  }
  return pattern.test(trimmed) ? null : patternMessage
}

/**
 * 必填校验：空值返回带字段名的定向文案，否则返回 null。
 * 与 requiredError 的区别是它会真的检查值，可直接放进 collectFieldErrors 的条目。
 */
export function requiredField(value: string, label: string): string | null {
  return value.trim() === "" ? requiredError(label) : null
}

/** 字段校验失败的统一文案：每条都带字段名和可执行的下一步。 */
export function requiredError(label: string): string {
  return i18n.t("field.required", { label })
}

/** 长度错误：0 下限时只说上限，避免“0–64 个字符”这种噪音。 */
export function lengthError(label: string, min: number, max: number): string {
  if (min <= 0) {
    return i18n.t("field.maxLength", { label, max })
  }
  return i18n.t("field.length", { label, min, max })
}

/** 整数范围错误：闭区间，单位由调用方写进 label（例如“重复间隔（秒）”）。 */
export function rangeError(
  label: string,
  min: number,
  max: number
): string {
  return i18n.t("field.range", { label, min, max })
}

/** 格式错误：hint 来自 field.hint.*，说明期望格式和例子。 */
export function formatError(label: string, hint: string): string {
  return i18n.t("field.pattern", { label, hint })
}

/** 正则格式约束的提示文案（含字段名），用于 patternError / requirePattern 的 message。 */
export function formatMessage(label: string, hint: string): string {
  return formatError(label, hint)
}

/** 常用格式提示。调用方仍可传任意本地化提示覆盖。 */
export function patternHint(
  kind:
    | "format"
    | "email"
    | "url"
    | "endpoint"
    | "cron"
    | "phone"
    | "slug"
    | "labelKey"
    | "labelValue"
    | "seconds"
    | "smtp"
    | "host"
    | "fromEmail"
    | "nonNegative"
): string {
  return i18n.t(`field.hint.${kind}`)
}

export interface FieldCheck {
  /** 字段显示名（已本地化），错误文案会原样带回，例如“名称”“端口”。 */
  label: string
  value: string
  required?: boolean
  /** 去空白后的最少字符数。 */
  min?: number
  /** 去空白后的最多字符数。 */
  max?: number
  /** 正则格式约束；与 hint 一起决定“期望的格式”。 */
  pattern?: RegExp
  /** pattern 不匹配时的格式说明，例如“以 http:// 或 https:// 开头”。 */
  hint?: string
  /** 整数闭区间，例如端口 1–65535。 */
  integerRange?: readonly [number, number]
}

/**
 * 单字段校验的统一入口：按 必填 → 整数范围 → 正则格式 → 长度 的顺序返回第一条
 * 可执行的中文/英文错误，文案里始终包含字段名和期望的格式或范围。
 * 返回 null 表示通过。
 */
export function checkField(check: FieldCheck): string | null {
  const value = check.value.trim()
  if (value === "") {
    return check.required ? requiredError(check.label) : null
  }
  if (check.integerRange) {
    const [min, max] = check.integerRange
    if (!/^-?\d+$/.test(value)) {
      return rangeError(check.label, min, max)
    }
    const parsed = Number(value)
    if (parsed < min || parsed > max) {
      return rangeError(check.label, min, max)
    }
  }
  if (check.pattern && !check.pattern.test(value)) {
    return formatError(check.label, check.hint ?? i18n.t("field.hint.format"))
  }
  const length = Array.from(value).length
  if (typeof check.min === "number" && length < check.min) {
    return lengthError(check.label, check.min, check.max ?? check.min)
  }
  if (typeof check.max === "number" && length > check.max) {
    return lengthError(check.label, check.min ?? 0, check.max)
  }
  return null
}

/** 字段名到错误原因的映射，用于在对应输入框下展示错误。 */
export type FieldErrors = Record<string, string>

/** 收集字段错误；同一字段只保留第一条原因。 */
export function collectFieldErrors(
  entries: Array<[field: string, message: string | null]>
): FieldErrors {
  const errors: FieldErrors = {}
  for (const [field, message] of entries) {
    if (message && !errors[field]) {
      errors[field] = message
    }
  }
  return errors
}

/** 是否存在任意字段错误。 */
export function hasFieldErrors(errors: FieldErrors): boolean {
  return Object.keys(errors).length > 0
}
