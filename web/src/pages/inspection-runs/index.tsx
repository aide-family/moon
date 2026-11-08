import { Copy, Download, Link2, Search } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getInspectionReport,
  listInspectionRuns,
  resendInspectionReport,
  listRunReports,
  revokeInspectionReportShare,
  shareInspectionReport,
} from "@/api/inspection"
import type {
  InspectionReport,
  InspectionRun,
  InspectionRunStatus,
} from "@/api/inspection/types"
import { useSession } from "@/auth/session"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
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
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_COLUMN_WIDTH, TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

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

function runLabel(t: (key: string) => string, status: InspectionRunStatus): string {
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

function ReportDialog({
  open,
  run,
  onOpenChange,
}: {
  open: boolean
  run: InspectionRun | null
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const [reports, setReports] = useState<InspectionReport[]>([])
  const [active, setActive] = useState<InspectionReport | null>(null)
  const [content, setContent] = useState("")
  const [loading, setLoading] = useState(false)
  const [shareToken, setShareToken] = useState("")
  const [sharing, setSharing] = useState(false)
  const [resending, setResending] = useState(false)

  useEffect(() => {
    if (!open || !run) {
      return
    }
    let cancelled = false
    listRunReports(run.uid)
      .then((result) => {
        if (cancelled) {
          return
        }
        setReports(result.items)
        const first = result.items.find((item) => item.format === "HTML") ?? result.items[0]
        if (!first) {
          setLoading(false)
          return
        }
        setActive(first)
        return getInspectionReport(first.uid).then((response) => {
          if (!cancelled) {
            setContent(response.report.content)
          }
        })
      })
      .catch((err) => {
        if (!cancelled) {
          toast.error(formErrorMessage(err, t("state.failed")))
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })
    return () => {
      cancelled = true
    }
  }, [open, run, t])

  function switchReport(report: InspectionReport) {
    setActive(report)
    setContent("")
    setLoading(true)
    getInspectionReport(report.uid)
      .then((response) => setContent(response.report.content))
      .catch((err) => toast.error(formErrorMessage(err, t("state.failed"))))
      .finally(() => setLoading(false))
  }

  function download() {
    if (!active || !content) {
      return
    }
    const blob = new Blob([content], {
      type: active.format === "HTML" ? "text/html" : "text/markdown",
    })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement("a")
    anchor.href = url
    anchor.download = `${active.title || "inspection-report"}.${active.format === "HTML" ? "html" : "md"}`
    anchor.click()
    URL.revokeObjectURL(url)
  }

  async function resend() {
    if (!active || resending) {
      return
    }
    setResending(true)
    try {
      await resendInspectionReport(active.uid)
      toast.success(t("inspection.resent"))
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setResending(false)
    }
  }

  async function share() {
    if (!active || sharing) {
      return
    }
    setSharing(true)
    try {
      const result = await shareInspectionReport(active.uid, 7, "")
      setShareToken(result.token)
      toast.success(t("inspection.shared"))
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setSharing(false)
    }
  }

  async function revoke() {
    if (!active) {
      return
    }
    try {
      await revokeInspectionReportShare(active.uid)
      setShareToken("")
      toast.success(t("inspection.shareRevoked"))
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  async function copy() {
    if (!shareToken) {
      return
    }
    const url = `${window.location.origin}/share/report/${shareToken}`
    try {
      await navigator.clipboard.writeText(url)
      toast.success(t("inspection.shareCopied"))
    } catch {
      toast.error(t("action.copyFailed"))
    }
  }

  const isHtml = active?.format === "HTML"

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-4xl">
        <DialogHeader>
          <DialogTitle>{run?.planName || t("inspection.reportTitle")}</DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-3">
          {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
          {!loading && reports.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("inspection.noReport")}</p>
          ) : null}
          {!loading && reports.length > 0 ? (
            <>
              <div className="flex flex-wrap items-center gap-2">
                {reports.map((report) => (
                  <Button
                    key={report.uid}
                    type="button"
                    size="sm"
                    variant={active?.uid === report.uid ? "secondary" : "outline"}
                    onClick={() => switchReport(report)}
                  >
                    {report.format}
                  </Button>
                ))}
                <span className="flex-1" />
                <Button type="button" size="sm" variant="outline" onClick={download}>
                  <Download data-icon="inline-start" />
                  {t("inspection.download")}
                </Button>
                <Button type="button" size="sm" variant="outline" onClick={() => void resend()} disabled={resending}>
                  {t("inspection.resend")}
                </Button>
                <Button type="button" size="sm" variant="outline" onClick={() => void share()} disabled={sharing}>
                  <Link2 data-icon="inline-start" />
                  {t("inspection.share")}
                </Button>
              </div>
              {shareToken ? (
                <div className="flex flex-wrap items-center gap-2">
                  <Input readOnly value={`${window.location.origin}/share/report/${shareToken}`} className="min-w-0 flex-1 font-mono text-xs" />
                  <Button type="button" size="sm" variant="outline" onClick={() => void copy()}>
                    <Copy data-icon="inline-start" />
                    {t("inspection.shareCopy")}
                  </Button>
                  <Button type="button" size="sm" variant="ghost" className="text-destructive" onClick={() => void revoke()}>
                    {t("inspection.shareDisable")}
                  </Button>
                </div>
              ) : null}
              {isHtml ? (
                <iframe
                  title={t("inspection.reportPreview")}
                  sandbox=""
                  srcDoc={content}
                  className="h-[520px] w-full rounded-lg border border-border bg-white"
                />
              ) : (
                <pre className="max-h-[520px] overflow-auto rounded-lg border border-border bg-muted/40 p-3 font-mono text-xs whitespace-pre-wrap">
                  {content}
                </pre>
              )}
            </>
          ) : null}
        </DialogBody>
      </DialogContent>
    </Dialog>
  )
}

export function InspectionRunPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [reportRun, setReportRun] = useState<InspectionRun | null>(null)
  const [reportOpen, setReportOpen] = useState(false)
  const status =
    ["PENDING", "RUNNING", "SUCCEEDED", "PARTIAL", "FAILED", "CANCELED"].includes(query.status)
      ? (query.status as InspectionRunStatus)
      : ""

  const { items, total, loading, failed, error, reload } = usePagedList<InspectionRun>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${status}`,
    fetch: (page, pageSize) => listInspectionRuns({ page, pageSize, status }),
  })

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("inspection.runs")} description={t("inspection.runsDesc")} />
      <form
        className="flex flex-wrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3"
        onSubmit={(event) => {
          event.preventDefault()
        }}
      >
        <Select
          value={status || ALL_STATUS}
          onValueChange={(value) => query.update({ status: value === ALL_STATUS ? "" : value })}
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
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{error || t("inspection.failed")}</p>
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
                <TableActionsHead>{t("inspection.report")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((run) => (
                <TableRow key={run.uid}>
                  <TableCell>
                    <div className="grid gap-1">
                      <span className="font-medium">{run.planName}</span>
                      <span className="font-mono text-xs text-muted-foreground">{run.uid}</span>
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
                    {run.trigger === "MANUAL" ? t("inspection.manual") : t("inspection.scheduled")}
                  </TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>
                    {formatDateTime(run.startedAt) || "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{runDuration(run)}</TableCell>
                  <TableCell className="font-mono text-xs">
                    {run.itemOk} / {run.itemWarn} / {run.itemFailed}
                  </TableCell>
                  <TableCell>
                    <Button
                      type="button"
                      size="sm"
                      variant="link"
                      onClick={() => {
                        setReportRun(run)
                        setReportOpen(true)
                      }}
                    >
                      {t("inspection.viewReport")}
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <ReportDialog
        key={`${reportRun?.uid ?? "none"}-${reportOpen}`}
        open={reportOpen}
        run={reportRun}
        onOpenChange={setReportOpen}
      />
    </section>
  )
}
