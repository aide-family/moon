import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  deleteKubernetesJob,
  getKubernetesJobDetail,
  listKubernetesJobs,
} from "@/api/cluster"
import type {
  GetKubernetesJobDetailResult,
  KubernetesJobItem,
  KubernetesPodItem,
} from "@/api/cluster"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailCopyText, DetailTagList } from "@/components/detail-values"
import { ListFrame } from "@/components/list-frame"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
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
import {
  formatClusterAge,
  formatClusterTime,
  formatJobStatus,
  readJobStatusTone,
  readPodPhaseTone,
} from "./format"
import {
  ClusterRelationBody,
  ClusterRelationList,
  useClusterRelationJump,
} from "./relation"
import { ClusterResourceState } from "./resource-state"
import { useClusterResourceDetail } from "./use-resource-detail"

interface ClusterJobsTabProps {
  clusterUid: string
  namespace: string
  name: string
  canWrite: boolean
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

/**
 * Job 列表：执行状态、完成数与成功/失败/活跃计数、所属控制器与时间。
 * 行操作是「详情 + 更多（删除）」；详情抽屉展示字段与 Job 创建的 Pod。
 */
export function ClusterJobsTab({
  clusterUid,
  namespace,
  name,
  canWrite,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: ClusterJobsTabProps) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<KubernetesJobItem | null>(null)
  const [deleting, setDeleting] = useState<KubernetesJobItem | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const jumpTo = useClusterRelationJump()
  const jobRelation = useClusterResourceDetail<GetKubernetesJobDetailResult>({
    enabled: detail !== null,
    clusterUid,
    namespace: detail?.namespace ?? "",
    name: detail?.name ?? "",
    fetch: (uid, ns, nm) =>
      getKubernetesJobDetail(uid, { namespace: ns, name: nm }),
  })

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesJobItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${namespace}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesJobs(clusterUid, {
          namespace: namespace || undefined,
          name: name || undefined,
          page: nextPage,
          pageSize: nextPageSize,
        }),
    })

  function openPod(pod: KubernetesPodItem) {
    setDetail(null)
    jumpTo({ tab: "pods", namespace: pod.namespace, name: pod.name })
  }

  function jobActions(job: KubernetesJobItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      {
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setDeleteError("")
          setDeleting(job)
        },
      },
    ]
  }

  async function runDelete() {
    if (!deleting) {
      return
    }
    setDeletePending(true)
    setDeleteError("")
    try {
      await deleteKubernetesJob(clusterUid, deleting.namespace, deleting.name)
      setDeleting(null)
      reload()
      toast.success(t("kube.actionDone"))
    } catch (err) {
      setDeleteError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeletePending(false)
    }
  }

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
            title={t("kube.jobsEmpty")}
            hint={t("kube.jobsEmptyHint")}
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
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.status")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.completions")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.jobProgress")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.ownerCronJob")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("kube.startTime")}
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
              {items.map((job) => (
                <TableRow key={`${job.namespace}/${job.name}`}>
                  <TableCell className="font-mono text-xs font-medium break-all">
                    {job.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {job.namespace}
                  </TableCell>
                  <TableCell>
                    <StatusBadge
                      value={job.status}
                      label={formatJobStatus(t, job.status)}
                      tone={readJobStatusTone(job.status)}
                    />
                  </TableCell>
                  <TableCell className="font-mono text-xs tabular-nums">
                    {job.completions}
                  </TableCell>
                  <TableCell>
                    <span className="font-mono text-xs tabular-nums">
                      <span aria-hidden="true">
                        {job.succeeded}/{job.failed}/{job.active}
                      </span>
                      <span className="sr-only">
                        {t("kube.succeeded")} {job.succeeded},{" "}
                        {t("kube.failed")} {job.failed}, {t("kube.active")}{" "}
                        {job.active}
                      </span>
                    </span>
                  </TableCell>
                  <TableCell className="max-w-[14rem] text-xs break-words whitespace-normal text-muted-foreground">
                    {job.ownerName
                      ? job.ownerKind
                        ? `${job.ownerKind}/${job.ownerName}`
                        : job.ownerName
                      : "—"}
                  </TableCell>
                  <TableCell className="text-xs tabular-nums">
                    <span className="whitespace-nowrap">
                      {formatClusterTime(job.startTime) || "—"}
                    </span>
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(job.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(job)}
                      actions={jobActions(job)}
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
              <StatusBadge
                value={detail.status}
                label={formatJobStatus(t, detail.status)}
                tone={readJobStatusTone(detail.status)}
              />
              <span className="font-mono text-xs">{detail.namespace}</span>
            </>
          ) : null
        }
        info={
          detail
            ? [
                {
                  title: t("kube.sectionBasic"),
                  fields: [
                    {
                      label: t("kube.namespace"),
                      value: (
                        <DetailCopyText
                          value={detail.namespace}
                          label={t("kube.namespace")}
                        />
                      ),
                    },
                    {
                      label: t("kube.kind"),
                      value: t("kube.job"),
                    },
                    {
                      label: t("kube.status"),
                      value: (
                        <StatusBadge
                          value={detail.status}
                          label={formatJobStatus(t, detail.status)}
                          tone={readJobStatusTone(detail.status)}
                        />
                      ),
                    },
                    {
                      label: t("kube.ownerCronJob"),
                      value: detail.ownerName ? (
                        <DetailCopyText
                          value={
                            detail.ownerKind
                              ? `${detail.ownerKind}/${detail.ownerName}`
                              : detail.ownerName
                          }
                          label={t("kube.ownerCronJob")}
                        />
                      ) : null,
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionJobProgress"),
                  fields: [
                    {
                      label: t("kube.completions"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.completions}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.parallelism"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.parallelism}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.active"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.active}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.succeeded"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.succeeded}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.failed"),
                      value: (
                        <span
                          className={
                            detail.failed > 0
                              ? "font-mono text-xs text-danger tabular-nums"
                              : "font-mono text-xs tabular-nums"
                          }
                        >
                          {detail.failed}
                        </span>
                      ),
                    },
                  ],
                },
                {
                  title: t("kube.images"),
                  description: detail.images?.length
                    ? t("kube.imagesCount", { count: detail.images.length })
                    : undefined,
                  content: (
                    <DetailTagList
                      items={detail.images ?? []}
                      copyable
                      copyLabel={t("kube.images")}
                    />
                  ),
                },
                {
                  title: t("kube.sectionPods"),
                  description: jobRelation.data?.pods?.length
                    ? t("kube.relationsCount", {
                        count: jobRelation.data.pods.length,
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={jobRelation.loading}
                      failed={jobRelation.failed}
                      error={jobRelation.error}
                      onRetry={jobRelation.retry}
                      empty={
                        !jobRelation.loading &&
                        !jobRelation.failed &&
                        (jobRelation.data?.pods?.length ?? 0) === 0
                      }
                    >
                      <ClusterRelationList
                        items={(jobRelation.data?.pods ?? []).map((pod) => ({
                          key: `${pod.namespace}/${pod.name}`,
                          title: (
                            <span className="font-mono text-xs break-all">
                              {pod.name}
                            </span>
                          ),
                          status: (
                            <span className="flex items-center gap-1.5">
                              <StatusBadge
                                value={pod.phase}
                                tone={readPodPhaseTone(pod.phase)}
                              />
                              <span className="font-mono tabular-nums">
                                {pod.ready || "—"}
                              </span>
                            </span>
                          ),
                          subtitle: pod.nodeName ? (
                            <span>{pod.nodeName}</span>
                          ) : undefined,
                          meta: (
                            <span>
                              {t("kube.restarts")}: {pod.restarts}
                            </span>
                          ),
                          onSelect: () => openPod(pod),
                          ariaLabel: t("kube.openResource", {
                            name: pod.name,
                          }),
                        }))}
                      />
                    </ClusterRelationBody>
                  ),
                },
                {
                  title: t("kube.sectionTime"),
                  fields: [
                    {
                      label: t("kube.startTime"),
                      value: formatClusterTime(detail.startTime) || "—",
                    },
                    {
                      label: t("kube.completionTime"),
                      value: formatClusterTime(detail.completionTime) || "—",
                    },
                    {
                      label: t("kube.createdAt"),
                      value: formatClusterTime(detail.createdAt) || "—",
                    },
                    {
                      label: t("kube.age"),
                      value: formatClusterAge(detail.createdAt, t),
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
                kind: "Job",
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
      <ConfirmAction
        open={deleting !== null}
        title={t("kube.deleteJobTitle")}
        description={
          deleting
            ? t("kube.deleteJobDescription", {
                namespace: deleting.namespace,
                name: deleting.name,
              })
            : ""
        }
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={deletePending}
        error={deleteError}
        onConfirm={() => void runDelete()}
        onOpenChange={(open) => {
          if (!open) {
            setDeleting(null)
            setDeleteError("")
          }
        }}
      />
    </>
  )
}
