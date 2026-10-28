import { useState } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  deleteKubernetesCronJob,
  getKubernetesCronJobDetail,
  listKubernetesCronJobs,
  setKubernetesCronJobSuspend,
  triggerKubernetesCronJob,
} from "@/api/cluster"
import type {
  GetKubernetesCronJobDetailResult,
  KubernetesCronJobItem,
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
} from "./format"
import {
  ClusterRelationBody,
  ClusterRelationList,
  useClusterRelationJump,
} from "./relation"
import { ClusterResourceState } from "./resource-state"
import { useClusterResourceDetail } from "./use-resource-detail"

/** cronjob-detail 最多返回 10 条历史 Job，列表按这个上限展示。 */
const MAX_RECENT_JOBS = 10

interface ClusterCronJobsTabProps {
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

function concurrencyLabel(t: TFunction, policy: string): string {
  switch (policy) {
    case "Allow":
      return t("kube.concurrencyAllow")
    case "Forbid":
      return t("kube.concurrencyForbid")
    case "Replace":
      return t("kube.concurrencyReplace")
    default:
      return policy || "—"
  }
}

/**
 * CronJob 列表：调度、时区、挂起状态、并发策略与活跃 Job。
 * 行操作是「详情 + 更多（暂停/恢复、立即运行、删除）」；详情抽屉展示字段与最近 Job。
 */
export function ClusterCronJobsTab({
  clusterUid,
  namespace,
  name,
  canWrite,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: ClusterCronJobsTabProps) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<KubernetesCronJobItem | null>(null)
  const [suspendTarget, setSuspendTarget] =
    useState<KubernetesCronJobItem | null>(null)
  const [suspendPending, setSuspendPending] = useState(false)
  const [suspendError, setSuspendError] = useState("")
  const [triggerTarget, setTriggerTarget] =
    useState<KubernetesCronJobItem | null>(null)
  const [triggerPending, setTriggerPending] = useState(false)
  const [triggerError, setTriggerError] = useState("")
  const [deleting, setDeleting] = useState<KubernetesCronJobItem | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const jumpTo = useClusterRelationJump()
  const cronJobRelation =
    useClusterResourceDetail<GetKubernetesCronJobDetailResult>({
      enabled: detail !== null,
      clusterUid,
      namespace: detail?.namespace ?? "",
      name: detail?.name ?? "",
      fetch: (uid, ns, nm) =>
        getKubernetesCronJobDetail(uid, { namespace: ns, name: nm }),
    })

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesCronJobItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${namespace}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesCronJobs(clusterUid, {
          namespace: namespace || undefined,
          name: name || undefined,
          page: nextPage,
          pageSize: nextPageSize,
        }),
    })

  function openRecentJob(namespace: string, jobName: string) {
    setDetail(null)
    jumpTo({ tab: "jobs", namespace, name: jobName })
  }

  function cronJobActions(cronjob: KubernetesCronJobItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      {
        key: "suspend",
        label: cronjob.suspend ? t("kube.resume") : t("kube.suspend"),
        onSelect: () => {
          setSuspendError("")
          setSuspendTarget(cronjob)
        },
      },
      {
        key: "trigger",
        label: t("kube.trigger"),
        onSelect: () => {
          setTriggerError("")
          setTriggerTarget(cronjob)
        },
      },
      {
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setDeleteError("")
          setDeleting(cronjob)
        },
      },
    ]
  }

  async function runSuspend() {
    if (!suspendTarget) {
      return
    }
    setSuspendPending(true)
    setSuspendError("")
    try {
      await setKubernetesCronJobSuspend(clusterUid, {
        namespace: suspendTarget.namespace,
        name: suspendTarget.name,
        suspend: !suspendTarget.suspend,
      })
      setSuspendTarget(null)
      reload()
      toast.success(t("kube.actionDone"))
    } catch (err) {
      setSuspendError(formErrorMessage(err, t("state.failed")))
    } finally {
      setSuspendPending(false)
    }
  }

  async function runTrigger() {
    if (!triggerTarget) {
      return
    }
    setTriggerPending(true)
    setTriggerError("")
    try {
      const result = await triggerKubernetesCronJob(clusterUid, {
        namespace: triggerTarget.namespace,
        name: triggerTarget.name,
      })
      setTriggerTarget(null)
      reload()
      toast.success(
        t("kube.triggerDone"),
        result.job?.name ? { description: result.job.name } : undefined
      )
    } catch (err) {
      setTriggerError(formErrorMessage(err, t("state.failed")))
    } finally {
      setTriggerPending(false)
    }
  }

  async function runDelete() {
    if (!deleting) {
      return
    }
    setDeletePending(true)
    setDeleteError("")
    try {
      await deleteKubernetesCronJob(
        clusterUid,
        deleting.namespace,
        deleting.name
      )
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
            title={t("kube.cronjobsEmpty")}
            hint={t("kube.cronjobsEmptyHint")}
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
                  {t("kube.cronSchedule")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.identity} ${TABLE_NOWRAP}`}
                >
                  {t("kube.timeZone")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.suspend")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.concurrencyPolicy")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("kube.lastSchedule")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.activeJobs")}
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
              {items.map((cronjob) => (
                <TableRow key={`${cronjob.namespace}/${cronjob.name}`}>
                  <TableCell className="font-medium break-words">
                    {cronjob.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {cronjob.namespace}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {cronjob.schedule || "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {cronjob.timeZone || "—"}
                  </TableCell>
                  <TableCell>
                    <StatusBadge
                      value={cronjob.suspend ? "SUSPENDED" : "ACTIVE"}
                      label={cronjob.suspend ? t("kube.yes") : t("kube.no")}
                      tone={cronjob.suspend ? "warning" : "success"}
                    />
                  </TableCell>
                  <TableCell className="text-xs whitespace-nowrap">
                    {concurrencyLabel(t, cronjob.concurrencyPolicy)}
                  </TableCell>
                  <TableCell className="text-xs tabular-nums">
                    <span className="whitespace-nowrap">
                      {formatClusterTime(cronjob.lastScheduleTime) || "—"}
                    </span>
                  </TableCell>
                  <TableCell className="font-mono text-xs tabular-nums">
                    {cronjob.activeCount}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(cronjob.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(cronjob)}
                      actions={cronJobActions(cronjob)}
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
                value={detail.suspend ? "SUSPENDED" : "ACTIVE"}
                label={detail.suspend ? t("kube.suspended") : t("kube.running")}
                tone={detail.suspend ? "warning" : "success"}
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
                      value: t("kube.cronjob"),
                    },
                    {
                      label: t("kube.suspend"),
                      value: (
                        <StatusBadge
                          value={detail.suspend ? "SUSPENDED" : "ACTIVE"}
                          label={
                            detail.suspend
                              ? t("kube.suspended")
                              : t("kube.running")
                          }
                          tone={detail.suspend ? "warning" : "success"}
                        />
                      ),
                    },
                    {
                      label: t("kube.timeZone"),
                      value: detail.timeZone || null,
                    },
                    {
                      label: t("kube.concurrencyPolicy"),
                      value: concurrencyLabel(t, detail.concurrencyPolicy),
                    },
                  ],
                },
                {
                  title: t("kube.cronSchedule"),
                  fields: [
                    {
                      label: t("kube.cronExpression"),
                      value: (
                        <DetailCopyText
                          value={detail.schedule}
                          label={t("kube.cronExpression")}
                        />
                      ),
                    },
                    {
                      label: t("kube.lastSchedule"),
                      value: formatClusterTime(detail.lastScheduleTime) || "—",
                    },
                    {
                      label: t("kube.activeJobs"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.activeCount}
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
                  title: t("kube.recentJobs"),
                  description: cronJobRelation.data?.jobs?.length
                    ? t("kube.relationsCount", {
                        count: Math.min(
                          cronJobRelation.data.jobs.length,
                          MAX_RECENT_JOBS
                        ),
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={cronJobRelation.loading}
                      failed={cronJobRelation.failed}
                      error={cronJobRelation.error}
                      onRetry={cronJobRelation.retry}
                      empty={
                        !cronJobRelation.loading &&
                        !cronJobRelation.failed &&
                        (cronJobRelation.data?.jobs?.length ?? 0) === 0
                      }
                    >
                      <ClusterRelationList
                        items={(cronJobRelation.data?.jobs ?? [])
                          .slice(0, MAX_RECENT_JOBS)
                          .map((job) => ({
                            key: `${job.namespace}/${job.name}`,
                            title: (
                              <span className="font-mono text-xs break-all">
                                {job.name}
                              </span>
                            ),
                            status: (
                              <StatusBadge
                                value={job.status}
                                label={formatJobStatus(t, job.status)}
                                tone={readJobStatusTone(job.status)}
                              />
                            ),
                            meta: (
                              <>
                                <span>
                                  {t("kube.succeeded")}: {job.succeeded}
                                </span>
                                <span>
                                  {t("kube.failed")}: {job.failed}
                                </span>
                                <span>
                                  {t("kube.active")}: {job.active}
                                </span>
                                <span>
                                  {formatClusterAge(job.createdAt, t)}
                                </span>
                              </>
                            ),
                            onSelect: () =>
                              openRecentJob(job.namespace, job.name),
                            ariaLabel: t("kube.openResource", {
                              name: job.name,
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
                kind: "CronJob",
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
        open={suspendTarget !== null}
        title={
          suspendTarget?.suspend
            ? t("kube.resumeTitle")
            : t("kube.suspendTitle")
        }
        description={
          suspendTarget
            ? t(
                suspendTarget.suspend
                  ? "kube.resumeDescription"
                  : "kube.suspendDescription",
                {
                  namespace: suspendTarget.namespace,
                  name: suspendTarget.name,
                }
              )
            : ""
        }
        confirmLabel={
          suspendTarget?.suspend ? t("kube.resume") : t("kube.suspend")
        }
        cancelLabel={t("action.cancel")}
        pending={suspendPending}
        error={suspendError}
        onConfirm={() => void runSuspend()}
        onOpenChange={(open) => {
          if (!open) {
            setSuspendTarget(null)
            setSuspendError("")
          }
        }}
      />
      <ConfirmAction
        open={triggerTarget !== null}
        title={t("kube.triggerTitle")}
        description={
          triggerTarget
            ? t("kube.triggerDescription", {
                namespace: triggerTarget.namespace,
                name: triggerTarget.name,
              })
            : ""
        }
        confirmLabel={t("kube.trigger")}
        cancelLabel={t("action.cancel")}
        pending={triggerPending}
        error={triggerError}
        onConfirm={() => void runTrigger()}
        onOpenChange={(open) => {
          if (!open) {
            setTriggerTarget(null)
            setTriggerError("")
          }
        }}
      />
      <ConfirmAction
        open={deleting !== null}
        title={t("kube.deleteCronJobTitle")}
        description={
          deleting
            ? t("kube.deleteCronJobDescription", {
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
