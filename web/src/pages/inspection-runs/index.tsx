import { FileText, Search } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { listInspectionRuns } from "@/api/inspection"
import type { InspectionRun, InspectionRunStatus } from "@/api/inspection/types"
import { useSession } from "@/auth/session"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
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
  const canShare = session.can("inspection:share")
  const status =
    ["PENDING", "RUNNING", "SUCCEEDED", "PARTIAL", "FAILED", "CANCELED"].includes(
      query.status
    )
      ? (query.status as InspectionRunStatus)
      : ""

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

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("inspection.runs")}
        description={t("inspection.runsDesc")}
      />
      <form
        className="flex flex-wrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3"
        onSubmit={(event) => {
          event.preventDefault()
        }}
      >
        <Select
          value={status || ALL_STATUS}
          onValueChange={(value) =>
            query.update({ status: value === ALL_STATUS ? "" : value })
          }
        >
          <SelectTrigger className="w-40" aria-label={t("inspection.runStatus")}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL_STATUS}>{t("inspection.all")}</SelectItem>
            <SelectItem value="SUCCEEDED">{t("inspection.runSucceeded")}</SelectItem>
            <SelectItem value="PARTIAL">{t("inspection.runPartial")}</SelectItem>
            <SelectItem value="FAILED">{t("inspection.runFailed")}</SelectItem>
            <SelectItem value="RUNNING">{t("inspection.runRunning")}</SelectItem>
          </SelectContent>
        </Select>
        <Button type="submit" variant="outline">
          <Search data-icon="inline-start" />
          {t("filter.apply")}
        </Button>
      </form>
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
                    <div className="flex min-w-0 items-start gap-1">
                      <div className="grid min-w-0 gap-1">
                        <span className="font-medium break-words">{run.planName}</span>
                        <span className="font-mono text-xs text-muted-foreground">
                          {run.uid}
                        </span>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-xs"
                        className="mt-0.5 shrink-0"
                        aria-label={t("inspection.viewReport")}
                        title={t("inspection.viewReport")}
                        onClick={() => openReport(run)}
                      >
                        <FileText />
                      </Button>
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
                    <div className="flex justify-center">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="link" size="sm" className="h-auto px-0">
                            {t("action.more")}
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onSelect={() => openReport(run)}>
                            {t("inspection.viewReport")}
                          </DropdownMenuItem>
                          {canShare ? (
                            <DropdownMenuItem onSelect={() => openShare(run)}>
                              {t("inspection.share")}
                            </DropdownMenuItem>
                          ) : null}
                        </DropdownMenuContent>
                      </DropdownMenu>
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
