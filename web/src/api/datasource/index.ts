import { del, get, post, put } from "@/api/http"

import type {
  CreateDatasourceParams,
  DatasourceItem,
  ListDatasourcesParams,
  ListDatasourcesResult,
  SelectItem,
  UpdateDatasourceParams,
} from "./types"

/** GET /v1/datasources */
export function listDatasources(params: ListDatasourcesParams = {}): Promise<ListDatasourcesResult> {
  return get<ListDatasourcesResult>("/datasources", params)
}

/** GET /v1/datasources/select */
export function selectDatasources(params: { limit?: number; lastUid?: string; type?: string } = {}): Promise<{ items: SelectItem[]; lastUid: string }> {
  return get<{ items: SelectItem[]; lastUid: string }>("/datasources/select", params)
}

/** POST /v1/datasource */
export function createDatasource(body: CreateDatasourceParams): Promise<{ datasource: DatasourceItem }> {
  return post<{ datasource: DatasourceItem }>("/datasource", body)
}

/** PUT /v1/datasource/{uid} */
export function updateDatasource(body: UpdateDatasourceParams): Promise<{ datasource: DatasourceItem }> {
  return put<{ datasource: DatasourceItem }>(`/datasource/${body.uid}`, body)
}

/** PUT /v1/datasource/{uid}/status */
export function updateDatasourceStatus(uid: string, status: string): Promise<{ datasource: DatasourceItem }> {
  return put<{ datasource: DatasourceItem }>(`/datasource/${uid}/status`, { uid, status })
}

/** DELETE /v1/datasource/{uid} */
export function deleteDatasource(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/datasource/${uid}`)
}

/** GET /v1/datasource/{uid} */
export function getDatasource(uid: string): Promise<{ datasource: DatasourceItem }> {
  return get<{ datasource: DatasourceItem }>(`/datasource/${uid}`)
}

/** GET /v1/datasource/{uid}/status */
export function getDatasourceStatus(uid: string): Promise<{ status: string }> {
  return get<{ status: string }>(`/datasource/${uid}/status`)
}

/** GET /v1/datasource/{uid}/metrics */
export function listDatasourceMetrics(uid: string): Promise<{ names: string[] }> {
  return get<{ names: string[] }>(`/datasource/${uid}/metrics`)
}

/** GET /v1/datasource/{uid}/metrics/{name} */
export function getDatasourceMetric(uid: string, name: string): Promise<{ body: string }> {
  return get<{ body: string }>(`/datasource/${uid}/metrics/${encodeURIComponent(name)}`)
}

export type { CreateDatasourceParams, DatasourceItem, ListDatasourcesParams, ListDatasourcesResult, UpdateDatasourceParams } from "./types"
