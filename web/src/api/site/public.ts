import type { SiteConfig } from "./types"

/**
 * 公开站点配置请求失败。带上 HTTP 状态码，调用方可以区分 404 与普通网络错误。
 */
export class SiteConfigError extends Error {
  status: number

  constructor(status: number, message = "") {
    super(message)
    this.name = "SiteConfigError"
    this.status = status
  }
}

/**
 * GET /v1/site/config：官网、登录页与注册页共用的公开配置。
 *
 * 这个接口不需要登录与租户头，所以不能走 `@/api/http`：那边的封装会加
 * Authorization / X-Namespace，并在 401 时跳登录、在 403 时跳 /forbidden。
 * 这里用原生 fetch 直接请求相对路径，失败由调用方决定是否回落默认值。
 */
export async function fetchSiteConfig(): Promise<SiteConfig> {
  const response = await fetch("/v1/site/config", {
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
    throw new SiteConfigError(response.status, message)
  }
  const data = (await response.json()) as Partial<SiteConfig>
  if (!data.site || !data.loginMethods) {
    throw new SiteConfigError(500)
  }
  return { site: data.site, loginMethods: data.loginMethods }
}
