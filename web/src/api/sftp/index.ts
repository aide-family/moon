import { get, post } from "@/api/http"

import type { SFTPAuditItem, SFTPEntry } from "./types"

/** POST /v1/sftp/list */
export function listSFTP(body: { agentUid: string; path: string }) {
  return post<{ entries: SFTPEntry[]; audit: SFTPAuditItem }>("/sftp/list", body)
}

/** POST /v1/sftp/download */
export function downloadSFTP(body: { agentUid: string; path: string }) {
  return post<{ content: string; audit: SFTPAuditItem }>("/sftp/download", body)
}

/** POST /v1/sftp/upload */
export function uploadSFTP(body: { agentUid: string; path: string; content: string }) {
  return post<{ audit: SFTPAuditItem }>("/sftp/upload", body)
}

/** POST /v1/sftp/mkdir */
export function mkdirSFTP(body: { agentUid: string; path: string }) {
  return post<{ audit: SFTPAuditItem }>("/sftp/mkdir", body)
}

/** POST /v1/sftp/delete */
export function deleteSFTP(body: { agentUid: string; path: string }) {
  return post<{ audit: SFTPAuditItem }>("/sftp/delete", body)
}

/** GET /v1/sftp-audits */
export function listSFTPAudits(params: { page?: number; pageSize?: number; agentUid?: string } = {}) {
  return get<{ items: SFTPAuditItem[]; total: number; page: number; pageSize: number }>("/sftp-audits", params)
}

export type { SFTPAuditItem, SFTPEntry } from "./types"
