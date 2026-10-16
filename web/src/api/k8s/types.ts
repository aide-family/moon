export interface K8sProxyAuditItem {
  uid: string
  agentUid: string
  operatorUid: string
  method: string
  resource: string
  kubeNamespace: string
  result: string
  detail: string
  createdAt: string
}
