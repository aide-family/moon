import { useTranslation } from "react-i18next"

import type { ClusterItem } from "@/api/cluster"
import { StatusDot } from "@/components/status-dot"
import { cn } from "@/lib/class-names"

import { formatClusterTime } from "../format"

interface ClusterSummaryProps {
  cluster: ClusterItem
}

/**
 * 集群元数据条：连接状态、版本、Context 与最近检测时间。
 * 只做只读展示，操作入口留在页头，避免信息块里再出现按钮。
 */
export function ClusterSummary({ cluster }: ClusterSummaryProps) {
  const { t } = useTranslation()
  const checked = Number(cluster.lastCheckedAt) > 0
  const failed = cluster.lastError !== ""
  const stateLabel = !checked
    ? t("kube.connectionUnknown")
    : failed
      ? t("kube.connectionFailed")
      : t("kube.connectionOk")

  return (
    <div className="flex shrink-0 flex-col gap-2 rounded-xl border border-border/60 bg-card/55 px-4 py-3">
      <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-xs xl:grid-cols-5">
        <div className="flex min-w-0 flex-col gap-0.5">
          <dt className="text-[11px] text-muted-foreground">
            {t("kube.connection")}
          </dt>
          <dd
            className={cn(
              "flex min-w-0 items-center gap-1.5 font-medium",
              failed ? "text-destructive" : ""
            )}
          >
            {checked ? (
              <StatusDot active={!failed} />
            ) : (
              <span
                aria-hidden="true"
                className="size-2 shrink-0 rounded-full bg-muted-foreground/40"
              />
            )}
            <span className="truncate">{stateLabel}</span>
          </dd>
        </div>
        <div className="flex min-w-0 flex-col gap-0.5">
          <dt className="text-[11px] text-muted-foreground">
            {t("cluster.version")}
          </dt>
          <dd className="truncate font-mono text-[11px]">
            {cluster.version || "—"}
          </dd>
        </div>
        <div className="flex min-w-0 flex-col gap-0.5">
          <dt className="text-[11px] text-muted-foreground">
            {t("cluster.clusterName")} / {t("cluster.contextName")}
          </dt>
          <dd
            className="truncate font-mono text-[11px]"
            title={`${cluster.clusterName || "—"} / ${cluster.contextName || "—"}`}
          >
            {cluster.clusterName || "—"} / {cluster.contextName || "—"}
          </dd>
        </div>
        <div className="flex min-w-0 flex-col gap-0.5">
          <dt className="text-[11px] text-muted-foreground">
            {t("cluster.lastCheck")}
          </dt>
          <dd className="truncate font-mono text-[11px] tabular-nums">
            {formatClusterTime(cluster.lastCheckedAt) ||
              t("cluster.neverChecked")}
          </dd>
        </div>
        <div className="hidden min-w-0 flex-col gap-0.5 sm:flex">
          <dt className="text-[11px] text-muted-foreground">
            {t("cluster.createdAt")}
          </dt>
          <dd className="truncate font-mono text-[11px] tabular-nums">
            {formatClusterTime(cluster.createdAt) || "—"}
          </dd>
        </div>
      </dl>
      {cluster.remark ? (
        <p className="text-xs break-words text-muted-foreground">
          {cluster.remark}
        </p>
      ) : null}
      {cluster.lastError ? (
        <p className="rounded-md border border-destructive/20 bg-destructive/5 px-2 py-1.5 text-xs break-words text-destructive">
          {t("cluster.lastError")}: {cluster.lastError}
        </p>
      ) : null}
    </div>
  )
}
