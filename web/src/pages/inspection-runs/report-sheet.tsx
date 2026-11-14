import { Download, RefreshCw } from "lucide-react"
import { useEffect, useState } from "react"
import ReactMarkdown from "react-markdown"
import { useTranslation } from "react-i18next"
import remarkGfm from "remark-gfm"
import { toast } from "sonner"

import {
  getInspectionReport,
  listRunReports,
  resendInspectionReport,
} from "@/api/inspection"
import type { InspectionReport, InspectionRun } from "@/api/inspection/types"
import { DetailSheet } from "@/components/detail-sheet"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { formatDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"

/**
 * 报告阅读抽屉：宽 80vw，HTML 报告用 iframe 保留原样式，Markdown 报告按文档排版。
 * 版式变更只影响阅读，不改报告快照本身。
 */
export function InspectionReportSheet({
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
  const [loading, setLoading] = useState(true)
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
        const first =
          result.items.find((item) => item.format === "HTML") ?? result.items[0]
        if (!first) {
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
    anchor.download = `${active.title || "inspection-report"}.${
      active.format === "HTML" ? "html" : "md"
    }`
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

  const isHtml = active?.format === "HTML"

  return (
    <DetailSheet
      open={open}
      onOpenChange={onOpenChange}
      size="80vw"
      title={active?.title || run?.planName || t("inspection.reportTitle")}
      description={
        active ? (
          <>
            <StatusBadge value={active.format} label={active.format} tone={isHtml ? "info" : "violet"} />
            <span>{formatDateTime(active.createdAt) || "—"}</span>
            {run ? <span className="font-mono text-xs">{run.uid}</span> : null}
          </>
        ) : null
      }
      actions={
        active ? (
          <>
            {reports.length > 1
              ? reports.map((report) => (
                  <Button
                    key={report.uid}
                    type="button"
                    size="sm"
                    variant={active.uid === report.uid ? "secondary" : "outline"}
                    onClick={() => switchReport(report)}
                  >
                    {report.format}
                  </Button>
                ))
              : null}
            <Button type="button" size="sm" variant="outline" onClick={download}>
              <Download data-icon="inline-start" />
              {t("inspection.download")}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => void resend()}
              disabled={resending}
            >
              <RefreshCw data-icon="inline-start" />
              {t("inspection.resend")}
            </Button>
          </>
        ) : null
      }
      contentMode={isHtml ? "fill" : "scroll"}
    >
      {loading ? (
        <p className="py-10 text-center text-sm text-muted-foreground">{t("state.loading")}</p>
      ) : null}
      {!loading && reports.length === 0 ? (
        <p className="py-10 text-center text-sm text-muted-foreground">{t("inspection.noReport")}</p>
      ) : null}
      {!loading && active && isHtml ? (
        <iframe
          title={active.title || t("inspection.reportPreview")}
          sandbox=""
          srcDoc={content}
          className="min-h-0 w-full flex-1 border-0 bg-white"
        />
      ) : null}
      {!loading && active && !isHtml ? (
        <article className="prose prose-sm mx-auto w-full max-w-[880px] rounded-xl border border-border bg-card px-8 py-8 text-sm dark:prose-invert">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
        </article>
      ) : null}
    </DetailSheet>
  )
}
