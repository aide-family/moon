import {
  formatResultTimestamp,
  type ResultLabelPair,
  type ResultStatGroup,
  type ResultStatValue,
} from "./result-data"

export interface ElasticsearchHit {
  _index?: string
  _id?: string
  _score?: number
  _source?: Record<string, unknown>
  sort?: unknown[]
}

export interface ElasticsearchApiResponse {
  took?: number
  timed_out?: boolean
  hits?: {
    total?: number | { value?: number; relation?: string }
    hits?: ElasticsearchHit[]
  }
  aggregations?: Record<string, unknown>
  error?: unknown
  status?: number
}

export interface ElasticsearchLogRow {
  key: string
  index: string
  id: string
  timestamp: number
  source: Record<string, unknown>
  line: string
  labels: ResultLabelPair[]
}

/** 解析 Elasticsearch 的 JSON 响应。上游返回非 JSON 时返回 null，由页面展示原文。 */
export function parseElasticsearchBody(
  body: string
): ElasticsearchApiResponse | null {
  try {
    return JSON.parse(body) as ElasticsearchApiResponse
  } catch {
    return null
  }
}

/** 按点号路径读取 _source 中的字段，例如 event.time。 */
function lookupField(source: Record<string, unknown>, path: string): unknown {
  if (path in source) {
    return source[path]
  }
  let current: unknown = source
  for (const segment of path.split(".")) {
    if (
      typeof current !== "object" ||
      current === null ||
      Array.isArray(current)
    ) {
      return undefined
    }
    current = (current as Record<string, unknown>)[segment]
  }
  return current
}

/** Elasticsearch 的时间可能是 date 字符串、毫秒或秒级时间戳。统一转成毫秒。 */
function toMillis(value: unknown): number {
  if (typeof value === "number") {
    return value > 1e12 ? value : value * 1000
  }
  if (typeof value === "string") {
    const parsed = Date.parse(value)
    return Number.isNaN(parsed) ? 0 : parsed
  }
  return 0
}

function hitTimestamp(hit: ElasticsearchHit, timeField: string): number {
  const fromSource = toMillis(lookupField(hit._source ?? {}, timeField))
  if (fromSource > 0) {
    return fromSource
  }
  return toMillis(hit.sort?.[0])
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

function textValue(value: unknown): string {
  if (value == null) {
    return ""
  }
  if (typeof value === "string") {
    return value
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value)
  }
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}

// 常见日志索引的消息字段。不同采集器命名不统一，按优先级取第一个非空值；
// 都没有时页面回退到原有文档表格，而不是把整份 _source 当成日志正文。
const ELASTICSEARCH_MESSAGE_FIELDS = [
  "message",
  "msg",
  "log",
  "@message",
  "event.original",
  "body",
  "content",
  "text",
]

function findElasticsearchMessage(
  source: Record<string, unknown>,
  timeField: string
): { field: string; line: string } | null {
  for (const field of ELASTICSEARCH_MESSAGE_FIELDS) {
    if (field === timeField) {
      continue
    }
    const value = lookupField(source, field)
    if (value == null) {
      continue
    }
    const line = textValue(value)
    if (line.trim() === "") {
      continue
    }
    return { field, line }
  }
  return null
}

/** 命中文档里至少有一条能识别出日志正文时，才用折叠日志行展示。 */
export function elasticsearchHasLogMessage(
  response: ElasticsearchApiResponse | null,
  timeField = "@timestamp"
): boolean {
  return (response?.hits?.hits ?? []).some(
    (hit) =>
      findElasticsearchMessage(hit._source ?? {}, timeField || "@timestamp") !=
      null
  )
}

function flattenSource(
  value: unknown,
  key: string,
  out: ResultLabelPair[]
): void {
  if (Array.isArray(value)) {
    out.push({ key, value: JSON.stringify(value) })
    return
  }
  if (isRecord(value)) {
    const entries = Object.entries(value)
    if (entries.length === 0) {
      out.push({ key, value: "{}" })
      return
    }
    for (const [childKey, childValue] of entries) {
      flattenSource(childValue, key ? `${key}.${childKey}` : childKey, out)
    }
    return
  }
  out.push({ key, value: textValue(value) })
}

function elasticsearchLogLabels(
  hit: ElasticsearchHit,
  timeField: string,
  messageField?: string
): ResultLabelPair[] {
  const labels: ResultLabelPair[] = []
  if (hit._index) {
    labels.push({ key: "_index", value: hit._index })
  }
  if (hit._id) {
    labels.push({ key: "_id", value: hit._id })
  }
  if (typeof hit._score === "number" && Number.isFinite(hit._score)) {
    labels.push({ key: "_score", value: String(hit._score) })
  }
  const sourceLabels: ResultLabelPair[] = []
  flattenSource(hit._source ?? {}, "", sourceLabels)
  for (const pair of sourceLabels) {
    if (pair.key === timeField || pair.key === messageField) {
      continue
    }
    labels.push(pair)
  }
  return labels
}

/** 把命中的文档摊平成日志行。direction 决定时间顺序，和查询参数保持一致。 */
export function toElasticsearchRows(
  response: ElasticsearchApiResponse | null,
  direction: "backward" | "forward" = "backward",
  timeField = "@timestamp"
): ElasticsearchLogRow[] {
  if (!response || response.error) {
    return []
  }
  const resolvedTimeField = timeField || "@timestamp"
  const rows = (response.hits?.hits ?? []).map((hit, index) => {
    const source = hit._source ?? {}
    const message = findElasticsearchMessage(source, resolvedTimeField)
    return {
      key: `${hit._index ?? ""}-${hit._id ?? index}`,
      index: hit._index ?? "",
      id: hit._id ?? "",
      timestamp: hitTimestamp(hit, resolvedTimeField),
      source,
      line: message?.line ?? "",
      labels: elasticsearchLogLabels(hit, resolvedTimeField, message?.field),
    }
  })
  rows.sort((left, right) =>
    direction === "backward"
      ? right.timestamp - left.timestamp
      : left.timestamp - right.timestamp
  )
  return rows
}

/** JSON 页签只展示命中文档；错误响应保留 error 与 status。 */
export function extractElasticsearchResult(
  response: ElasticsearchApiResponse | null
): unknown {
  if (!response) {
    return null
  }
  if (response.error) {
    return { error: response.error, status: response.status }
  }
  return response.hits?.hits ?? response
}

export type ElasticsearchResultView = "logs" | "value" | "table"

export interface ElasticsearchResultStat {
  fields: string[]
  groups: ResultStatGroup[]
}

type NumericLeaf = ResultStatValue

interface BucketAggregation {
  name: string
  buckets: Record<string, unknown>[]
}

// 这些键是聚合的元数据或嵌套桶，不能当成数值列。
const AGGREGATION_SKIP_KEYS = new Set([
  "buckets",
  "hits",
  "meta",
  "key",
  "key_as_string",
  "from",
  "from_as_string",
  "to",
  "to_as_string",
  "doc_count",
  "doc_count_error_upper_bound",
  "sum_other_doc_count",
  "value_as_string",
])

function collectMetricLeaves(
  value: unknown,
  path: string[],
  out: NumericLeaf[],
  depth = 0
): void {
  if (!isRecord(value) || depth > 6) {
    return
  }
  const numericValue = value.value
  if (typeof numericValue === "number" && Number.isFinite(numericValue)) {
    const formatted = value.value_as_string
    out.push({
      field: path.join("."),
      raw: typeof formatted === "string" ? formatted : String(numericValue),
      value: numericValue,
    })
    return
  }
  for (const [key, child] of Object.entries(value)) {
    if (AGGREGATION_SKIP_KEYS.has(key)) {
      continue
    }
    if (typeof child === "number" && Number.isFinite(child)) {
      out.push({
        field: [...path, key].join("."),
        raw: String(child),
        value: child,
      })
      continue
    }
    if (isRecord(child)) {
      collectMetricLeaves(child, [...path, key], out, depth + 1)
    }
  }
}

function findBucketAggregations(
  value: unknown,
  path: string[],
  out: BucketAggregation[],
  depth = 0
): void {
  if (!isRecord(value) || depth > 6) {
    return
  }
  if (Array.isArray(value.buckets)) {
    const buckets = value.buckets.filter(isRecord)
    if (buckets.length > 0) {
      out.push({ name: path.join(".") || "buckets", buckets })
    }
    return
  }
  for (const [key, child] of Object.entries(value)) {
    if (isRecord(child)) {
      findBucketAggregations(child, [...path, key], out, depth + 1)
    }
  }
}

function rangeEndpoint(value: unknown, fallback: string): string {
  if (typeof value === "string" && value) {
    const timestamp = Date.parse(value)
    return Number.isFinite(timestamp) ? formatResultTimestamp(timestamp) : value
  }
  if (typeof value === "number" && Number.isFinite(value)) {
    return Math.abs(value) >= 1e11
      ? formatResultTimestamp(value)
      : String(value)
  }
  return fallback
}

function bucketLabel(bucket: Record<string, unknown>, index: number): string {
  const asString = bucket.key_as_string
  if (typeof asString === "string" && asString) {
    const timestamp = Date.parse(asString)
    return Number.isFinite(timestamp)
      ? formatResultTimestamp(timestamp)
      : asString
  }
  const key = bucket.key
  if (typeof key === "number" && Number.isFinite(key)) {
    // date_histogram 的 key 是毫秒时间戳；terms 的数字 key 通常远小于该阈值。
    return Math.abs(key) >= 1e11 ? formatResultTimestamp(key) : String(key)
  }
  if (typeof key === "string" || typeof key === "boolean") {
    return String(key)
  }
  if (bucket.from !== undefined || bucket.to !== undefined) {
    const from = rangeEndpoint(bucket.from_as_string ?? bucket.from, "*")
    const to = rangeEndpoint(bucket.to_as_string ?? bucket.to, "*")
    return `${from} – ${to}`
  }
  return `#${index + 1}`
}

function bucketLeaves(bucket: Record<string, unknown>): NumericLeaf[] {
  const leaves: NumericLeaf[] = []
  const docCount = bucket.doc_count
  if (typeof docCount === "number" && Number.isFinite(docCount)) {
    leaves.push({ field: "doc_count", raw: String(docCount), value: docCount })
  }
  for (const [key, child] of Object.entries(bucket)) {
    if (key === "doc_count" || AGGREGATION_SKIP_KEYS.has(key)) {
      continue
    }
    if (isRecord(child)) {
      collectMetricLeaves(child, [key], leaves)
    }
  }
  return leaves
}

function bucketAggregationStat(
  aggregation: BucketAggregation
): ElasticsearchResultStat | null {
  const leavesByBucket = aggregation.buckets.map(bucketLeaves)
  const fields: string[] = []
  for (const leaves of leavesByBucket) {
    for (const leaf of leaves) {
      if (!fields.includes(leaf.field)) {
        fields.push(leaf.field)
      }
    }
  }
  if (fields.length === 0) {
    return null
  }
  const groups: ResultStatGroup[] = aggregation.buckets.map((bucket, index) => {
    const leaves = leavesByBucket[index]
    const byField = new Map(leaves.map((leaf) => [leaf.field, leaf]))
    return {
      key: `${aggregation.name}-${index}`,
      label: bucketLabel(bucket, index),
      values: fields.map((field) => {
        const leaf = byField.get(field)
        return leaf ?? { field, raw: "", value: Number.NaN }
      }),
    }
  })
  return { fields, groups }
}

function metricAggregationStat(
  aggregations: Record<string, unknown>
): ElasticsearchResultStat | null {
  const leaves: NumericLeaf[] = []
  for (const [name, value] of Object.entries(aggregations)) {
    if (isRecord(value) && Array.isArray(value.buckets)) {
      continue
    }
    collectMetricLeaves(value, [name], leaves)
  }
  if (leaves.length === 0) {
    return null
  }
  const fields: string[] = []
  const values: ResultStatValue[] = []
  for (const leaf of leaves) {
    if (fields.includes(leaf.field)) {
      continue
    }
    fields.push(leaf.field)
    values.push(leaf)
  }
  return {
    fields,
    groups: [{ key: "aggregation", label: "#1", values }],
  }
}

/**
 * 把 date_histogram / terms 等桶聚合摊平成柱状图分组：每个桶一组，
 * 组内是 doc_count 和桶下数值子聚合；没有桶时按顶层指标聚合生成单组。
 */
export function elasticsearchAggregationStat(
  response: ElasticsearchApiResponse | null
): ElasticsearchResultStat | null {
  if (!response || response.error || !isRecord(response.aggregations)) {
    return null
  }
  const aggregations = response.aggregations
  const bucketAggregations: BucketAggregation[] = []
  findBucketAggregations(aggregations, [], bucketAggregations)
  let best: ElasticsearchResultStat | null = null
  for (const aggregation of bucketAggregations) {
    const parsed = bucketAggregationStat(aggregation)
    if (parsed && (!best || parsed.fields.length > best.fields.length)) {
      best = parsed
    }
  }
  return best ?? metricAggregationStat(aggregations)
}

/** hits 展示折叠日志行；聚合有数值时展示单值或柱状图；其余保留原文档表格。 */
export function elasticsearchResultView(
  response: ElasticsearchApiResponse | null,
  timeField = "@timestamp"
): ElasticsearchResultView {
  if (!response) {
    return "table"
  }
  if (elasticsearchAggregationStat(response)) {
    return "value"
  }
  const hits = response.hits?.hits ?? []
  if (hits.length > 0) {
    return elasticsearchHasLogMessage(response, timeField) ? "logs" : "table"
  }
  return response.hits ? "logs" : "table"
}
