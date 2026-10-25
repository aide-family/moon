import i18n from "@/i18n"
import { toast } from "sonner"

const baseURL = "/v1"

const messageKeys: Record<string, string> = {
  "captcha is invalid": "error.captchaInvalid",
  "email already exists": "error.emailExists",
  "invalid email or code": "error.invalidCode",
  "email is not registered": "error.emailNotRegistered",
  "invalid username": "register.usernameInvalid",
  "send code too frequently": "error.sendTooFrequent",
  "oauth login failed": "error.oauthFailed",
  "oauth provider is not configured": "error.oauthNotConfigured",
  "oauth email is not verified": "error.oauthEmailNotVerified",
  "member already exists": "error.memberExists",
  "invitation expired": "error.invitationExpired",
  "invitation is not pending": "error.invitationHandled",
  "invitation belongs to another user": "error.invitationNotYours",
  "invitation email does not match": "error.invitationNotYours",
  "invalid datasource endpoint": "form.invalidUrl",
  "datasource connection already exists": "datasource.connectionExists",
  "name is too long": "form.nameTooLong",
}

const fieldByMessage: Record<string, string> = {
  "captcha is invalid": "captchaCode",
  "email already exists": "email",
  "email is not registered": "email",
  "invalid username": "name",
}

export class ApiError extends Error {
  status: number
  fields: Record<string, string>
  /** 服务端原始 message（未本地化）。需要按业务含义分支时用它，不要比对本地化后的文本。 */
  raw: string
  /** 服务端错误携带的附加信息，例如发码节流的 retryAfterSeconds。 */
  metadata: Record<string, string>

  constructor(
    status: number,
    message: string,
    fields: Record<string, string> = {},
    raw = "",
    metadata: Record<string, string> = {}
  ) {
    super(message)
    this.status = status
    this.fields = fields
    this.raw = raw
    this.metadata = metadata
  }
}

export function localizeMessage(message: string): string {
  const key = messageKeys[message]
  return key ? i18n.t(key) : message
}

function fieldName(raw: string): string {
  const snake = raw
    .replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
    .replace(/^_/, "")
  return snake.replace(/_([a-z])/g, (_, letter: string) => letter.toUpperCase())
}

export function parseFieldErrors(message: string): Record<string, string> {
  const known = fieldByMessage[message]
  if (known) {
    return { [known]: localizeMessage(message) }
  }
  const fields: Record<string, string> = {}
  const pattern = /invalid\s+[\w.]+\.(\w+)\s*:\s*([^;\n]+)/gi
  for (const match of message.matchAll(pattern)) {
    const name = fieldName(match[1])
    fields[name] = localizeMessage(match[2].trim())
  }
  return fields
}

function staysOnAuthForm(path: string): boolean {
  return path.startsWith("/auth/") || path === "/captcha"
}

function notifyStatus(status: number, message: string) {
  if (status === 404) {
    toast(message || i18n.t("error.notFound"))
    return
  }
  if (status >= 500) {
    toast.error(i18n.t("error.server"), {
      description: message || undefined,
      duration: 8000,
    })
  }
}

let refreshing: Promise<boolean> | null = null

function clearSession(): void {
  localStorage.removeItem("token")
  localStorage.removeItem("refreshToken")
  if (window.location.pathname !== "/login") {
    window.location.assign("/login")
  }
}

// refreshAccessToken 用 refresh token 换新的 access token。
// 并发 401 共享同一个刷新请求，避免同时发出多次刷新。
async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = localStorage.getItem("refreshToken")
  if (!refreshToken) {
    return false
  }
  if (!refreshing) {
    refreshing = (async () => {
      try {
        const headers = new Headers({ Authorization: `Bearer ${refreshToken}` })
        const namespace = localStorage.getItem("namespace")
        if (namespace) {
          headers.set("X-Namespace", namespace)
        }
        const response = await fetch(`${baseURL}/self/refresh-token`, {
          headers,
        })
        if (!response.ok) {
          return false
        }
        const data = (await response.json()) as { accessToken?: string }
        if (!data.accessToken) {
          return false
        }
        localStorage.setItem("token", data.accessToken)
        return true
      } catch {
        return false
      }
    })()
  }
  try {
    return await refreshing
  } finally {
    refreshing = null
  }
}

export async function request(
  path: string,
  init: RequestInit = {},
  retry = true
): Promise<Response> {
  const headers = new Headers(init.headers)
  const token = localStorage.getItem("token") ?? ""
  const namespace = localStorage.getItem("namespace") ?? ""
  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }
  // 切换租户时会显式传入目标租户，不能被 localStorage 里的旧值覆盖。
  if (namespace && !headers.has("X-Namespace")) {
    headers.set("X-Namespace", namespace)
  }
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }

  const response = await fetch(`${baseURL}${path}`, { ...init, headers })
  if (
    response.status === 401 &&
    retry &&
    !staysOnAuthForm(path) &&
    path !== "/self/refresh-token"
  ) {
    if (await refreshAccessToken()) {
      return request(path, init, false)
    }
    clearSession()
  }
  if (
    response.status === 403 &&
    !path.startsWith("/invitation/") &&
    window.location.pathname !== "/forbidden"
  ) {
    window.location.assign("/forbidden")
  }
  return response
}

export interface RequestOptions {
  // 后台轮询和列表加载由页面自己展示错误，不弹全局 toast。
  silent?: boolean
}

export async function send<T>(
  path: string,
  init: RequestInit = {},
  options: RequestOptions = {}
): Promise<T> {
  const response = await request(path, init)
  const text = await response.text()
  let data: { message?: string; metadata?: Record<string, string> } = {}
  if (text) {
    try {
      data = JSON.parse(text) as {
        message?: string
        metadata?: Record<string, string>
      }
    } catch {
      data = { message: text }
    }
  }
  if (!response.ok) {
    const raw = data.message || `request failed: ${response.status}`
    const message = localizeMessage(raw)
    const fields = response.status === 400 ? parseFieldErrors(raw) : {}
    if (
      !options.silent &&
      (response.status === 404 || response.status >= 500)
    ) {
      notifyStatus(response.status, message)
    }
    throw new ApiError(
      response.status,
      message,
      fields,
      raw,
      data.metadata ?? {}
    )
  }
  return data as T
}

export function get<T>(
  path: string,
  params?: object,
  init?: RequestInit,
  options?: RequestOptions
): Promise<T> {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(params ?? {})) {
    if (
      typeof value !== "string" &&
      typeof value !== "number" &&
      typeof value !== "boolean"
    ) {
      continue
    }
    if (value === "" || value === false) {
      continue
    }
    query.set(key, String(value))
  }
  const suffix = query.size > 0 ? `?${query.toString()}` : ""
  return send<T>(`${path}${suffix}`, init, options)
}

export function post<T>(path: string, body?: unknown): Promise<T> {
  return send<T>(path, { method: "POST", body: JSON.stringify(body ?? {}) })
}

export function put<T>(path: string, body?: unknown): Promise<T> {
  return send<T>(path, { method: "PUT", body: JSON.stringify(body ?? {}) })
}

export function del<T>(path: string): Promise<T> {
  return send<T>(path, { method: "DELETE" })
}
