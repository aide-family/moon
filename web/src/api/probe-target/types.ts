/** 探测类型。一期只有 PROBE_HTTP；TCP / ICMP 在后续版本提供。 */
export type ProbeDriver = "PROBE_HTTP"

/**
 * 组内地址共用的 HTTP 请求参数。
 * headers 是 JSON 对象文本（与 proto 一致），由页面序列化。
 */
export interface ProbeRequestConfig {
  /** 请求方法，GET 或 HEAD。 */
  method: string
  /** 期望状态码表达式：200、200-299、200,204,301；留空按 200-299。 */
  expectStatus: string
  /** 单次超时毫秒数，100..60000。 */
  timeoutMs: number
  /** 是否跟随重定向，最多 5 跳。 */
  followRedirects: boolean
  /** 是否跳过 TLS 证书校验。 */
  skipTlsVerify: boolean
  /** 是否校验证书有效期并计算剩余天数。 */
  verifyCert: boolean
  /** 自定义请求头，JSON 对象文本。 */
  headers: string
  /** 响应体关键字，命中即成功；留空只按状态码判定。 */
  bodyKeyword: string
}

/** 目标组里的一个地址及其最近一次探测结果。 */
export interface ProbeTargetItem {
  uid: string
  spec: string
  labels: string
  status: string
  sort: number
  /** 最近一次探测结果；还没有探测记录时为空。 */
  lastResult?: ProbeResultItem
}

/** 一条探测记录。试探测结果的 uid/ruleUid/slot 为空。 */
export interface ProbeResultItem {
  targetUid: string
  targetSpec: string
  ruleUid: string
  ruleName: string
  ok: boolean
  value: string
  durationMs: number
  statusCode: number
  errorKind: string
  errorMessage: string
  dnsMs: number
  tcpMs: number
  tlsMs: number
  ttfbMs: number
  certDays: number
  resolvedIp: string
  slot: string
  probedAt: string
  executor: string
}

/** 探测目标组。 */
export interface ProbeTargetGroupItem {
  uid: string
  name: string
  driver: ProbeDriver | string
  status: string
  labels: string
  /** 执行器标签 JSON，空对象表示集群内执行。 */
  executorLabels: string
  request: ProbeRequestConfig
  remark: string
  targets: ProbeTargetItem[]
  targetCount: number
  activeTargetCount: number
  /** 最近一次探测时间，RFC3339；还没有探测过时为空。 */
  lastProbedAt: string
}

export interface ListProbeTargetGroupsParams {
  page?: number
  pageSize?: number
  name?: string
  /** 探测类型。留空表示全部。 */
  driver?: string
}

export interface ListProbeTargetGroupsResult {
  items: ProbeTargetGroupItem[]
  total: number
  page: number
  pageSize: number
}

/** 保存目标组时提交的目标；uid 为空表示新增。 */
export interface ProbeTargetInput {
  uid?: string
  spec: string
  labels?: string
  status?: string
  sort?: number
}

export interface CreateProbeTargetGroupParams {
  name: string
  driver: string
  labels?: string
  executorLabels?: string
  request: ProbeRequestConfig
  remark?: string
  targets: ProbeTargetInput[]
}

export interface UpdateProbeTargetGroupParams
  extends CreateProbeTargetGroupParams {
  uid: string
}

export interface ListProbeTargetResultsParams {
  /** 目标标识。留空表示组内全部目标。 */
  targetUid?: string
  page?: number
  pageSize?: number
}

export interface ListProbeTargetResultsResult {
  items: ProbeResultItem[]
  total: number
  page: number
  pageSize: number
}

/** 试探测参数：不落库、不判定、不产生事件。 */
export interface ProbeTrialParams {
  spec: string
  request?: ProbeRequestConfig
}

export interface ProbeTrialResult {
  result: ProbeResultItem
}
