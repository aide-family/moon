import { useTranslation } from "react-i18next"

import type { ClusterProbe } from "@/api/cluster"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

interface ClusterProbeDialogProps {
  open: boolean
  /** 被测集群名称；探测结果本身不带 name。 */
  name: string
  probe: ClusterProbe | null
  error: string
  onOpenChange: (open: boolean) => void
}

/** 已保存集群的连通性测试结果。成功展示探测到的元数据，失败展示服务端原因。 */
export function ClusterProbeDialog({
  open,
  name,
  probe,
  error,
  onOpenChange,
}: ClusterProbeDialogProps) {
  const { t } = useTranslation()
  const rows = probe
    ? [
        { label: t("cluster.server"), value: probe.server, mono: true },
        { label: t("cluster.version"), value: probe.version, mono: true },
        {
          label: t("cluster.clusterName"),
          value: probe.clusterName,
          mono: true,
        },
        {
          label: t("cluster.contextName"),
          value: probe.contextName,
          mono: true,
        },
        {
          label: t("cluster.nodeCount"),
          value: String(probe.nodeCount),
          mono: true,
        },
      ]
    : []

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("cluster.probeTitle")}</DialogTitle>
        </DialogHeader>
        <DialogBody className="grid min-w-0 gap-3">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <span className="min-w-0 truncate font-medium">{name}</span>
            <StatusBadge
              value={probe ? "OK" : "FAILED"}
              label={
                probe ? t("cluster.probeSuccess") : t("cluster.testFailed")
              }
              tone={probe ? "success" : "danger"}
            />
          </div>
          {probe ? (
            <dl className="grid gap-x-4 gap-y-2 text-sm sm:grid-cols-[9rem_minmax(0,1fr)]">
              {rows.map((row) => (
                <div key={row.label} className="contents">
                  <dt className="text-muted-foreground">{row.label}</dt>
                  <dd
                    className={
                      row.mono
                        ? "min-w-0 font-mono text-xs break-all"
                        : "min-w-0 break-words"
                    }
                  >
                    {row.value || "—"}
                  </dd>
                </div>
              ))}
            </dl>
          ) : (
            <p className="text-sm break-words text-destructive">
              {error || t("cluster.testFailed")}
            </p>
          )}
        </DialogBody>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("action.close")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
