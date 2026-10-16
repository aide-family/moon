import { del, get, post, put } from "@/api/http"

import type {
  CreateSSHCommandParams,
  ListSSHCommandAuditsParams,
  ListSSHCommandAuditsResult,
  ListSSHCommandsParams,
  ListSSHCommandsResult,
  SSHCommandAuditItem,
  SSHCommandItem,
  UpdateSSHCommandParams,
} from "./types"

/** GET /v1/ssh-commands */
export function listSSHCommands(params: ListSSHCommandsParams = {}): Promise<ListSSHCommandsResult> {
  return get<ListSSHCommandsResult>("/ssh-commands", params)
}

/** POST /v1/ssh-command */
export function createSSHCommand(body: CreateSSHCommandParams): Promise<{ command: SSHCommandItem }> {
  return post<{ command: SSHCommandItem }>("/ssh-command", body)
}

/** PUT /v1/ssh-command/{uid} */
export function updateSSHCommand(body: UpdateSSHCommandParams): Promise<{ command: SSHCommandItem }> {
  return put<{ command: SSHCommandItem }>(`/ssh-command/${body.uid}`, body)
}

/** POST /v1/ssh-command/{uid}/approve */
export function approveSSHCommand(uid: string): Promise<{ command: SSHCommandItem }> {
  return post<{ command: SSHCommandItem }>(`/ssh-command/${uid}/approve`, {})
}

/** POST /v1/ssh-command/{uid}/reject */
export function rejectSSHCommand(uid: string): Promise<{ command: SSHCommandItem }> {
  return post<{ command: SSHCommandItem }>(`/ssh-command/${uid}/reject`, {})
}

/** DELETE /v1/ssh-command/{uid} */
export function deleteSSHCommand(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/ssh-command/${uid}`)
}

/** GET /v1/ssh-command/{uid} */
export function getSSHCommand(uid: string): Promise<{ command: SSHCommandItem }> {
  return get<{ command: SSHCommandItem }>(`/ssh-command/${uid}`)
}

/** POST /v1/ssh-command/{uid}/execute */
export function executeSSHCommand(uid: string, agentUid: string): Promise<{ audit: SSHCommandAuditItem }> {
  return post<{ audit: SSHCommandAuditItem }>(`/ssh-command/${uid}/execute`, { uid, agentUid })
}

/** GET /v1/ssh-command-audits */
export function listSSHCommandAudits(params: ListSSHCommandAuditsParams = {}): Promise<ListSSHCommandAuditsResult> {
  return get<ListSSHCommandAuditsResult>("/ssh-command-audits", params)
}

export type {
  CreateSSHCommandParams,
  ListSSHCommandAuditsParams,
  ListSSHCommandAuditsResult,
  ListSSHCommandsParams,
  ListSSHCommandsResult,
  SSHCommandAuditItem,
  SSHCommandItem,
  UpdateSSHCommandParams,
} from "./types"
