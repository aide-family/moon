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
  operatorName: string
  operatorNickname: string
  operatorAvatar: string
  kind: string
  status: string
  command: string
  exitCode: number
  output: string
  truncated: boolean
  detail: string
  createdAt: string
  /** 机器展示名；机器被删除后为空。 */
  agentName: string
  /** 命令名称；命令被删除后为空。 */
  commandName: string
}

export interface ListSSHCommandsParams {
  page?: number
  pageSize?: number
  name?: string
  reviewStatus?: string
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
  /** 审计种类：CREATE / UPDATE / APPROVE / REJECT / EXECUTE；留空表示全部。 */
  kind?: string
  /** 状态：PENDING / APPROVED / REJECTED / DONE / FAILED；留空表示全部。 */
  status?: string
  /** 操作者标识。 */
  operatorUid?: string
  /** 命令正文关键字，大小写不敏感的子串匹配。 */
  keyword?: string
  /** 创建时间下界，Unix 秒，含；0 表示不限。 */
  startTime?: number
  /** 创建时间上界，Unix 秒，不含；0 表示不限。 */
  endTime?: number
}

export interface ListSSHCommandAuditsResult {
  items: SSHCommandAuditItem[]
  total: number
  page: number
  pageSize: number
}
