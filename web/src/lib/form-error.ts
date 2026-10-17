import { ApiError } from "@/api/http"

export interface FormFailure {
  form: string
  fields: Record<string, string>
}

export function readFormError(err: unknown, fallback: string): FormFailure {
  if (err instanceof ApiError) {
    if (err.status === 400) {
      const mapped = Object.keys(err.fields).length > 0
      return { form: mapped ? "" : err.message || fallback, fields: err.fields }
    }
    if (err.status === 404 || err.status >= 500) {
      return { form: "", fields: {} }
    }
    if (err.status === 409) {
      return { form: err.message || fallback, fields: {} }
    }
    return { form: err.message || fallback, fields: {} }
  }
  if (err instanceof Error && err.message) {
    return { form: err.message, fields: {} }
  }
  return { form: fallback, fields: {} }
}

export function formMessage(err: unknown, fallback: string): string {
  const failure = readFormError(err, fallback)
  return failure.form || Object.values(failure.fields)[0] || ""
}

export function failureText(err: unknown, fallback: string): string {
  if (err instanceof Error && err.message) {
    return err.message
  }
  return fallback
}
