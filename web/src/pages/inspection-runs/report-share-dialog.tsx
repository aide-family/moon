import { Copy, Link2 } from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  listInspectionReportShares,
  listRunReports,
  revokeInspectionReportShare,
  revokeInspectionReportShareLink,
  shareInspectionReport,
} from "@/api/inspection"
import type {
  InspectionReport,
  InspectionReportShare,
  InspectionRun,
} from "@/api/inspection/types"
import { useSession } from "@/auth/session"
import { StatusBadge } from "@/components/status-badge"
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
import { formatDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"

function shareStatusLabel(
  t: (key: string) => string,
  status: InspectionReportShare["status"]
): string {
  switch (status) {
    case "ACTIVE":
      return t("inspection.shareStatusActive")
    case "EXPIRED":
      return t("inspection.shareStatusExpired")
    default:
      return t("inspection.shareStatusRevoked")
  }
}

/**
 * 分享链接管理模态框：从执行记录的操作菜单打开，列出这份报告的全部分享记录，
 * 可新建、复制新链接并逐条停用。token 只在创建时返回一次。
 */
export function InspectionReportShareDialog({
  open,
  run,
  onOpenChange,
}: {
  open: boolean
  run: InspectionRun | null
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const session = useSession()
  const canShare = session.can("inspection:share")
  const [active, setActive] = useState<InspectionReport | null>(null)
  const [shares, setShares] = useState<InspectionReportShare[]>([])
  const [loading, setLoading] = useState(true)
  const [sharesLoading, setSharesLoading] = useState(false)
  const [token, setToken] = useState("")
  const [days, setDays] = useState("7")
  const [password, setPassword] = useState("")
  const [pending, setPending] = useState(false)

  const loadShares = useCallback(
    (report: InspectionReport) => {
      setSharesLoading(true)
      listInspectionReportShares(report.uid)
        .then((result) => setShares(result.items))
        .catch((err) => toast.error(formErrorMessage(err, t("state.failed"))))
        .finally(() => setSharesLoading(false))
    },
    [t]
  )

  useEffect(() => {
    if (!open || !run || !canShare) {
      return
    }
    let cancelled = false
    listRunReports(run.uid)
      .then((result) => {
        if (cancelled) {
          return
        }
        const first =
          result.items.find((item) => item.format === "HTML") ?? result.items[0]
        if (!first) {
          return
        }
        setActive(first)
        loadShares(first)
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
  }, [open, run, canShare, t, loadShares])

  async function createShare() {
    if (!active || pending) {
      return
    }
    setPending(true)
    try {
      const result = await shareInspectionReport(active.uid, Number(days), password)
      setToken(result.token)
      setPassword("")
      loadShares(active)
      toast.success(t("inspection.shared"))
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function revokeShare(share: InspectionReportShare) {
    if (!active) {
      return
    }
    try {
      if (share.uid) {
        await revokeInspectionReportShareLink(active.uid, share.uid)
      } else {
        await revokeInspectionReportShare(active.uid)
      }
      setToken("")
      toast.success(t("inspection.shareRevoked"))
      loadShares(active)
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  async function copy() {
    if (!token) {
      return
    }
    try {
      await navigator.clipboard.writeText(
        `${window.location.origin}/share/report/${token}`
      )
      toast.success(t("inspection.shareCopied"))
    } catch {
      toast.error(t("action.copyFailed"))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("inspection.shareList")}</DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-3">
          {loading ? (
            <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
          ) : null}
          {!loading && !active ? (
            <p className="text-sm text-muted-foreground">{t("inspection.noReport")}</p>
          ) : null}
          {active ? (
            <>
              <p className="text-xs text-muted-foreground">
                {active.title} · {active.format} · {t("inspection.shareDialogDesc")}
              </p>
              <div className="grid gap-2 rounded-lg border border-border bg-muted/30 p-3 sm:grid-cols-[160px_minmax(0,1fr)_auto]">
                <Select value={days} onValueChange={setDays}>
                  <SelectTrigger aria-label={t("inspection.shareExpiry")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="7">{t("inspection.shareDays7")}</SelectItem>
                    <SelectItem value="30">{t("inspection.shareDays30")}</SelectItem>
                    <SelectItem value="0">{t("inspection.sharePermanent")}</SelectItem>
                  </SelectContent>
                </Select>
                <Input
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder={t("inspection.sharePassword")}
                  aria-label={t("inspection.sharePassword")}
                />
                <Button type="button" size="sm" onClick={() => void createShare()} disabled={pending}>
                  <Link2 data-icon="inline-start" />
                  {t("inspection.shareGenerate")}
                </Button>
              </div>
              {token ? (
                <div className="grid gap-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <Input
                      readOnly
                      value={`${window.location.origin}/share/report/${token}`}
                      className="min-w-0 flex-1 font-mono text-xs"
                    />
                    <Button type="button" size="sm" variant="outline" onClick={() => void copy()}>
                      <Copy data-icon="inline-start" />
                      {t("inspection.shareCopy")}
                    </Button>
                  </div>
                  <p className="text-[11px] text-muted-foreground">{t("inspection.shareTokenOnce")}</p>
                </div>
              ) : null}
              <div className="overflow-hidden rounded-lg border border-border">
                <div className="flex flex-wrap items-center gap-x-2 gap-y-1 border-b border-border bg-muted/30 px-3 py-2">
                  <span className="text-sm font-medium">{t("inspection.shareList")}</span>
                  <span className="text-xs text-muted-foreground">{t("inspection.shareListHint")}</span>
                </div>
                {sharesLoading ? (
                  <p className="px-3 py-3 text-xs text-muted-foreground">{t("state.loading")}</p>
                ) : null}
                {!sharesLoading && shares.length === 0 ? (
                  <p className="px-3 py-3 text-xs text-muted-foreground">{t("inspection.shareListEmpty")}</p>
                ) : null}
                {!sharesLoading && shares.length > 0 ? (
                  <ul className="divide-y divide-border">
                    {shares.map((share, index) => (
                      <li
                        key={share.uid || `legacy-${index}`}
                        className="flex flex-wrap items-center gap-3 px-3 py-2"
                      >
                        <StatusBadge
                          value={share.status}
                          label={shareStatusLabel(t, share.status)}
                        />
                        <div className="min-w-0 flex-1">
                          <p className="text-xs font-medium">
                            {share.uid ? t("inspection.share") : t("inspection.shareLegacy")}
                            {share.hasPassword ? ` · ${t("inspection.shareProtected")}` : ""}
                          </p>
                          <p className="text-[11px] text-muted-foreground">
                            {t("inspection.shareCreatedAt")} {formatDateTime(share.createdAt) || "—"}
                            {" · "}
                            {t("inspection.shareExpires")}{" "}
                            {share.expiresAt
                              ? formatDateTime(share.expiresAt)
                              : t("inspection.shareNeverExpires")}
                            {" · "}
                            {t("inspection.shareViews")} {share.viewCount}
                            {share.lastViewedAt
                              ? ` · ${t("inspection.shareLastViewed")} ${formatDateTime(share.lastViewedAt)}`
                              : ""}
                          </p>
                          {!share.uid ? (
                            <p className="text-[11px] text-muted-foreground">
                              {t("inspection.shareLegacyHint")}
                            </p>
                          ) : null}
                        </div>
                        {share.status === "ACTIVE" ? (
                          <Button
                            type="button"
                            size="sm"
                            variant="ghost"
                            className="text-destructive"
                            onClick={() => void revokeShare(share)}
                          >
                            {t("inspection.shareRevokeLink")}
                          </Button>
                        ) : null}
                      </li>
                    ))}
                  </ul>
                ) : null}
              </div>
            </>
          ) : null}
        </DialogBody>
      </DialogContent>
    </Dialog>
  )
}
