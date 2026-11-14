import { useState } from "react"
import { useTranslation } from "react-i18next"

import { listInspectionRuns } from "@/api/inspection"
import type { InspectionRun, InspectionRunStatus } from "@/api/inspection/types"
import { useSession } from "@/auth/session"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import { Field, FieldLabel } from "@/components/ui/field"
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
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { InspectionReportSheet } from "./report-sheet"
import { InspectionReportShareDialog } from "./report-share-dialog"

const ALL_STATUS = "ALL"
const RUN_STATUSES: InspectionRunStatus[] = [
  "PENDING",
  "RUNNING",
  "SUCCEEDED",
  "PARTIAL",
  "FAILED",
  "CANCELED",
]

// URL 里的状态只认执行记录的真实枚举，非法值按不筛选处理。
function normalizeRunStatus(value: string): InspectionRunStatus | "" {
  return RUN_STATUSES.includes(value as InspectionRunStatus)
    ? (value as InspectionRunStatus)
    : ""
}

function runTone(status: InspectionRunStatus) {
  switch (status) {
    case "SUCCEEDED":
      return "success" as const
    case "PARTIAL":
      return "warning" as const
    case "FAILED":
      return "danger" as const
    case "RUNNING":
      return "info" as const
    default:
      return "warning" as const
  }
}

function runLabel(
  t: (key: string) => string,
  status: InspectionRunStatus
): string {
  switch (status) {
    case "SUCCEEDED":
      return t("inspection.runSucceeded")
    case "PARTIAL":
      return t("inspection.runPartial")
    case "FAILED":
      return t("inspection.runFailed")
    case "RUNNING":
      return t("inspection.runRunning")
    default:
      return t("inspection.runPending")
  }
}

function runDuration(run: InspectionRun): string {
  if (!run.startedAt || !run.finishedAt) {
    return "—"
  }
  const ms = Date.parse(run.finishedAt) - Date.parse(run.startedAt)
  if (!Number.isFinite(ms) || ms < 0) {
    return "—"
  }
  return `${(ms / 1000).toFixed(1)}s`
}

export function InspectionRunPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [reportRun, setReportRun] = useState<InspectionRun | null>(null)
  const [reportOpen, setReportOpen] = useState(false)
  const [shareRun, setShareRun] = useState<InspectionRun | null>(null)
  const [shareOpen, setShareOpen] = useState(false)
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({
    status: normalizeRunStatus(query.status),
  })
  const canShare = session.can("inspection:share")
  const status = normalizeRunStatus(query.status)

  const { items, total, loading, failed, error, reload } =
    usePagedList<InspectionRun>({
      page: query.page,
      pageSize: query.pageSize,
      filterKey: `${session.namespaceUid}|${status}`,
      fetch: (page, pageSize) => listInspectionRuns({ page, pageSize, status }),
    })

  function openReport(run: InspectionRun) {
    setReportRun(run)
    setReportOpen(true)
  }

  function openShare(run: InspectionRun) {
    setShareRun(run)
    setShareOpen(true)
  }

  const filterChips: FilterChip[] = []
  if (status) {
    filterChips.push({
      key: "status",
      label: `${t("inspection.runStatus")}: ${runLabel(t, status)}`,
      onRemove: () => query.update({ status: "" }),
    })
  }

  function applyFilters() {
    query.update({ status: draft.status })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ status: "" })
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("inspection.runs")}
        description={t("inspection.runsDesc")}
      />
      <ListFilter
        id="inspection-run-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={filterChips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={filterChips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="w-40 shrink-0">
          <FieldLabel
            htmlFor="inspection-run-status"
            className="text-xs text-muted-foreground"
          >
            {t("inspection.runStatus")}
          </FieldLabel>
          <Select
            value={draft.status || ALL_STATUS}
            onValueChange={(value) =>
              update(
                "status",
                normalizeRunStatus(value === ALL_STATUS ? "" : value)
              )
            }
          >
            <SelectTrigger id="inspection-run-status" className="w-full">
              <SelectValue placeholder={t("inspection.all")}>
                <span className="flex items-center gap-2">
                  <EnumIcon value={draft.status} />
                  <span>
                    {draft.status ? runLabel(t, draft.status) : t("inspection.all")}
                  </span>
                </span>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_STATUS}>{t("inspection.all")}</SelectItem>
              <SelectItem value="SUCCEEDED">
                <span className="flex items-center gap-2">
                  <EnumIcon value="SUCCEEDED" />
                  <span>{t("inspection.runSucceeded")}</span>
                </span>
              </SelectItem>
              <SelectItem value="PARTIAL">
                <span className="flex items-center gap-2">
                  <EnumIcon value="PARTIAL" />
                  <span>{t("inspection.runPartial")}</span>
                </span>
              </SelectItem>
              <SelectItem value="FAILED">
                <span className="flex items-center gap-2">
                  <EnumIcon value="FAILED" />
                  <span>{t("inspection.runFailed")}</span>
                </span>
              </SelectItem>
              <SelectItem value="RUNNING">
                <span className="flex items-center gap-2">
                  <EnumIcon value="RUNNING" />
                  <span>{t("inspection.runRunning")}</span>
                </span>
              </SelectItem>
            </SelectContent>
          </Select>
        </Field>
      </ListFilter>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? (
          <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
        ) : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">
              {error || t("inspection.failed")}
            </p>
            <Button variant="outline" onClick={() => reload()}>
              {t("inspection.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("inspection.empty")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("inspection.planName")}</TableHead>
                <TableStatusHead>{t("inspection.runStatus")}</TableStatusHead>
                <TableHead className={TABLE_NOWRAP}>{t("inspection.trigger")}</TableHead>
                <TableHead className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}>
                  {t("inspection.startedAt")}
                </TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("inspection.duration")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("inspection.itemCounts")}</TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((run) => (
                <TableRow key={run.uid}>
                  <TableCell>
                    <div className="grid gap-1">
                      <span className="font-medium break-words">{run.planName}</span>
                      <span className="font-mono text-xs text-muted-foreground">
                        {run.uid}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <StatusBadge
                      value={run.status}
                      label={runLabel(t, run.status)}
                      tone={runTone(run.status)}
                    />
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {run.trigger === "MANUAL"
                      ? t("inspection.manual")
                      : t("inspection.scheduled")}
                  </TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>
                    {formatDateTime(run.startedAt) || "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {runDuration(run)}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {run.itemOk} / {run.itemWarn} / {run.itemFailed}
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap justify-center gap-2">
                      <Button
                        variant="link"
                        size="sm"
                        className="h-auto px-0"
                        onClick={() => openReport(run)}
                      >
                        {t("inspection.viewReport")}
                      </Button>
                      {canShare ? (
                        <Button
                          variant="link"
                          size="sm"
                          className="h-auto px-0"
                          onClick={() => openShare(run)}
                        >
                          {t("inspection.share")}
                        </Button>
                      ) : null}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <InspectionReportSheet
        key={`${reportRun?.uid ?? "none"}-${reportOpen}`}
        open={reportOpen}
        run={reportRun}
        onOpenChange={setReportOpen}
      />
      <InspectionReportShareDialog
        key={`${shareRun?.uid ?? "none"}-${shareOpen}`}
        open={shareOpen}
        run={shareRun}
        onOpenChange={setShareOpen}
      />
    </section>
  )
}
