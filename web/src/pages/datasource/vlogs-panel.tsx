import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { getDatasourceStatus } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { queryLogs } from "@/api/logs-query"
import { FormField } from "@/components/form-field"
import { PromqlEditor } from "@/components/promql-editor"
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
import { failureText } from "@/lib/form-error"
import { resolveRange, type RangePreset } from "@/pages/datasource/range"
import {
  formatVlogsTimestamp,
  parseVlogsBody,
  toVlogsRows,
  type VlogsLogRow,
} from "@/pages/datasource/vlogs"

type LogDirection = "backward" | "forward"
type ResultTab = "logs" | "json"

const MAX_LOG_LIMIT = 5000

function LogTable({ rows }: { rows: VlogsLogRow[] }) {
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
              {t("datasource.logsStream")}
            </TableHead>
            <TableHead className={`${stickyHead} whitespace-nowrap`}>
              {t("datasource.logsLine")}
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.key} className="align-top">
              <TableCell className="align-top font-mono text-xs whitespace-nowrap text-muted-foreground">
                {formatVlogsTimestamp(row.timestamp)}
              </TableCell>
              <TableCell className="max-w-64 align-top font-mono text-xs text-muted-foreground">
                <pre className="break-all whitespace-pre-wrap">
                  {row.stream || "—"}
                </pre>
              </TableCell>
              <TableCell className="align-top font-mono text-xs">
                <pre className="break-all whitespace-pre-wrap">
                  {row.message || "—"}
                </pre>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

export function VlogsPanel({ datasource }: { datasource: DatasourceItem }) {
  const { t } = useTranslation()
  const [expr, setExpr] = useState("*")
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

  const entries = useMemo(() => parseVlogsBody(body), [body])
  const rows = useMemo(
    () => toVlogsRows(entries, direction),
    [entries, direction]
  )
  const jsonText = useMemo(() => {
    if (!body) {
      return ""
    }
    return entries.length > 0 ? JSON.stringify(entries, null, 2) : body
  }, [body, entries])

  async function runQuery() {
    const trimmed = expr.trim()
    if (!trimmed) {
      setExprError(t("form.required"))
      setQueryError("")
      return
    }
    setExprError("")
    const range = resolveRange(rangePreset, customStart, customEnd)
    if (
      !Number.isFinite(range.start) ||
      !Number.isFinite(range.end) ||
      range.end <= range.start
    ) {
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
      setBody(response.body)
      setResultTab("logs")
    } catch (err) {
      setBody("")
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
          htmlFor="vlogs-expr"
          hint={t("datasource.vlogsExprHint")}
          required
          error={exprError}
        >
          <PromqlEditor
            value={expr}
            variant="logsql"
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
            <Select
              value={rangePreset}
              onValueChange={(value) => setRangePreset(value as RangePreset)}
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
              onChange={(event) => setLimit(event.target.value)}
            />
            <Select
              value={direction}
              onValueChange={(value) => setDirection(value as LogDirection)}
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
            <TabsTrigger value="logs">{t("enum.LOGS")}</TabsTrigger>
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
          {body ? <LogTable rows={rows} /> : null}
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
