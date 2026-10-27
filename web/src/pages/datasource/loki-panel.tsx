import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { getDatasourceStatus } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { queryLogs } from "@/api/logs-query"
import { CodeBlock } from "@/components/code-block"
import { FormField } from "@/components/form-field"
import { MetricGraph } from "@/components/metric-graph"
import { useQueryDraft } from "@/hooks/use-query-draft"
import { useQueryLibrary } from "@/hooks/use-query-library"
import { LogQLEditor } from "@/components/query-expression-editor"
import { Button } from "@/components/ui/button"
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { failureMessage } from "@/lib/form-error"
import type { QuerySnapshot } from "@/lib/query-store"
import { DatasourceQueryToolbar } from "@/pages/datasource/components/query-toolbar"
import {
  defaultSnapshot,
  restoreSnapshot,
} from "@/pages/datasource/query-dialect"
import {
  LogResultTable,
  StatResult,
  type ResultLogRow,
} from "@/pages/datasource/result-view"
import { formatResultTimestamp } from "@/pages/datasource/result-data"
import {
  extractLokiResult,
  lokiResultStat,
  lokiResultView,
  lokiTableRows,
  parseLokiBody,
  toLokiRows,
  type LokiLogRow,
  type LokiTableRow,
} from "@/pages/datasource/loki"
import { resolveRange, type RangePreset } from "@/pages/datasource/range"

type LogDirection = "backward" | "forward"
type ResultTab = "logs" | "graph" | "json"

const MAX_LOG_LIMIT = 5000

function lokiLogRows(rows: LokiLogRow[]): ResultLogRow[] {
  return rows.map((row) => ({
    key: row.key,
    time: formatResultTimestamp(row.timestamp),
    line: row.line || "—",
    labels: Object.entries(row.labels).map(([key, value]) => ({ key, value })),
  }))
}

// 无法归类的返回形状（例如非数值 matrix）保留表格兜底，
// 日志和数值结果分别交给共用视图。
function LokiResultTable({ rows }: { rows: LokiTableRow[] }) {
  const { t } = useTranslation()
  const stickyHead =
    "sticky top-0 z-20 bg-background shadow-[inset_0_-1px_0_0_hsl(var(--border))]"

  if (rows.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("datasource.queryEmpty")}
      </p>
    )
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-auto rounded-lg border">
      <Table className="w-full">
        <TableHeader>
          <TableRow>
            <TableHead className={`${stickyHead} w-64 whitespace-nowrap`}>
              {t("datasource.resultSeries")}
            </TableHead>
            <TableHead className={`${stickyHead} w-48 whitespace-nowrap`}>
              {t("datasource.resultTime")}
            </TableHead>
            <TableHead className={`${stickyHead} whitespace-nowrap`}>
              {t("datasource.resultValue")}
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.key} className="align-top">
              <TableCell className="max-w-64 align-top font-mono text-xs text-muted-foreground">
                <pre className="break-all whitespace-pre-wrap">
                  {row.series}
                </pre>
              </TableCell>
              <TableCell className="align-top font-mono text-xs whitespace-nowrap text-muted-foreground">
                {row.time || "—"}
              </TableCell>
              <TableCell className="text-right align-top font-mono text-xs tabular-nums">
                {row.value || "—"}
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
  const defaults = useMemo(() => defaultSnapshot("logql"), [])
  const draft = useQueryDraft(datasource.uid, "logql", defaults)
  const library = useQueryLibrary(datasource.uid, "logql")
  const { expr, rangePreset, limit, direction } = draft.state
  const [customStart, setCustomStart] = useState("")
  const [customEnd, setCustomEnd] = useState("")
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
    [direction, parsedResponse]
  )
  const resultView = useMemo(
    () => lokiResultView(parsedResponse),
    [parsedResponse]
  )
  const statResult = useMemo(
    () => lokiResultStat(parsedResponse),
    [parsedResponse]
  )
  const logRows = useMemo(() => lokiLogRows(rows), [rows])
  const tableRows = useMemo(
    () => lokiTableRows(parsedResponse, direction),
    [direction, parsedResponse]
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

  const resultTabLabel = !body
    ? t("enum.LOGS")
    : resultView === "value"
      ? t("datasource.statsValue")
      : resultView === "table"
        ? t("datasource.resultTable")
        : t("enum.LOGS")

  async function runQuery(overrides?: QuerySnapshot) {
    const active = overrides ? restoreSnapshot(overrides) : draft.state
    if (overrides) {
      draft.replace(active)
    }
    const trimmed = active.expr.trim()
    if (!trimmed) {
      setExprError(t("form.required"))
      setQueryError("")
      return
    }
    setExprError("")
    const range = resolveRange(
      active.rangePreset as RangePreset,
      customStart,
      customEnd
    )
    if (
      !Number.isFinite(range.start) ||
      !Number.isFinite(range.end) ||
      range.end <= range.start
    ) {
      setQueryError(t("datasource.rangeInvalid"))
      return
    }
    const size = Number(active.limit)
    if (!Number.isInteger(size) || size < 1 || size > MAX_LOG_LIMIT) {
      setQueryError(t("datasource.logsLimitInvalid"))
      return
    }
    setPending(true)
    setQueryError("")
    setStatusText("")
    const startedAt = Date.now()
    try {
      const response = await queryLogs({
        datasourceUid: datasource.uid,
        query: trimmed,
        start: range.start,
        end: range.end,
        limit: size,
        direction: active.direction,
      })
      const queryResponse = parseLokiBody(response.body)
      setBody(response.body)
      setResultTab(queryResponse?.status === "success" ? "logs" : "json")
      draft.markUsed(active)
      library.record({
        statement: trimmed,
        snapshot: active,
        ok: true,
        durationMs: Date.now() - startedAt,
      })
    } catch (err) {
      const message = failureMessage(err, t("state.failed"))
      setBody("")
      setResultTab("logs")
      setQueryError(message)
      draft.markUsed(active)
      library.record({
        statement: trimmed,
        snapshot: active,
        ok: false,
        durationMs: Date.now() - startedAt,
        error: message,
      })
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

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
      <div className="shrink-0">
        <FormField
          label={t("datasource.expr")}
          htmlFor="loki-expr"
          hint={t("datasource.logsExprHint")}
          required
          error={exprError}
          actions={
            <DatasourceQueryToolbar
              dialect="logql"
              library={library}
              snapshot={draft.state}
              defaultSnapshot={defaults}
              lastUsedAt={draft.lastUsedAt}
              pending={pending}
              onFill={(snapshot) => draft.replace(snapshot)}
              onRun={(snapshot) => void runQuery(snapshot)}
            />
          }
        >
          <LogQLEditor
            value={expr}
            datasourceUid={datasource.uid}
            placeholder={t("datasource.lokiQueryExprPlaceholder")}
            onChange={(value) => {
              draft.patch({ expr: value })
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
            <Select
              value={rangePreset}
              onValueChange={(value) => draft.patch({ rangePreset: value })}
            >
              <SelectTrigger
                className="w-36 shrink-0"
                aria-label={t("datasource.rangePreset")}
              >
                <SelectValue placeholder={t("datasource.rangePreset")} />
              </SelectTrigger>
              <SelectContent>
                {(
                  [
                    "1h",
                    "3h",
                    "6h",
                    "12h",
                    "1d",
                    "7d",
                    "custom",
                  ] as RangePreset[]
                ).map((item) => (
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
              onChange={(event) => draft.patch({ limit: event.target.value })}
            />
            <Select
              value={direction}
              onValueChange={(value) =>
                draft.patch({ direction: value as LogDirection })
              }
            >
              <SelectTrigger
                className="w-32 shrink-0"
                aria-label={t("datasource.logsDirection")}
              >
                <SelectValue placeholder={t("datasource.logsDirection")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="backward">
                  {t("datasource.logsDirectionBackward")}
                </SelectItem>
                <SelectItem value="forward">
                  {t("datasource.logsDirectionForward")}
                </SelectItem>
              </SelectContent>
            </Select>
            <Button
              type="button"
              className="shrink-0"
              disabled={pending}
              onClick={() => void runQuery()}
            >
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
            <TabsTrigger value="logs">{resultTabLabel}</TabsTrigger>
            {isMetricMatrix ? (
              <TabsTrigger value="graph">
                {t("datasource.resultGraph")}
              </TabsTrigger>
            ) : null}
            <TabsTrigger value="json">{t("datasource.resultJson")}</TabsTrigger>
          </TabsList>
        </div>
        {statusText ? (
          <p className="shrink-0 text-sm text-muted-foreground">
            {t("datasource.statusResult")}: {statusText}
          </p>
        ) : null}
        {queryError ? (
          <p className="shrink-0 text-sm text-destructive">{queryError}</p>
        ) : null}
        <TabsContent
          value="logs"
          className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden"
        >
          {body ? (
            resultView === "logs" ? (
              <LogResultTable key={`${direction}-${body}`} rows={logRows} />
            ) : resultView === "value" && statResult ? (
              <StatResult
                key={`${direction}-${body}`}
                groups={statResult.groups}
                fields={statResult.fields}
              />
            ) : (
              <LokiResultTable rows={tableRows} />
            )
          ) : null}
        </TabsContent>
        <TabsContent
          value="graph"
          className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden"
        >
          {isMetricMatrix ? (
            <MetricGraph
              key={body}
              response={parsedResponse}
              emptyText={t("datasource.queryGraphEmpty")}
            />
          ) : null}
        </TabsContent>
        <TabsContent
          value="json"
          className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden"
        >
          {jsonText ? (
            <CodeBlock
              fill
              code={jsonText}
              language="json"
              label={t("datasource.resultJson")}
              copyable
            />
          ) : null}
        </TabsContent>
      </Tabs>
    </div>
  )
}
