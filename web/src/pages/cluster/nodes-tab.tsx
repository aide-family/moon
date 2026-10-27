import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  listKubernetesNodes,
  setKubernetesNodeUnschedulable,
} from "@/api/cluster"
import type { KubernetesNodeItem } from "@/api/cluster"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailCopyText, DetailTagList } from "@/components/detail-values"
import { ListFrame } from "@/components/list-frame"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge, type BadgeTone } from "@/components/status-badge"
import { TableActionsHead } from "@/components/table-heads"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { usePagedList } from "@/hooks/use-paged-list"
import { formErrorMessage } from "@/lib/form-error"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { ClusterDetailSheet } from "./cluster-detail-sheet"
import { formatClusterAge, formatClusterTime } from "./format"
import { ClusterResourceState } from "./resource-state"

interface KubernetesNodesTabProps {
  clusterUid: string
  /** 服务端节点名过滤；空表示不过滤。 */
  name: string
  canWrite: boolean
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

function nodeTone(status: string): BadgeTone {
  if (status === "Ready") {
    return "success"
  }
  if (status === "NotReady") {
    return "danger"
  }
  return "warning"
}

/** 节点列表：状态、调度状态、角色与容量，支持 cordon / uncordon。 */
export function KubernetesNodesTab({
  clusterUid,
  name,
  canWrite,
  refreshKey = 0,
}: KubernetesNodesTabProps) {
  const { t } = useTranslation()
  const [confirm, setConfirm] = useState<{
    node: KubernetesNodeItem
    unschedulable: boolean
  } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<KubernetesNodeItem | null>(null)

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesNodeItem>({
      page: 1,
      pageSize: 100,
      filterKey: `${clusterUid}|${name}`,
      refreshKey,
      fetch: () => listKubernetesNodes(clusterUid, { name: name || undefined }),
    })

  function nodeActions(node: KubernetesNodeItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      {
        key: "schedule",
        label: node.schedulable ? t("kube.cordon") : t("kube.uncordon"),
        onSelect: () => {
          setConfirmError("")
          setConfirm({
            node,
            unschedulable: node.schedulable,
          })
        },
      },
    ]
  }

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await setKubernetesNodeUnschedulable(
        clusterUid,
        confirm.node.name,
        confirm.unschedulable
      )
      setConfirm(null)
      reload()
      toast.success(t("kube.actionDone"))
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  return (
    <>
      <ListFrame
        total={total}
        page={1}
        pageSize={Math.max(total, 1)}
        showPager={false}
        onPrev={() => undefined}
        onNext={() => undefined}
      >
        {loading ? (
          <ClusterResourceState state="loading" title={t("state.loading")} />
        ) : null}
        {!loading && failed ? (
          <ClusterResourceState
            state="error"
            title={error || t("state.failed")}
            hint={t("kube.resourceErrorHint")}
            onRetry={() => reload()}
          />
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <ClusterResourceState
            state="empty"
            title={t("kube.nodesEmpty")}
            hint={t("kube.nodesEmptyHint")}
          />
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("kube.nodeName")}</TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.status")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.schedule")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.roles")}
                </TableHead>
                <TableHead className={TABLE_NOWRAP}>
                  {t("kube.internalIp")}
                </TableHead>
                <TableHead className={TABLE_NOWRAP}>
                  {t("kube.kubeletVersion")}
                </TableHead>
                <TableHead className={TABLE_NOWRAP}>
                  {t("kube.capacity")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.podCount")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("kube.age")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((node) => (
                <TableRow key={node.name}>
                  <TableCell className="font-mono text-xs font-medium break-all">
                    {node.name}
                  </TableCell>
                  <TableCell>
                    <StatusBadge
                      value={node.status}
                      tone={nodeTone(node.status)}
                    />
                  </TableCell>
                  <TableCell>
                    <StatusBadge
                      value={node.schedulable ? "SCHEDULABLE" : "UNSCHEDULABLE"}
                      label={
                        node.schedulable
                          ? t("kube.schedulable")
                          : t("kube.unschedulable")
                      }
                      tone={node.schedulable ? "success" : "warning"}
                    />
                  </TableCell>
                  <TableCell className="max-w-[12rem] text-xs break-words whitespace-normal text-muted-foreground">
                    {node.roles?.length ? node.roles.join(", ") : "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {node.internalIp || "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {node.kubeletVersion || "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {node.cpuCapacity || "—"} / {node.memoryCapacity || "—"}
                  </TableCell>
                  <TableCell className="text-xs tabular-nums">
                    {node.podCount}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(node.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(node)}
                      actions={nodeActions(node)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <ConfirmAction
        open={confirm !== null}
        title={
          confirm?.unschedulable
            ? t("kube.cordonTitle")
            : t("kube.uncordonTitle")
        }
        description={
          confirm?.unschedulable
            ? t("kube.cordonDescription", { name: confirm.node.name })
            : t("kube.uncordonDescription", {
                name: confirm?.node.name ?? "",
              })
        }
        confirmLabel={
          confirm?.unschedulable ? t("kube.cordon") : t("kube.uncordon")
        }
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runConfirm()}
        onOpenChange={(open) => {
          if (!open) {
            setConfirm(null)
            setConfirmError("")
          }
        }}
      />
      <ClusterDetailSheet
        open={detail !== null}
        title={detail?.name ?? ""}
        size="wide"
        info={
          detail
            ? [
                {
                  title: t("kube.sectionStatus"),
                  fields: [
                    {
                      label: t("kube.status"),
                      value: (
                        <StatusBadge
                          value={detail.status}
                          tone={nodeTone(detail.status)}
                        />
                      ),
                    },
                    {
                      label: t("kube.schedule"),
                      value: (
                        <StatusBadge
                          value={
                            detail.schedulable ? "SCHEDULABLE" : "UNSCHEDULABLE"
                          }
                          label={
                            detail.schedulable
                              ? t("kube.schedulable")
                              : t("kube.unschedulable")
                          }
                          tone={detail.schedulable ? "success" : "warning"}
                        />
                      ),
                    },
                    {
                      label: t("kube.roles"),
                      value: (
                        <DetailTagList
                          items={detail.roles ?? []}
                          mono={false}
                        />
                      ),
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionCapacity"),
                  fields: [
                    {
                      label: t("kube.capacity"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.cpuCapacity || "—"}
                          <span className="mx-1 text-muted-foreground">/</span>
                          {detail.memoryCapacity || "—"}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.podCount"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.podCount}
                        </span>
                      ),
                    },
                  ],
                },
                {
                  title: t("kube.sectionBasic"),
                  fields: [
                    {
                      label: t("kube.internalIp"),
                      value: (
                        <DetailCopyText
                          value={detail.internalIp}
                          label={t("kube.internalIp")}
                        />
                      ),
                    },
                    {
                      label: t("kube.kubeletVersion"),
                      value: (
                        <DetailCopyText
                          value={detail.kubeletVersion}
                          label={t("kube.kubeletVersion")}
                        />
                      ),
                    },
                    {
                      label: t("kube.osImage"),
                      value: detail.osImage ? (
                        <span className="text-xs break-words text-muted-foreground">
                          {detail.osImage}
                        </span>
                      ) : null,
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionTime"),
                  fields: [
                    {
                      label: t("kube.age"),
                      value: formatClusterAge(detail.createdAt, t),
                    },
                    {
                      label: t("kube.createdAt"),
                      value: formatClusterTime(detail.createdAt) || "—",
                    },
                  ],
                },
              ]
            : []
        }
        yaml={detail ? { clusterUid, kind: "Node", name: detail.name } : null}
        onOpenChange={(open) => {
          if (!open) {
            setDetail(null)
          }
        }}
      />
    </>
  )
}
