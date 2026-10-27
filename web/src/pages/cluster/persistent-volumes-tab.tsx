import { useState } from "react"
import { useTranslation } from "react-i18next"

import { listKubernetesPersistentVolumes } from "@/api/cluster"
import type { KubernetesPersistentVolumeItem } from "@/api/cluster"
import { DetailCopyText, DetailTagList } from "@/components/detail-values"
import { ListFrame } from "@/components/list-frame"
import { RowActions } from "@/components/row-actions"
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
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { ClusterDetailSheet } from "./cluster-detail-sheet"
import { formatClusterAge, formatClusterTime } from "./format"
import { ClusterResourceState } from "./resource-state"

interface KubernetesPersistentVolumesTabProps {
  clusterUid: string
  name: string
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

function volumeTone(phase: string): BadgeTone {
  switch (phase) {
    case "Bound":
      return "success"
    case "Available":
      return "info"
    case "Released":
      return "warning"
    case "Failed":
      return "danger"
    default:
      return "neutral"
  }
}

function joinOrDash(values: string[] | undefined): string {
  return values?.length ? values.join(", ") : "—"
}

/** PersistentVolume 列表：集群级存储资源，展示容量、绑定关系与回收策略。 */
export function KubernetesPersistentVolumesTab({
  clusterUid,
  name,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: KubernetesPersistentVolumesTabProps) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<KubernetesPersistentVolumeItem | null>(
    null
  )

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesPersistentVolumeItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesPersistentVolumes(clusterUid, {
          name: name || undefined,
          page: nextPage,
          pageSize: nextPageSize,
        }),
    })

  return (
    <>
      <ListFrame
        total={total}
        page={page}
        pageSize={pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => onPageChange(Math.max(1, page - 1))}
        onNext={() => onPageChange(page + 1)}
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
            title={t("kube.pvsEmpty")}
            hint={t("kube.pvsEmptyHint")}
          />
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("kube.name")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>
                  {t("kube.volumeCapacity")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.identity} ${TABLE_NOWRAP}`}
                >
                  {t("kube.storageClass")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.accessModes")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.claim")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.phase")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.reclaimPolicy")}
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
              {items.map((volume) => (
                <TableRow key={volume.name}>
                  <TableCell className="font-mono text-xs font-medium break-all">
                    {volume.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {volume.capacity || "—"}
                  </TableCell>
                  <TableCell className="max-w-[12rem] font-mono text-xs break-all whitespace-normal">
                    {volume.storageClass || "—"}
                  </TableCell>
                  <TableCell className="max-w-[12rem] text-xs break-words whitespace-normal">
                    {joinOrDash(volume.accessModes)}
                  </TableCell>
                  <TableCell className="max-w-[14rem] font-mono text-xs break-all whitespace-normal">
                    {volume.claim || "—"}
                  </TableCell>
                  <TableCell>
                    {volume.phase ? (
                      <StatusBadge
                        value={volume.phase}
                        tone={volumeTone(volume.phase)}
                      />
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell className="text-xs">
                    {volume.reclaimPolicy || "—"}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(volume.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(volume)}
                      actions={[]}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <ClusterDetailSheet
        open={detail !== null}
        title={detail?.name ?? ""}
        size="wide"
        description={
          detail?.phase ? (
            <StatusBadge value={detail.phase} tone={volumeTone(detail.phase)} />
          ) : null
        }
        info={
          detail
            ? [
                {
                  title: t("kube.sectionBasic"),
                  fields: [
                    {
                      label: t("kube.phase"),
                      value: detail.phase ? (
                        <StatusBadge
                          value={detail.phase}
                          tone={volumeTone(detail.phase)}
                        />
                      ) : null,
                    },
                    {
                      label: t("kube.reclaimPolicy"),
                      value: detail.reclaimPolicy || null,
                    },
                  ],
                },
                {
                  title: t("kube.sectionStorage"),
                  fields: [
                    {
                      label: t("kube.volumeCapacity"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.capacity || "—"}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.storageClass"),
                      value: (
                        <DetailCopyText
                          value={detail.storageClass}
                          label={t("kube.storageClass")}
                        />
                      ),
                    },
                    {
                      label: t("kube.accessModes"),
                      value: (
                        <DetailTagList
                          items={detail.accessModes ?? []}
                          mono={false}
                        />
                      ),
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionBinding"),
                  fields: [
                    {
                      label: t("kube.claim"),
                      value: (
                        <DetailCopyText
                          value={detail.claim}
                          label={t("kube.claim")}
                        />
                      ),
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
        yaml={
          detail
            ? { clusterUid, kind: "PersistentVolume", name: detail.name }
            : null
        }
        onOpenChange={(open) => {
          if (!open) {
            setDetail(null)
          }
        }}
      />
    </>
  )
}
