import { useState } from "react"
import { useTranslation } from "react-i18next"

import {
  getKubernetesPersistentVolumeClaimDetail,
  listKubernetesPersistentVolumeClaims,
} from "@/api/cluster"
import type {
  GetKubernetesPersistentVolumeClaimDetailResult,
  KubernetesPersistentVolumeClaimItem,
  KubernetesWorkloadRefItem,
} from "@/api/cluster"
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
import {
  ClusterRelatedWorkloadList,
  ClusterRelationBody,
  useClusterRelationJump,
} from "./relation"
import { ClusterResourceState } from "./resource-state"
import { useClusterResourceDetail } from "./use-resource-detail"

interface KubernetesPersistentVolumeClaimsTabProps {
  clusterUid: string
  namespace: string
  name: string
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

function claimTone(phase: string): BadgeTone {
  switch (phase) {
    case "Bound":
      return "success"
    case "Pending":
      return "warning"
    case "Lost":
      return "danger"
    default:
      return "neutral"
  }
}

function joinOrDash(values: string[] | undefined): string {
  return values?.length ? values.join(", ") : "—"
}

/** PersistentVolumeClaim 列表：命名空间存储申请，展示容量、绑定 PV 与阶段。 */
export function KubernetesPersistentVolumeClaimsTab({
  clusterUid,
  namespace,
  name,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: KubernetesPersistentVolumeClaimsTabProps) {
  const { t } = useTranslation()
  const [detail, setDetail] =
    useState<KubernetesPersistentVolumeClaimItem | null>(null)
  const claimRelation =
    useClusterResourceDetail<GetKubernetesPersistentVolumeClaimDetailResult>({
      enabled: detail !== null,
      clusterUid,
      namespace: detail?.namespace ?? "",
      name: detail?.name ?? "",
      fetch: (uid, ns, nm) =>
        getKubernetesPersistentVolumeClaimDetail(uid, {
          namespace: ns,
          name: nm,
        }),
    })
  const jumpTo = useClusterRelationJump()

  function openClaimWorkload(ref: KubernetesWorkloadRefItem) {
    setDetail(null)
    jumpTo({
      tab: "workloads",
      namespace: ref.namespace,
      name: ref.name,
      kind: ref.kind,
    })
  }

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesPersistentVolumeClaimItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${namespace}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesPersistentVolumeClaims(clusterUid, {
          namespace: namespace || undefined,
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
            title={t("kube.pvcsEmpty")}
            hint={t("kube.pvcsEmptyHint")}
          />
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("kube.name")}</TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.identity} ${TABLE_NOWRAP}`}
                >
                  {t("kube.namespace")}
                </TableHead>
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
                  {t("kube.volumeName")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.phase")}
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
              {items.map((claim) => (
                <TableRow key={`${claim.namespace}/${claim.name}`}>
                  <TableCell className="font-medium break-words">
                    {claim.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {claim.namespace}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {claim.capacity || "—"}
                  </TableCell>
                  <TableCell className="max-w-[12rem] font-mono text-xs break-all whitespace-normal">
                    {claim.storageClass || "—"}
                  </TableCell>
                  <TableCell className="max-w-[12rem] text-xs break-words whitespace-normal">
                    {joinOrDash(claim.accessModes)}
                  </TableCell>
                  <TableCell className="max-w-[14rem] font-mono text-xs break-all whitespace-normal">
                    {claim.volumeName || "—"}
                  </TableCell>
                  <TableCell>
                    {claim.phase ? (
                      <StatusBadge
                        value={claim.phase}
                        tone={claimTone(claim.phase)}
                      />
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(claim.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(claim)}
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
          detail ? (
            <>
              <span className="font-mono text-xs">{detail.namespace}</span>
              {detail.phase ? (
                <StatusBadge
                  value={detail.phase}
                  tone={claimTone(detail.phase)}
                />
              ) : null}
            </>
          ) : null
        }
        info={
          detail
            ? [
                {
                  title: t("kube.sectionBasic"),
                  fields: [
                    { label: t("kube.namespace"), value: detail.namespace },
                    {
                      label: t("kube.phase"),
                      value: detail.phase ? (
                        <StatusBadge
                          value={detail.phase}
                          tone={claimTone(detail.phase)}
                        />
                      ) : null,
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
                      label: t("kube.volumeName"),
                      value: (
                        <DetailCopyText
                          value={detail.volumeName}
                          label={t("kube.volumeName")}
                        />
                      ),
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionRelatedWorkloads"),
                  description: claimRelation.data?.relatedWorkloads?.length
                    ? t("kube.relationsCount", {
                        count: claimRelation.data.relatedWorkloads.length,
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={claimRelation.loading}
                      failed={claimRelation.failed}
                      error={claimRelation.error}
                      onRetry={claimRelation.retry}
                      empty={
                        !claimRelation.loading &&
                        !claimRelation.failed &&
                        (claimRelation.data?.relatedWorkloads?.length ?? 0) ===
                          0
                      }
                    >
                      <ClusterRelatedWorkloadList
                        items={claimRelation.data?.relatedWorkloads ?? []}
                        onOpen={openClaimWorkload}
                      />
                    </ClusterRelationBody>
                  ),
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
            ? {
                clusterUid,
                kind: "PersistentVolumeClaim",
                namespace: detail.namespace,
                name: detail.name,
              }
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
