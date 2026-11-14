import {
  ChevronDown,
  ChevronRight,
  EllipsisVertical,
  Search,
  Trash2,
} from "lucide-react"
import { Fragment, useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import {
  createProbeTargetGroup,
  deleteProbeTargetGroup,
  getProbeTargetGroup,
  listProbeTargetGroups,
  listProbeTargetResults,
  probeTargetGroupNow,
  probeTrial,
  updateProbeTargetGroup,
  updateProbeTargetGroupStatus,
} from "@/api/probe-target"
import type {
  CreateProbeTargetGroupParams,
  ProbeRequestConfig,
  ProbeResultItem,
  ProbeTargetGroupItem,
  ProbeTargetInput,
} from "@/api/probe-target/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { FormField } from "@/components/form-field"
import { LabelTags } from "@/components/label-tags"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { StepForm } from "@/components/step-form"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Field, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Textarea } from "@/components/ui/textarea"
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import { enumLabel } from "@/lib/enum-label"
import { authorizedProbeDrivers } from "@/lib/feature-catalog"
import {
  collectFieldErrors,
  hasFieldErrors,
  maxLength,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { readFormError } from "@/lib/form-error"
import { parseLabelPairs, serializeLabelPairs, type LabelPair } from "@/lib/labels"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

// 探测类型。页面按套餐授权展示可选类型，五种类型的目标写法与判定口径不同。
const PROBE_HTTP = "PROBE_HTTP"
const PROBE_TCP = "PROBE_TCP"
const PROBE_ICMP = "PROBE_ICMP"
const PROBE_DNS = "PROBE_DNS"
const PROBE_TLS = "PROBE_TLS"
const dnsRecordTypes = ["A", "AAAA", "CNAME", "MX", "TXT", "NS"]
const MAX_TARGETS = 50
const MAX_TARGET_SPEC = 2048
const MAX_TIMEOUT_MS = 60000
const MIN_TIMEOUT_MS = 100
const MAX_BODY_KEYWORD = 256
const MAX_BODY_REGEX = 512
const MAX_EXPECT_HEADERS = 4096
const MAX_DNS_EXPECT_VALUES = 1024
const MAX_DNS_SERVER = 255
const MAX_TLS_SERVER_NAME = 255
const MAX_LABELS = 4096
const MAX_HEADERS = 8192
// 与 pkg/probe 的受控头清单一致：由客户端计算，探测不能覆盖。
const controlledHeaders = new Set([
  "host",
  "content-length",
  "transfer-encoding",
  "connection",
  "upgrade",
])

const defaultRequest: ProbeRequestConfig = {
  method: "GET",
  expectStatus: "200-299",
  timeoutMs: 3000,
  followRedirects: true,
  skipTlsVerify: false,
  verifyCert: true,
  headers: "{}",
  bodyKeyword: "",
  expectHeaders: "{}",
  bodyRegex: "",
  icmpCount: 1,
  icmpPacketSize: 56,
  dnsRecordType: "A",
  dnsExpectValues: "",
  dnsServer: "",
  tlsServerName: "",
}

/** 目标行草稿：uid 为空表示新增，labels 保留原值避免更新时被清空。 */
interface TargetDraft {
  uid: string
  spec: string
  labels: string
  enabled: boolean
}

interface ProbeDraft {
  name: string
  driver: string
  labels: LabelPair[]
  remark: string
  targets: TargetDraft[]
  method: string
  expectStatus: string
  timeoutMs: string
  followRedirects: boolean
  skipTlsVerify: boolean
  verifyCert: boolean
  bodyKeyword: string
  headers: LabelPair[]
  expectHeaders: LabelPair[]
  bodyRegex: string
  icmpCount: string
  icmpPacketSize: string
  dnsRecordType: string
  dnsExpectValues: string
  dnsServer: string
  tlsServerName: string
}

interface TrialEntry {
  spec: string
  result: ProbeResultItem | null
}

function emptyTarget(): TargetDraft {
  return { uid: "", spec: "", labels: "{}", enabled: true }
}

function emptyDraft(): ProbeDraft {
  return {
    name: "",
    driver: PROBE_HTTP,
    labels: [{ key: "", value: "" }],
    remark: "",
    targets: [emptyTarget()],
    method: defaultRequest.method,
    expectStatus: defaultRequest.expectStatus,
    timeoutMs: String(defaultRequest.timeoutMs),
    followRedirects: defaultRequest.followRedirects,
    skipTlsVerify: defaultRequest.skipTlsVerify,
    verifyCert: defaultRequest.verifyCert,
    bodyKeyword: defaultRequest.bodyKeyword,
    headers: [],
    expectHeaders: [],
    bodyRegex: defaultRequest.bodyRegex,
    icmpCount: String(defaultRequest.icmpCount),
    icmpPacketSize: String(defaultRequest.icmpPacketSize),
    dnsRecordType: defaultRequest.dnsRecordType,
    dnsExpectValues: defaultRequest.dnsExpectValues,
    dnsServer: defaultRequest.dnsServer,
    tlsServerName: defaultRequest.tlsServerName,
  }
}

function draftFromItem(item: ProbeTargetGroupItem): ProbeDraft {
  const request = { ...defaultRequest, ...(item.request ?? {}) }
  return {
    name: item.name,
    driver: item.driver || PROBE_HTTP,
    labels: parseLabelPairs(item.labels),
    remark: item.remark ?? "",
    targets: (item.targets ?? [])
      .slice()
      .sort((left, right) => left.sort - right.sort)
      .map((target) => ({
        uid: target.uid,
        spec: target.spec,
        labels: target.labels || "{}",
        enabled: target.status !== "DISABLED",
      })),
    method: request.method || defaultRequest.method,
    expectStatus: request.expectStatus ?? defaultRequest.expectStatus,
    timeoutMs: String(request.timeoutMs || defaultRequest.timeoutMs),
    followRedirects: Boolean(request.followRedirects),
    skipTlsVerify: Boolean(request.skipTlsVerify),
    verifyCert: Boolean(request.verifyCert),
    bodyKeyword: request.bodyKeyword ?? "",
    headers: parseHeaderPairs(request.headers),
    expectHeaders: parseHeaderPairs(request.expectHeaders),
    bodyRegex: request.bodyRegex ?? "",
    // 0 在后端按 1 / 56 处理，编辑时直接展示生效值。
    icmpCount: String(request.icmpCount || defaultRequest.icmpCount),
    icmpPacketSize: String(
      request.icmpPacketSize || defaultRequest.icmpPacketSize
    ),
    dnsRecordType: request.dnsRecordType || defaultRequest.dnsRecordType,
    dnsExpectValues: request.dnsExpectValues ?? "",
    dnsServer: request.dnsServer ?? "",
    tlsServerName: request.tlsServerName ?? "",
  }
}

function parseHeaderPairs(raw: string): LabelPair[] {
  try {
    const parsed: unknown = JSON.parse(raw || "{}")
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return []
    }
    return Object.entries(parsed as Record<string, unknown>).map(
      ([key, value]) => ({ key, value: String(value) })
    )
  } catch {
    return []
  }
}

function serializeHeaders(pairs: LabelPair[]): string {
  const headers: Record<string, string> = {}
  for (const pair of pairs) {
    const key = pair.key.trim()
    if (key === "") {
      continue
    }
    headers[key] = pair.value.trim()
  }
  return JSON.stringify(headers)
}

/** 只接受带 host 的 http / https 地址，与 pkg/probe.NormalizeSpec 的要求一致。 */
function isHttpTarget(value: string): boolean {
  try {
    const url = new URL(value.trim())
    return (
      (url.protocol === "http:" || url.protocol === "https:") &&
      url.hostname !== ""
    )
  } catch {
    return false
  }
}

/** 组内 HTTP 去重键：scheme/host 大小写归一，路径与查询串保持原样。 */
function httpTargetKey(value: string): string {
  try {
    const url = new URL(value.trim())
    const port = url.port ? `:${url.port}` : ""
    return `${url.protocol}//${url.hostname.toLowerCase()}${port}${url.pathname}${url.search}`
  } catch {
    return value.trim()
  }
}

/** IP 字面量：IPv4 校验每段（拒绝前导零），IPv6 交给 URL 解析，避免漏掉压缩写法。 */
function isIpLiteral(value: string): boolean {
  if (/^\d{1,3}(\.\d{1,3}){3}$/.test(value)) {
    return value
      .split(".")
      .every((part) => /^(0|[1-9]\d{0,2})$/.test(part) && Number(part) <= 255)
  }
  if (!value.includes(":") || !/^[0-9a-fA-F:]+$/.test(value)) {
    return false
  }
  try {
    new URL(`http://[${value}]/`)
    return true
  } catch {
    return false
  }
}

/**
 * TCP 目标必须是 host:port：不接受 scheme、路径与查询串，端口 1..65535；
 * IPv6 字面量必须带方括号，与 pkg/probe.normalizeTCPTarget 的规则一致。
 */
function splitTCPTarget(value: string): { host: string; port: number } | null {
  const text = value.trim()
  if (text === "" || text.includes("://") || /[/?#]/.test(text)) {
    return null
  }
  let host: string
  let portText: string
  if (text.startsWith("[")) {
    const closing = text.indexOf("]:")
    if (closing <= 1) {
      return null
    }
    host = text.slice(1, closing)
    portText = text.slice(closing + 2)
    if (!isIpLiteral(host)) {
      return null
    }
  } else {
    const parts = text.split(":")
    if (parts.length !== 2) {
      return null
    }
    host = parts[0]
    portText = parts[1]
  }
  const trimmedHost = host.trim()
  // 主机里出现空白、@ 或方括号说明不是 host:port（IPv6 方括号已在上面的分支剥掉）。
  if (
    trimmedHost === "" ||
    /[\s@]/.test(trimmedHost) ||
    trimmedHost.includes("[") ||
    trimmedHost.includes("]") ||
    !/^\d{1,5}$/.test(portText)
  ) {
    return null
  }
  const port = Number(portText)
  if (port < 1 || port > 65535) {
    return null
  }
  return { host: trimmedHost, port }
}

function tcpTargetKey(value: string): string {
  const parsed = splitTCPTarget(value)
  return parsed ? `${parsed.host.toLowerCase()}:${parsed.port}` : value.trim()
}

/**
 * ICMP 目标只接受主机名或 IP：拒绝 scheme、路径、端口；IPv6 字面量可带方括号，
 * 与 pkg/probe.normalizeICMPTarget 的规则一致。
 */
function isICMPTarget(value: string): boolean {
  let text = value.trim()
  if (text === "" || text.includes("://") || /[/?#]/.test(text)) {
    return false
  }
  if (text.startsWith("[") && text.endsWith("]")) {
    text = text.slice(1, -1)
  }
  if (isIpLiteral(text)) {
    return true
  }
  // 带冒号却不是 IP，通常是误填了 host:port。
  return !text.includes(":")
}

function icmpTargetKey(value: string): string {
  let text = value.trim()
  if (text.startsWith("[") && text.endsWith("]")) {
    text = text.slice(1, -1)
  }
  return text.toLowerCase()
}

// 单个 DNS 标签：字母/数字/连字符，不以连字符开头或结尾，1–63 字符。
const dnsLabelPattern = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/

/** 归一化 DNS 域名：转小写、去末尾点；返回 null 表示非法，与后端规则一致。 */
function normalizeDnsTarget(value: string): string | null {
  let text = value.trim().toLowerCase()
  if (text.endsWith(".")) {
    text = text.slice(0, -1)
  }
  if (text === "" || text.length > 253) {
    return null
  }
  // scheme、端口、路径、空白与邮箱式写法都不是合法域名。
  if (/[\s@:/?#]/.test(text)) {
    return null
  }
  // IP 字面量（含保留网段）不算域名。
  if (isIpLiteral(text)) {
    return null
  }
  if (!text.split(".").every((label) => dnsLabelPattern.test(label))) {
    return null
  }
  return text
}

function isDnsTarget(value: string): boolean {
  return normalizeDnsTarget(value) !== null
}

function dnsTargetKey(value: string): string {
  return normalizeDnsTarget(value) ?? value.trim().toLowerCase()
}

/** 按探测类型校验目标写法。 */
function isTargetForDriver(driver: string, value: string): boolean {
  if (driver === PROBE_TCP || driver === PROBE_TLS) {
    const parsed = splitTCPTarget(value)
    return parsed !== null && !isBlockedProbeHost(parsed.host)
  }
  if (driver === PROBE_ICMP) {
    return isICMPTarget(value)
  }
  if (driver === PROBE_DNS) {
    return isDnsTarget(value)
  }
  return isHttpTarget(value)
}

/** 按探测类型生成组内去重键。 */
function targetKey(driver: string, value: string): string {
  if (driver === PROBE_TCP || driver === PROBE_TLS) {
    return tcpTargetKey(value)
  }
  if (driver === PROBE_ICMP) {
    return icmpTargetKey(value)
  }
  if (driver === PROBE_DNS) {
    return dnsTargetKey(value)
  }
  return httpTargetKey(value)
}

/** 驱动对应的文案后缀；未知驱动按 HTTP 处理，避免拼出空翻译 key。 */
function driverCopySuffix(
  driver: string
): "Http" | "Tcp" | "Icmp" | "Dns" | "Tls" {
  if (driver === PROBE_TCP) {
    return "Tcp"
  }
  if (driver === PROBE_ICMP) {
    return "Icmp"
  }
  if (driver === PROBE_DNS) {
    return "Dns"
  }
  if (driver === PROBE_TLS) {
    return "Tls"
  }
  return "Http"
}

/** 保留网段地址与后端 pkg/probe.BlockedHost 对齐（元数据、链路本地、组播、未指定）。 */
function isBlockedProbeHost(host: string): boolean {
  const text = host.replace(/^\[|\]$/g, "").toLowerCase()
  if (text === "0.0.0.0" || text === "::") {
    return true
  }
  if (/^\d{1,3}(\.\d{1,3}){3}$/.test(text)) {
    const [first, second] = text.split(".").map(Number)
    return (
      (first === 169 && second === 254) || (first >= 224 && first <= 239)
    )
  }
  return (
    text.startsWith("fe8") ||
    text.startsWith("fe9") ||
    text.startsWith("fea") ||
    text.startsWith("feb") ||
    text.startsWith("ff")
  )
}

/** DNS 服务器可选：留空用系统 resolver，填写时必须是 host:port 且不是保留网段。 */
function dnsServerInvalid(value: string): boolean {
  const text = value.trim()
  return (
    text !== "" &&
    (text.length > MAX_DNS_SERVER || splitTCPTarget(text) === null)
  )
}

function dnsServerBlocked(value: string): boolean {
  const text = value.trim()
  if (text === "") {
    return false
  }
  const parsed = splitTCPTarget(text)
  return parsed !== null && isBlockedProbeHost(parsed.host)
}

/** TLS SNI 只能是主机名：拒绝端口、scheme、路径与空白。 */
function tlsServerNameInvalid(value: string): boolean {
  const text = value.trim()
  if (text === "") {
    return false
  }
  return text.length > MAX_TLS_SERVER_NAME || /[\s@:/?#]/.test(text)
}

/** 整数区间校验；留空表示用后端默认值。 */
function intRangeInvalid(value: string, min: number, max: number): boolean {
  const text = value.trim()
  if (text === "") {
    return false
  }
  if (!/^\d+$/.test(text)) {
    return true
  }
  const parsed = Number(text)
  return parsed < min || parsed > max
}

/** 正则可编译性校验；断言值按完整匹配正则使用。 */
function regexInvalid(value: string): boolean {
  try {
    new RegExp(value.trim())
    return false
  } catch {
    return true
  }
}

/** 非 HTTP 驱动的超时字段标题。 */
function timeoutLabelKey(driver: string): string {
  if (driver === PROBE_TCP) {
    return "probeTarget.timeoutConnect"
  }
  if (driver === PROBE_ICMP) {
    return "probeTarget.timeoutEcho"
  }
  if (driver === PROBE_DNS) {
    return "probeTarget.timeoutDns"
  }
  if (driver === PROBE_TLS) {
    return "probeTarget.timeoutTls"
  }
  return "probeTarget.timeoutMs"
}

/** 非 HTTP 驱动的超时字段说明。 */
function timeoutHintKey(driver: string): string {
  if (driver === PROBE_TCP) {
    return "probeTarget.timeoutConnectHint"
  }
  if (driver === PROBE_ICMP) {
    return "probeTarget.timeoutEchoHint"
  }
  if (driver === PROBE_DNS) {
    return "probeTarget.timeoutDnsHint"
  }
  if (driver === PROBE_TLS) {
    return "probeTarget.timeoutTlsHint"
  }
  return "probeTarget.timeoutInvalid"
}

/**
 * 非 HTTP 探测结果的「探测详情」摘要：DNS 答案与记录数、ICMP 丢包/抖动/RTT 极值、
 * TLS 协议与套件/证书主体。HTTP 保留状态码与证书列，不使用这个单元格。
 */
function ProbeDetailSummary({
  driver,
  result,
}: {
  driver: string
  result?: ProbeResultItem | null
}) {
  const { t } = useTranslation()
  if (!result) {
    return <>—</>
  }
  if (driver === PROBE_DNS) {
    if (!result.dnsAnswers && !result.answerCount) {
      return <>—</>
    }
    return (
      <span className="break-words">
        {t("probeTarget.detailDns", {
          count: result.answerCount,
          answers: result.dnsAnswers || "—",
        })}
      </span>
    )
  }
  if (driver === PROBE_ICMP) {
    return (
      <span className="break-words tabular-nums">
        {t("probeTarget.detailIcmp", {
          loss: result.packetLoss,
          jitter: result.jitterMs,
          min: result.minRttMs,
          max: result.maxRttMs,
        })}
      </span>
    )
  }
  if (driver === PROBE_TLS) {
    if (!result.tlsVersion && !result.tlsCipher && !result.certSubject) {
      return <>—</>
    }
    return (
      <span className="grid min-w-0 gap-0.5">
        <span className="break-words tabular-nums">
          {t("probeTarget.detailTlsProtocol", {
            version: result.tlsVersion || "—",
            cipher: result.tlsCipher || "—",
          })}
        </span>
        <span className="text-xs break-words text-muted-foreground">
          {result.certSubject
            ? t("probeTarget.detailTlsCert", {
                subject: result.certSubject,
              })
            : "—"}
        </span>
      </span>
    )
  }
  return <>—</>
}

/** 期望状态码表达式是否非法：单值、区间与逗号列表，值域 100..599。 */
function expectStatusInvalid(value: string): boolean {
  const text = value.replace(/\s+/g, "")
  if (text === "") {
    return false
  }
  let matched = false
  for (const part of text.split(",")) {
    if (part === "") {
      continue
    }
    matched = true
    const dash = part.indexOf("-")
    if (dash > 0) {
      const min = Number(part.slice(0, dash))
      const max = Number(part.slice(dash + 1))
      if (
        !Number.isInteger(min) ||
        !Number.isInteger(max) ||
        min < 100 ||
        max > 599 ||
        min > max
      ) {
        return true
      }
      continue
    }
    if (!/^\d{3}$/.test(part)) {
      return true
    }
    const code = Number(part)
    if (code < 100 || code > 599) {
      return true
    }
  }
  return !matched
}

function timeoutInvalid(value: string): boolean {
  const text = value.trim()
  if (!/^\d+$/.test(text)) {
    return true
  }
  const timeout = Number(text)
  return timeout < MIN_TIMEOUT_MS || timeout > MAX_TIMEOUT_MS
}

function buildRequest(form: ProbeDraft): ProbeRequestConfig {
  // 先用满足 proto 校验的默认值兜底，再按驱动只填该类型真正使用的字段。
  const request: ProbeRequestConfig = {
    method: defaultRequest.method,
    expectStatus: defaultRequest.expectStatus,
    timeoutMs: Number(form.timeoutMs || "0"),
    followRedirects: false,
    skipTlsVerify: false,
    verifyCert: false,
    headers: "{}",
    bodyKeyword: "",
    expectHeaders: "{}",
    bodyRegex: "",
    icmpCount: 0,
    icmpPacketSize: 0,
    dnsRecordType: "",
    dnsExpectValues: "",
    dnsServer: "",
    tlsServerName: "",
  }
  switch (form.driver) {
    case PROBE_HTTP:
      return {
        ...request,
        method: form.method,
        expectStatus: form.expectStatus.trim(),
        followRedirects: form.followRedirects,
        skipTlsVerify: form.skipTlsVerify,
        verifyCert: form.verifyCert,
        headers: serializeHeaders(form.headers),
        bodyKeyword: form.bodyKeyword.trim(),
        expectHeaders: serializeHeaders(form.expectHeaders),
        bodyRegex: form.bodyRegex.trim(),
      }
    case PROBE_ICMP:
      return {
        ...request,
        icmpCount: Number(form.icmpCount || "0"),
        icmpPacketSize: Number(form.icmpPacketSize || "0"),
      }
    case PROBE_DNS:
      return {
        ...request,
        dnsRecordType: form.dnsRecordType,
        dnsExpectValues: form.dnsExpectValues.trim(),
        dnsServer: form.dnsServer.trim(),
      }
    case PROBE_TLS:
      return {
        ...request,
        skipTlsVerify: form.skipTlsVerify,
        verifyCert: form.verifyCert,
        tlsServerName: form.tlsServerName.trim(),
      }
    default:
      // TCP 只用超时。
      return request
  }
}

function buildTargets(form: ProbeDraft): ProbeTargetInput[] {
  return form.targets.map((target, index) => ({
    uid: target.uid || undefined,
    spec: target.spec.trim(),
    labels: target.labels || "{}",
    status: target.enabled ? "ACTIVE" : "DISABLED",
    sort: index + 1,
  }))
}

export function ProbeTargetPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [params, setParams] = useSearchParams()
  const selected = params.get("uid") || ""
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({
    name: query.name,
    type: query.type,
  })
  const [current, setCurrent] = useState<ProbeTargetGroupItem | null>(null)
  // 探测记录按目标展开：一行目标点开后只拉这个目标的记录，按目标缓存。
  const [expandedTarget, setExpandedTarget] = useState("")
  const [targetRecords, setTargetRecords] = useState<
    Record<string, { items: ProbeResultItem[]; total: number; loading: boolean }>
  >({})
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailVersion, setDetailVersion] = useState(0)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState(false)
  const [formStep, setFormStep] = useState(0)
  const [form, setForm] = useState<ProbeDraft>(emptyDraft)
  const [bulkOpen, setBulkOpen] = useState(false)
  const [bulkText, setBulkText] = useState("")
  const [trialResults, setTrialResults] = useState<TrialEntry[]>([])
  const [trialPending, setTrialPending] = useState(false)
  const [probing, setProbing] = useState(false)
  const [note, setNote] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)
  const [confirm, setConfirm] = useState<{
    action: "disable" | "delete"
    uid: string
  } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const [shownAction, setShownAction] = useState<"disable" | "delete">(
    "disable"
  )
  const canWrite = session.can("probe:write")
  // 探测类型按套餐授权展示：未授权的能力不出现在筛选与新建/编辑表单里。
  const authorizedDrivers = useMemo(
    () => authorizedProbeDrivers(session.features),
    [session.features]
  )
  // 编辑存量目标组时保留它的驱动，即使套餐已不再授权，避免改坏已有数据。
  const formDriverChoices = useMemo(() => {
    const choices = [...authorizedDrivers]
    if (form.driver && !choices.includes(form.driver)) {
      choices.push(form.driver)
    }
    return choices
  }, [authorizedDrivers, form.driver])
  const filterDriverChoices = useMemo(() => {
    const choices = [...authorizedDrivers]
    if (query.type && !choices.includes(query.type)) {
      choices.push(query.type)
    }
    return choices
  }, [authorizedDrivers, query.type])
  // 目标地址的占位、提示与校验文案随探测类型切换。
  const formCopySuffix = driverCopySuffix(form.driver)
  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    version,
    reload,
    fail,
  } = usePagedList<ProbeTargetGroupItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${query.type}`,
    fetch: (page, pageSize) =>
      listProbeTargetGroups({
        page,
        pageSize,
        name: query.name,
        driver: query.type,
      }),
  })

  function setParam(
    key: string,
    value: string,
    options: { replace?: boolean } = {}
  ) {
    const next = new URLSearchParams(params)
    if (value) {
      next.set(key, value)
    } else {
      next.delete(key)
    }
    if (key !== "uid" && key !== "type") {
      next.set("page", "1")
    }
    setParams(next, { replace: options.replace })
  }

  function askConfirm(action: "disable" | "delete", uid: string) {
    setShownAction(action)
    setConfirm({ action, uid })
  }

  useEffect(() => {
    let active = true
    async function load() {
      setExpandedTarget("")
      setTargetRecords({})
      if (!selected) {
        setCurrent(null)
        return
      }
      setDetailLoading(true)
      try {
        const groupResult = await getProbeTargetGroup(selected)
        if (active) {
          setCurrent(groupResult.group)
        }
      } catch {
        if (active) {
          setCurrent(null)
        }
      }
      if (active) {
        setDetailLoading(false)
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [selected, version, detailVersion, session.namespaceUid])

  // 进入菜单时默认选中第一个目标组，右侧详情直接可用；只有 URL 没有 uid 时才补选。
  useEffect(() => {
    if (selected || loading || failed || items.length === 0) {
      return
    }
    setParam("uid", items[0].uid, { replace: true })
    // setParam 每次渲染都是新函数；写入 uid 后 selected 非空，本 effect 自然停下。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, loading, failed, items])

  // loadTargetRecords 拉一个目标的探测记录；force 用于「刷新」按钮。
  async function loadTargetRecords(targetUid: string, force = false) {
    if (!selected) {
      return
    }
    if (!force && targetRecords[targetUid] && !targetRecords[targetUid].loading) {
      return
    }
    setTargetRecords((prev) => ({
      ...prev,
      [targetUid]: {
        items: prev[targetUid]?.items ?? [],
        total: prev[targetUid]?.total ?? 0,
        loading: true,
      },
    }))
    try {
      const result = await listProbeTargetResults(selected, {
        targetUid,
        page: 1,
        pageSize: 20,
      })
      setTargetRecords((prev) => ({
        ...prev,
        [targetUid]: {
          items: result.items ?? [],
          total: result.total ?? 0,
          loading: false,
        },
      }))
    } catch {
      setTargetRecords((prev) => ({
        ...prev,
        [targetUid]: { items: [], total: 0, loading: false },
      }))
    }
  }

  // toggleTargetRecords 展开当前目标的记录行，第一次展开时拉数据。
  function toggleTargetRecords(targetUid: string) {
    if (expandedTarget === targetUid) {
      setExpandedTarget("")
      return
    }
    setExpandedTarget(targetUid)
    void loadTargetRecords(targetUid)
  }

  function openCreate() {
    setEditing(false)
    // 默认选中第一个已开通的探测类型；一个都没开通时留空，由校验拦住保存。
    setForm({ ...emptyDraft(), driver: authorizedDrivers[0] ?? "" })
    setBulkOpen(false)
    setBulkText("")
    setTrialResults([])
    setError("")
    setFieldErrors({})
    setFormStep(0)
    setNote("")
    setOpen(true)
  }

  function openEdit(item: ProbeTargetGroupItem) {
    // 列表接口的 targets 可能被裁掉；编辑必须以详情为准，避免保存时整体替换成空列表。
    void (async () => {
      try {
        const result = await getProbeTargetGroup(item.uid)
        setParam("uid", item.uid)
        setEditing(true)
        setForm(draftFromItem(result.group))
        setBulkOpen(false)
        setBulkText("")
        setTrialResults([])
        setError("")
        setFieldErrors({})
        setFormStep(0)
        setNote("")
        setOpen(true)
      } catch (err) {
        const failure = readFormError(err, t("state.failed"))
        if (failure.form) {
          fail(failure.form)
        }
      }
    })()
  }

  function updateTarget(index: number, patch: Partial<TargetDraft>) {
    setForm({
      ...form,
      targets: form.targets.map((target, currentIndex) =>
        currentIndex === index ? { ...target, ...patch } : target
      ),
    })
  }

  function updateHeader(index: number, patch: Partial<LabelPair>) {
    setForm({
      ...form,
      headers: form.headers.map((pair, currentIndex) =>
        currentIndex === index ? { ...pair, ...patch } : pair
      ),
    })
  }

  function updateExpectHeader(index: number, patch: Partial<LabelPair>) {
    setForm({
      ...form,
      expectHeaders: form.expectHeaders.map((pair, currentIndex) =>
        currentIndex === index ? { ...pair, ...patch } : pair
      ),
    })
  }

  function validateBasicStep(): FieldErrors {
    const errors = collectFieldErrors([
      ["name", requireTrimmed(form.name, t("form.required"))],
      ["name", maxLength(form.name, 64, t("form.nameTooLong"))],
      ["driver", requireTrimmed(form.driver, t("form.required"))],
      [
        "labels",
        serializeLabelPairs(form.labels).length > MAX_LABELS
          ? t("probeTarget.labelsTooLong")
          : null,
      ],
      ["remark", maxLength(form.remark, 200, t("probeTarget.remarkTooLong"))],
    ])
    if (!formDriverChoices.includes(form.driver)) {
      errors.driver = t("form.required")
    }
    return errors
  }

  function validateTargetsStep(): FieldErrors {
    const checks: Array<[string, string | null]> = []
    if (form.targets.length === 0) {
      checks.push(["targets", t("probeTarget.targetRequired")])
    }
    if (form.targets.length > MAX_TARGETS) {
      checks.push(["targets", t("probeTarget.tooManyTargets")])
    }
    const seen = new Set<string>()
    form.targets.forEach((target, index) => {
      const spec = target.spec.trim()
      if (spec === "") {
        checks.push([`target.${index}.spec`, t("form.required")])
        return
      }
      if (spec.length > MAX_TARGET_SPEC) {
        checks.push([
          `target.${index}.spec`,
          t("probeTarget.invalidTargetLength"),
        ])
        return
      }
      if (!isTargetForDriver(form.driver, spec)) {
        checks.push([
          `target.${index}.spec`,
          t(`probeTarget.invalidTarget${driverCopySuffix(form.driver)}`),
        ])
        return
      }
      const key = targetKey(form.driver, spec)
      if (seen.has(key)) {
        checks.push([`target.${index}.spec`, t("probeTarget.duplicateTarget")])
      }
      seen.add(key)
    })
    return collectFieldErrors(checks)
  }

  function validateRequestStep(): FieldErrors {
    const checks: Array<[string, string | null]> = []
    // HTTP 探针校验完整请求参数与断言；其它驱动只校验各自使用的字段。
    if (form.driver === PROBE_HTTP) {
      if (form.method !== "GET" && form.method !== "HEAD") {
        checks.push(["method", t("form.required")])
      }
      if (expectStatusInvalid(form.expectStatus)) {
        checks.push(["expectStatus", t("probeTarget.expectStatusInvalid")])
      }
      if (form.bodyKeyword.trim().length > MAX_BODY_KEYWORD) {
        checks.push(["bodyKeyword", t("probeTarget.bodyKeywordTooLong")])
      }
      form.headers.forEach((pair, index) => {
        const key = pair.key.trim()
        if (key === "" && pair.value.trim() === "") {
          return
        }
        if (key === "") {
          checks.push([
            `header.${index}.key`,
            t("probeTarget.headerKeyRequired"),
          ])
          return
        }
        if (controlledHeaders.has(key.toLowerCase())) {
          checks.push([
            `header.${index}.key`,
            t("probeTarget.headerControlled", { key }),
          ])
        }
        if (/[\r\n]/.test(key) || /[\r\n]/.test(pair.value)) {
          checks.push([
            `header.${index}.value`,
            t("probeTarget.headerNewline"),
          ])
        }
      })
      if (serializeHeaders(form.headers).length > MAX_HEADERS) {
        checks.push(["headers", t("probeTarget.headersTooLong")])
      }
      // 响应头断言：每个响应头的值都要能编译为完整匹配正则。
      form.expectHeaders.forEach((pair, index) => {
        const key = pair.key.trim()
        if (key === "" && pair.value.trim() === "") {
          return
        }
        if (key === "") {
          checks.push([
            `expectHeader.${index}.key`,
            t("probeTarget.expectHeaderKeyRequired"),
          ])
          return
        }
        if (/[\r\n]/.test(key) || /[\r\n]/.test(pair.value)) {
          checks.push([
            `expectHeader.${index}.value`,
            t("probeTarget.expectHeaderNewline"),
          ])
          return
        }
        if (regexInvalid(pair.value)) {
          checks.push([
            `expectHeader.${index}.value`,
            t("probeTarget.expectHeaderInvalidRegex"),
          ])
        }
      })
      if (serializeHeaders(form.expectHeaders).length > MAX_EXPECT_HEADERS) {
        checks.push(["expectHeaders", t("probeTarget.expectHeadersTooLong")])
      }
      const bodyRegex = form.bodyRegex.trim()
      if (bodyRegex.length > MAX_BODY_REGEX) {
        checks.push(["bodyRegex", t("probeTarget.bodyRegexTooLong")])
      } else if (bodyRegex !== "" && regexInvalid(bodyRegex)) {
        checks.push(["bodyRegex", t("probeTarget.bodyRegexInvalid")])
      }
    }
    if (timeoutInvalid(form.timeoutMs)) {
      checks.push(["timeoutMs", t("probeTarget.timeoutInvalid")])
    }
    if (form.driver === PROBE_ICMP) {
      if (intRangeInvalid(form.icmpCount, 0, 10)) {
        checks.push(["icmpCount", t("probeTarget.icmpCountInvalid")])
      }
      if (intRangeInvalid(form.icmpPacketSize, 0, 1400)) {
        checks.push(["icmpPacketSize", t("probeTarget.icmpPacketSizeInvalid")])
      }
    }
    if (form.driver === PROBE_DNS) {
      const expectValues = form.dnsExpectValues.trim()
      if (expectValues.length > MAX_DNS_EXPECT_VALUES) {
        checks.push([
          "dnsExpectValues",
          t("probeTarget.dnsExpectValuesTooLong"),
        ])
      } else {
        // 每个期望值都是完整匹配正则，任何一个非法都拦下。
        expectValues
          .split(",")
          .map((value) => value.trim())
          .filter((value) => value !== "")
          .forEach((value, index) => {
            if (regexInvalid(value)) {
              checks.push([
                "dnsExpectValues",
                t("probeTarget.dnsExpectValuesInvalid", { index: index + 1 }),
              ])
            }
          })
      }
      if (dnsServerInvalid(form.dnsServer)) {
        checks.push(["dnsServer", t("probeTarget.dnsServerInvalid")])
      } else if (dnsServerBlocked(form.dnsServer)) {
        checks.push(["dnsServer", t("probeTarget.dnsServerBlocked")])
      }
    }
    if (form.driver === PROBE_TLS && tlsServerNameInvalid(form.tlsServerName)) {
      checks.push(["tlsServerName", t("probeTarget.tlsServerNameInvalid")])
    }
    return collectFieldErrors(checks)
  }

  function validateStep(step: number): FieldErrors {
    if (step === 0) {
      return validateBasicStep()
    }
    if (step === 1) {
      return validateTargetsStep()
    }
    return validateRequestStep()
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const stepErrors = [0, 1, 2].map((step) => validateStep(step))
    const errors = stepErrors.reduce<FieldErrors>(
      (all, currentErrors) => ({ ...all, ...currentErrors }),
      {}
    )
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setError("")
      const firstInvalidStep = stepErrors.findIndex(hasFieldErrors)
      if (firstInvalidStep >= 0) {
        setFormStep(firstInvalidStep)
      }
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload: CreateProbeTargetGroupParams = {
        name: form.name.trim(),
        driver: form.driver,
        labels: serializeLabelPairs(form.labels),
        // 一期只支持集群内执行器，执行器标签固定为空对象。
        executorLabels: "{}",
        request: buildRequest(form),
        remark: form.remark.trim(),
        targets: buildTargets(form),
      }
      if (editing && current) {
        await updateProbeTargetGroup({ uid: current.uid, ...payload })
      } else {
        const created = await createProbeTargetGroup(payload)
        setParam("uid", created.group.uid, { replace: true })
      }
      setOpen(false)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      if (confirm.action === "delete") {
        await deleteProbeTargetGroup(confirm.uid)
        if (selected === confirm.uid) {
          setParam("uid", "")
        }
      } else {
        await updateProbeTargetGroupStatus(confirm.uid, "DISABLED")
      }
      setConfirm(null)
      setDetailVersion((value) => value + 1)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setConfirmError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function enableItem(item: ProbeTargetGroupItem) {
    try {
      await updateProbeTargetGroupStatus(item.uid, "ACTIVE")
      setDetailVersion((value) => value + 1)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      if (failure.form) {
        fail(failure.form)
      }
    }
  }

  async function toggleCurrentStatus() {
    if (!current) {
      return
    }
    setPending(true)
    setNote("")
    try {
      const nextStatus = current.status === "ACTIVE" ? "DISABLED" : "ACTIVE"
      await updateProbeTargetGroupStatus(current.uid, nextStatus)
      setDetailVersion((value) => value + 1)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      if (failure.form) {
        fail(failure.form)
      }
    } finally {
      setPending(false)
    }
  }

  async function probeNow() {
    if (!current) {
      return
    }
    setProbing(true)
    setNote("")
    try {
      const result = await probeTargetGroupNow(current.uid)
      setNote(
        t("probeTarget.probeNowDone", { count: result.results?.length ?? 0 })
      )
      setDetailVersion((value) => value + 1)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      if (failure.form) {
        fail(failure.form)
      }
    } finally {
      setProbing(false)
    }
  }

  function applyBulk() {
    const lines = bulkText
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter((line) => line !== "")
    if (lines.length === 0) {
      setFieldErrors({
        ...fieldErrors,
        bulk: t("probeTarget.bulkPasteEmpty"),
      })
      return
    }
    const seen = new Set(
      form.targets
        .map((target) => target.spec.trim())
        .filter((spec) => spec !== "")
        .map((spec) => targetKey(form.driver, spec))
    )
    const additions: TargetDraft[] = []
    for (let index = 0; index < lines.length; index += 1) {
      const spec = lines[index]
      if (!isTargetForDriver(form.driver, spec)) {
        setFieldErrors({
          ...fieldErrors,
          bulk: t(
            `probeTarget.bulkPasteInvalidLine${driverCopySuffix(form.driver)}`,
            { line: index + 1 }
          ),
        })
        return
      }
      const key = targetKey(form.driver, spec)
      if (seen.has(key)) {
        setFieldErrors({
          ...fieldErrors,
          bulk: t("probeTarget.bulkPasteDuplicateLine", { line: index + 1 }),
        })
        return
      }
      seen.add(key)
      additions.push({ uid: "", spec, labels: "{}", enabled: true })
    }
    // 批量导入时丢掉还没填写内容的空行，避免留下无意义的必填错误。
    const kept = form.targets.filter((target) => target.spec.trim() !== "")
    setForm({ ...form, targets: [...kept, ...additions] })
    setFieldErrors({})
    setBulkOpen(false)
    setBulkText("")
  }

  async function runTrial() {
    const errors = {
      ...validateTargetsStep(),
      ...validateRequestStep(),
    }
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    const specs = form.targets
      .map((target) => target.spec.trim())
      .filter((spec) => spec !== "")
    if (specs.length === 0) {
      return
    }
    setTrialPending(true)
    setTrialResults([])
    try {
      const request = buildRequest(form)
      const entries: TrialEntry[] = []
      for (const spec of specs) {
        try {
          const response = await probeTrial({
            driver: form.driver,
            spec,
            request,
          })
          entries.push({ spec, result: response.result })
        } catch {
          entries.push({ spec, result: null })
        }
        setTrialResults([...entries])
      }
    } finally {
      setTrialPending(false)
    }
  }

  const filterChips: FilterChip[] = []
  if (query.name) {
    filterChips.push({
      key: "name",
      label: `${t("probeTarget.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (query.type) {
    filterChips.push({
      key: "type",
      label: `${t("probeTarget.type")}: ${enumLabel(t, query.type)}`,
      onRemove: () => query.update({ type: "" }),
    })
  }

  function applyFilters() {
    query.update({ name: draft.name.trim(), type: draft.type })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "", type: "" })
  }

  const targets = current?.targets ?? []
  // HTTP 专属列（状态码、证书剩余天数）只对 HTTP 目标组展示。
  const currentIsHttp = current?.driver === PROBE_HTTP

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("probeTarget.title")}
        description={t("probeTarget.description")}
      >
        {canWrite ? (
          <Button onClick={openCreate}>{t("action.create")}</Button>
        ) : null}
      </PageHeader>
      <ListFilter
        id="probe-target-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={filterChips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={filterChips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="min-w-0 w-full max-w-[500px] flex-1 basis-56">
          <FieldLabel
            htmlFor="probe-target-filter-name"
            className="text-xs text-muted-foreground"
          >
            {t("probeTarget.name")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="probe-target-filter-name"
              className="pl-9"
              value={draft.name}
              placeholder={t("probeTarget.namePlaceholder")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
        <Field className="w-40 shrink-0">
          <FieldLabel
            htmlFor="probe-target-filter-type"
            className="text-xs text-muted-foreground"
          >
            {t("probeTarget.type")}
          </FieldLabel>
          <Select
            value={draft.type || "ALL"}
            onValueChange={(value) =>
              update("type", value === "ALL" ? "" : value)
            }
          >
            <SelectTrigger id="probe-target-filter-type" className="w-full">
              <SelectValue placeholder={t("probeTarget.typePlaceholder")}>
                <span className="flex items-center gap-2">
                  {draft.type ? <EnumIcon value={draft.type} /> : null}
                  <span>
                    {draft.type
                      ? enumLabel(t, draft.type)
                      : t("probeTarget.typeAll")}
                  </span>
                </span>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">{t("probeTarget.typeAll")}</SelectItem>
              {filterDriverChoices.map((driver) => (
                <SelectItem key={driver} value={driver}>
                  <span className="flex items-center gap-2">
                    <EnumIcon value={driver} />
                    <span>{enumLabel(t, driver)}</span>
                  </span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>
      </ListFilter>
      <div className="grid min-h-0 min-w-0 flex-1 gap-6 lg:grid-cols-[minmax(0,18rem)_minmax(0,1fr)]">
        <ListFrame
          total={total}
          page={query.page}
          pageSize={query.pageSize}
          showPager={!loading && !failed && items.length > 0}
          onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
          onNext={() => query.update({ page: query.page + 1 })}
        >
          {loading ? (
            <p className="text-sm text-muted-foreground">
              {t("state.loading")}
            </p>
          ) : null}
          {!loading && failed ? (
            <div className="flex items-center gap-2">
              <p className="text-sm text-destructive">
                {loadError || t("state.failed")}
              </p>
              <Button variant="outline" onClick={() => reload()}>
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {!loading && !failed && items.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {t("probeTarget.empty")}
            </p>
          ) : null}
          <div className="grid min-h-0 min-w-0 flex-1 content-start gap-2 overflow-x-hidden overflow-y-auto">
            {items.map((item) => (
              <div key={item.uid} className="flex min-w-0 items-center gap-1">
                <Button
                  variant={item.uid === selected ? "secondary" : "ghost"}
                  className="h-auto min-w-0 flex-1 justify-start overflow-hidden px-2 py-2"
                  onClick={() => setParam("uid", item.uid)}
                >
                  <span className="flex w-full min-w-0 flex-col items-start gap-1">
                    <span className="w-full truncate text-left" title={item.name}>
                      {item.name}
                    </span>
                    <span className="flex flex-wrap gap-1">
                      <StatusBadge value={item.driver} />
                      <StatusBadge value={item.status} />
                    </span>
                  </span>
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={t("action.more")}
                    >
                      <EllipsisVertical />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    {canWrite ? (
                      <DropdownMenuItem onSelect={() => openEdit(item)}>
                        {t("action.edit")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite && item.status !== "ACTIVE" ? (
                      <DropdownMenuItem onSelect={() => void enableItem(item)}>
                        {t("action.enable")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite ? <DropdownMenuSeparator /> : null}
                    {canWrite && item.status === "ACTIVE" ? (
                      <DropdownMenuItem
                        variant="destructive"
                        onSelect={() => askConfirm("disable", item.uid)}
                      >
                        {t("action.disable")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite ? (
                      <DropdownMenuItem
                        variant="destructive"
                        onSelect={() => askConfirm("delete", item.uid)}
                      >
                        {t("action.delete")}
                      </DropdownMenuItem>
                    ) : null}
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))}
          </div>
        </ListFrame>
        <Card className="flex min-h-0 flex-1 flex-col overflow-hidden">
          <CardContent className="flex min-h-0 flex-1 flex-col gap-3 overflow-hidden">
            {!current ? (
              <p className="text-sm text-muted-foreground">
                {detailLoading ? t("state.loading") : t("probeTarget.pick")}
              </p>
            ) : null}
            {current ? (
              <div className="min-h-0 min-w-0 flex-1 overflow-y-auto">
                <div className="grid min-w-0 gap-4">
                  <div className="grid gap-2">
                    <h2 className="flex min-w-0 flex-wrap items-center gap-2 text-lg font-semibold">
                      <span className="min-w-0 truncate" title={current.name}>
                        {current.name}
                      </span>
                      <Badge variant="secondary">
                        {t("probeTarget.targetCountBadge", {
                          count: current.targetCount,
                        })}
                      </Badge>
                      <StatusBadge value={current.driver} />
                      <StatusBadge value={current.status} />
                    </h2>
                  </div>
                  <dl className="grid min-w-0 gap-4 sm:grid-cols-2">
                    <div className="grid min-w-0 gap-1">
                      <dt className="text-sm text-muted-foreground">
                        {t("probeTarget.type")}
                      </dt>
                      <dd>
                        <StatusBadge value={current.driver} />
                      </dd>
                    </div>
                    <div className="grid min-w-0 gap-1">
                      <dt className="text-sm text-muted-foreground">
                        {t("probeTarget.targetCountLabel")}
                      </dt>
                      <dd className="tabular-nums">
                        {t("probeTarget.targetCountValue", {
                          total: current.targetCount,
                          active: current.activeTargetCount,
                        })}
                      </dd>
                    </div>
                    <div className="grid min-w-0 gap-1">
                      <dt className="text-sm text-muted-foreground">
                        {t("probeTarget.executor")}
                      </dt>
                      <dd className="grid gap-1">
                        <span>{t("probeTarget.executorCluster")}</span>
                        <span className="text-xs text-muted-foreground">
                          {t("probeTarget.executorHint")}
                        </span>
                      </dd>
                    </div>
                    <div className="grid min-w-0 gap-1">
                      <dt className="text-sm text-muted-foreground">
                        {t("probeTarget.lastProbedAt")}
                      </dt>
                      <dd className="tabular-nums break-words">
                        {current.lastProbedAt
                          ? formatDateTime(current.lastProbedAt)
                          : t("probeTarget.neverProbed")}
                      </dd>
                    </div>
                    <div className="grid min-w-0 gap-1 sm:col-span-full">
                      <dt className="text-sm text-muted-foreground">
                        {t("probeTarget.labels")}
                      </dt>
                      <dd className="min-w-0">
                        <LabelTags labels={current.labels} />
                      </dd>
                    </div>
                    <div className="grid min-w-0 gap-1 sm:col-span-full">
                      <dt className="text-sm text-muted-foreground">
                        {t("probeTarget.remark")}
                      </dt>
                      <dd className="min-w-0 break-words whitespace-pre-wrap">
                        {current.remark || "—"}
                      </dd>
                    </div>
                  </dl>
                  <section className="grid min-w-0 gap-2 border-t pt-4">
                    <h3 className="text-sm font-medium">
                      {t("probeTarget.targetsSection")}
                    </h3>
                    <div className="flex max-h-[18rem] min-h-0 flex-col">
                    <Table className={TABLE_MIN_WIDTH.default}>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-10" />
                          <TableHead>{t("probeTarget.targetUrl")}</TableHead>
                          <TableHead className={TABLE_COLUMN_WIDTH.status}>
                            {t("namespace.status")}
                          </TableHead>
                          <TableHead className={TABLE_COLUMN_WIDTH.status}>
                            {t("probeTarget.lastResult")}
                          </TableHead>
                          {currentIsHttp ? (
                            <TableHead className={TABLE_COLUMN_WIDTH.type}>
                              {t("probeTarget.statusCode")}
                            </TableHead>
                          ) : (
                            <TableHead className="min-w-[10rem]">
                              {t("probeTarget.detailColumn")}
                            </TableHead>
                          )}
                          <TableHead className={TABLE_COLUMN_WIDTH.type}>
                            {t("probeTarget.duration")}
                          </TableHead>
                          {currentIsHttp ? (
                            <TableHead className={TABLE_COLUMN_WIDTH.type}>
                              {t("probeTarget.certDays")}
                            </TableHead>
                          ) : null}
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {targets.map((target) => {
                          const targetKey = target.uid || target.spec
                          const expanded = expandedTarget === targetKey
                          const records = targetRecords[targetKey]
                          return (
                          <Fragment key={targetKey}>
                          <TableRow>
                            <TableCell className={TABLE_NOWRAP}>
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                aria-expanded={expanded}
                                aria-controls={`probe-records-${targetKey}`}
                                aria-label={
                                  expanded
                                    ? t("probeTarget.recordsCollapse")
                                    : t("probeTarget.recordsExpand")
                                }
                                onClick={() => toggleTargetRecords(targetKey)}
                              >
                                {expanded ? <ChevronDown /> : <ChevronRight />}
                              </Button>
                            </TableCell>
                            <TableCell className="min-w-0 break-all">
                              {target.spec}
                            </TableCell>
                            <TableCell className={TABLE_NOWRAP}>
                              <StatusBadge value={target.status} />
                            </TableCell>
                            <TableCell className={TABLE_NOWRAP}>
                              {target.lastResult ? (
                                <StatusBadge
                                  value={target.lastResult.ok ? "OK" : "FAILED"}
                                  label={
                                    target.lastResult.ok
                                      ? t("probeTarget.ok")
                                      : t("probeTarget.failed")
                                  }
                                />
                              ) : (
                                "—"
                              )}
                            </TableCell>
                            {currentIsHttp ? (
                              <TableCell className={TABLE_NOWRAP}>
                                {target.lastResult?.statusCode || "—"}
                              </TableCell>
                            ) : (
                              <TableCell className="max-w-[16rem] min-w-0 align-top">
                                <ProbeDetailSummary
                                  driver={current.driver}
                                  result={target.lastResult}
                                />
                              </TableCell>
                            )}
                            <TableCell className={TABLE_NOWRAP}>
                              {target.lastResult
                                ? t("probeTarget.durationValue", {
                                    ms: target.lastResult.durationMs,
                                  })
                                : "—"}
                            </TableCell>
                            {currentIsHttp ? (
                              <TableCell className={TABLE_NOWRAP}>
                                {target.lastResult?.certDays
                                  ? t("probeTarget.certDaysValue", {
                                      days: target.lastResult.certDays,
                                    })
                                  : "—"}
                              </TableCell>
                            ) : null}
                          </TableRow>
                          {expanded ? (
                            <TableRow
                              id={`probe-records-${targetKey}`}
                              className="bg-muted/20 hover:bg-muted/20"
                            >
                              <TableCell
                                colSpan={currentIsHttp ? 7 : 6}
                                className="min-w-0"
                              >
                                <div className="grid min-w-0 gap-2 py-1">
                                  <div className="flex flex-wrap items-center justify-between gap-2">
                                    <span className="text-xs text-muted-foreground">
                                      {t("probeTarget.recordsTotal", {
                                        count: records?.total ?? 0,
                                      })}
                                    </span>
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      disabled={records?.loading}
                                      onClick={() =>
                                        void loadTargetRecords(targetKey, true)
                                      }
                                    >
                                      {t("action.refresh")}
                                    </Button>
                                  </div>
                                  {records?.loading && !records.items.length ? (
                                    <p className="text-sm text-muted-foreground">
                                      {t("state.loading")}
                                    </p>
                                  ) : null}
                                  {!records?.loading &&
                                  (records?.items.length ?? 0) === 0 ? (
                                    <p className="text-sm text-muted-foreground">
                                      {t("probeTarget.recordsEmpty")}
                                    </p>
                                  ) : null}
                                  {(records?.items.length ?? 0) > 0 ? (
                                    <div className="flex max-h-64 min-h-0 flex-col">
                                    <Table className={TABLE_MIN_WIDTH.wide}>
                                      <TableHeader>
                                        <TableRow>
                                          <TableHead
                                            className={TABLE_COLUMN_WIDTH.time}
                                          >
                                            {t("probeTarget.resultTime")}
                                          </TableHead>
                                          <TableHead>
                                            {t("probeTarget.rule")}
                                          </TableHead>
                                          <TableHead
                                            className={TABLE_COLUMN_WIDTH.status}
                                          >
                                            {t("probeTarget.resultStatus")}
                                          </TableHead>
                                          {currentIsHttp ? (
                                            <TableHead
                                              className={
                                                TABLE_COLUMN_WIDTH.type
                                              }
                                            >
                                              {t("probeTarget.statusCode")}
                                            </TableHead>
                                          ) : (
                                            <TableHead className="min-w-[10rem]">
                                              {t("probeTarget.detailColumn")}
                                            </TableHead>
                                          )}
                                          <TableHead
                                            className={TABLE_COLUMN_WIDTH.type}
                                          >
                                            {t("probeTarget.duration")}
                                          </TableHead>
                                          <TableHead>
                                            {t("probeTarget.errorMessage")}
                                          </TableHead>
                                        </TableRow>
                                      </TableHeader>
                                      <TableBody>
                                        {records.items.map((item) => (
                                          <TableRow
                                            key={`${item.targetUid}-${item.slot || item.probedAt}`}
                                          >
                                            <TableCell
                                              className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP} tabular-nums`}
                                            >
                                              {formatDateTime(
                                                item.probedAt || item.slot
                                              )}
                                            </TableCell>
                                            <TableCell className="min-w-0 break-words">
                                              {item.ruleName ||
                                                (item.ruleUid
                                                  ? "—"
                                                  : t("probeTarget.manualProbe"))}
                                            </TableCell>
                                            <TableCell className={TABLE_NOWRAP}>
                                              <StatusBadge
                                                value={item.ok ? "OK" : "FAILED"}
                                                label={
                                                  item.ok
                                                    ? t("probeTarget.ok")
                                                    : t("probeTarget.failed")
                                                }
                                              />
                                            </TableCell>
                                            {currentIsHttp ? (
                                              <TableCell
                                                className={TABLE_NOWRAP}
                                              >
                                                {item.statusCode || "—"}
                                              </TableCell>
                                            ) : (
                                              <TableCell className="max-w-[16rem] min-w-0 align-top">
                                                <ProbeDetailSummary
                                                  driver={current.driver}
                                                  result={item}
                                                />
                                              </TableCell>
                                            )}
                                            <TableCell className={TABLE_NOWRAP}>
                                              {t("probeTarget.durationValue", {
                                                ms: item.durationMs,
                                              })}
                                            </TableCell>
                                            <TableCell className="min-w-0 break-words">
                                              {item.ok
                                                ? item.errorMessage || "—"
                                                : `${item.errorKind ? `${item.errorKind}: ` : ""}${item.errorMessage || "—"}`}
                                            </TableCell>
                                          </TableRow>
                                        ))}
                                      </TableBody>
                                    </Table>
                                    </div>
                                  ) : null}
                                </div>
                              </TableCell>
                            </TableRow>
                          ) : null}
                          </Fragment>
                          )
                        })}
                      </TableBody>
                    </Table>
                    </div>
                  </section>
                </div>
              </div>
            ) : null}
            {current ? (
              <div className="flex shrink-0 flex-wrap items-center justify-end gap-2 border-t pt-3">
                {note ? (
                  <p className="mr-auto text-sm text-muted-foreground">
                    {note}
                  </p>
                ) : null}
                {canWrite ? (
                  <>
                    <Button
                      variant="outline"
                      disabled={pending}
                      onClick={() => void toggleCurrentStatus()}
                    >
                      {current.status === "ACTIVE"
                        ? t("action.disable")
                        : t("action.enable")}
                    </Button>
                    <Button
                      disabled={
                        probing || current.activeTargetCount === 0
                      }
                      onClick={() => void probeNow()}
                    >
                      {probing
                        ? t("probeTarget.probing")
                        : t("probeTarget.probeNow")}
                    </Button>
                  </>
                ) : null}
              </div>
            ) : null}
          </CardContent>
        </Card>
      </div>
      <Dialog
        open={open}
        onOpenChange={(value) => {
          setOpen(value)
          if (!value) {
            setError("")
          }
        }}
      >
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editing ? t("action.edit") : t("action.create")}
            </DialogTitle>
          </DialogHeader>
          <StepForm
            step={formStep}
            stepCount={3}
            pending={pending}
            error={error}
            fieldsClassName="sm:grid-cols-2"
            onBack={() => setFormStep((value) => Math.max(0, value - 1))}
            onNext={() => {
              const errors = validateStep(formStep)
              if (hasFieldErrors(errors)) {
                setFieldErrors(errors)
                setError("")
                return
              }
              setError("")
              setFieldErrors({})
              setFormStep((value) => Math.min(2, value + 1))
            }}
            onSubmit={onSave}
          >
            {formStep === 0 ? (
              <>
                <FormField
                  label={t("probeTarget.name")}
                  htmlFor="probe-name"
                  className="sm:col-span-full"
                  required
                  error={fieldErrors.name}
                >
                  <Input
                    id="probe-name"
                    value={form.name}
                    maxLength={64}
                    placeholder={t("probeTarget.namePlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, name: event.target.value })
                    }
                    required
                  />
                </FormField>
                <FormField
                  label={t("probeTarget.type")}
                  hint={t("probeTarget.typeHint")}
                  required
                  error={fieldErrors.driver}
                >
                  <Select
                    value={form.driver}
                    onValueChange={(value) => {
                      if (value === form.driver) {
                        return
                      }
                      // 目标写法随探测类型变化：切换类型后清空目标列表，
                      // 避免留下另一种类型的地址被一起保存。
                      setForm({
                        ...form,
                        driver: value,
                        targets: [emptyTarget()],
                      })
                      setTrialResults([])
                      setFieldErrors({})
                    }}
                  >
                    <SelectTrigger aria-label={t("probeTarget.type")}>
                      <SelectValue
                        placeholder={t("probeTarget.typePlaceholder")}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {formDriverChoices.map((driver) => (
                        <SelectItem key={driver} value={driver}>
                          {enumLabel(t, driver)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField
                  label={t("probeTarget.remark")}
                  htmlFor="probe-remark"
                  className="sm:col-span-full"
                  error={fieldErrors.remark}
                >
                  <Input
                    id="probe-remark"
                    value={form.remark}
                    maxLength={200}
                    placeholder={t("probeTarget.remark")}
                    onChange={(event) =>
                      setForm({ ...form, remark: event.target.value })
                    }
                  />
                </FormField>
                <div className="grid gap-2 sm:col-span-full">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-medium">
                      {t("probeTarget.labels")}
                    </span>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        setForm({
                          ...form,
                          labels: [...form.labels, { key: "", value: "" }],
                        })
                      }
                    >
                      {t("action.add")}
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {t("probeTarget.labelsHint")}
                  </p>
                  {form.labels.map((pair, index) => (
                    <div key={index} className="grid gap-2 sm:grid-cols-2">
                      <Input
                        value={pair.key}
                        placeholder={t("datasource.labelKeyPlaceholder")}
                        onChange={(event) => {
                          const labels = [...form.labels]
                          labels[index] = { ...pair, key: event.target.value }
                          setForm({ ...form, labels })
                        }}
                      />
                      <Input
                        value={pair.value}
                        placeholder={t("datasource.labelValuePlaceholder")}
                        onChange={(event) => {
                          const labels = [...form.labels]
                          labels[index] = {
                            ...pair,
                            value: event.target.value,
                          }
                          setForm({ ...form, labels })
                        }}
                      />
                    </div>
                  ))}
                  {fieldErrors.labels ? (
                    <p role="alert" className="text-sm text-destructive">
                      {fieldErrors.labels}
                    </p>
                  ) : null}
                </div>
              </>
            ) : null}
            {formStep === 1 ? (
              <>
                <div className="flex flex-wrap items-center justify-between gap-2 sm:col-span-full">
                  <span className="text-sm font-medium">
                    {t("probeTarget.targetsStep")}
                  </span>
                  <div className="flex flex-wrap gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        setForm({
                          ...form,
                          targets: [...form.targets, emptyTarget()],
                        })
                      }
                    >
                      {t("probeTarget.addTarget")}
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setBulkOpen((value) => !value)
                        setFieldErrors({})
                      }}
                    >
                      {t("probeTarget.bulkPaste")}
                    </Button>
                  </div>
                </div>
                <p className="text-xs text-muted-foreground sm:col-span-full">
                  {t(`probeTarget.targetsHint${formCopySuffix}`)}
                </p>
                {bulkOpen ? (
                  <div className="grid gap-2 rounded-lg border p-3 sm:col-span-full">
                    <p className="text-xs text-muted-foreground">
                      {t(`probeTarget.bulkPasteHint${formCopySuffix}`)}
                    </p>
                    <Textarea
                      value={bulkText}
                      placeholder={t(
                        `probeTarget.bulkPastePlaceholder${formCopySuffix}`
                      )}
                      onChange={(event) => setBulkText(event.target.value)}
                      aria-label={t("probeTarget.bulkPaste")}
                    />
                    {fieldErrors.bulk ? (
                      <p role="alert" className="text-sm text-destructive">
                        {fieldErrors.bulk}
                      </p>
                    ) : null}
                    <div className="flex justify-end gap-2">
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setBulkOpen(false)
                          setBulkText("")
                          setFieldErrors({})
                        }}
                      >
                        {t("probeTarget.bulkPasteCancel")}
                      </Button>
                      <Button type="button" size="sm" onClick={applyBulk}>
                        {t("probeTarget.bulkPasteApply")}
                      </Button>
                    </div>
                  </div>
                ) : null}
                {form.targets.map((target, index) => (
                  <FormField
                    key={index}
                    label={t("probeTarget.targetLabel", { index: index + 1 })}
                    className="sm:col-span-full"
                    required
                    error={fieldErrors[`target.${index}.spec`]}
                  >
                    <div className="flex items-center gap-2">
                      <Input
                        className="min-w-0 flex-1"
                        value={target.spec}
                        maxLength={MAX_TARGET_SPEC}
                        placeholder={t(
                          `probeTarget.targetSpecPlaceholder${formCopySuffix}`
                        )}
                        onChange={(event) =>
                          updateTarget(index, { spec: event.target.value })
                        }
                      />
                      <label className="flex shrink-0 items-center gap-1.5 text-sm">
                        <Checkbox
                          checked={target.enabled}
                          aria-label={t("probeTarget.targetEnabled")}
                          onCheckedChange={(value) =>
                            updateTarget(index, { enabled: value === true })
                          }
                        />
                        <span>{t("probeTarget.targetEnabled")}</span>
                      </label>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        aria-label={t("probeTarget.removeTarget")}
                        onClick={() =>
                          setForm({
                            ...form,
                            targets: form.targets.filter(
                              (_, currentIndex) => currentIndex !== index
                            ),
                          })
                        }
                      >
                        <Trash2 />
                      </Button>
                    </div>
                  </FormField>
                ))}
                {fieldErrors.targets ? (
                  <p
                    role="alert"
                    className="text-sm text-destructive sm:col-span-full"
                  >
                    {fieldErrors.targets}
                  </p>
                ) : null}
              </>
            ) : null}
            {formStep === 2 ? (
              <>
                {form.driver === PROBE_HTTP ? (
                  <>
                    <FormField
                      label={t("probeTarget.method")}
                      required
                      error={fieldErrors.method}
                    >
                      <Select
                        value={form.method}
                        onValueChange={(value) =>
                          setForm({ ...form, method: value })
                        }
                      >
                        <SelectTrigger aria-label={t("probeTarget.method")}>
                          <SelectValue placeholder="GET" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="GET">GET</SelectItem>
                          <SelectItem value="HEAD">HEAD</SelectItem>
                        </SelectContent>
                      </Select>
                    </FormField>
                    <FormField
                      label={t("probeTarget.expectStatus")}
                      htmlFor="probe-expect-status"
                      hint={t("probeTarget.expectStatusHint")}
                      error={fieldErrors.expectStatus}
                    >
                      <Input
                        id="probe-expect-status"
                        value={form.expectStatus}
                        maxLength={128}
                        placeholder="200-299"
                        onChange={(event) =>
                          setForm({ ...form, expectStatus: event.target.value })
                        }
                      />
                    </FormField>
                    <FormField
                      label={t("probeTarget.timeoutMs")}
                      htmlFor="probe-timeout"
                      required
                      error={fieldErrors.timeoutMs}
                    >
                      <Input
                        id="probe-timeout"
                        type="number"
                        min={MIN_TIMEOUT_MS}
                        max={MAX_TIMEOUT_MS}
                        step={1}
                        inputMode="numeric"
                        value={form.timeoutMs}
                        onChange={(event) =>
                          setForm({ ...form, timeoutMs: event.target.value })
                        }
                        required
                      />
                    </FormField>
                    <FormField
                      label={t("probeTarget.bodyKeyword")}
                      htmlFor="probe-keyword"
                      hint={t("probeTarget.bodyKeywordHint")}
                      error={fieldErrors.bodyKeyword}
                    >
                      <Input
                        id="probe-keyword"
                        value={form.bodyKeyword}
                        maxLength={MAX_BODY_KEYWORD}
                        onChange={(event) =>
                          setForm({ ...form, bodyKeyword: event.target.value })
                        }
                      />
                    </FormField>
                    <FormField
                      label={t("probeTarget.bodyRegex")}
                      htmlFor="probe-body-regex"
                      hint={t("probeTarget.bodyRegexHint")}
                      error={fieldErrors.bodyRegex}
                    >
                      <Input
                        id="probe-body-regex"
                        value={form.bodyRegex}
                        maxLength={MAX_BODY_REGEX}
                        placeholder="ok|healthy"
                        onChange={(event) =>
                          setForm({ ...form, bodyRegex: event.target.value })
                        }
                      />
                    </FormField>
                    <div className="grid gap-3 sm:col-span-full">
                      <label className="flex items-start gap-2">
                        <Checkbox
                          className="mt-0.5"
                          checked={form.followRedirects}
                          onCheckedChange={(value) =>
                            setForm({ ...form, followRedirects: value === true })
                          }
                        />
                        <span className="grid gap-1">
                          <span className="text-sm font-medium">
                            {t("probeTarget.followRedirects")}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {t("probeTarget.followRedirectsHint")}
                          </span>
                        </span>
                      </label>
                      <label className="flex items-start gap-2">
                        <Checkbox
                          className="mt-0.5"
                          checked={form.skipTlsVerify}
                          onCheckedChange={(value) =>
                            setForm({ ...form, skipTlsVerify: value === true })
                          }
                        />
                        <span className="grid gap-1">
                          <span className="text-sm font-medium">
                            {t("probeTarget.skipTlsVerify")}
                          </span>
                        </span>
                      </label>
                      <label className="flex items-start gap-2">
                        <Checkbox
                          className="mt-0.5"
                          checked={form.verifyCert}
                          onCheckedChange={(value) =>
                            setForm({ ...form, verifyCert: value === true })
                          }
                        />
                        <span className="grid gap-1">
                          <span className="text-sm font-medium">
                            {t("probeTarget.verifyCert")}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {t("probeTarget.verifyCertHint")}
                          </span>
                        </span>
                      </label>
                    </div>
                    <div className="grid min-w-0 gap-2 sm:col-span-full">
                      <div className="flex items-center justify-between gap-2">
                        <span className="text-sm font-medium">
                          {t("probeTarget.headers")}
                        </span>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            setForm({
                              ...form,
                              headers: [...form.headers, { key: "", value: "" }],
                            })
                          }
                        >
                          {t("probeTarget.addHeader")}
                        </Button>
                      </div>
                      <p className="text-xs text-muted-foreground">
                        {t("probeTarget.headersHint")}
                      </p>
                      {form.headers.map((pair, index) => (
                        <FormField
                          key={index}
                          label={t("probeTarget.headerLabel", { index: index + 1 })}
                          error={
                            fieldErrors[`header.${index}.key`] ??
                            fieldErrors[`header.${index}.value`]
                          }
                        >
                          <div className="flex items-center gap-2">
                            <Input
                              className="min-w-0 flex-1"
                              value={pair.key}
                              placeholder={t("probeTarget.headerKeyPlaceholder")}
                              onChange={(event) =>
                                updateHeader(index, { key: event.target.value })
                              }
                            />
                            <Input
                              className="min-w-0 flex-1"
                              value={pair.value}
                              placeholder={t("probeTarget.headerValuePlaceholder")}
                              onChange={(event) =>
                                updateHeader(index, { value: event.target.value })
                              }
                            />
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              aria-label={t("probeTarget.removeHeader")}
                              onClick={() =>
                                setForm({
                                  ...form,
                                  headers: form.headers.filter(
                                    (_, currentIndex) => currentIndex !== index
                                  ),
                                })
                              }
                            >
                              <Trash2 />
                            </Button>
                          </div>
                        </FormField>
                      ))}
                      {fieldErrors.headers ? (
                        <p role="alert" className="text-sm text-destructive">
                          {fieldErrors.headers}
                        </p>
                      ) : null}
                    </div>
                    <div className="grid min-w-0 gap-2 sm:col-span-full">
                      <div className="flex items-center justify-between gap-2">
                        <span className="text-sm font-medium">
                          {t("probeTarget.expectHeaders")}
                        </span>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            setForm({
                              ...form,
                              expectHeaders: [
                                ...form.expectHeaders,
                                { key: "", value: "" },
                              ],
                            })
                          }
                        >
                          {t("probeTarget.addExpectHeader")}
                        </Button>
                      </div>
                      <p className="text-xs text-muted-foreground">
                        {t("probeTarget.expectHeadersHint")}
                      </p>
                      {form.expectHeaders.map((pair, index) => (
                        <FormField
                          key={index}
                          label={t("probeTarget.expectHeaderLabel", {
                            index: index + 1,
                          })}
                          error={
                            fieldErrors[`expectHeader.${index}.key`] ??
                            fieldErrors[`expectHeader.${index}.value`]
                          }
                        >
                          <div className="flex items-center gap-2">
                            <Input
                              className="min-w-0 flex-1"
                              value={pair.key}
                              placeholder={t(
                                "probeTarget.expectHeaderKeyPlaceholder"
                              )}
                              onChange={(event) =>
                                updateExpectHeader(index, {
                                  key: event.target.value,
                                })
                              }
                            />
                            <Input
                              className="min-w-0 flex-1"
                              value={pair.value}
                              placeholder={t(
                                "probeTarget.expectHeaderValuePlaceholder"
                              )}
                              onChange={(event) =>
                                updateExpectHeader(index, {
                                  value: event.target.value,
                                })
                              }
                            />
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              aria-label={t("probeTarget.removeExpectHeader")}
                              onClick={() =>
                                setForm({
                                  ...form,
                                  expectHeaders: form.expectHeaders.filter(
                                    (_, currentIndex) => currentIndex !== index
                                  ),
                                })
                              }
                            >
                              <Trash2 />
                            </Button>
                          </div>
                        </FormField>
                      ))}
                      {fieldErrors.expectHeaders ? (
                        <p role="alert" className="text-sm text-destructive">
                          {fieldErrors.expectHeaders}
                        </p>
                      ) : null}
                    </div>
                  </>
                ) : (
                  <>
                    <p className="text-xs text-muted-foreground sm:col-span-full">
                      {t("probeTarget.nonHttpRequestHint")}
                    </p>
                    <FormField
                      label={t(timeoutLabelKey(form.driver))}
                      htmlFor="probe-timeout"
                      hint={t(timeoutHintKey(form.driver))}
                      required
                      error={fieldErrors.timeoutMs}
                    >
                      <Input
                        id="probe-timeout"
                        type="number"
                        min={MIN_TIMEOUT_MS}
                        max={MAX_TIMEOUT_MS}
                        step={1}
                        inputMode="numeric"
                        value={form.timeoutMs}
                        onChange={(event) =>
                          setForm({ ...form, timeoutMs: event.target.value })
                        }
                        required
                      />
                    </FormField>
                    {form.driver === PROBE_ICMP ? (
                      <>
                        <FormField
                          label={t("probeTarget.icmpCount")}
                          htmlFor="probe-icmp-count"
                          hint={t("probeTarget.icmpCountHint")}
                          error={fieldErrors.icmpCount}
                        >
                          <Input
                            id="probe-icmp-count"
                            type="number"
                            min={0}
                            max={10}
                            step={1}
                            inputMode="numeric"
                            value={form.icmpCount}
                            onChange={(event) =>
                              setForm({
                                ...form,
                                icmpCount: event.target.value,
                              })
                            }
                          />
                        </FormField>
                        <FormField
                          label={t("probeTarget.icmpPacketSize")}
                          htmlFor="probe-icmp-size"
                          hint={t("probeTarget.icmpPacketSizeHint")}
                          error={fieldErrors.icmpPacketSize}
                        >
                          <Input
                            id="probe-icmp-size"
                            type="number"
                            min={0}
                            max={1400}
                            step={1}
                            inputMode="numeric"
                            value={form.icmpPacketSize}
                            onChange={(event) =>
                              setForm({
                                ...form,
                                icmpPacketSize: event.target.value,
                              })
                            }
                          />
                        </FormField>
                      </>
                    ) : null}
                    {form.driver === PROBE_DNS ? (
                      <>
                        <FormField
                          label={t("probeTarget.dnsRecordType")}
                          hint={t("probeTarget.dnsRecordTypeHint")}
                        >
                          <Select
                            value={form.dnsRecordType}
                            onValueChange={(value) =>
                              setForm({ ...form, dnsRecordType: value })
                            }
                          >
                            <SelectTrigger
                              aria-label={t("probeTarget.dnsRecordType")}
                            >
                              <SelectValue placeholder="A" />
                            </SelectTrigger>
                            <SelectContent>
                              {dnsRecordTypes.map((item) => (
                                <SelectItem key={item} value={item}>
                                  {item}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </FormField>
                        <FormField
                          label={t("probeTarget.dnsExpectValues")}
                          htmlFor="probe-dns-expect"
                          hint={t("probeTarget.dnsExpectValuesHint")}
                          error={fieldErrors.dnsExpectValues}
                        >
                          <Input
                            id="probe-dns-expect"
                            value={form.dnsExpectValues}
                            maxLength={MAX_DNS_EXPECT_VALUES}
                            placeholder="192\\.0\\.2\\.\\d+"
                            onChange={(event) =>
                              setForm({
                                ...form,
                                dnsExpectValues: event.target.value,
                              })
                            }
                          />
                        </FormField>
                        <FormField
                          label={t("probeTarget.dnsServer")}
                          htmlFor="probe-dns-server"
                          hint={t("probeTarget.dnsServerHint")}
                          error={fieldErrors.dnsServer}
                        >
                          <Input
                            id="probe-dns-server"
                            value={form.dnsServer}
                            maxLength={MAX_DNS_SERVER}
                            placeholder="8.8.8.8:53"
                            onChange={(event) =>
                              setForm({
                                ...form,
                                dnsServer: event.target.value,
                              })
                            }
                          />
                        </FormField>
                      </>
                    ) : null}
                    {form.driver === PROBE_TLS ? (
                      <>
                        <FormField
                          label={t("probeTarget.tlsServerName")}
                          htmlFor="probe-tls-sni"
                          hint={t("probeTarget.tlsServerNameHint")}
                          error={fieldErrors.tlsServerName}
                        >
                          <Input
                            id="probe-tls-sni"
                            value={form.tlsServerName}
                            maxLength={MAX_TLS_SERVER_NAME}
                            placeholder="api.example.com"
                            onChange={(event) =>
                              setForm({
                                ...form,
                                tlsServerName: event.target.value,
                              })
                            }
                          />
                        </FormField>
                        <div className="grid gap-3 sm:col-span-full">
                          <label className="flex items-start gap-2">
                            <Checkbox
                              className="mt-0.5"
                              checked={form.skipTlsVerify}
                              onCheckedChange={(value) =>
                                setForm({
                                  ...form,
                                  skipTlsVerify: value === true,
                                })
                              }
                            />
                            <span className="grid gap-1">
                              <span className="text-sm font-medium">
                                {t("probeTarget.skipTlsVerify")}
                              </span>
                            </span>
                          </label>
                          <label className="flex items-start gap-2">
                            <Checkbox
                              className="mt-0.5"
                              checked={form.verifyCert}
                              onCheckedChange={(value) =>
                                setForm({ ...form, verifyCert: value === true })
                              }
                            />
                            <span className="grid gap-1">
                              <span className="text-sm font-medium">
                                {t("probeTarget.verifyCert")}
                              </span>
                              <span className="text-xs text-muted-foreground">
                                {t("probeTarget.verifyCertHint")}
                              </span>
                            </span>
                          </label>
                        </div>
                      </>
                    ) : null}
                  </>
                )}
                <section className="grid min-w-0 gap-3 border-t pt-4 sm:col-span-full">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <h3 className="text-sm font-medium">
                      {t("probeTarget.trial")}
                    </h3>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={trialPending}
                      onClick={() => void runTrial()}
                    >
                      {trialPending
                        ? t("probeTarget.trialRunning")
                        : t("probeTarget.trialRun")}
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {t("probeTarget.trialHint")}
                  </p>
                  {trialResults.length === 0 ? (
                    <p className="text-xs text-muted-foreground">
                      {t("probeTarget.trialEmpty")}
                    </p>
                  ) : (
                    <Table className={TABLE_MIN_WIDTH.default}>
                      <TableHeader>
                        <TableRow>
                          <TableHead>{t("probeTarget.trialTarget")}</TableHead>
                          <TableHead className={TABLE_COLUMN_WIDTH.status}>
                            {t("probeTarget.trialResult")}
                          </TableHead>
                          {form.driver === PROBE_HTTP ? (
                            <TableHead className={TABLE_COLUMN_WIDTH.type}>
                              {t("probeTarget.statusCode")}
                            </TableHead>
                          ) : (
                            <TableHead className="min-w-[10rem]">
                              {t("probeTarget.detailColumn")}
                            </TableHead>
                          )}
                          <TableHead className={TABLE_COLUMN_WIDTH.type}>
                            {t("probeTarget.duration")}
                          </TableHead>
                          {form.driver === PROBE_HTTP ? (
                            <TableHead className={TABLE_COLUMN_WIDTH.type}>
                              {t("probeTarget.certDays")}
                            </TableHead>
                          ) : null}
                          <TableHead>{t("probeTarget.errorMessage")}</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {trialResults.map((entry) => (
                          <TableRow key={entry.spec}>
                            <TableCell className="min-w-0 break-all">
                              {entry.spec}
                            </TableCell>
                            <TableCell className={TABLE_NOWRAP}>
                              {entry.result ? (
                                <StatusBadge
                                  value={entry.result.ok ? "OK" : "FAILED"}
                                  label={
                                    entry.result.ok
                                      ? t("probeTarget.ok")
                                      : t("probeTarget.failed")
                                  }
                                />
                              ) : (
                                <StatusBadge
                                  value="FAILED"
                                  tone="danger"
                                  label={t("probeTarget.trialFailed")}
                                />
                              )}
                            </TableCell>
                            {form.driver === PROBE_HTTP ? (
                              <TableCell className={TABLE_NOWRAP}>
                                {entry.result?.statusCode || "—"}
                              </TableCell>
                            ) : (
                              <TableCell className="max-w-[16rem] min-w-0 align-top">
                                <ProbeDetailSummary
                                  driver={form.driver}
                                  result={entry.result}
                                />
                              </TableCell>
                            )}
                            <TableCell className={TABLE_NOWRAP}>
                              {entry.result
                                ? t("probeTarget.durationValue", {
                                    ms: entry.result.durationMs,
                                  })
                                : "—"}
                            </TableCell>
                            {form.driver === PROBE_HTTP ? (
                              <TableCell className={TABLE_NOWRAP}>
                                {entry.result?.certDays
                                  ? t("probeTarget.certDaysValue", {
                                      days: entry.result.certDays,
                                    })
                                  : "—"}
                              </TableCell>
                            ) : null}
                            <TableCell className="min-w-0 break-words">
                              {entry.result?.errorMessage || "—"}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  )}
                </section>
              </>
            ) : null}
          </StepForm>
        </DialogContent>
      </Dialog>
      <ConfirmAction
        open={confirm !== null}
        title={
          shownAction === "delete"
            ? t("confirm.deleteTitle")
            : t("confirm.disableTitle")
        }
        description={
          shownAction === "delete"
            ? t("confirm.deleteDescription")
            : t("confirm.disableDescription")
        }
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runConfirm()}
        onOpenChange={(value) => {
          if (!value) {
            setConfirm(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
