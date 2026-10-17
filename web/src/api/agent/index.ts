import { del, get, post, put } from "@/api/http"

import type {
  AgentItem,
  CreateAgentParams,
  ListAgentsParams,
  ListAgentsResult,
  UpdateAgentParams,
} from "./types"

/** GET /v1/agents */
export function listAgents(params: ListAgentsParams = {}): Promise<ListAgentsResult> {
  return get<ListAgentsResult>("/agents", params)
}

/** POST /v1/agent */
export function createAgent(body: CreateAgentParams): Promise<{ agent: AgentItem; token: string }> {
  return post<{ agent: AgentItem; token: string }>("/agent", body)
}

/** PUT /v1/agent/{uid} */
export function updateAgent(body: UpdateAgentParams): Promise<{ agent: AgentItem }> {
  return put<{ agent: AgentItem }>(`/agent/${body.uid}`, body)
}

/** POST /v1/agent/{uid}/token */
export function rotateAgentToken(uid: string): Promise<{ agent: AgentItem; token: string }> {
  return post<{ agent: AgentItem; token: string }>(`/agent/${uid}/token`, {})
}

/** POST /v1/agent/{uid}/revoke */
export function revokeAgent(uid: string): Promise<{ agent: AgentItem }> {
  return post<{ agent: AgentItem }>(`/agent/${uid}/revoke`, {})
}

/** DELETE /v1/agent/{uid} */
export function deleteAgent(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/agent/${uid}`)
}

/** GET /v1/agent/{uid} */
export function getAgent(uid: string): Promise<{ agent: AgentItem }> {
  return get<{ agent: AgentItem }>(`/agent/${uid}`)
}

/** POST /v1/agent/{uid}/shell-ticket */
export function issueShellTicket(uid: string): Promise<{ ticket: string }> {
  return post<{ ticket: string }>(`/agent/${uid}/shell-ticket`, {})
}

export type {
  AgentItem,
  CreateAgentParams,
  ListAgentsParams,
  ListAgentsResult,
  UpdateAgentParams,
} from "./types"
