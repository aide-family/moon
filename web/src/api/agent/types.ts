export interface AgentItem {
  uid: string
  name: string
  hostname: string
  machineId: string
  localIp: string
  version: string
  online: boolean
  status: string
  sshHost: string
  sshPort: number
  sshUser: string
  hasPassword: boolean
  hasPrivateKey: boolean
  remark: string
  os: string
  arch: string
  osVersion: string
  cpu: string
  disk: string
  memory: string
  bootTime: string
  timeServer: string
}

export interface ListAgentsParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListAgentsResult {
  items: AgentItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateAgentParams {
  name: string
  sshHost?: string
  sshPort?: number
  sshUser?: string
  sshPassword?: string
  sshPrivateKey?: string
  remark?: string
}

export interface UpdateAgentParams extends CreateAgentParams {
  uid: string
}
