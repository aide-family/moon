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
  /** 自定义标签，JSON 对象文本；"{}" 表示没有标签。 */
  labels: string
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
  /** 自定义标签，JSON 对象文本，例如 {"env":"prod"}。 */
  labels?: string
}

export interface UpdateAgentParams extends CreateAgentParams {
  uid: string
}

/** 按标签匹配机器：每个键值对都必须命中（AND 精确匹配）。 */
export interface MatchAgentsParams {
  labels: Record<string, string>
}

/** 命中的 ACTIVE 机器，online 由页面提示。 */
export interface MatchAgentsResult {
  items: AgentItem[]
  total: number
}
