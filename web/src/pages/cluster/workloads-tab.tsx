import { Minus, Plus } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getKubernetesWorkloadDetail,
  listKubernetesWorkloads,
  restartKubernetesWorkload,
  scaleKubernetesWorkload,
} from "@/api/cluster"
import type {
  GetKubernetesWorkloadDetailResult,
  KubernetesPodItem,
  KubernetesWorkloadItem,
} from "@/api/cluster"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailTagList } from "@/components/detail-values"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge, type BadgeTone } from "@/components/status-badge"
import { TableActionsHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { usePagedList } from "@/hooks/use-paged-list"
import { formErrorMessage, readFormError } from "@/lib/form-error"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { ClusterDetailSheet } from "./cluster-detail-sheet"
import { formatClusterAge, formatClusterTime } from "./format"
import {
  ClusterRelationBody,
  ClusterRelationList,
  useClusterRelationJump,
} from "./relation"
import { ClusterResourceState } from "./resource-state"
import { useClusterResourceDetail } from "./use-resource-detail"
import { WorkloadImageDialog } from "./workload-image-dialog"
import { WorkloadRollbackDialog } from "./workload-rollback-dialog"

// 副本数上限由产品约束固定为 100；下限 0 表示缩容到零。
const MIN_REPLICAS = 0
const MAX_REPLICAS = 100

function workloadTone(kind: string): BadgeTone {
  if (kind === "Deployment") {
    return "info"
  }
  if (kind === "StatefulSet") {
    return "primary"
  }
  return "warning"
}

function podPhaseTone(phase: string): BadgeTone {
  switch (phase) {
    case "Running":
      return "success"
    case "Pending":
      return "warning"
    case "Failed":
      return "danger"
    case "Succeeded":
      return "info"
    default:
      return "neutral"
  }
}

// 副本数只接受非负整数；0 是合法值（缩容到零），不能用「大于 0」的通用校验。
function readReplicas(value: string): number | null {
  const trimmed = value.trim()
  if (!/^\d+$/.test(trimmed)) {
    return null
  }
  const parsed = Number(trimmed)
  if (
    !Number.isInteger(parsed) ||
    parsed < MIN_REPLICAS ||
    parsed > MAX_REPLICAS
  ) {
    return null
  }
  return parsed
}

/** 步进按钮：以当前输入为基准加减 1，并收敛到 [0, 100]。 */
function stepReplicas(value: string, delta: number): string {
  const parsed = Number.parseInt(value, 10)
  const base = Number.isFinite(parsed) ? parsed : MIN_REPLICAS
  return String(Math.min(MAX_REPLICAS, Math.max(MIN_REPLICAS, base + delta)))
}

interface KubernetesWorkloadsTabProps {
  clusterUid: string
  namespace: string
  kind: string
  name: string
  canWrite: boolean
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

/** 工作负载列表：Deployment / StatefulSet / DaemonSet 的副本状态、镜像与副本调整。 */
export function KubernetesWorkloadsTab({
  clusterUid,
  namespace,
  kind,
  name,
  canWrite,
  refreshKey = 0,
}: KubernetesWorkloadsTabProps) {
  const { t } = useTranslation()
  const [scaleTarget, setScaleTarget] = useState<KubernetesWorkloadItem | null>(
    null
  )
  const [replicas, setReplicas] = useState("")
  const [scaleFieldError, setScaleFieldError] = useState("")
  const [scaleError, setScaleError] = useState("")
  const [scalePending, setScalePending] = useState(false)
  const [restartTarget, setRestartTarget] =
    useState<KubernetesWorkloadItem | null>(null)
  const [restartError, setRestartError] = useState("")
  const [restartPending, setRestartPending] = useState(false)
  const [imageTarget, setImageTarget] = useState<KubernetesWorkloadItem | null>(
    null
  )
  const [rollbackTarget, setRollbackTarget] =
    useState<KubernetesWorkloadItem | null>(null)
  const [detail, setDetail] = useState<KubernetesWorkloadItem | null>(null)
  const workloadRelation =
    useClusterResourceDetail<GetKubernetesWorkloadDetailResult>({
      enabled: detail !== null,
      clusterUid,
      namespace: detail?.namespace ?? "",
      name: detail?.name ?? "",
      fetch: (uid, ns, nm) =>
        getKubernetesWorkloadDetail(uid, {
          kind: detail?.kind ?? "",
          namespace: ns,
          name: nm,
        }),
    })
  const jumpTo = useClusterRelationJump()

  function openPod(pod: KubernetesPodItem) {
    setDetail(null)
    jumpTo({ tab: "pods", namespace: pod.namespace, name: pod.name })
  }

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesWorkloadItem>({
      page: 1,
      pageSize: 100,
      filterKey: `${clusterUid}|${namespace}|${kind}|${name}`,
      refreshKey,
      fetch: () =>
        listKubernetesWorkloads(clusterUid, {
          namespace: namespace || undefined,
          kind: kind || undefined,
          name: name || undefined,
        }),
    })

  function workloadActions(item: KubernetesWorkloadItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    const actions: RowAction[] = []
    // DaemonSet 没有副本数概念，服务端会拒绝 scale，入口直接不提供。
    if (item.kind !== "DaemonSet") {
      actions.push({
        key: "scale",
        label: t("kube.scale"),
        onSelect: () => openScale(item),
      })
    }
    actions.push({
      key: "restart",
      label: t("kube.restart"),
      onSelect: () => {
        setRestartError("")
        setRestartTarget(item)
      },
    })
    // 镜像与回滚都只改 Pod 模板，三种控制器都支持。
    actions.push({
      key: "image",
      label: t("kube.image"),
      onSelect: () => setImageTarget(item),
    })
    actions.push({
      key: "rollback",
      label: t("kube.rollback"),
      onSelect: () => setRollbackTarget(item),
    })
    return actions
  }

  function openScale(item: KubernetesWorkloadItem) {
    setScaleTarget(item)
    setReplicas(String(item.desiredReplicas))
    setScaleFieldError("")
    setScaleError("")
  }

  function adjustReplicas(delta: number) {
    setReplicas((value) => stepReplicas(value, delta))
    setScaleFieldError("")
  }

  async function onScale(event: React.FormEvent) {
    event.preventDefault()
    if (!scaleTarget) {
      return
    }
    const value = readReplicas(replicas)
    if (value === null) {
      setScaleFieldError(t("kube.replicasInvalid"))
      return
    }
    setScalePending(true)
    setScaleFieldError("")
    setScaleError("")
    try {
      await scaleKubernetesWorkload(clusterUid, {
        namespace: scaleTarget.namespace,
        kind: scaleTarget.kind,
        name: scaleTarget.name,
        replicas: value,
      })
      setScaleTarget(null)
      reload()
      toast.success(t("kube.actionDone"))
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setScaleFieldError(failure.fields.replicas ?? "")
      setScaleError(failure.form)
    } finally {
      setScalePending(false)
    }
  }

  async function runRestart() {
    if (!restartTarget) {
      return
    }
    setRestartPending(true)
    setRestartError("")
    try {
      await restartKubernetesWorkload(clusterUid, {
        namespace: restartTarget.namespace,
        kind: restartTarget.kind,
        name: restartTarget.name,
      })
      setRestartTarget(null)
      reload()
      toast.success(t("kube.actionDone"))
    } catch (err) {
      setRestartError(formErrorMessage(err, t("state.failed")))
    } finally {
      setRestartPending(false)
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
            title={t("kube.workloadsEmpty")}
            hint={t("kube.workloadsEmptyHint")}
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
                  {t("kube.kind")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.ready")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.uid}>
                  {t("kube.images")}
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
              {items.map((item) => (
                <TableRow key={`${item.namespace}/${item.kind}/${item.name}`}>
                  <TableCell className="font-medium break-words">
                    {item.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {item.namespace}
                  </TableCell>
                  <TableCell>
                    <StatusBadge
                      value={item.kind}
                      tone={workloadTone(item.kind)}
                    />
                  </TableCell>
                  <TableCell className="font-mono text-xs tabular-nums">
                    {item.readyReplicas}/{item.desiredReplicas}
                  </TableCell>
                  <TableCell className="whitespace-normal">
                    {item.images?.length ? (
                      <ul className="grid gap-0.5">
                        {item.images.map((image) => (
                          <li
                            key={image}
                            className="font-mono text-xs break-all text-muted-foreground"
                          >
                            {image}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(item.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(item)}
                      actions={workloadActions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <Dialog
        open={scaleTarget !== null}
        onOpenChange={(open) => {
          if (!open) {
            setScaleTarget(null)
            setScaleFieldError("")
            setScaleError("")
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("kube.scaleTitle")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onScale}>
            <DialogBody className="grid min-w-0 content-start gap-4">
              <p className="text-sm break-all text-muted-foreground">
                {scaleTarget
                  ? `${scaleTarget.kind} · ${scaleTarget.namespace}/${scaleTarget.name}`
                  : ""}
              </p>
              <FormField
                label={t("kube.replicas")}
                htmlFor="workload-replicas"
                hint={t("kube.replicasHint")}
                required
                error={scaleFieldError}
              >
                <div className="flex w-full items-center gap-2">
                  {/* 步进按钮在可滚动的 DialogBody 内；按钮按下的 1px 位移会撑出纵向滚动条，这里关掉。 */}
                  <Button
                    className="active:translate-y-0!"
                    type="button"
                    variant="outline"
                    size="icon"
                    aria-label={t("kube.replicasDecrease")}
                    disabled={
                      scalePending || readReplicas(replicas) === MIN_REPLICAS
                    }
                    onClick={() => adjustReplicas(-1)}
                  >
                    <Minus aria-hidden="true" />
                  </Button>
                  <Input
                    id="workload-replicas"
                    className="min-w-0 flex-1 text-center font-mono tabular-nums"
                    type="text"
                    inputMode="numeric"
                    autoComplete="off"
                    pattern="[0-9]*"
                    placeholder={t("kube.replicasPlaceholder")}
                    value={replicas}
                    aria-invalid={scaleFieldError ? true : undefined}
                    onChange={(event) => {
                      // 只接受数字，最多 3 位（副本数上限 100）。
                      setReplicas(
                        event.target.value.replace(/\D/g, "").slice(0, 3)
                      )
                      setScaleFieldError("")
                    }}
                  />
                  <Button
                    className="active:translate-y-0!"
                    type="button"
                    variant="outline"
                    size="icon"
                    aria-label={t("kube.replicasIncrease")}
                    disabled={
                      scalePending || readReplicas(replicas) === MAX_REPLICAS
                    }
                    onClick={() => adjustReplicas(1)}
                  >
                    <Plus aria-hidden="true" />
                  </Button>
                </div>
              </FormField>
            </DialogBody>
            {scaleError ? (
              <p className="text-sm break-words text-destructive">
                {scaleError}
              </p>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                disabled={scalePending}
                onClick={() => setScaleTarget(null)}
              >
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={scalePending}>
                {scalePending ? t("action.saving") : t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <ConfirmAction
        open={restartTarget !== null}
        title={t("kube.restartTitle")}
        description={
          restartTarget
            ? t("kube.restartDescription", {
                kind: restartTarget.kind,
                namespace: restartTarget.namespace,
                name: restartTarget.name,
              })
            : ""
        }
        confirmLabel={t("kube.restart")}
        cancelLabel={t("action.cancel")}
        pending={restartPending}
        error={restartError}
        onConfirm={() => void runRestart()}
        onOpenChange={(open) => {
          if (!open) {
            setRestartTarget(null)
            setRestartError("")
          }
        }}
      />
      <WorkloadImageDialog
        key={
          imageTarget
            ? `${imageTarget.kind}/${imageTarget.namespace}/${imageTarget.name}`
            : "no-image-target"
        }
        open={imageTarget !== null}
        clusterUid={clusterUid}
        workload={imageTarget}
        onOpenChange={(open) => {
          if (!open) {
            setImageTarget(null)
          }
        }}
        onSaved={() => {
          reload()
          workloadRelation.retry()
        }}
      />
      <WorkloadRollbackDialog
        key={
          rollbackTarget
            ? `${rollbackTarget.kind}/${rollbackTarget.namespace}/${rollbackTarget.name}`
            : "no-rollback-target"
        }
        open={rollbackTarget !== null}
        clusterUid={clusterUid}
        workload={rollbackTarget}
        onOpenChange={(open) => {
          if (!open) {
            setRollbackTarget(null)
          }
        }}
        onRolledBack={() => {
          reload()
          workloadRelation.retry()
        }}
      />
      <ClusterDetailSheet
        open={detail !== null}
        title={detail?.name ?? ""}
        size="wide"
        description={
          detail ? (
            <>
              <StatusBadge
                value={detail.kind}
                tone={workloadTone(detail.kind)}
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
                    { label: t("kube.namespace"), value: detail.namespace },
                    {
                      label: t("kube.kind"),
                      value: (
                        <StatusBadge
                          value={detail.kind}
                          tone={workloadTone(detail.kind)}
                        />
                      ),
                    },
                  ],
                },
                {
                  title: t("kube.sectionReplicas"),
                  description: (
                    <span
                      className={
                        detail.readyReplicas < detail.desiredReplicas
                          ? "text-warning"
                          : undefined
                      }
                    >
                      {detail.readyReplicas}/{detail.desiredReplicas}
                    </span>
                  ),
                  fields: [
                    {
                      label: t("kube.desiredReplicas"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.desiredReplicas}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.readyReplicas"),
                      value: (
                        <span
                          className={
                            detail.readyReplicas < detail.desiredReplicas
                              ? "font-mono text-xs text-warning tabular-nums"
                              : "font-mono text-xs tabular-nums"
                          }
                        >
                          {detail.readyReplicas}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.updatedReplicas"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.updatedReplicas}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.availableReplicas"),
                      value: (
                        <span className="font-mono text-xs tabular-nums">
                          {detail.availableReplicas}
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
                  title: t("kube.sectionReplicasets"),
                  description: workloadRelation.data?.replicasets?.length
                    ? t("kube.relationsCount", {
                        count: workloadRelation.data.replicasets.length,
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={workloadRelation.loading}
                      failed={workloadRelation.failed}
                      error={workloadRelation.error}
                      onRetry={workloadRelation.retry}
                      empty={
                        !workloadRelation.loading &&
                        !workloadRelation.failed &&
                        (workloadRelation.data?.replicasets?.length ?? 0) === 0
                      }
                    >
                      <ClusterRelationList
                        items={(workloadRelation.data?.replicasets ?? []).map(
                          (replicaSet) => ({
                            key: replicaSet.name,
                            title: (
                              <span className="font-mono text-xs break-all">
                                {replicaSet.name}
                              </span>
                            ),
                            status: (
                              <span className="font-mono text-xs tabular-nums">
                                {replicaSet.readyReplicas}/
                                {replicaSet.desiredReplicas}
                              </span>
                            ),
                            subtitle: replicaSet.images?.length
                              ? replicaSet.images.join(" · ")
                              : undefined,
                            meta: (
                              <>
                                <span>
                                  {t("kube.desiredReplicas")}:{" "}
                                  {replicaSet.desiredReplicas}
                                </span>
                                <span>
                                  {t("kube.readyReplicas")}:{" "}
                                  {replicaSet.readyReplicas}
                                </span>
                                <span>
                                  {t("kube.currentReplicas")}:{" "}
                                  {replicaSet.currentReplicas}
                                </span>
                                <span>
                                  {formatClusterAge(replicaSet.createdAt, t)}
                                </span>
                              </>
                            ),
                          })
                        )}
                      />
                    </ClusterRelationBody>
                  ),
                },
                {
                  title: t("kube.sectionPods"),
                  description: workloadRelation.data?.pods?.length
                    ? t("kube.relationsCount", {
                        count: workloadRelation.data.pods.length,
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={workloadRelation.loading}
                      failed={workloadRelation.failed}
                      error={workloadRelation.error}
                      onRetry={workloadRelation.retry}
                      empty={
                        !workloadRelation.loading &&
                        !workloadRelation.failed &&
                        (workloadRelation.data?.pods?.length ?? 0) === 0
                      }
                    >
                      <ClusterRelationList
                        items={(workloadRelation.data?.pods ?? []).map(
                          (pod) => ({
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
                                  tone={podPhaseTone(pod.phase)}
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
                          })
                        )}
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
                kind: detail.kind,
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
