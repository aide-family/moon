import { Plus, RotateCcw, Save, Trash2 } from "lucide-react"
import { useState } from "react"
import ReactMarkdown from "react-markdown"
import { useTranslation } from "react-i18next"

import type { DatasourceItem } from "@/api/datasource/types"
import { testInspectionItem } from "@/api/inspection"
import type {
  InspectionAlertParams,
  InspectionItemDisplay,
  InspectionItemInput,
  InspectionItemKind,
  InspectionLogParams,
  InspectionLogSample,
  InspectionMetricParams,
  InspectionProbeParams,
  InspectionSection,
  InspectionTable,
  TestInspectionItemResult,
} from "@/api/inspection/types"
import { FormField } from "@/components/form-field"
import { MultiSelect, type MultiSelectOption } from "@/components/multi-select"
import { QueryExpressionEditor } from "@/components/query-expression-editor"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { readFormError } from "@/lib/form-error"

export interface DraftItem extends InspectionItemInput {
  key: string
}

interface InspectionItemEditorProps {
  item: DraftItem
  sources: { metrics: DatasourceItem[]; logs: DatasourceItem[] }
  probeGroups: MultiSelectOption[]
  alertLevels: MultiSelectOption[]
  dirty: boolean
  onChange: (patch: Partial<DraftItem>) => void
  onSave: () => void
  onRestore: () => void
  onRemove: () => void
}

const KINDS: InspectionItemKind[] = [
  "METRICS_QUERY",
  "LOGS_QUERY",
  "PROBE_TARGET",
  "ALERT_EVENTS",
  "NOTE",
]

const WINDOWS = [
  { value: "3600", labelKey: "inspection.window1h" },
  { value: "21600", labelKey: "inspection.window6h" },
  { value: "86400", labelKey: "inspection.window24h" },
  { value: "604800", labelKey: "inspection.window7d" },
] as const

function parseJSON<T>(raw: string, fallback: T): T {
  try {
    const value = JSON.parse(raw) as T
    return value && typeof value === "object" ? value : fallback
  } catch {
    return fallback
  }
}

function metricParams(item: DraftItem): InspectionMetricParams {
  return {
    unit: "",
    empty_is_warn: false,
    thresholds: [],
    ...parseJSON<InspectionMetricParams>(item.params, {}),
  }
}

function logParams(item: DraftItem): InspectionLogParams {
  return {
    levels: [],
    limit: 50,
    direction: "BACKWARD",
    warn_above: 0,
    ...parseJSON<InspectionLogParams>(item.params, {}),
  }
}

function probeParams(item: DraftItem): InspectionProbeParams {
  const parsed = parseJSON<Partial<InspectionProbeParams>>(item.params, {})
  return {
    metric: "AVAILABILITY",
    availability_below: 99.5,
    latency_above_ms: 500,
    cert_days_below: 14,
    warn_on_down: true,
    ...parsed,
    group_uids: parsed.group_uids ?? [],
  }
}

function alertParams(item: DraftItem): InspectionAlertParams {
  return {
    status: "FIRING",
    level_uids: [],
    labels: "",
    limit: 50,
    warn_levels: ["CRITICAL"],
    ...parseJSON<InspectionAlertParams>(item.params, {}),
  }
}

function defaultParams(kind: InspectionItemKind): string {
  switch (kind) {
    case "METRICS_QUERY":
      return JSON.stringify({ unit: "", empty_is_warn: false, thresholds: [] })
    case "LOGS_QUERY":
      return JSON.stringify({ levels: [], limit: 50, direction: "BACKWARD", warn_above: 0 })
    case "PROBE_TARGET":
      return JSON.stringify({
        group_uids: [],
        metric: "AVAILABILITY",
        availability_below: 99.5,
        latency_above_ms: 500,
        cert_days_below: 14,
        warn_on_down: true,
      })
    case "ALERT_EVENTS":
      return JSON.stringify({ status: "FIRING", level_uids: [], labels: "", limit: 50, warn_levels: ["CRITICAL"] })
    default:
      return "{}"
  }
}

function defaultDisplay(kind: InspectionItemKind): InspectionItemDisplay {
  switch (kind) {
    case "METRICS_QUERY":
      return "METRIC"
    case "LOGS_QUERY":
      return "LOGS"
    case "NOTE":
      return "TEXT"
    default:
      return "TABLE"
  }
}

function toInput(item: DraftItem): InspectionItemInput {
  return {
    kind: item.kind,
    title: item.title,
    datasourceUid: item.datasourceUid,
    query: item.query,
    windowSeconds: item.windowSeconds,
    stepSeconds: item.stepSeconds,
    display: item.display,
    params: item.params,
    enabled: item.enabled,
  }
}

function statusTone(status: string) {
  if (status === "OK") return "success" as const
  if (status === "WARN") return "warning" as const
  return "danger" as const
}

function statusLabel(t: (key: string) => string, status: string) {
  if (status === "OK") return t("inspection.statusOk")
  if (status === "WARN") return t("inspection.statusWarn")
  return t("inspection.statusFailed")
}

function logVariant(source?: DatasourceItem): "logql" | "logsql" | "json" {
  if (!source) return "logql"
  if (source.driver.includes("ELASTIC")) return "json"
  if (source.driver.includes("VLOGS")) return "logsql"
  return "logql"
}

/** 分类型检查项编辑器：左侧配置、右侧真实测试预览，保存只写计划草稿。 */
export function InspectionItemEditor({
  item,
  sources,
  probeGroups,
  alertLevels,
  dirty,
  onChange,
  onSave,
  onRestore,
  onRemove,
}: InspectionItemEditorProps) {
  const { t } = useTranslation()
  const [testing, setTesting] = useState(false)
  const [testError, setTestError] = useState("")
  const [result, setResult] = useState<TestInspectionItemResult | null>(null)
  const [section, setSection] = useState<InspectionSection | null>(null)

  function setParams(params: object) {
    onChange({ params: JSON.stringify(params) })
  }

  function changeKind(kind: InspectionItemKind) {
    if (dirty && !window.confirm(t("inspection.discardChanges"))) {
      return
    }
    const source =
      kind === "METRICS_QUERY"
        ? sources.metrics[0]?.uid ?? ""
        : kind === "LOGS_QUERY"
          ? sources.logs[0]?.uid ?? ""
          : ""
    onChange({
      kind,
      datasourceUid: source,
      query: "",
      display: defaultDisplay(kind),
      params: defaultParams(kind),
    })
  }

  async function runTest() {
    setTesting(true)
    setTestError("")
    try {
      const response = await testInspectionItem(toInput(item))
      setResult(response)
      setSection(JSON.parse(response.resultJson) as InspectionSection)
    } catch (err) {
      setResult(null)
      setSection(null)
      setTestError(readFormError(err, t("inspection.testError")).form)
    } finally {
      setTesting(false)
    }
  }

  const metrics = metricParams(item)
  const logs = logParams(item)
  const probe = probeParams(item)
  const alerts = alertParams(item)
  const selectedLogSource = sources.logs.find((source) => source.uid === item.datasourceUid)

  return (
    <div className="grid gap-4">
      <div className="flex flex-wrap items-center gap-1 rounded-lg border border-border p-1">
        {KINDS.map((kind) => (
          <Button
            key={kind}
            type="button"
            size="sm"
            variant={item.kind === kind ? "default" : "ghost"}
            onClick={() => changeKind(kind)}
          >
            {kindLabel(t, kind)}
          </Button>
        ))}
      </div>

      <div className="grid min-w-0 items-start gap-4 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="grid min-w-0 content-start gap-4">
          <FormField label={t("inspection.itemTitle")} required>
            <Input
              value={item.title}
              onChange={(event) => onChange({ title: event.target.value })}
              placeholder={t("inspection.itemTitle")}
            />
          </FormField>
          <label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={item.enabled}
              onCheckedChange={(value) => onChange({ enabled: value === true })}
            />
            {t("inspection.enabled")}
          </label>

          {item.kind === "METRICS_QUERY" ? (
            <MetricsConfig
              item={item}
              params={metrics}
              sources={sources.metrics}
              onChange={onChange}
              setParams={setParams}
            />
          ) : null}
          {item.kind === "LOGS_QUERY" ? (
            <LogsConfig
              item={item}
              params={logs}
              sources={sources.logs}
              selected={selectedLogSource}
              onChange={onChange}
              setParams={setParams}
            />
          ) : null}
          {item.kind === "PROBE_TARGET" ? (
            <ProbeConfig
              item={item}
              params={probe}
              groups={probeGroups}
              onChange={onChange}
              setParams={setParams}
            />
          ) : null}
          {item.kind === "ALERT_EVENTS" ? (
            <AlertConfig
              item={item}
              params={alerts}
              levels={alertLevels}
              onChange={onChange}
              setParams={setParams}
            />
          ) : null}
          {item.kind === "NOTE" ? (
            <FormField label={t("inspection.noteContent")} hint={t("inspection.markdownHint")} required>
              <Textarea
                rows={7}
                className="font-mono text-xs"
                value={item.query}
                onChange={(event) => onChange({ query: event.target.value })}
              />
            </FormField>
          ) : null}

          <div className="flex flex-wrap items-center gap-2 border-t border-border pt-3">
            <span className={`text-xs ${dirty ? "text-warning" : "text-muted-foreground"}`}>
              {dirty ? t("inspection.itemDirty") : t("inspection.itemSaved")}
            </span>
            <span className="flex-1" />
            <Button type="button" size="sm" variant="outline" onClick={onRestore}>
              <RotateCcw data-icon="inline-start" />
              {t("inspection.restoreItem")}
            </Button>
            <Button type="button" size="sm" onClick={onSave}>
              <Save data-icon="inline-start" />
              {t("inspection.saveItem")}
            </Button>
            <Button type="button" size="sm" variant="ghost" className="text-destructive" onClick={onRemove}>
              <Trash2 data-icon="inline-start" />
              {t("inspection.remove")}
            </Button>
          </div>
        </div>

        <div className="grid min-w-0 gap-3 self-start rounded-xl border border-border bg-muted/20 p-3 lg:sticky lg:top-0">
          <div className="flex items-center gap-2">
            <span className="text-[11px] font-semibold tracking-wide text-muted-foreground uppercase">
              {t("inspection.preview")}
            </span>
            <span className="flex-1" />
            {item.kind === "NOTE" ? null : (
              <Button type="button" size="sm" variant="outline" onClick={() => void runTest()} disabled={testing}>
                {testing ? t("inspection.testing") : t("inspection.test")}
              </Button>
            )}
          </div>
          <div className="grid min-h-48 gap-3" aria-live="polite">
            {item.kind === "NOTE" ? <MarkdownPreview content={item.query} /> : null}
            {item.kind !== "NOTE" && !result && !testError ? (
              <div className="m-auto grid max-w-xs gap-1 text-center text-xs text-muted-foreground">
                <span className="font-medium text-foreground">{t("inspection.untested")}</span>
                <span>{t("inspection.untestedHint")}</span>
              </div>
            ) : null}
            {testError ? <p className="text-sm text-destructive">{testError}</p> : null}
            {result && section ? (
              <>
                <div className="flex flex-wrap items-center gap-2">
                  <StatusBadge value={section.status} label={statusLabel(t, section.status)} tone={statusTone(section.status)} />
                  <span className="text-xs text-muted-foreground">
                    {Number(result.elapsedMs)}ms
                  </span>
                </div>
                {section.error ? <p className="text-sm text-destructive">{section.error}</p> : null}
                <SectionPreview section={section} />
              </>
            ) : null}
          </div>
          <p className="text-[11px] leading-relaxed text-muted-foreground">{t("inspection.testHint")}</p>
        </div>
      </div>
    </div>
  )
}

function kindLabel(t: (key: string) => string, kind: InspectionItemKind): string {
  switch (kind) {
    case "METRICS_QUERY":
      return t("inspection.kindMetrics")
    case "LOGS_QUERY":
      return t("inspection.kindLogs")
    case "PROBE_TARGET":
      return t("inspection.kindProbe")
    case "ALERT_EVENTS":
      return t("inspection.kindAlert")
    default:
      return t("inspection.kindNote")
  }
}

function displayLabel(t: (key: string) => string, display: InspectionItemDisplay): string {
  switch (display) {
    case "METRIC":
      return t("inspection.displayMetric")
    case "TREND":
      return t("inspection.displayTrend")
    case "LOGS":
      return t("inspection.displayLogs")
    case "TEXT":
      return t("inspection.displayText")
    default:
      return t("inspection.displayTable")
  }
}

function WindowField({
  item,
  onChange,
}: {
  item: DraftItem
  onChange: (patch: Partial<DraftItem>) => void
}) {
  const { t } = useTranslation()
  return (
    <FormField label={t("inspection.window")}>
      <Select value={String(item.windowSeconds)} onValueChange={(value) => onChange({ windowSeconds: Number(value) })}>
        <SelectTrigger>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {WINDOWS.map((window) => (
            <SelectItem key={window.value} value={window.value}>
              {t(window.labelKey)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </FormField>
  )
}

function DatasourceField({
  item,
  sources,
  type,
  onChange,
}: {
  item: DraftItem
  sources: DatasourceItem[]
  type: "metrics" | "logs"
  onChange: (patch: Partial<DraftItem>) => void
}) {
  const { t } = useTranslation()
  return (
    <FormField label={t("inspection.datasource")} required>
      <Select value={item.datasourceUid} onValueChange={(value) => onChange({ datasourceUid: value })}>
        <SelectTrigger>
          <SelectValue placeholder={t("inspection.selectDatasource")} />
        </SelectTrigger>
        <SelectContent>
          {sources.map((source) => (
            <SelectItem key={source.uid} value={source.uid}>
              {source.name} · {source.driver}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {sources.length === 0 ? (
        <p className="text-xs text-muted-foreground">
          {t(type === "metrics" ? "inspection.noDatasource" : "inspection.noDatasource")}
        </p>
      ) : null}
    </FormField>
  )
}

function MetricsConfig({
  item,
  params,
  sources,
  onChange,
  setParams,
}: {
  item: DraftItem
  params: InspectionMetricParams
  sources: DatasourceItem[]
  onChange: (patch: Partial<DraftItem>) => void
  setParams: (params: object) => void
}) {
  const { t } = useTranslation()
  const thresholds = params.thresholds ?? []
  function updateThreshold(index: number, patch: Partial<NonNullable<InspectionMetricParams["thresholds"]>[number]>) {
    const next = thresholds.map((threshold, current) => (current === index ? { ...threshold, ...patch } : threshold))
    setParams({ ...params, thresholds: next })
  }
  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2">
        <DatasourceField item={item} sources={sources} type="metrics" onChange={onChange} />
        <WindowField item={item} onChange={onChange} />
      </div>
      <FormField label={t("inspection.query")} required>
        <div className="overflow-hidden rounded-lg border border-input">
          <div className="flex items-center gap-2 border-b border-border bg-muted/40 px-3 py-1.5">
            <span className="rounded border border-primary/30 bg-primary-soft px-1.5 font-mono text-[10px] text-primary-ink">
              PromQL
            </span>
            <span className="flex-1" />
            <span className="text-xs text-muted-foreground">{t("inspection.queryMode")}: PromQL</span>
          </div>
          <QueryExpressionEditor
            value={item.query}
            onChange={(value) => onChange({ query: value })}
            datasourceUid={item.datasourceUid}
            variant="promql"
            minHeight={150}
            placeholder="histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))"
          />
        </div>
      </FormField>
      <div className="grid gap-4 sm:grid-cols-3">
        <FormField label={t("inspection.step")}>
          <Select value={String(item.stepSeconds)} onValueChange={(value) => onChange({ stepSeconds: Number(value) })}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[15, 30, 60, 300].map((value) => (
                <SelectItem key={value} value={String(value)}>
                  {value}s
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </FormField>
        <FormField label={t("inspection.display")}>
          <Select value={item.display} onValueChange={(value) => onChange({ display: value as InspectionItemDisplay })}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(["METRIC", "TABLE", "TREND"] as InspectionItemDisplay[]).map((display) => (
                <SelectItem key={display} value={display}>
                  {displayLabel(t, display)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </FormField>
        <FormField label={t("inspection.unit")}>
          <Input value={params.unit ?? ""} onChange={(event) => setParams({ ...params, unit: event.target.value })} placeholder="s / % / MB" />
        </FormField>
      </div>
      <label className="flex items-center gap-2 text-sm">
        <Checkbox
          checked={params.empty_is_warn === true}
          onCheckedChange={(value) => setParams({ ...params, empty_is_warn: value === true })}
        />
        {t("inspection.emptyIsWarn")}
      </label>
      <div className="grid gap-2">
        <span className="text-sm font-medium">{t("inspection.thresholds")}</span>
        {thresholds.map((threshold, index) => (
          <div key={index} className="grid grid-cols-[5rem_7rem_7rem_2rem] items-center gap-2">
            <Select value={threshold.op} onValueChange={(value) => updateThreshold(index, { op: value })}>
              <SelectTrigger aria-label={t("inspection.thresholdValue")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[">", ">=", "<", "<=", "==", "!="].map((op) => (
                  <SelectItem key={op} value={op}>
                    {op}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Input
              type="number"
              value={threshold.value}
              onChange={(event) => updateThreshold(index, { value: Number(event.target.value) })}
              aria-label={t("inspection.thresholdValue")}
            />
            <Select value={threshold.status} onValueChange={(value) => updateThreshold(index, { status: value as "WARN" | "FAILED" })}>
              <SelectTrigger aria-label={t("inspection.warning")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="WARN">{t("inspection.warning")}</SelectItem>
                <SelectItem value="FAILED">{t("inspection.failedStatus")}</SelectItem>
              </SelectContent>
            </Select>
            <Button
              type="button"
              size="icon-sm"
              variant="ghost"
              aria-label={t("inspection.remove")}
              onClick={() => setParams({ ...params, thresholds: thresholds.filter((_, current) => current !== index) })}
            >
              ×
            </Button>
          </div>
        ))}
        <Button
          type="button"
          size="sm"
          variant="outline"
          className="w-fit"
          onClick={() => setParams({ ...params, thresholds: [...thresholds, { op: ">", value: 0, status: "WARN" }] })}
        >
          <Plus data-icon="inline-start" />
          {t("inspection.addThreshold")}
        </Button>
      </div>
    </>
  )
}

function LogsConfig({
  item,
  params,
  sources,
  selected,
  onChange,
  setParams,
}: {
  item: DraftItem
  params: InspectionLogParams
  sources: DatasourceItem[]
  selected?: DatasourceItem
  onChange: (patch: Partial<DraftItem>) => void
  setParams: (params: object) => void
}) {
  const { t } = useTranslation()
  const levels = ["ERROR", "WARN", "INFO", "DEBUG"]
  function toggleLevel(level: string) {
    const current = params.levels ?? []
    setParams({
      ...params,
      levels: current.includes(level) ? current.filter((value) => value !== level) : [...current, level],
    })
  }
  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2">
        <DatasourceField item={item} sources={sources} type="logs" onChange={onChange} />
        <WindowField item={item} onChange={onChange} />
      </div>
      <FormField label={t("inspection.query")} required>
        <div className="overflow-hidden rounded-lg border border-input">
          <div className="flex items-center gap-2 border-b border-border bg-muted/40 px-3 py-1.5">
            <span className="rounded border border-primary/30 bg-primary-soft px-1.5 font-mono text-[10px] text-primary-ink">
              {selected?.driver ?? "LogQL"}
            </span>
            <span className="flex-1" />
            <span className="text-xs text-muted-foreground">{t("inspection.queryMode")}</span>
          </div>
          <QueryExpressionEditor
            value={item.query}
            onChange={(value) => onChange({ query: value })}
            datasourceUid={item.datasourceUid}
            variant={logVariant(selected)}
            minHeight={150}
            placeholder={logVariant(selected) === "json" ? '{ "query": { "match_all": {} } }' : '{app="gateway"} |= "ERROR"'}
          />
        </div>
      </FormField>
      <div className="grid gap-2">
        <span className="text-sm font-medium">{t("inspection.logLevels")}</span>
        <div className="flex flex-wrap gap-2">
          {levels.map((level) => (
            <Button
              key={level}
              type="button"
              size="sm"
              variant={(params.levels ?? []).includes(level) ? "secondary" : "outline"}
              onClick={() => toggleLevel(level)}
            >
              {level}
            </Button>
          ))}
        </div>
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        <FormField label={t("inspection.limit")}>
          <Select value={String(params.limit ?? 50)} onValueChange={(value) => setParams({ ...params, limit: Number(value) })}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[20, 50, 100, 200].map((value) => (
                <SelectItem key={value} value={String(value)}>
                  {value}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </FormField>
        <FormField label={t("inspection.direction")}>
          <Select value={params.direction ?? "BACKWARD"} onValueChange={(value) => setParams({ ...params, direction: value as "BACKWARD" | "FORWARD" })}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="BACKWARD">{t("inspection.newestFirst")}</SelectItem>
              <SelectItem value="FORWARD">{t("inspection.oldestFirst")}</SelectItem>
            </SelectContent>
          </Select>
        </FormField>
        <FormField label={t("inspection.warnAbove")}>
          <Input
            type="number"
            min={0}
            value={params.warn_above ?? 0}
            onChange={(event) => setParams({ ...params, warn_above: Number(event.target.value) })}
          />
        </FormField>
      </div>
    </>
  )
}

function ProbeConfig({
  item,
  params,
  groups,
  onChange,
  setParams,
}: {
  item: DraftItem
  params: InspectionProbeParams
  groups: MultiSelectOption[]
  onChange: (patch: Partial<DraftItem>) => void
  setParams: (params: object) => void
}) {
  const { t } = useTranslation()
  const metric = params.metric ?? "AVAILABILITY"
  return (
    <>
      <FormField label={t("inspection.probeGroups")} required>
        <MultiSelect
          values={params.group_uids ?? []}
          options={groups}
          onChange={(values) => setParams({ ...params, group_uids: values })}
          placeholder={t("inspection.selectTargetGroup")}
          searchPlaceholder={t("multiSelect.searchPlaceholder")}
          emptyText={t("inspection.noProbeGroups")}
          allLabel={t("multiSelect.all")}
        />
        {groups.length === 0 ? <p className="text-xs text-muted-foreground">{t("inspection.noProbeGroups")}</p> : null}
      </FormField>
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField label={t("inspection.probeMetric")}>
          <Select value={metric} onValueChange={(value) => setParams({ ...params, metric: value })}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="AVAILABILITY">{t("inspection.probeAvailability")}</SelectItem>
              <SelectItem value="LATENCY">{t("inspection.probeLatency")}</SelectItem>
              <SelectItem value="P95">{t("inspection.probeP95")}</SelectItem>
              <SelectItem value="CERT_DAYS">{t("inspection.probeCertDays")}</SelectItem>
            </SelectContent>
          </Select>
        </FormField>
        <WindowField item={item} onChange={onChange} />
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        {metric === "AVAILABILITY" ? (
          <FormField label={t("inspection.availabilityBelow")}>
            <Input
              type="number"
              min={0}
              max={100}
              step="0.1"
              value={params.availability_below ?? 99.5}
              onChange={(event) => setParams({ ...params, availability_below: Number(event.target.value) })}
            />
          </FormField>
        ) : null}
        {metric === "LATENCY" || metric === "P95" ? (
          <FormField label={t("inspection.latencyAbove")}>
            <Input
              type="number"
              min={0}
              value={params.latency_above_ms ?? 500}
              onChange={(event) => setParams({ ...params, latency_above_ms: Number(event.target.value) })}
            />
          </FormField>
        ) : null}
        {metric === "CERT_DAYS" ? (
          <FormField label={t("inspection.certBelow")}>
            <Input
              type="number"
              min={0}
              value={params.cert_days_below ?? 14}
              onChange={(event) => setParams({ ...params, cert_days_below: Number(event.target.value) })}
            />
          </FormField>
        ) : null}
        <label className="flex items-center gap-2 self-end pb-2 text-sm">
          <Checkbox
            checked={params.warn_on_down !== false}
            onCheckedChange={(value) => setParams({ ...params, warn_on_down: value === true })}
          />
          {t("inspection.warnOnDown")}
        </label>
      </div>
    </>
  )
}

function AlertConfig({
  item,
  params,
  levels,
  onChange,
  setParams,
}: {
  item: DraftItem
  params: InspectionAlertParams
  levels: MultiSelectOption[]
  onChange: (patch: Partial<DraftItem>) => void
  setParams: (params: object) => void
}) {
  const { t } = useTranslation()
  function toggleWarn(level: string) {
    const current = params.warn_levels ?? []
    setParams({
      ...params,
      warn_levels: current.includes(level) ? current.filter((value) => value !== level) : [...current, level],
    })
  }
  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField label={t("inspection.alertStatus")}>
          <div className="flex gap-1">
            {[
              ["FIRING", t("inspection.unresolved")],
              ["ALL", t("inspection.allEvents")],
            ].map(([value, label]) => (
              <Button
                key={value}
                type="button"
                size="sm"
                variant={(params.status ?? "FIRING") === value ? "default" : "outline"}
                onClick={() => setParams({ ...params, status: value })}
              >
                {label}
              </Button>
            ))}
          </div>
        </FormField>
        <WindowField item={item} onChange={onChange} />
      </div>
      <FormField label={t("inspection.alertLevels")}>
        <MultiSelect
          values={params.level_uids ?? []}
          options={levels}
          onChange={(values) => setParams({ ...params, level_uids: values })}
          placeholder={t("inspection.alertLevels")}
          searchPlaceholder={t("multiSelect.searchPlaceholder")}
          emptyText={t("inspection.noAlertLevels")}
          allLabel={t("multiSelect.all")}
        />
        {levels.length === 0 ? <p className="text-xs text-muted-foreground">{t("inspection.noAlertLevels")}</p> : null}
      </FormField>
      <FormField label={t("inspection.labels")} hint={t("inspection.labelsHint")}>
        <Input
          value={params.labels ?? ""}
          onChange={(event) => setParams({ ...params, labels: event.target.value })}
          placeholder='service="order",env=prod'
        />
      </FormField>
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField label={t("inspection.limit")}>
          <Select value={String(params.limit ?? 50)} onValueChange={(value) => setParams({ ...params, limit: Number(value) })}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[20, 50, 100, 200].map((value) => (
                <SelectItem key={value} value={String(value)}>
                  {value}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </FormField>
        <FormField label={t("inspection.warnLevels")}>
          <div className="flex flex-wrap gap-2">
            {["CRITICAL", "WARNING"].map((level) => (
              <Button
                key={level}
                type="button"
                size="sm"
                variant={(params.warn_levels ?? []).includes(level) ? "secondary" : "outline"}
                onClick={() => toggleWarn(level)}
              >
                {t(level === "CRITICAL" ? "inspection.warnOnCritical" : "inspection.warnOnWarning")}
              </Button>
            ))}
          </div>
        </FormField>
      </div>
    </>
  )
}

function MarkdownPreview({ content }: { content: string }) {
  if (!content.trim()) {
    return <p className="text-xs text-muted-foreground">—</p>
  }
  return (
    <div className="prose prose-sm max-w-none text-sm dark:prose-invert">
      <ReactMarkdown>{content}</ReactMarkdown>
    </div>
  )
}

function SectionPreview({ section }: { section: InspectionSection }) {
  const { t } = useTranslation()
  return (
    <div className="grid min-w-0 gap-3">
      {section.metrics && section.metrics.length > 0 ? (
        <div className="grid gap-2 sm:grid-cols-2">
          {section.metrics.map((metric) => (
            <div key={metric.name} className="rounded-lg border border-border bg-background p-2">
              <p className="truncate text-[11px] text-muted-foreground">{metric.name}</p>
              <p className="font-mono text-base font-semibold">
                {metric.value}
                {metric.unit ? <span className="ml-1 text-xs font-normal text-muted-foreground">{metric.unit}</span> : null}
              </p>
            </div>
          ))}
        </div>
      ) : null}
      {section.series && section.series.length > 0 && section.series[0].points.length > 1 ? (
        <Sparkline series={section.series[0].points} />
      ) : null}
      {section.table ? <ResultTable table={section.table} /> : null}
      {section.logs && section.logs.length > 0 ? (
        <div className="overflow-hidden rounded-lg border border-border">
          {section.logs.map((log, index) => (
            <LogRow key={`${log.time}-${index}`} log={log} />
          ))}
        </div>
      ) : null}
      {section.alerts && section.alerts.length > 0 ? (
        <div className="overflow-x-auto rounded-lg border border-border">
          <table className="w-full text-left text-xs">
            <thead className="text-muted-foreground">
              <tr className="border-b border-border">
                <th className="p-2 font-medium">{t("inspection.itemTitle")}</th>
                <th className="p-2 font-medium">{t("inspection.alertLevels")}</th>
                <th className="p-2 font-medium">{t("inspection.alertStatus")}</th>
              </tr>
            </thead>
            <tbody>
              {section.alerts.map((alert, index) => (
                <tr key={`${alert.title}-${index}`} className="border-b border-border/60 last:border-0">
                  <td className="max-w-[16rem] p-2 break-words">{alert.title}</td>
                  <td className="p-2 whitespace-nowrap">{alert.level}</td>
                  <td className="p-2 whitespace-nowrap">{alert.status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
      {section.note ? <p className="text-xs leading-relaxed text-muted-foreground">{section.note}</p> : null}
      {section.raw ? (
        <pre className="max-h-48 overflow-auto rounded-lg bg-muted p-2 font-mono text-[10px] whitespace-pre-wrap">{section.raw}</pre>
      ) : null}
    </div>
  )
}

function ResultTable({ table }: { table: InspectionTable }) {
  return (
    <div className="overflow-x-auto rounded-lg border border-border">
      <table className="w-full text-left text-xs">
        <thead className="text-muted-foreground">
          <tr className="border-b border-border">
            {table.columns.map((column) => (
              <th key={column} className="p-2 font-medium whitespace-nowrap">
                {column}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {table.rows.map((row, index) => (
            <tr key={index} className="border-b border-border/60 last:border-0">
              {row.map((cell, cellIndex) => (
                <td key={cellIndex} className="p-2 break-words">
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function LogRow({ log }: { log: InspectionLogSample }) {
  return (
    <div className="grid grid-cols-[4.5rem_3.5rem_minmax(0,1fr)] gap-2 border-b border-border/60 p-2 font-mono text-[10px] last:border-0">
      <span className="text-muted-foreground">{log.time}</span>
      <span className={log.level === "ERROR" ? "text-danger" : log.level === "WARN" ? "text-warning" : "text-muted-foreground"}>
        {log.level}
      </span>
      <span className="break-words">{log.message}</span>
    </div>
  )
}

function Sparkline({ series }: { series: Array<{ value: number }> }) {
  const width = 320
  const height = 64
  const values = series.map((point) => point.value)
  const min = Math.min(...values)
  const max = Math.max(...values)
  const span = max - min || 1
  const path = values
    .map((value, index) => {
      const x = (index / Math.max(values.length - 1, 1)) * width
      const y = height - ((value - min) / span) * (height - 8) - 4
      return `${index === 0 ? "M" : "L"}${x.toFixed(1)} ${y.toFixed(1)}`
    })
    .join(" ")
  return (
    <svg viewBox={`0 0 ${width} ${height}`} className="h-16 w-full text-primary" aria-hidden="true">
      <path d={path} fill="none" stroke="currentColor" strokeWidth="2" />
    </svg>
  )
}
