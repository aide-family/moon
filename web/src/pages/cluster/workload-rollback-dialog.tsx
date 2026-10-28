import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  listKubernetesWorkloadRevisions,
  rollbackKubernetesWorkload,
} from "@/api/cluster"
import type {
  KubernetesWorkloadItem,
  KubernetesWorkloadRevisionItem,
} from "@/api/cluster"
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
import { cn } from "@/lib/class-names"
import { failureMessage, readFormError } from "@/lib/form-error"

import { formatClusterTime } from "./format"
import { ClusterResourceState } from "./resource-state"

interface WorkloadRollbackDialogProps {
  open: boolean
  clusterUid: string
  /** 目标工作负载；关闭时为空。 */
  workload: KubernetesWorkloadItem | null
  onOpenChange: (open: boolean) => void
  /** 回滚成功后刷新列表与详情。 */
  onRolledBack: () => void
}

/**
 * 工作负载回滚对话框：列出历史版本（版本号、镜像、副本数、创建时间），
 * 当前版本不能回滚；选中一个历史版本后确认，调用回滚接口。
 * 遵循全局对话框规范：头部与页脚固定，DialogBody 是唯一滚动区，
 * 版本用可点击的 radio 列表而不是表格，避免在滚动区内再嵌一层滚动条。
 */
export function WorkloadRollbackDialog({
  open,
  clusterUid,
  workload,
  onOpenChange,
  onRolledBack,
}: WorkloadRollbackDialogProps) {
  const { t } = useTranslation()
  const [revisions, setRevisions] = useState<
    KubernetesWorkloadRevisionItem[] | null
  >(null)
  const [selected, setSelected] = useState<number | null>(null)
  const [loading, setLoading] = useState(false)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState(false)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    if (!open || !workload) {
      return
    }
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      setLoadError("")
      setFormError("")
      setSelected(null)
      try {
        const result = await listKubernetesWorkloadRevisions(clusterUid, {
          kind: workload!.kind,
          namespace: workload!.namespace,
          name: workload!.name,
        })
        if (!active) {
          return
        }
        setRevisions(result.revisions ?? [])
      } catch (err) {
        if (!active) {
          return
        }
        setRevisions(null)
        setFailed(true)
        setLoadError(failureMessage(err, t("kube.revisionsLoadFailed")))
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [open, clusterUid, workload, attempt, t])

  const canSubmit =
    !pending && !loading && !failed && selected !== null && workload !== null

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (!workload || selected === null || pending) {
      return
    }
    setPending(true)
    setFormError("")
    try {
      await rollbackKubernetesWorkload(clusterUid, {
        namespace: workload.namespace,
        kind: workload.kind,
        name: workload.name,
        revision: selected,
      })
      toast.success(t("kube.actionDone"))
      onRolledBack()
      onOpenChange(false)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFormError(
        failure.form || Object.values(failure.fields)[0] || t("state.failed")
      )
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next && pending) {
          return
        }
        onOpenChange(next)
      }}
    >
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("kube.rollbackTitle")}</DialogTitle>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSubmit}
        >
          <DialogBody className="grid min-w-0 content-start gap-3">
            {workload ? (
              <p className="text-sm break-all text-muted-foreground">
                {workload.kind} · {workload.namespace}/{workload.name}
              </p>
            ) : null}
            {loading ? (
              <ClusterResourceState
                state="loading"
                title={t("state.loading")}
              />
            ) : null}
            {!loading && failed ? (
              <ClusterResourceState
                state="error"
                title={loadError || t("kube.revisionsLoadFailed")}
                hint={t("kube.resourceErrorHint")}
                onRetry={() => setAttempt((value) => value + 1)}
              />
            ) : null}
            {!loading &&
            !failed &&
            revisions !== null &&
            revisions.length === 0 ? (
              <ClusterResourceState
                state="empty"
                title={t("kube.revisionsEmpty")}
                hint={t("kube.revisionsEmptyHint")}
              />
            ) : null}
            {!loading &&
            !failed &&
            revisions !== null &&
            revisions.length > 0 ? (
              <>
                <div className="grid gap-1">
                  <h3 className="font-heading text-xs font-semibold">
                    {t("kube.revisions")}
                  </h3>
                  <p className="text-xs text-muted-foreground">
                    {t("kube.revisionSelect")}
                  </p>
                </div>
                <div
                  role="radiogroup"
                  aria-label={t("kube.revisions")}
                  className="grid min-w-0 gap-2"
                >
                  {revisions.map((revision) => {
                    const value = revision.revision
                    const checked = !revision.current && selected === value
                    return (
                      <label
                        key={`${value}-${revision.name}`}
                        data-state={checked ? "selected" : undefined}
                        className={cn(
                          "flex min-w-0 items-start gap-3 rounded-lg border border-border/60 p-3 transition-colors",
                          revision.current
                            ? "opacity-60"
                            : "cursor-pointer hover:bg-muted/40",
                          checked && "border-primary/60 ring-1 ring-primary/30"
                        )}
                      >
                        <input
                          type="radio"
                          name="workload-revision"
                          className="mt-0.5 size-4 shrink-0 accent-primary"
                          checked={checked}
                          disabled={revision.current || pending}
                          onChange={() => setSelected(value)}
                          aria-label={`${t("kube.revision")} ${value}`}
                        />
                        <span className="grid min-w-0 flex-1 gap-1.5">
                          <span className="flex flex-wrap items-center gap-2">
                            <span className="font-mono text-xs font-medium tabular-nums">
                              {value}
                            </span>
                            {revision.current ? (
                              <StatusBadge
                                value="CURRENT"
                                label={t("kube.revisionCurrent")}
                                tone="primary"
                              />
                            ) : null}
                          </span>
                          {revision.images?.length ? (
                            <span className="grid min-w-0 gap-0.5">
                              {revision.images.map((image) => (
                                <span
                                  key={image}
                                  className="font-mono text-xs break-all text-muted-foreground"
                                >
                                  {image}
                                </span>
                              ))}
                            </span>
                          ) : (
                            <span className="text-xs text-muted-foreground">
                              —
                            </span>
                          )}
                          <span className="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground tabular-nums">
                            <span>
                              {t("kube.revisionReplicas")}:{" "}
                              <span className="font-mono">
                                {revision.replicas}
                              </span>
                            </span>
                            <span>
                              {formatClusterTime(revision.createdAt) || "—"}
                            </span>
                          </span>
                        </span>
                      </label>
                    )
                  })}
                </div>
                {selected !== null && workload ? (
                  <p className="text-xs break-words text-muted-foreground">
                    {t("kube.rollbackDescription", {
                      kind: workload.kind,
                      namespace: workload.namespace,
                      name: workload.name,
                      revision: selected,
                    })}
                  </p>
                ) : null}
              </>
            ) : null}
          </DialogBody>
          {formError ? (
            <p className="shrink-0 text-sm break-words text-destructive">
              {formError}
            </p>
          ) : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => onOpenChange(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button type="submit" disabled={!canSubmit}>
              {pending ? t("kube.rollbackPending") : t("kube.rollbackConfirm")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
