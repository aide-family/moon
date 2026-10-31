import { get, post, send, sendBlob, type RequestOptions } from "@/api/http"

import type {
  ArchiveImportResult,
  MaintenanceStatus,
  SchemaMigrationResult,
} from "./types"

/** GET /v1/system/maintenance。需要系统权限 system:migration:read。 */
export function getMaintenanceStatus(
  options?: RequestOptions
): Promise<MaintenanceStatus> {
  return get<MaintenanceStatus>(
    "/system/maintenance",
    undefined,
    undefined,
    options
  )
}

/** POST /v1/system/maintenance/migrate。需要 system:migration:write（仅超管）。 */
export function runSchemaMigration(): Promise<SchemaMigrationResult> {
  return post<SchemaMigrationResult>("/system/maintenance/migrate", {})
}

/** GET /v1/system/maintenance/archive。流式下载迁移包，401/403 由 http.ts 统一处理。 */
export function downloadArchive(): Promise<{ blob: Blob; filename: string }> {
  return sendBlob("/system/maintenance/archive", { method: "GET" })
}

/** POST /v1/system/maintenance/archive。请求体就是迁移包文件本身。 */
export function importArchive(file: File): Promise<ArchiveImportResult> {
  return send<ArchiveImportResult>("/system/maintenance/archive", {
    method: "POST",
    body: file,
    headers: { "Content-Type": "application/gzip" },
  })
}
