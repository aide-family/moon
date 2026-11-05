import type { SiteHome, SitePageDetail } from "./types"

/** 预览接口失败：保留 HTTP 状态码，让预览页区分未登录、无权限与普通失败。 */
export class SitePreviewError extends Error {
  status: number

  constructor(status: number, message = "") {
    super(message)
    this.name = "SitePreviewError"
    this.status = status
  }
}

/**
 * 预览页专用请求：带 token 读后台草稿。
 *
 * 预览页允许未登录直接打开，401 必须留在页面上提示「请先登录」；`@/api/http`
 * 会在刷新失败后清理会话并跳 /login，所以这里用原生 fetch 单独处理状态码。
 */
async function previewJSON<T>(path: string): Promise<T> {
  const token = localStorage.getItem("token") ?? ""
  if (!token) {
    throw new SitePreviewError(401)
  }
  let response: Response
  try {
    response = await fetch(`/v1${path}`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        Authorization: `Bearer ${token}`,
      },
      cache: "no-store",
    })
  } catch {
    throw new SitePreviewError(0)
  }
  if (!response.ok) {
    let message = ""
    try {
      message = ((await response.json()) as { message?: string }).message ?? ""
    } catch {
      // 网关错误页不是 JSON：只保留状态码，页面按失败态展示。
    }
    throw new SitePreviewError(response.status, message)
  }
  return (await response.json()) as T
}

/** GET /v1/system/site/pages/{uid}：取后台草稿（不是线上已发布快照）。 */
export async function fetchSitePageDraft(
  uid: string
): Promise<SitePageDetail> {
  const data = await previewJSON<{ page?: SitePageDetail }>(
    `/system/site/pages/${encodeURIComponent(uid)}`
  )
  if (!data.page) {
    throw new SitePreviewError(500)
  }
  return data.page
}

/** GET /v1/system/site/home?locale=：首页文案按草稿语言读取。 */
export async function fetchSiteHomeDraft(locale: "zh" | "en"): Promise<SiteHome> {
  return previewJSON<SiteHome>(`/system/site/home?locale=${locale}`)
}
