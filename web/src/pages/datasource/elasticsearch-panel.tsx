import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { getDatasourceStatus } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { queryLogs } from "@/api/logs-query"
import { FormField } from "@/components/form-field"
import { PromQLEditor } from "@/components/query-expression-editor"
import { useQueryDraft, useQueryLibrary } from "@/hooks/use-query-store"
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
import { formatResultTimestamp } from "@/pages/datasource/result-data"
import {
  LogResultTable,
  StatResult,
  type ResultLogRow,
} from "@/pages/datasource/result-view"
import {
  elasticsearchAggregationStat,
  elasticsearchResultView,
  extractElasticsearchResult,
  parseElasticsearchBody,
  toElasticsearchRows,
  type ElasticsearchLogRow,
} from "@/pages/datasource/elasticsearch"
import { resolveRange, type RangePreset } from "@/pages/datasource/range"

type LogDirection = "backward" | "forward"
type ResultTab = "logs" | "json"

const MAX_LOG_LIMIT = 5000

function elasticsearchLogRows(rows: ElasticsearchLogRow[]): ResultLogRow[] {
  return rows.map((row) => ({
    key: row.key,
    time: formatResultTimestamp(row.timestamp),
    line: row.line || "—",
    labels: row.labels,
  }))
}

function elasticsearchFieldLabel(
  t: (key: string) => string,
  field: string
): string {
  if (field === "doc_count") {
    return t("datasource.esDocCount")
  }
  const parts = field.split(".")
  if (parts.length === 1) {
    return field
  }
  const last = parts[parts.length - 1]
  const prefix = parts.slice(0, -1).join(".")
  return last === "value" ? prefix : `${prefix} · ${last}`
}

// 无法识别日志正文的文档保留原有表格；日志和数值结果分别交给共用视图。
function HitsTable({ rows }: { rows: ElasticsearchLogRow[] }) {
  const { t } = useTranslation()
  const stickyHead =
    "sticky top-0 z-20 bg-background shadow-[inset_0_-1px_0_0_hsl(var(--border))]"

  if (rows.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("datasource.logsEmpty")}
      </p>
    )
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
              {t("datasource.esDocument")}
            </TableHead>
            <TableHead className={`${stickyHead} whitespace-nowrap`}>
              {t("datasource.esSource")}
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.key} className="align-top">
              <TableCell className="align-top font-mono text-xs whitespace-nowrap text-muted-foreground">
                {formatResultTimestamp(row.timestamp)}
              </TableCell>
              <TableCell className="max-w-64 align-top font-mono text-xs text-muted-foreground">
                <pre className="break-all whitespace-pre-wrap">
                  {row.index}
                  {row.id ? `\n${row.id}` : ""}
                </pre>
              </TableCell>
              <TableCell className="align-top font-mono text-xs">
                <pre className="break-all whitespace-pre-wrap">
                  {JSON.stringify(row.source, null, 2)}
                </pre>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

export function ElasticsearchPanel({
  datasource,
}: {
  datasource: DatasourceItem
}) {
  const { t } = useTranslation()
  const defaults = useMemo(() => defaultSnapshot("dsl"), [])
  const draft = useQueryDraft(datasource.uid, "dsl", defaults)
  const library = useQueryLibrary(datasource.uid, "dsl")
  const { expr, rangePreset, limit, direction } = draft.state
  const [customStart, setCustomStart] = useState("")
  const [customEnd, setCustomEnd] = useState("")
  const [pending, setPending] = useState(false)
  const [queryError, setQueryError] = useState("")
  const [exprError, setExprError] = useState("")
  const [statusText, setStatusText] = useState("")
  const [body, setBody] = useState("")
  const [resultTab, setResultTab] = useState<ResultTab>("logs")

  const timeField = datasource.timeField || "@timestamp"
  const parsed = useMemo(() => parseElasticsearchBody(body), [body])
  const rows = useMemo(
    () => toElasticsearchRows(parsed, direction, timeField),
    [parsed, direction, timeField]
  )
  const resultView = useMemo(
    () => elasticsearchResultView(parsed, timeField),
    [parsed, timeField]
  )
  const statResult = useMemo(
    () => elasticsearchAggregationStat(parsed),
    [parsed]
  )
  const logRows = useMemo(() => elasticsearchLogRows(rows), [rows])
  const jsonText = useMemo(() => {
    if (!body) {
      return ""
    }
    if (!parsed) {
      return body
    }
    try {
      return JSON.stringify(extractElasticsearchResult(parsed), null, 2)
    } catch {
      return body
    }
  }, [body, parsed])

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
    if (trimmed.startsWith("{") && !isJSONObject(trimmed)) {
      setExprError(t("datasource.esQueryInvalid"))
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
      setBody(response.body)
      setResultTab("logs")
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
          htmlFor="es-expr"
          hint={t("datasource.esExprHint")}
          required
          error={exprError}
          actions={
            <DatasourceQueryToolbar
              dialect="dsl"
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
          <PromQLEditor
            value={expr}
            variant="json"
            datasourceUid={datasource.uid}
            onChange={(value) => {
              draft.patch({ expr: value })
              if (exprError) {
                setExprError("")
              }
            }}
            onSubmit={() => void runQuery()}
          />
        </FormField>
        <p className="mt-2 text-xs text-muted-foreground">
          {t("datasource.indexPattern")}:{" "}
          {datasource.indexPattern || t("datasource.indexPatternAll")} ·{" "}
          {t("datasource.timeField")}: {timeField}
        </p>
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
                fieldLabel={(field) => elasticsearchFieldLabel(t, field)}
              />
            ) : (
              <HitsTable rows={rows} />
            )
          ) : null}
        </TabsContent>
        <TabsContent
          value="json"
          className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden"
        >
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

function isJSONObject(value: string): boolean {
  try {
    const parsed = JSON.parse(value) as unknown
    return (
      typeof parsed === "object" && parsed !== null && !Array.isArray(parsed)
    )
  } catch {
    return false
  }
}
