export interface WorkerInstanceItem {
  service: "MEGREZ" | "MERAK" | string
  instanceUid: string
  hostname: string
  busy: boolean
  taskUid: string
  taskName: string
  taskStartedAt: string
  lastSeenAt: string
  online: boolean
}

export interface ListWorkerInstancesResult {
  items: WorkerInstanceItem[]
  onlineCount: number
  busyCount: number
  idleCount: number
  observedAt: string
}
