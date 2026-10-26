import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { getDatasourceStatus } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { queryLogs } from "@/api/logs-query"
import { FormField } from "@/components/form-field"
import { LogsQLEditor } from "@/components/query-expression-editor"
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
import { resolveRange, type RangePreset } from "@/pages/datasource/range"
import {
  LogResultTable,
  StatResult,
  type ResultLogRow,
} from "@/pages/datasource/result-view"
import {
  formatVlogsTimestamp,
  isNumericVlogsColumn,
  parseVlogsBody,
  sortVlogsEntries,
  vlogsLabelFields,
  vlogsLogLabels,
  vlogsMetricFields,
  vlogsResultFields,
  vlogsResultView,
  vlogsStatGroups,
  type VlogsLogEntry,
} from "@/pages/datasource/vlogs"

type LogDirection = "backward" | "forward"
type ResultTab = "logs" | "json"

const MAX_LOG_LIMIT = 5000

// vlogsFieldLabel 复用普通日志的列名；stats 结果的新字段直接显示字段名，
// 这样 host.name、value 等统计列不会被固定三列表格吞掉。
function vlogsFieldLabel(t: (key: string) => string, field: string): string {
  switch (field) {
    case "_time":
      return t("datasource.resultTime")
    case "_stream":
      return t("datasource.logsStream")
    case "_msg":
      return t("datasource.logsLine")
    case "value":
      return t("datasource.statsValue")
    default:
      return field
  }
}

function vlogsTimeLabel(value: string | undefined): string {
  const raw = value ?? ""
  const timestamp = Date.parse(raw)
  return timestamp ? formatVlogsTimestamp(timestamp) : raw || "—"
}

// VlogsTable 只用于无法归类的返回形状（例如全部是字符串字段），
// 日志和数值结果分别交给上面的专用视图。
function VlogsTable({
  entries,
  fields,
}: {
  entries: VlogsLogEntry[]
  fields: string[]
}) {
  const { t } = useTranslation()
  const numericFields = useMemo(
    () =>
      new Set(fields.filter((field) => isNumericVlogsColumn(entries, field))),
    [entries, fields]
  )
  const stickyHead =
    "sticky top-0 z-20 bg-background shadow-[inset_0_-1px_0_0_hsl(var(--border))]"

  if (entries.length === 0) {
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
            {fields.map((field) => (
              <TableHead
                key={field}
                className={`${stickyHead} whitespace-nowrap ${
                  field === "_time"
                    ? "w-48"
                    : field === "_stream"
                      ? "w-64"
                      : numericFields.has(field)
                        ? "w-28 text-right"
                        : ""
                }`}
              >
                {vlogsFieldLabel(t, field)}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {entries.map((entry, index) => (
            <TableRow key={index} className="align-top">
              {fields.map((field) => {
                const value = entry[field] ?? ""
                const numeric = numericFields.has(field)
                if (field === "_time") {
                  const timestamp = Date.parse(value)
                  return (
                    <TableCell
                      key={field}
                      className="align-top font-mono text-xs whitespace-nowrap text-muted-foreground"
                    >
                      {timestamp
                        ? formatVlogsTimestamp(timestamp)
                        : value || "—"}
                    </TableCell>
                  )
                }
                if (numeric) {
                  return (
                    <TableCell
                      key={field}
                      className="text-right align-top font-mono text-xs tabular-nums"
                    >
                      {value || "—"}
                    </TableCell>
                  )
                }
                return (
                  <TableCell
                    key={field}
                    className={`align-top font-mono text-xs ${
                      field === "_msg" ? "" : "max-w-64 text-muted-foreground"
                    }`}
                  >
                    <pre className="break-all whitespace-pre-wrap">
                      {value || "—"}
                    </pre>
                  </TableCell>
                )
              })}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

export function VlogsPanel({ datasource }: { datasource: DatasourceItem }) {
  const { t } = useTranslation()
  const defaults = useMemo(() => defaultSnapshot("logsql"), [])
  const draft = useQueryDraft(datasource.uid, "logsql", defaults)
  const library = useQueryLibrary(datasource.uid, "logsql")
  const { expr, rangePreset, limit, direction } = draft.state
  const [customStart, setCustomStart] = useState("")
  const [customEnd, setCustomEnd] = useState("")
  const [pending, setPending] = useState(false)
  const [queryError, setQueryError] = useState("")
  const [exprError, setExprError] = useState("")
  const [statusText, setStatusText] = useState("")
  const [body, setBody] = useState("")
  const [resultTab, setResultTab] = useState<ResultTab>("logs")

  const entries = useMemo(() => parseVlogsBody(body), [body])
  const fields = useMemo(() => vlogsResultFields(entries), [entries])
  const orderedEntries = useMemo(
    () => sortVlogsEntries(entries, direction),
    [entries, direction]
  )
  const resultView = useMemo(
    () => vlogsResultView(entries, fields),
    [entries, fields]
  )
  const logRows = useMemo<ResultLogRow[]>(
    () =>
      resultView === "logs"
        ? orderedEntries.map((entry, index) => ({
            key: `${entry._time ?? ""}-${index}`,
            time: vlogsTimeLabel(entry._time),
            line: entry._msg || "—",
            labels: vlogsLogLabels(entry),
          }))
        : [],
    [orderedEntries, resultView]
  )
  const metricFields = useMemo(
    () => vlogsMetricFields(entries, fields),
    [entries, fields]
  )
  const statGroups = useMemo(
    () =>
      vlogsStatGroups(
        orderedEntries,
        metricFields,
        vlogsLabelFields(fields, metricFields)
      ),
    [orderedEntries, metricFields, fields]
  )
  const jsonText = useMemo(() => {
    if (!body) {
      return ""
    }
    return entries.length > 0 ? JSON.stringify(entries, null, 2) : body
  }, [body, entries])

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

  const resultTabLabel = !body
    ? t("enum.LOGS")
    : resultView === "value"
      ? t("datasource.statsValue")
      : resultView === "table"
        ? t("datasource.resultTable")
        : t("enum.LOGS")

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
      <div className="shrink-0">
        <FormField
          label={t("datasource.expr")}
          htmlFor="vlogs-expr"
          hint={t("datasource.vlogsExprHint")}
          required
          error={exprError}
          actions={
            <DatasourceQueryToolbar
              dialect="logsql"
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
          <LogsQLEditor
            value={expr}
            datasourceUid={datasource.uid}
            placeholder={t("datasource.vlogsQueryExprPlaceholder")}
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
            ) : resultView === "value" ? (
              <StatResult
                key={`${direction}-${body}`}
                groups={statGroups}
                fields={metricFields}
                fieldLabel={(field) => vlogsFieldLabel(t, field)}
              />
            ) : (
              <VlogsTable entries={orderedEntries} fields={fields} />
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
