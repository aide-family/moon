import { del, get, post, put, type RequestOptions } from "@/api/http"

import type {
  CreateSitePageParams,
  ListSitePagesParams,
  SitePageDetail,
  SitePageList,
  UpdateSitePageNavParams,
  UpdateSitePageParams,
} from "./types"

// 官网页面管理接口：走 @/api/http，401/403 与会话刷新和其它后台页面一致。
// 读接口固定 silent：页面自己渲染失败态与重试，不弹全局错误 toast。
const silent: RequestOptions = { silent: true }

/** GET /v1/system/site/pages */
export function listSitePages(
  params: ListSitePagesParams
): Promise<SitePageList> {
  return get<SitePageList>("/system/site/pages", params, undefined, silent)
}

/** GET /v1/system/site/pages/{uid} */
export function getSitePage(uid: string): Promise<{ page: SitePageDetail }> {
  return get<{ page: SitePageDetail }>(
    `/system/site/pages/${uid}`,
    undefined,
    undefined,
    silent
  )
}

/** POST /v1/system/site/pages */
export function createSitePage(
  body: CreateSitePageParams
): Promise<{ page: SitePageDetail }> {
  return post<{ page: SitePageDetail }>("/system/site/pages", body)
}

/** PUT /v1/system/site/pages/{uid}。updatedAt 是乐观锁，冲突返回 409。 */
export function updateSitePage({
  uid,
  ...body
}: UpdateSitePageParams): Promise<{ page: SitePageDetail }> {
  return put<{ page: SitePageDetail }>(`/system/site/pages/${uid}`, body)
}

/** POST /v1/system/site/pages/{uid}/publish */
export function publishSitePage(
  uid: string
): Promise<{ page: SitePageDetail }> {
  return post<{ page: SitePageDetail }>(
    `/system/site/pages/${uid}/publish`,
    {}
  )
}

/** POST /v1/system/site/pages/{uid}/unpublish */
export function unpublishSitePage(
  uid: string
): Promise<{ page: SitePageDetail }> {
  return post<{ page: SitePageDetail }>(
    `/system/site/pages/${uid}/unpublish`,
    {}
  )
}

/** DELETE /v1/system/site/pages/{uid} */
export function deleteSitePage(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/system/site/pages/${uid}`)
}

/** PUT /v1/system/site/pages/{uid}/nav */
export function updateSitePageNav({
  uid,
  ...body
}: UpdateSitePageNavParams): Promise<{ page: SitePageDetail }> {
  return put<{ page: SitePageDetail }>(
    `/system/site/pages/${uid}/nav`,
    body
  )
}
