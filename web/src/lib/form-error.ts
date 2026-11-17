import { ApiError } from "@/api/http"

export interface FormFailure {
  form: string
  fields: Record<string, string>
}

export function readFormError(err: unknown, fallback: string): FormFailure {
  if (err instanceof ApiError) {
    // 404/5xx 已经由全局 toast 提示（http.ts 的 notifyStatus），表单里不再重复。
    if (err.status === 404 || err.status >= 500) {
      return { form: "", fields: {} }
    }
    // 400 的字段校验、409 的字段冲突都优先落到对应字段的 label 行；
    // 无法定位到字段时才留到表单级错误区。
    const mapped = Object.keys(err.fields).length > 0
    return {
      form: mapped ? "" : err.message || fallback,
      fields: err.fields,
    }
  }
  if (err instanceof Error && err.message) {
    return { form: err.message, fields: {} }
  }
  return { form: fallback, fields: {} }
}

export function formErrorMessage(err: unknown, fallback: string): string {
  const failure = readFormError(err, fallback)
  return failure.form || Object.values(failure.fields)[0] || ""
}

export function failureMessage(err: unknown, fallback: string): string {
  if (err instanceof Error && err.message) {
    return err.message
  }
  return fallback
}
