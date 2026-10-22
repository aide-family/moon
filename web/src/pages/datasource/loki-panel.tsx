import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { getDatasourceStatus } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { queryLogs } from "@/api/logs-query"
import { FormField } from "@/components/form-field"
import { MetricGraph } from "@/components/metric-graph"
import { LogqlEditor } from "@/components/promql-editor"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { failureText } from "@/lib/form-error"
import {
  extractLokiResult,
  formatLokiLabels,
  parseLokiBody,
  toLokiRows,
  type LokiLogRow,
} from "@/pages/datasource/loki"
import { resolveRange, type RangePreset } from "@/pages/datasource/range"

type LogDirection = "backward" | "forward"
type ResultTab = "logs" | "graph" | "json"

const MAX_LOG_LIMIT = 5000

function formatTimestamp(timestamp: number): string {
  if (!timestamp) {
    return "—"
  }
  const date = new Date(timestamp)
  const pad = (value: number, size = 2) => String(value).padStart(size, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}.${pad(date.getMilliseconds(), 3)}`
}

function LogTable({ rows }: { rows: LokiLogRow[] }) {
  const { t } = useTranslation()
  const stickyHead = "sticky top-0 z-20 bg-background shadow-[inset_0_-1px_0_0_hsl(var(--border))]"

  if (rows.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("datasource.logsEmpty")}</p>
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-auto rounded-lg border">
      <Table className="w-full">
        <TableHeader>
          <TableRow>
            <TableHead className={`${stickyHead} w-48 whitespace-nowrap`}>
              {t("datasource.resultTime")}
            </TableHead>
            <TableHead className={`${stickyHead} w-64 whitespace-nowrap`}>
              {t("datasource.logsStream")}
            </TableHead>
            <TableHead className={`${stickyHead} whitespace-nowrap`}>{t("datasource.logsLine")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.key} className="align-top">
              <TableCell className="whitespace-nowrap align-top font-mono text-xs text-muted-foreground">
                {formatTimestamp(row.timestamp)}
              </TableCell>
              <TableCell className="max-w-64 align-top font-mono text-xs text-muted-foreground">
                <pre className="whitespace-pre-wrap break-all">{formatLokiLabels(row.labels)}</pre>
              </TableCell>
              <TableCell className="align-top font-mono text-xs">
                <pre className="whitespace-pre-wrap break-all">{row.line}</pre>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

export function LokiPanel({ datasource }: { datasource: DatasourceItem }) {
  const { t } = useTranslation()
  const [expr, setExpr] = useState('{job=~".+"}')
  const [rangePreset, setRangePreset] = useState<RangePreset>("1h")
  const [customStart, setCustomStart] = useState("")
  const [customEnd, setCustomEnd] = useState("")
  const [limit, setLimit] = useState("100")
  const [direction, setDirection] = useState<LogDirection>("backward")
  const [pending, setPending] = useState(false)
  const [queryError, setQueryError] = useState("")
  const [exprError, setExprError] = useState("")
  const [statusText, setStatusText] = useState("")
  const [body, setBody] = useState("")
  const [resultTab, setResultTab] = useState<ResultTab>("logs")

  const parsedResponse = useMemo(() => parseLokiBody(body), [body])
  const resultType = parsedResponse?.data?.resultType
  const isMetricMatrix =
    parsedResponse?.status === "success" && resultType === "matrix"
  const rows = useMemo(
    () => toLokiRows(parsedResponse, direction),
    [direction, parsedResponse],
  )
  const jsonText = useMemo(() => {
    if (!body) {
      return ""
    }
    if (!parsedResponse) {
      return body
    }
    try {
      return JSON.stringify(extractLokiResult(parsedResponse), null, 2)
    } catch {
      return body
    }
  }, [body, parsedResponse])

  async function runQuery() {
    const trimmed = expr.trim()
    if (!trimmed) {
      setExprError(t("form.required"))
      setQueryError("")
      return
    }
    setExprError("")
    const range = resolveRange(rangePreset, customStart, customEnd)
    if (!Number.isFinite(range.start) || !Number.isFinite(range.end) || range.end <= range.start) {
      setQueryError(t("datasource.rangeInvalid"))
      return
    }
    const size = Number(limit)
    if (!Number.isInteger(size) || size < 1 || size > MAX_LOG_LIMIT) {
      setQueryError(t("datasource.logsLimitInvalid"))
      return
    }
    setPending(true)
    setQueryError("")
    setStatusText("")
    try {
      const response = await queryLogs({
        datasourceUid: datasource.uid,
        query: trimmed,
        start: range.start,
        end: range.end,
        limit: size,
        direction,
      })
      const queryResponse = parseLokiBody(response.body)
      setBody(response.body)
      setResultTab(
        queryResponse?.status === "success" && queryResponse.data?.resultType === "streams"
          ? "logs"
          : queryResponse?.status === "success" && queryResponse.data?.resultType === "matrix"
            ? "graph"
            : "json",
      )
    } catch (err) {
      setBody("")
      setResultTab("logs")
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

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
      <div className="shrink-0">
        <FormField
          label={t("datasource.expr")}
          htmlFor="loki-expr"
          hint={t("datasource.logsExprHint")}
          required
          error={exprError}
        >
          <LogqlEditor
            value={expr}
            datasourceUid={datasource.uid}
            onChange={(value) => {
              setExpr(value)
              if (exprError) {
                setExprError("")
              }
            }}
            onSubmit={() => void runQuery()}
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
              className="max-w-24"
              value={limit}
              aria-label={t("datasource.logsLimit")}
              placeholder={t("datasource.logsLimit")}
              onChange={(event) => setLimit(event.target.value)}
            />
            <Select value={direction} onValueChange={(value) => setDirection(value as LogDirection)}>
              <SelectTrigger className="w-32 shrink-0" aria-label={t("datasource.logsDirection")}>
                <SelectValue placeholder={t("datasource.logsDirection")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="backward">{t("datasource.logsDirectionBackward")}</SelectItem>
                <SelectItem value="forward">{t("datasource.logsDirectionForward")}</SelectItem>
              </SelectContent>
            </Select>
            <Button type="button" className="shrink-0" disabled={pending} onClick={() => void runQuery()}>
              {t("datasource.runQuery")}
            </Button>
            <Button
              type="button"
              variant="outline"
              className="shrink-0"
              disabled={pending}
              onClick={() => void runStatusCheck()}
            >
              {t("datasource.check")}
            </Button>
          </div>
          <TabsList className="shrink-0">
            {!body || resultType === "streams" ? (
              <TabsTrigger value="logs">{t("enum.LOGS")}</TabsTrigger>
            ) : null}
            {isMetricMatrix ? (
              <TabsTrigger value="graph">{t("datasource.resultGraph")}</TabsTrigger>
            ) : null}
            <TabsTrigger value="json">{t("datasource.resultJson")}</TabsTrigger>
          </TabsList>
        </div>
        {statusText ? (
          <p className="shrink-0 text-sm text-muted-foreground">
            {t("datasource.statusResult")}: {statusText}
          </p>
        ) : null}
        {queryError ? <p className="shrink-0 text-sm text-destructive">{queryError}</p> : null}
        <TabsContent value="logs" className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          {body ? <LogTable rows={rows} /> : null}
        </TabsContent>
        <TabsContent value="graph" className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          {isMetricMatrix ? (
            <MetricGraph
              key={body}
              response={parsedResponse}
              emptyText={t("datasource.queryGraphEmpty")}
            />
          ) : null}
        </TabsContent>
        <TabsContent value="json" className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          {jsonText ? (
            <pre className="min-h-0 flex-1 overflow-auto rounded-lg border bg-background p-3 font-mono text-xs">
              {jsonText}
            </pre>
          ) : null}
        </TabsContent>
      </Tabs>
    </div>
  )
}
