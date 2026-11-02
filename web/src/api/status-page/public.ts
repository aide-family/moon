import type { PublicStatusPage } from "./types"

/**
 * 公开状态页请求失败。带上 HTTP 状态码，页面据此区分「链接无效」（404）、
 * 「访问太频繁」（429）与普通网络错误。
 */
export class PublicStatusPageError extends Error {
  status: number

  constructor(status: number, message = "") {
    super(message)
    this.name = "PublicStatusPageError"
    this.status = status
  }
}

/**
 * 按 token 读取公开状态页。
 *
 * 这个接口不需要登录与租户头，所以不能走 `@/api/http`：那边的封装会加
 * Authorization / X-Namespace，并在 401 时跳登录、在 403 时跳 /forbidden。
 * 这里用原生 fetch 直接请求相对路径，失败由页面自己展示。
 */
export async function fetchPublicStatusPage(
  token: string
): Promise<PublicStatusPage> {
  const response = await fetch(`/v1/status/${encodeURIComponent(token)}`, {
    method: "GET",
    headers: { Accept: "application/json" },
    cache: "no-store",
  })
  if (!response.ok) {
    let message = ""
    try {
      message = ((await response.json()) as { message?: string }).message ?? ""
    } catch {
      // 非 JSON 响应（网关错误页）沿用状态码判断。
    }
    throw new PublicStatusPageError(response.status, message)
  }
  const data = (await response.json()) as { statusPage?: PublicStatusPage }
  if (!data.statusPage) {
    throw new PublicStatusPageError(404)
  }
  return data.statusPage
}
