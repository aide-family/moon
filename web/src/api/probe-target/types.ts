/** 探测类型：HTTP、TCP、ICMP、DNS、TLS；目标写法与观测指标各不相同。 */
export type ProbeDriver =
  | "PROBE_HTTP"
  | "PROBE_TCP"
  | "PROBE_ICMP"
  | "PROBE_DNS"
  | "PROBE_TLS"

/**
 * 组内地址共用的请求参数。HTTP 使用方法、请求头与断言字段；TCP 只使用 timeoutMs；
 * ICMP / DNS / TLS 使用各自的统计或解析字段，不适用的字段会被后端忽略。
 * headers / expectHeaders 是 JSON 对象文本（与 proto 一致），由页面序列化。
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
  /** HTTP 响应头断言，JSON 对象文本：响应头名 → 完整匹配正则。 */
  expectHeaders: string
  /** HTTP 响应体完整匹配正则；与 bodyKeyword 同时配置时两者都要满足。 */
  bodyRegex: string
  /** ICMP 单次探测发送的 echo 数量，0 按 1 处理，最多 10。 */
  icmpCount: number
  /** ICMP payload 字节数，0 按 56 处理，最多 1400。 */
  icmpPacketSize: number
  /** DNS 记录类型：A / AAAA / CNAME / MX / TXT / NS；留空按 A 处理。 */
  dnsRecordType: string
  /** DNS 期望值，逗号分隔的完整匹配正则；留空表示只要有记录即可。 */
  dnsExpectValues: string
  /** DNS 服务器 host:port；留空使用系统 resolver。 */
  dnsServer: string
  /** TLS SNI；留空取目标 host。 */
  tlsServerName: string
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
  /** HTTP 状态码；非 HTTP 探测为 0。 */
  statusCode: number
  errorKind: string
  errorMessage: string
  dnsMs: number
  tcpMs: number
  tlsMs: number
  ttfbMs: number
  /** TLS 证书剩余天数；HTTP / TLS 探测有意义。 */
  certDays: number
  resolvedIp: string
  slot: string
  probedAt: string
  executor: string
  /** ICMP 丢包百分比 0–100。 */
  packetLoss: number
  /** ICMP 抖动（相邻 RTT 平均绝对差，毫秒）。 */
  jitterMs: number
  /** ICMP 最小 RTT，毫秒。 */
  minRttMs: number
  /** ICMP 最大 RTT，毫秒。 */
  maxRttMs: number
  /** DNS 记录数。 */
  answerCount: number
  /** DNS 答案文本，逗号分隔。 */
  dnsAnswers: string
  /** TLS 协议版本，如 TLSv1.3。 */
  tlsVersion: string
  /** TLS 加密套件。 */
  tlsCipher: string
  /** 证书 Subject。 */
  certSubject: string
  /** 证书 Issuer。 */
  certIssuer: string
  /** 证书 SAN，逗号分隔。 */
  certSans: string
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
  /** 探测类型。留空表示后端返回租户已授权的全部类型。 */
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
  /** 探测类型，决定 spec 的写法与超时语义。 */
  driver: string
  spec: string
  /** 请求参数。按驱动使用对应字段，不适用的字段会被后端忽略。 */
  request?: ProbeRequestConfig
}

export interface ProbeTrialResult {
  result: ProbeResultItem
}
