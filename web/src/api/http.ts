const baseURL = "/v1"

export async function request(path: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers)
  const token = localStorage.getItem("token") ?? ""
  const namespace = localStorage.getItem("namespace") ?? ""
  headers.set("Authorization", `Bearer ${token}`)
  headers.set("X-Namespace", namespace)

  const response = await fetch(`${baseURL}${path}`, { ...init, headers })
  if (response.status === 401) {
    localStorage.removeItem("token")
    if (window.location.pathname !== "/login") {
      window.location.assign("/login")
    }
  }
  return response
}
