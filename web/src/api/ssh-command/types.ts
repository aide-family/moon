export interface SSHCommandItem {
  uid: string
  name: string
  command: string
  workdir: string
  env: string
  reviewStatus: string
  remark: string
}

export interface SSHCommandAuditItem {
  uid: string
  commandUid: string
  agentUid: string
  operatorUid: string
  kind: string
  status: string
  command: string
  exitCode: number
  output: string
  truncated: boolean
  detail: string
  createdAt: string
}

export interface ListSSHCommandsParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListSSHCommandsResult {
  items: SSHCommandItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateSSHCommandParams {
  name: string
  command: string
  workdir?: string
  env?: string
  remark?: string
}

export interface UpdateSSHCommandParams extends CreateSSHCommandParams {
  uid: string
}

export interface ListSSHCommandAuditsParams {
  page?: number
  pageSize?: number
  commandUid?: string
  agentUid?: string
}

export interface ListSSHCommandAuditsResult {
  items: SSHCommandAuditItem[]
  total: number
  page: number
  pageSize: number
}
