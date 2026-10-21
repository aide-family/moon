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

/** First non-null message from checks, or null if all pass. */
export function firstError(...checks: Array<string | null>): string | null {
  for (const check of checks) {
    if (check) {
      return check
    }
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
