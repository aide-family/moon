export interface QueryLogsParams {
  datasourceUid: string
  query: string
  start: number
  end: number
  limit: number
  direction: "backward" | "forward"
}
