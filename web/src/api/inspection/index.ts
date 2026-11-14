import { del, get, post, put } from "@/api/http"

import type {
  TestInspectionItemResult,
  InspectionItemInput,
  InspectionPlan,
  InspectionPlanInput,
  InspectionPlanStatus,
  InspectionReport,
  InspectionRun,
  ListInspectionPlansParams,
  ListInspectionPlansResult,
  ListInspectionRunsParams,
  ListInspectionRunsResult,
  ListInspectionTemplatesParams,
  ListInspectionTemplatesResult,
  SaveInspectionTemplateParams,
  ShareInspectionReportResult,
  InspectionReportShare,
} from "./types"

/** GET /v1/inspection/plans */
export function listInspectionPlans(
  params: ListInspectionPlansParams = {}
): Promise<ListInspectionPlansResult> {
  return get<ListInspectionPlansResult>("/inspection/plans", params)
}

/** GET /v1/inspection/plans/{uid} */
export function getInspectionPlan(
  uid: string
): Promise<{ plan: InspectionPlan }> {
  return get<{ plan: InspectionPlan }>(`/inspection/plans/${uid}`)
}

/** POST /v1/inspection/plans */
export function createInspectionPlan(
  body: InspectionPlanInput
): Promise<{ plan: InspectionPlan }> {
  return post<{ plan: InspectionPlan }>("/inspection/plans", { plan: body })
}

/** PUT /v1/inspection/plans/{uid} */
export function updateInspectionPlan(
  uid: string,
  body: InspectionPlanInput
): Promise<{ plan: InspectionPlan }> {
  return put<{ plan: InspectionPlan }>(`/inspection/plans/${uid}`, { plan: body })
}

/** PUT /v1/inspection/plans/{uid}/status */
export function updateInspectionPlanStatus(
  uid: string,
  status: InspectionPlanStatus
): Promise<{ plan: InspectionPlan }> {
  return put<{ plan: InspectionPlan }>(`/inspection/plans/${uid}/status`, {
    uid,
    status,
  })
}

/** DELETE /v1/inspection/plans/{uid} */
export function deleteInspectionPlan(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/inspection/plans/${uid}`)
}

/** POST /v1/inspection/plans/{uid}/run */
export function runInspectionPlan(uid: string): Promise<{ run: InspectionRun }> {
  return post<{ run: InspectionRun }>(`/inspection/plans/${uid}/run`, {})
}

/** GET /v1/inspection/templates */
export function listInspectionTemplates(
  params: ListInspectionTemplatesParams = {}
): Promise<ListInspectionTemplatesResult> {
  return get<ListInspectionTemplatesResult>("/inspection/templates", params)
}

/** POST /v1/inspection/templates */
export function createInspectionTemplate(
  body: SaveInspectionTemplateParams
): Promise<{ template: unknown }> {
  return post<{ template: unknown }>("/inspection/templates", body)
}

/** PUT /v1/inspection/templates/{uid} */
export function updateInspectionTemplate(
  uid: string,
  body: SaveInspectionTemplateParams
): Promise<{ template: unknown }> {
  return put<{ template: unknown }>(`/inspection/templates/${uid}`, body)
}

/** DELETE /v1/inspection/templates/{uid} */
export function deleteInspectionTemplate(
  uid: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/inspection/templates/${uid}`)
}

/** GET /v1/inspection/runs */
export function listInspectionRuns(
  params: ListInspectionRunsParams = {}
): Promise<ListInspectionRunsResult> {
  return get<ListInspectionRunsResult>("/inspection/runs", params)
}

/** GET /v1/inspection/runs/{uid} */
export function getInspectionRun(uid: string): Promise<{ run: InspectionRun }> {
  return get<{ run: InspectionRun }>(`/inspection/runs/${uid}`)
}

/** GET /v1/inspection/runs/{uid}/reports */
export function listRunReports(
  uid: string
): Promise<{ items: InspectionReport[] }> {
  return get<{ items: InspectionReport[] }>(`/inspection/runs/${uid}/reports`)
}

/** GET /v1/inspection/reports/{uid} */
export function getInspectionReport(
  uid: string
): Promise<{ report: InspectionReport }> {
  return get<{ report: InspectionReport }>(`/inspection/reports/${uid}`)
}

/** POST /v1/inspection/test-item：真实执行一次检查项，不落库。 */
export function testInspectionItem(
  item: InspectionItemInput
): Promise<TestInspectionItemResult> {
  return post<TestInspectionItemResult>("/inspection/test-item", { item })
}

/** POST /v1/inspection/reports/{uid}/resend */
export function resendInspectionReport(
  uid: string
): Promise<{ created: number }> {
  return post<{ created: number }>(`/inspection/reports/${uid}/resend`, {})
}

/** POST /v1/inspection/reports/{uid}/share */
export function shareInspectionReport(
  uid: string,
  days: number,
  password: string
): Promise<ShareInspectionReportResult> {
  return post<ShareInspectionReportResult>(
    `/inspection/reports/${uid}/share`,
    { uid, days, password }
  )
}

/** DELETE /v1/inspection/reports/{uid}/share */
export function revokeInspectionReportShare(
  uid: string
): Promise<{ report: InspectionReport }> {
  return del<{ report: InspectionReport }>(`/inspection/reports/${uid}/share`)
}

/** GET /v1/inspection/reports/{uid}/shares */
export function listInspectionReportShares(
  uid: string
): Promise<{ items: InspectionReportShare[] }> {
  return get<{ items: InspectionReportShare[] }>(
    `/inspection/reports/${uid}/shares`
  )
}

/** DELETE /v1/inspection/reports/{uid}/shares/{shareUid} */
export function revokeInspectionReportShareLink(
  uid: string,
  shareUid: string
): Promise<{ share: InspectionReportShare }> {
  return del<{ share: InspectionReportShare }>(
    `/inspection/reports/${uid}/shares/${shareUid}`
  )
}

/** GET /v1/public/inspection-reports/{token}：公开页使用，不走登录态。 */
export function getSharedInspectionReport(
  token: string,
  password = ""
): Promise<{
  title: string
  format: string
  content: string
  createdAt: string
}> {
  return get(`/public/inspection-reports/${encodeURIComponent(token)}`, {
    password,
  })
}

export type * from "./types"
