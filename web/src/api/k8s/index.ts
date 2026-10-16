import { get, post } from "@/api/http"

import type { K8sProxyAuditItem } from "./types"

/** POST /v1/k8s/proxy */
export function proxyK8s(body: {
  agentUid: string
  method: string
  path: string
  query?: string
  headersJson?: string
  body?: string
}) {
  return post<{ statusCode: number; headersJson: string; body: string; audit: K8sProxyAuditItem }>("/k8s/proxy", body)
}

/** GET /v1/k8s-proxy-audits */
export function listK8sProxyAudits(params: { page?: number; pageSize?: number; agentUid?: string } = {}) {
  return get<{ items: K8sProxyAuditItem[]; total: number; page: number; pageSize: number }>("/k8s-proxy-audits", params)
}

export type { K8sProxyAuditItem } from "./types"
