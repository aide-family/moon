import { del, get, post, put } from "@/api/http"

import type {
  AlertMaintenanceWindowItem,
  ListAlertMaintenanceWindowsParams,
  ListAlertMaintenanceWindowsResult,
  SaveAlertMaintenanceWindowParams,
} from "./types"

/** GET /v1/alert-maintenance-windows */
export function listAlertMaintenanceWindows(
  params: ListAlertMaintenanceWindowsParams = {}
): Promise<ListAlertMaintenanceWindowsResult> {
  return get<ListAlertMaintenanceWindowsResult>(
    "/alert-maintenance-windows",
    params
  )
}

/** POST /v1/alert-maintenance-window */
export function createAlertMaintenanceWindow(
  body: SaveAlertMaintenanceWindowParams
): Promise<{ window: AlertMaintenanceWindowItem }> {
  return post<{ window: AlertMaintenanceWindowItem }>(
    "/alert-maintenance-window",
    body
  )
}

/** PUT /v1/alert-maintenance-window/{uid} */
export function updateAlertMaintenanceWindow(
  body: SaveAlertMaintenanceWindowParams & { uid: string }
): Promise<{ window: AlertMaintenanceWindowItem }> {
  return put<{ window: AlertMaintenanceWindowItem }>(
    `/alert-maintenance-window/${body.uid}`,
    body
  )
}

/** PUT /v1/alert-maintenance-window/{uid}/status */
export function updateAlertMaintenanceWindowStatus(
  uid: string,
  status: string
): Promise<{ window: AlertMaintenanceWindowItem }> {
  return put<{ window: AlertMaintenanceWindowItem }>(
    `/alert-maintenance-window/${uid}/status`,
    { uid, status }
  )
}

/** DELETE /v1/alert-maintenance-window/{uid} */
export function deleteAlertMaintenanceWindow(
  uid: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/alert-maintenance-window/${uid}`)
}

/** GET /v1/alert-maintenance-window/{uid} */
export function getAlertMaintenanceWindow(
  uid: string
): Promise<{ window: AlertMaintenanceWindowItem }> {
  return get<{ window: AlertMaintenanceWindowItem }>(
    `/alert-maintenance-window/${uid}`
  )
}

export type {
  AlertMaintenanceWindowItem,
  AlertMaintenanceWindowStatus,
  ListAlertMaintenanceWindowsParams,
  ListAlertMaintenanceWindowsResult,
  SaveAlertMaintenanceWindowParams,
} from "./types"
