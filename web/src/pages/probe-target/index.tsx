import {
  ChevronDown,
  ChevronRight,
  EllipsisVertical,
  Trash2,
} from "lucide-react"
import { Fragment, useEffect, useState } from "react"
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
import { FormField } from "@/components/form-field"
import { LabelTags } from "@/components/label-tags"
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
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import { enumLabel } from "@/lib/enum-label"
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

// 一期只有 HTTP；TCP / ICMP 只作为灰态选项展示。
const probeTypes = ["TCP", "ICMP"]
const PROBE_HTTP = "PROBE_HTTP"
const MAX_TARGETS = 50
const MAX_TARGET_SPEC = 2048
const MAX_TIMEOUT_MS = 60000
const MIN_TIMEOUT_MS = 100
const MAX_BODY_KEYWORD = 256
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

/** 组内去重用的比较键：scheme/host 大小写归一，路径与查询串保持原样。 */
function targetKey(value: string): string {
  try {
    const url = new URL(value.trim())
    const port = url.port ? `:${url.port}` : ""
    return `${url.protocol}//${url.hostname.toLowerCase()}${port}${url.pathname}${url.search}`
  } catch {
    return value.trim()
  }
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
  return {
    method: form.method,
    expectStatus: form.expectStatus.trim(),
    timeoutMs: Number(form.timeoutMs || "0"),
    followRedirects: form.followRedirects,
    skipTlsVerify: form.skipTlsVerify,
    verifyCert: form.verifyCert,
    headers: serializeHeaders(form.headers),
    bodyKeyword: form.bodyKeyword.trim(),
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
  const [draftName, setDraftName] = useState(query.name)
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
    setForm(emptyDraft())
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
    if (form.driver !== PROBE_HTTP) {
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
      if (!isHttpTarget(spec)) {
        checks.push([`target.${index}.spec`, t("probeTarget.invalidTarget")])
        return
      }
      const key = targetKey(spec)
      if (seen.has(key)) {
        checks.push([`target.${index}.spec`, t("probeTarget.duplicateTarget")])
      }
      seen.add(key)
    })
    return collectFieldErrors(checks)
  }

  function validateRequestStep(): FieldErrors {
    const checks: Array<[string, string | null]> = []
    if (form.method !== "GET" && form.method !== "HEAD") {
      checks.push(["method", t("form.required")])
    }
    if (expectStatusInvalid(form.expectStatus)) {
      checks.push(["expectStatus", t("probeTarget.expectStatusInvalid")])
    }
    if (timeoutInvalid(form.timeoutMs)) {
      checks.push(["timeoutMs", t("probeTarget.timeoutInvalid")])
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
        .map((spec) => targetKey(spec))
    )
    const additions: TargetDraft[] = []
    for (let index = 0; index < lines.length; index += 1) {
      const spec = lines[index]
      if (!isHttpTarget(spec)) {
        setFieldErrors({
          ...fieldErrors,
          bulk: t("probeTarget.bulkPasteInvalidLine", { line: index + 1 }),
        })
        return
      }
      const key = targetKey(spec)
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
          const response = await probeTrial({ spec, request })
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

  const targets = current?.targets ?? []

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
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draftName })
        }}
      >
        <Input
          className="max-w-sm"
          value={draftName}
          placeholder={t("probeTarget.namePlaceholder")}
          onChange={(event) => setDraftName(event.target.value)}
        />
        <Select
          value={query.type || "ALL"}
          onValueChange={(value) =>
            query.update({ type: value === "ALL" ? "" : value })
          }
        >
          <SelectTrigger className="w-40" aria-label={t("probeTarget.type")}>
            <SelectValue placeholder={t("probeTarget.typePlaceholder")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ALL">{t("probeTarget.typeAll")}</SelectItem>
            <SelectItem value={PROBE_HTTP}>
              {enumLabel(t, PROBE_HTTP)}
            </SelectItem>
          </SelectContent>
        </Select>
        <Button type="submit" variant="outline">
          {t("action.search")}
        </Button>
      </form>
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
                          <TableHead className={TABLE_COLUMN_WIDTH.type}>
                            {t("probeTarget.statusCode")}
                          </TableHead>
                          <TableHead className={TABLE_COLUMN_WIDTH.type}>
                            {t("probeTarget.duration")}
                          </TableHead>
                          <TableHead className={TABLE_COLUMN_WIDTH.type}>
                            {t("probeTarget.certDays")}
                          </TableHead>
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
                            <TableCell className={TABLE_NOWRAP}>
                              {target.lastResult?.statusCode || "—"}
                            </TableCell>
                            <TableCell className={TABLE_NOWRAP}>
                              {target.lastResult
                                ? t("probeTarget.durationValue", {
                                    ms: target.lastResult.durationMs,
                                  })
                                : "—"}
                            </TableCell>
                            <TableCell className={TABLE_NOWRAP}>
                              {target.lastResult?.certDays
                                ? t("probeTarget.certDaysValue", {
                                    days: target.lastResult.certDays,
                                  })
                                : "—"}
                            </TableCell>
                          </TableRow>
                          {expanded ? (
                            <TableRow
                              id={`probe-records-${targetKey}`}
                              className="bg-muted/20 hover:bg-muted/20"
                            >
                              <TableCell colSpan={7} className="min-w-0">
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
                                          <TableHead
                                            className={TABLE_COLUMN_WIDTH.type}
                                          >
                                            {t("probeTarget.statusCode")}
                                          </TableHead>
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
                                            <TableCell className={TABLE_NOWRAP}>
                                              {item.statusCode || "—"}
                                            </TableCell>
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
                    onValueChange={(value) =>
                      setForm({ ...form, driver: value })
                    }
                  >
                    <SelectTrigger aria-label={t("probeTarget.type")}>
                      <SelectValue
                        placeholder={t("probeTarget.typePlaceholder")}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value={PROBE_HTTP}>
                        {enumLabel(t, PROBE_HTTP)}
                      </SelectItem>
                      {probeTypes.map((item) => (
                        <SelectItem key={item} value={item} disabled>
                          {`${enumLabel(t, item)}（${t("probeTarget.comingSoon")}）`}
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
                  {t("probeTarget.targetsHint")}
                </p>
                {bulkOpen ? (
                  <div className="grid gap-2 rounded-lg border p-3 sm:col-span-full">
                    <p className="text-xs text-muted-foreground">
                      {t("probeTarget.bulkPasteHint")}
                    </p>
                    <Textarea
                      value={bulkText}
                      placeholder={t("probeTarget.bulkPastePlaceholder")}
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
                        placeholder={t("probeTarget.targetSpecPlaceholder")}
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
                          <TableHead className={TABLE_COLUMN_WIDTH.type}>
                            {t("probeTarget.statusCode")}
                          </TableHead>
                          <TableHead className={TABLE_COLUMN_WIDTH.type}>
                            {t("probeTarget.duration")}
                          </TableHead>
                          <TableHead className={TABLE_COLUMN_WIDTH.type}>
                            {t("probeTarget.certDays")}
                          </TableHead>
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
                            <TableCell className={TABLE_NOWRAP}>
                              {entry.result?.statusCode || "—"}
                            </TableCell>
                            <TableCell className={TABLE_NOWRAP}>
                              {entry.result
                                ? t("probeTarget.durationValue", {
                                    ms: entry.result.durationMs,
                                  })
                                : "—"}
                            </TableCell>
                            <TableCell className={TABLE_NOWRAP}>
                              {entry.result?.certDays
                                ? t("probeTarget.certDaysValue", {
                                    days: entry.result.certDays,
                                  })
                                : "—"}
                            </TableCell>
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
