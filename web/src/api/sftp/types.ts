export interface SFTPEntry {
  name: string
  path: string
  isDir: boolean
  size: number
  mode: string
  modTime: string
}

export interface SFTPAuditItem {
  uid: string
  agentUid: string
  operatorUid: string
  path: string
  action: string
  result: string
  detail: string
  createdAt: string
}
