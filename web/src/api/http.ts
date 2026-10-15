const baseURL = "/v1"

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

export async function request(path: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers)
  const token = localStorage.getItem("token") ?? ""
  const namespace = localStorage.getItem("namespace") ?? ""
  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }
  if (namespace) {
    headers.set("X-Namespace", namespace)
  }
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }

  const response = await fetch(`${baseURL}${path}`, { ...init, headers })
  if (response.status === 401 && !path.startsWith("/auth/") && path !== "/captcha") {
    localStorage.removeItem("token")
    localStorage.removeItem("refreshToken")
    if (window.location.pathname !== "/login") {
      window.location.assign("/login")
    }
  }
  return response
}

export async function send<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await request(path, init)
  const text = await response.text()
  const data = text ? (JSON.parse(text) as { message?: string }) : {}
  if (!response.ok) {
    throw new ApiError(response.status, data.message || `request failed: ${response.status}`)
  }
  return data as T
}

export function get<T>(path: string, params?: Record<string, string | number | undefined>): Promise<T> {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(params ?? {})) {
    if (value !== undefined && value !== "") {
      query.set(key, String(value))
    }
  }
  const suffix = query.size > 0 ? `?${query.toString()}` : ""
  return send<T>(`${path}${suffix}`)
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
