import { editor as monacoEditor } from "monaco-editor"
import { Check, Copy } from "lucide-react"
import { useEffect, useMemo, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getDatasourceStatus } from "@/api/datasource"
import { proxyQuery, queryMetric, queryMetricRange } from "@/api/metric-query"
import type { DatasourceItem } from "@/api/datasource/types"
import { FormField } from "@/components/form-field"
import { PromqlEditor } from "@/components/promql-editor"
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
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import { failureText } from "@/lib/form-error"
import {
  collectLabelKeys,
  extractPrometheusResultData,
  parsePrometheusBody,
  toGraphSeries,
  toTableRows,
  type PrometheusApiResponse,
} from "@/pages/datasource/prometheus"
import { ElasticsearchPanel } from "@/pages/datasource/elasticsearch-panel"
import { LokiPanel } from "@/pages/datasource/loki-panel"
import { resolveRange, type RangePreset } from "@/pages/datasource/range"

const areaClass =
  "min-h-24 w-full rounded-lg border border-input bg-transparent px-2.5 py-1 text-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"

type ResultTab = "table" | "graph" | "json"

const CHART_COLORS = [
  "#1677ff",
  "#52c41a",
  "#faad14",
  "#eb2f96",
  "#13c2c2",
  "#722ed1",
  "#fa541c",
  "#2f54eb",
]

function formatTime(ts: number): string {
  if (!ts) {
    return "—"
  }
  const date = new Date(ts * 1000)
  const pad = (value: number) => String(value).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

function formatClock(ts: number): string {
  const date = new Date(ts * 1000)
  const pad = (value: number) => String(value).padStart(2, "0")
  return `${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function findNearestPoint(points: { time: number; value: number }[], targetTime: number) {
  if (points.length === 0) {
    return null
  }
  let nearest = points[0]
  let minDiff = Math.abs(points[0].time - targetTime)
  for (let index = 1; index < points.length; index += 1) {
    const diff = Math.abs(points[index].time - targetTime)
    if (diff < minDiff) {
      minDiff = diff
      nearest = points[index]
    }
  }
  return nearest
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

function ResultGraph({ response }: { response: PrometheusApiResponse | null }) {
  const { t } = useTranslation()
  const containerRef = useRef<HTMLDivElement>(null)
  const [size, setSize] = useState({ width: 480, height: 280 })
  const [selected, setSelected] = useState<Set<string>>(() => new Set())
  const [hover, setHover] = useState<{
    time: number
    plotX: number
    items: { name: string; value: number; color: string }[]
  } | null>(null)

  const series = useMemo(() => (response ? toGraphSeries(response) : []), [response])
  const visibleSeries = useMemo(
    () => (selected.size === 0 ? series : series.filter((item) => selected.has(item.name))),
    [series, selected],
  )
  const width = size.width
  const height = Math.max(size.height, 160)
  const padding = { top: 24, right: 20, bottom: 44, left: 56 }
  const plotW = Math.max(width - padding.left - padding.right, 1)
  const plotH = height - padding.top - padding.bottom

  useEffect(() => {
    const node = containerRef.current
    if (!node) {
      return
    }
    const observer = new ResizeObserver((entries) => {
      const box = entries[0]?.contentRect
      if (!box) {
        return
      }
      setSize({
        width: Math.max(box.width, 240),
        height: Math.max(box.height, 160),
      })
    })
    observer.observe(node)
    return () => observer.disconnect()
  }, [])

  const chartMeta = useMemo(() => {
    const allPoints = visibleSeries.flatMap((item) => item.points)
    if (allPoints.length === 0) {
      return null
    }
    const minTime = Math.min(...allPoints.map((point) => point.time))
    const maxTime = Math.max(...allPoints.map((point) => point.time))
    const minVal = Math.min(...allPoints.map((point) => point.value))
    const maxVal = Math.max(...allPoints.map((point) => point.value))
    const valSpan = maxVal - minVal || 1
    const timeSpan = maxTime - minTime || 1
    return {
      minTime,
      maxTime,
      minVal,
      valSpan,
      timeSpan,
      x: (time: number) => padding.left + ((time - minTime) / timeSpan) * plotW,
      y: (value: number) => padding.top + plotH - ((value - minVal) / valSpan) * plotH,
    }
  }, [visibleSeries, plotW, plotH, padding.left, padding.top])

  function toggleSeries(name: string) {
    setSelected((current) => {
      const next = new Set(current)
      if (next.has(name)) {
        next.delete(name)
      } else {
        next.add(name)
      }
      return next
    })
  }

  function seriesShown(name: string) {
    return selected.size === 0 || selected.has(name)
  }

  if (!response || series.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("datasource.queryGraphEmpty")}</p>
  }

  const yTicks = chartMeta
    ? Array.from({ length: 6 }, (_, index) => {
        const ratio = index / 5
        return {
          value: chartMeta.minVal + chartMeta.valSpan * (1 - ratio),
          py: padding.top + plotH * ratio,
        }
      })
    : []
  const xLabels = chartMeta
    ? Array.from({ length: 5 }, (_, index) => {
        const ratio = index / 4
        return {
          time: chartMeta.minTime + chartMeta.timeSpan * ratio,
          x: padding.left + plotW * ratio,
        }
      })
    : []

  return (
    <div className="flex min-h-0 min-w-0 flex-1 gap-3">
      <div ref={containerRef} className="relative min-h-0 min-w-0 flex-1 overflow-hidden rounded-lg border bg-muted/20">
        {hover && chartMeta ? (
          <div
            className="pointer-events-none absolute z-10 max-w-[min(100%,20rem)] rounded-md border bg-background px-3 py-2 text-xs shadow-sm"
            style={{
              left: Math.min(Math.max(hover.plotX - 120, 8), width - 260),
              top: 8,
            }}
          >
            <div className="mb-1.5 font-medium">{formatTime(Math.round(hover.time))}</div>
            <div className="flex max-h-40 flex-col gap-1 overflow-auto">
              {hover.items.map((item) => (
                <div key={item.name} className="flex min-w-0 items-center gap-2">
                  <span className="size-2 shrink-0 rounded-full" style={{ backgroundColor: item.color }} />
                  <span className="min-w-0 flex-1 truncate text-muted-foreground">{item.name}</span>
                  <span className="shrink-0 font-mono">{item.value}</span>
                </div>
              ))}
            </div>
          </div>
        ) : null}
        {chartMeta ? (
          <svg
            viewBox={`0 0 ${width} ${height}`}
            className="block w-full select-none"
            role="img"
            aria-label={t("datasource.queryGraphAria")}
          >
            {yTicks.map((tick) => (
              <g key={tick.py}>
                <line
                  x1={padding.left}
                  y1={tick.py}
                  x2={width - padding.right}
                  y2={tick.py}
                  className="stroke-border"
                  strokeDasharray="4 4"
                />
                <text
                  x={padding.left - 8}
                  y={tick.py + 4}
                  textAnchor="end"
                  fontSize={11}
                  className="fill-muted-foreground"
                >
                  {tick.value.toPrecision(4)}
                </text>
              </g>
            ))}
            {xLabels.map((label) => (
              <text
                key={label.x}
                x={label.x}
                y={height - 12}
                textAnchor="middle"
                fontSize={11}
                className="fill-muted-foreground"
              >
                {formatClock(label.time)}
              </text>
            ))}
            <line
              x1={padding.left}
              y1={padding.top + plotH}
              x2={width - padding.right}
              y2={padding.top + plotH}
              className="stroke-border"
            />
            <line
              x1={padding.left}
              y1={padding.top}
              x2={padding.left}
              y2={padding.top + plotH}
              className="stroke-border"
            />
            {hover ? (
              <line
                x1={hover.plotX}
                y1={padding.top}
                x2={hover.plotX}
                y2={padding.top + plotH}
                className="stroke-muted-foreground"
                strokeDasharray="4 2"
              />
            ) : null}
            {series.map((item, index) => {
              if (!seriesShown(item.name) || item.points.length === 0) {
                return null
              }
              const color = CHART_COLORS[index % CHART_COLORS.length]
              const path = item.points
                .map((point, pointIndex) => {
                  const command = pointIndex === 0 ? "M" : "L"
                  return `${command}${chartMeta.x(point.time)} ${chartMeta.y(point.value)}`
                })
                .join(" ")
              return <path key={item.name} d={path} fill="none" stroke={color} strokeWidth={1.5} />
            })}
            <rect
              x={padding.left}
              y={padding.top}
              width={plotW}
              height={plotH}
              fill="transparent"
              onMouseMove={(event) => {
                const svg = event.currentTarget.ownerSVGElement
                if (!svg) {
                  return
                }
                const rect = svg.getBoundingClientRect()
                const scaleX = width / rect.width
                const relX = (event.clientX - rect.left) * scaleX - padding.left
                if (relX < 0 || relX > plotW) {
                  setHover(null)
                  return
                }
                const targetTime = chartMeta.minTime + (relX / plotW) * chartMeta.timeSpan
                const items = series
                  .map((item, index) => {
                    if (!seriesShown(item.name)) {
                      return null
                    }
                    const nearest = findNearestPoint(item.points, targetTime)
                    if (!nearest) {
                      return null
                    }
                    return {
                      name: item.name,
                      value: nearest.value,
                      color: CHART_COLORS[index % CHART_COLORS.length],
                    }
                  })
                  .filter((item): item is NonNullable<typeof item> => item != null)
                setHover({ time: targetTime, plotX: padding.left + relX, items })
              }}
              onMouseLeave={() => setHover(null)}
            />
          </svg>
        ) : (
          <div className="flex h-full min-h-40 items-center justify-center px-4">
            <p className="text-sm text-muted-foreground">{t("datasource.queryGraphEmpty")}</p>
          </div>
        )}
      </div>
      <TooltipProvider delayDuration={1500} skipDelayDuration={0}>
      <div className="flex h-full w-48 shrink-0 flex-col overflow-hidden rounded-lg border">
        <p className="shrink-0 px-2 py-1.5 text-xs font-medium text-muted-foreground">{t("datasource.legend")}</p>
        <div className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto p-2 pt-0">
          {series.map((item, index) => {
            const color = CHART_COLORS[index % CHART_COLORS.length]
            const dimmed = selected.size > 0 && !selected.has(item.name)
            return (
              <Tooltip key={item.name}>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    className={cn(
                      "flex w-full items-center gap-2 rounded-md px-1.5 py-1 text-left text-xs hover:bg-muted/60",
                      dimmed && "opacity-40",
                    )}
                    onClick={() => toggleSeries(item.name)}
                  >
                    <span className="size-2.5 shrink-0 rounded-sm" style={{ backgroundColor: color }} />
                    <span className="min-w-0 flex-1 truncate font-mono">{item.name}</span>
                  </button>
                </TooltipTrigger>
                <TooltipContent>{item.name}</TooltipContent>
              </Tooltip>
            )
          })}
        </div>
      </div>
      </TooltipProvider>
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
      setQueryError(failureText(err, t("state.failed")))
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
      setQueryError(failureText(err, t("state.failed")))
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
      setQueryError(failureText(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  if (datasource.type === "LOGS" && datasource.driver === "LOGS_LOKI") {
    return <LokiPanel datasource={datasource} />
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
          <PromqlEditor
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
          {rangeBody ? <ResultGraph key={rangeBody} response={graphResponse} /> : null}
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
