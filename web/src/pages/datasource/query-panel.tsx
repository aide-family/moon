import { editor as monacoEditor } from "monaco-editor"
import { Check, Copy } from "lucide-react"
import { useEffect, useMemo, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getDatasourceStatus } from "@/api/datasource"
import { proxyQuery, queryMetric, queryMetricRange } from "@/api/metric-query"
import type { DatasourceItem } from "@/api/datasource/types"
import { FormField } from "@/components/form-field"
import { MetricGraph } from "@/components/metric-graph"
import { PromQLEditor } from "@/components/query-expression-editor"
import { useTheme } from "@/components/theme-provider"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { cn } from "@/lib/class-names"
import { failureMessage } from "@/lib/form-error"
import {
  collectLabelKeys,
  extractPrometheusResultData,
  parsePrometheusBody,
  toTableRows,
  type PrometheusApiResponse,
} from "@/pages/datasource/prometheus"
import { ElasticsearchPanel } from "@/pages/datasource/elasticsearch-panel"
import { LokiPanel } from "@/pages/datasource/loki-panel"
import { resolveRange, type RangePreset } from "@/pages/datasource/range"
import { VlogsPanel } from "@/pages/datasource/vlogs-panel"

const areaClass =
  "min-h-24 w-full rounded-lg border border-input bg-transparent px-2.5 py-1 text-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"

type ResultTab = "table" | "graph" | "json"

function formatTime(ts: number): string {
  if (!ts) {
    return "—"
  }
  const date = new Date(ts * 1000)
  const pad = (value: number) => String(value).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

function stickyFirstKey(labelKeys: string[]): string | null {
  if (labelKeys.includes("__name__")) {
    return "__name__"
  }
  return labelKeys[0] ?? null
}

function ResultTable({ response }: { response: PrometheusApiResponse | null }) {
  const { t } = useTranslation()
  const rows = useMemo(() => (response ? toTableRows(response) : []), [response])
  const labelKeys = useMemo(() => collectLabelKeys(rows), [rows])
  const firstKey = stickyFirstKey(labelKeys)
  const otherKeys = firstKey ? labelKeys.filter((key) => key !== firstKey) : labelKeys
  const stickyHead =
    "sticky top-0 z-20 bg-background shadow-[inset_0_-1px_0_0_hsl(var(--border))]"
  const stickyCorner =
    "sticky left-0 top-0 z-30 bg-background shadow-[inset_-1px_0_0_0_hsl(var(--border)),inset_0_-1px_0_0_hsl(var(--border))]"
  const stickyCell =
    "sticky left-0 z-10 bg-background shadow-[inset_-1px_0_0_0_hsl(var(--border))]"

  if (rows.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("datasource.queryEmpty")}</p>
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-lg border">
      <Table className="w-max min-w-full">
        <TableHeader>
          <TableRow>
            {firstKey ? (
              <TableHead className={cn(stickyCorner, "min-w-[10rem] whitespace-nowrap")}>
                {firstKey}
              </TableHead>
            ) : (
              <TableHead className={cn(stickyCorner, "min-w-[10rem] whitespace-nowrap")}>
                {t("datasource.resultSeries")}
              </TableHead>
            )}
            {otherKeys.map((key) => (
              <TableHead key={key} className={cn(stickyHead, "whitespace-nowrap")}>
                {key}
              </TableHead>
            ))}
            <TableHead className={cn(stickyHead, "whitespace-nowrap")}>
              {t("datasource.resultValue")}
            </TableHead>
            <TableHead className={cn(stickyHead, "whitespace-nowrap")}>
              {t("datasource.resultTime")}
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.key}>
              <TableCell className={cn(stickyCell, "whitespace-nowrap font-mono text-xs")}>
                {firstKey ? (row.labels[firstKey] ?? "—") : row.series}
              </TableCell>
              {otherKeys.map((key) => (
                <TableCell key={key} className="whitespace-nowrap font-mono text-xs">
                  {row.labels[key] ?? "—"}
                </TableCell>
              ))}
              <TableCell className="whitespace-nowrap font-mono text-xs">{row.value}</TableCell>
              <TableCell className="whitespace-nowrap text-xs text-muted-foreground">
                {formatTime(row.timestamp)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

function ResultJson({
  response,
  raw,
  active,
}: {
  response: PrometheusApiResponse | null
  raw: string
  active: boolean
}) {
  const { t } = useTranslation()
  const { theme } = useTheme()
  const content = useMemo(() => {
    if (response) {
      return JSON.stringify(extractPrometheusResultData(response), null, 2)
    }
    return raw
  }, [raw, response])

  const [copied, setCopied] = useState(false)
  const [html, setHtml] = useState("")
  const copiedTimer = useRef<number | null>(null)
  const editorTheme =
    theme === "dark" || (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)
      ? "vs-dark"
      : "vs"

  useEffect(() => {
    let active = true
    async function paint() {
      if (!content) {
        if (active) {
          setHtml("")
        }
        return
      }
      monacoEditor.setTheme(editorTheme)
      const colored = await monacoEditor.colorize(content, "json", { tabSize: 2 })
      if (active) {
        setHtml(colored)
      }
    }
    void paint()
    return () => {
      active = false
    }
  }, [content, editorTheme])

  useEffect(() => {
    return () => {
      if (copiedTimer.current !== null) {
        window.clearTimeout(copiedTimer.current)
      }
    }
  }, [])

  async function copyJson() {
    if (!content) {
      return
    }
    try {
      await navigator.clipboard.writeText(content)
      setCopied(true)
      if (copiedTimer.current !== null) {
        window.clearTimeout(copiedTimer.current)
      }
      copiedTimer.current = window.setTimeout(() => setCopied(false), 2000)
      toast.success(t("action.copied"))
    } catch {
      toast.error(t("action.copyFailed"))
    }
  }

  if (!content) {
    return <p className="text-sm text-muted-foreground">{t("datasource.queryEmpty")}</p>
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border">
      <div className="flex shrink-0 items-center justify-end border-b px-2 py-1.5">
        <Button type="button" variant="outline" size="icon" aria-label={t("action.copy")} onClick={() => void copyJson()}>
          {copied ? <Check /> : <Copy />}
        </Button>
      </div>
      <pre className="min-h-0 flex-1 overflow-auto bg-background p-3 font-mono text-xs">
        <code dangerouslySetInnerHTML={{ __html: active ? html : "" }} />
      </pre>
    </div>
  )
}

export function QueryPanel({ datasource }: { datasource: DatasourceItem }) {
  const { t } = useTranslation()
  const [expr, setExpr] = useState("up")
  const [rangePreset, setRangePreset] = useState<RangePreset>("1h")
  const [customStart, setCustomStart] = useState("")
  const [customEnd, setCustomEnd] = useState("")
  const [step, setStep] = useState("60")
  const [pending, setPending] = useState(false)
  const [queryError, setQueryError] = useState("")
  const [exprError, setExprError] = useState("")
  const [statusText, setStatusText] = useState("")
  const [instantBody, setInstantBody] = useState("")
  const [rangeBody, setRangeBody] = useState("")
  const [resultTab, setResultTab] = useState<ResultTab>("table")
  const [proxyMethod, setProxyMethod] = useState("GET")
  const [proxyPath, setProxyPath] = useState("/")
  const [proxyBody, setProxyBody] = useState("")
  const [proxyResult, setProxyResult] = useState("")

  const instantResponse = useMemo(() => parsePrometheusBody(instantBody), [instantBody])
  const rangeResponse = useMemo(() => parsePrometheusBody(rangeBody), [rangeBody])
  const tableResponse = instantResponse ?? rangeResponse
  const graphResponse = rangeResponse
  const exprHint =
    datasource.driver === "METRICS_VICTORIA_METRICS"
      ? t("datasource.vmExprHint")
      : t("datasource.exprHint")

  function resolveQueryRange(): { start: number; end: number } {
    return resolveRange(rangePreset, customStart, customEnd)
  }

  async function runMetricsQuery() {
    const trimmed = expr.trim()
    if (!trimmed) {
      setExprError(t("form.required"))
      setQueryError("")
      return
    }
    setExprError("")
    const range = resolveQueryRange()
    const stepSeconds = Number(step)
    if (!Number.isFinite(range.start) || !Number.isFinite(range.end) || range.end <= range.start) {
      setQueryError(t("datasource.rangeInvalid"))
      return
    }
    if (!Number.isFinite(stepSeconds) || stepSeconds <= 0) {
      setQueryError(t("datasource.stepInvalid"))
      return
    }
    setPending(true)
    setQueryError("")
    setStatusText("")
    try {
      const [instant, ranged] = await Promise.all([
        queryMetric({ datasourceUid: datasource.uid, expr: trimmed, time: 0 }),
        queryMetricRange({
          datasourceUid: datasource.uid,
          expr: trimmed,
          start: range.start,
          end: range.end,
          step: stepSeconds,
        }),
      ])
      setInstantBody(instant.body)
      setRangeBody(ranged.body)
      setResultTab("table")
    } catch (err) {
      setInstantBody("")
      setRangeBody("")
      setQueryError(failureMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function runStatusCheck() {
    setPending(true)
    setQueryError("")
    try {
      const response = await getDatasourceStatus(datasource.uid)
      setStatusText(response.status)
    } catch (err) {
      setStatusText("")
      setQueryError(failureMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function runProxy() {
    setPending(true)
    setQueryError("")
    try {
      const response = await proxyQuery({
        datasourceUid: datasource.uid,
        method: proxyMethod,
        path: proxyPath,
        body: proxyBody,
      })
      setProxyResult(response.body)
    } catch (err) {
      setProxyResult("")
      setQueryError(failureMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  if (datasource.type === "LOGS" && datasource.driver === "LOGS_LOKI") {
    return <LokiPanel datasource={datasource} />
  }

  if (datasource.type === "LOGS" && datasource.driver === "LOGS_VLOGS") {
    return <VlogsPanel datasource={datasource} />
  }

  if (datasource.type === "LOGS" && datasource.driver === "LOGS_ELASTICSEARCH") {
    return <ElasticsearchPanel datasource={datasource} />
  }

  if (datasource.type !== "METRICS") {
    return (
      <div className="grid gap-4">
        <div className="flex flex-wrap gap-2">
          <Select value={proxyMethod} onValueChange={setProxyMethod}>
            <SelectTrigger className="w-28" aria-label={t("datasource.method")}>
              <SelectValue placeholder={t("datasource.method")} />
            </SelectTrigger>
            <SelectContent>
              {["GET", "POST", "PUT"].map((item) => (
                <SelectItem key={item} value={item}>
                  {item}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Input
            className="max-w-sm"
            value={proxyPath}
            placeholder={t("datasource.pathPlaceholder")}
            onChange={(event) => setProxyPath(event.target.value)}
          />
          <Button type="button" disabled={pending} onClick={() => void runProxy()}>
            {t("datasource.proxy")}
          </Button>
        </div>
        <textarea
          className={areaClass}
          value={proxyBody}
          placeholder={t("datasource.bodyPlaceholder")}
          onChange={(event) => setProxyBody(event.target.value)}
        />
        {queryError ? <p className="text-sm text-destructive">{queryError}</p> : null}
        {proxyResult ? <pre className="overflow-auto rounded-lg border p-3 text-sm">{proxyResult}</pre> : null}
      </div>
    )
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
      <div className="shrink-0">
        <FormField
          label={t("datasource.expr")}
          htmlFor="metric-expr"
          hint={exprHint}
          required
          error={exprError}
        >
          <PromQLEditor
            value={expr}
            datasourceUid={datasource.uid}
            onChange={(value) => {
              setExpr(value)
              if (exprError) {
                setExprError("")
              }
            }}
            onSubmit={() => void runMetricsQuery()}
          />
        </FormField>
      </div>
      <Tabs
        value={resultTab}
        onValueChange={(value) => setResultTab(value as ResultTab)}
        className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden"
      >
        <div className="flex shrink-0 flex-wrap items-center justify-between gap-2">
          <div className="flex flex-nowrap items-center gap-2">
            <Select value={rangePreset} onValueChange={(value) => setRangePreset(value as RangePreset)}>
              <SelectTrigger className="w-36 shrink-0" aria-label={t("datasource.rangePreset")}>
                <SelectValue placeholder={t("datasource.rangePreset")} />
              </SelectTrigger>
              <SelectContent>
                {(["1h", "3h", "6h", "12h", "1d", "7d", "custom"] as RangePreset[]).map((item) => (
                  <SelectItem key={item} value={item}>
                    {t(`datasource.preset.${item}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {rangePreset === "custom" ? (
              <>
                <Input
                  className="max-w-40"
                  value={customStart}
                  placeholder={t("datasource.start")}
                  onChange={(event) => setCustomStart(event.target.value)}
                />
                <Input
                  className="max-w-40"
                  value={customEnd}
                  placeholder={t("datasource.end")}
                  onChange={(event) => setCustomEnd(event.target.value)}
                />
              </>
            ) : null}
            <Input
              className="max-w-28"
              value={step}
              placeholder={t("datasource.step")}
              onChange={(event) => setStep(event.target.value)}
            />
            <Button type="button" className="shrink-0" disabled={pending} onClick={() => void runMetricsQuery()}>
              {t("datasource.runQuery")}
            </Button>
            <Button type="button" variant="outline" className="shrink-0" disabled={pending} onClick={() => void runStatusCheck()}>
              {t("datasource.check")}
            </Button>
          </div>
          <TabsList className="shrink-0">
            <TabsTrigger value="table">{t("datasource.resultTable")}</TabsTrigger>
            <TabsTrigger value="graph">{t("datasource.resultGraph")}</TabsTrigger>
            <TabsTrigger value="json">{t("datasource.resultJson")}</TabsTrigger>
          </TabsList>
        </div>
        {statusText ? (
          <p className="shrink-0 text-sm text-muted-foreground">
            {t("datasource.statusResult")}: {statusText}
          </p>
        ) : null}
        {queryError ? <p className="shrink-0 text-sm text-destructive">{queryError}</p> : null}
        <TabsContent value="table" className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          {instantBody || rangeBody ? <ResultTable response={tableResponse} /> : null}
        </TabsContent>
        <TabsContent value="graph" className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          {rangeBody ? <MetricGraph key={rangeBody} response={graphResponse} /> : null}
        </TabsContent>
        <TabsContent value="json" className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          {instantBody || rangeBody ? (
            <ResultJson
              response={rangeResponse ?? instantResponse}
              raw={rangeBody || instantBody}
              active={resultTab === "json"}
            />
          ) : null}
        </TabsContent>
      </Tabs>
    </div>
  )
}
