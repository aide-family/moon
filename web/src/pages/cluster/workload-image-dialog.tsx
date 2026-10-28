import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getKubernetesWorkloadImages,
  setKubernetesWorkloadImage,
} from "@/api/cluster"
import type {
  KubernetesWorkloadImageItem,
  KubernetesWorkloadItem,
} from "@/api/cluster"
import { FormField } from "@/components/form-field"
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
import { Input } from "@/components/ui/input"
import { failureMessage, readFormError } from "@/lib/form-error"

import { ClusterResourceState } from "./resource-state"

interface WorkloadImageDialogProps {
  open: boolean
  clusterUid: string
  /** 目标工作负载；关闭时为空。 */
  workload: KubernetesWorkloadItem | null
  onOpenChange: (open: boolean) => void
  /** 全部容器设置成功后刷新列表与详情。 */
  onSaved: () => void
}

// 普通容器与 init 容器可能重名，草稿和错误都按「类型 + 名称」分组。
function containerKey(image: KubernetesWorkloadImageItem): string {
  return `${image.init ? "init" : "main"}/${image.container}`
}

/**
 * 工作负载镜像对话框：按容器展示当前镜像与历史版本，只提交改动过的容器。
 * 多个容器的设置请求在同一个对话框内串行执行，任一失败就停下并保留错误。
 */
export function WorkloadImageDialog({
  open,
  clusterUid,
  workload,
  onOpenChange,
  onSaved,
}: WorkloadImageDialogProps) {
  const { t } = useTranslation()
  const [images, setImages] = useState<KubernetesWorkloadImageItem[] | null>(
    null
  )
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
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
      try {
        const result = await getKubernetesWorkloadImages(clusterUid, {
          kind: workload!.kind,
          namespace: workload!.namespace,
          name: workload!.name,
        })
        if (!active) {
          return
        }
        const list = result.images ?? []
        const nextDrafts: Record<string, string> = {}
        for (const image of list) {
          nextDrafts[containerKey(image)] = image.image
        }
        setImages(list)
        setDrafts(nextDrafts)
        setFieldErrors({})
      } catch (err) {
        if (!active) {
          return
        }
        setImages(null)
        setFailed(true)
        setLoadError(failureMessage(err, t("kube.imageLoadFailed")))
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

  const changed =
    images?.filter(
      (image) => (drafts[containerKey(image)] ?? "").trim() !== image.image
    ) ?? []
  const canSubmit =
    !pending && !loading && !failed && images !== null && changed.length > 0

  function setDraft(key: string, value: string) {
    setDrafts((current) => ({ ...current, [key]: value }))
  }

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (!workload || !images || pending) {
      return
    }
    const nextErrors: Record<string, string> = {}
    const targets: KubernetesWorkloadImageItem[] = []
    for (const image of images) {
      const key = containerKey(image)
      const value = (drafts[key] ?? "").trim()
      if (!value) {
        nextErrors[key] = t("kube.imageRequired")
        continue
      }
      if (value !== image.image) {
        targets.push(image)
      }
    }
    setFieldErrors(nextErrors)
    if (Object.keys(nextErrors).length > 0) {
      return
    }
    if (targets.length === 0) {
      return
    }
    setPending(true)
    setFormError("")
    let current = ""
    try {
      // 串行提交：顺序与容器顺序一致，前一个失败就不再动后面的容器。
      for (const image of targets) {
        current = image.container
        await setKubernetesWorkloadImage(clusterUid, {
          namespace: workload.namespace,
          kind: workload.kind,
          name: workload.name,
          container: image.container,
          init: image.init,
          image: (drafts[containerKey(image)] ?? "").trim(),
        })
      }
      toast.success(t("kube.actionDone"))
      onSaved()
      onOpenChange(false)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      const reason =
        failure.form || Object.values(failure.fields)[0] || t("state.failed")
      setFormError(
        t("kube.imageSetFailed", { container: current, error: reason })
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
          <DialogTitle>{t("kube.imageTitle")}</DialogTitle>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSubmit}
        >
          <DialogBody className="grid min-w-0 content-start gap-3">
            <div className="grid gap-1">
              {workload ? (
                <p className="text-sm break-all text-muted-foreground">
                  {workload.kind} · {workload.namespace}/{workload.name}
                </p>
              ) : null}
              <p className="text-xs break-words text-muted-foreground">
                {t("kube.imageContainerHint")}
              </p>
            </div>
            {loading ? (
              <ClusterResourceState
                state="loading"
                title={t("state.loading")}
              />
            ) : null}
            {!loading && failed ? (
              <ClusterResourceState
                state="error"
                title={loadError || t("kube.imageLoadFailed")}
                hint={t("kube.resourceErrorHint")}
                onRetry={() => setAttempt((value) => value + 1)}
              />
            ) : null}
            {!loading && !failed && images !== null && images.length === 0 ? (
              <ClusterResourceState
                state="empty"
                title={t("kube.imageEmpty")}
                hint={t("kube.imageEmptyHint")}
              />
            ) : null}
            {!loading && !failed && images !== null && images.length > 0 ? (
              <ul className="grid gap-3">
                {images.map((image) => {
                  const key = containerKey(image)
                  return (
                    <li
                      key={key}
                      className="grid gap-2 rounded-lg border border-border/60 p-3"
                    >
                      <div className="flex min-w-0 flex-wrap items-center gap-2">
                        <span className="min-w-0 font-mono text-xs font-medium break-all">
                          {image.container}
                        </span>
                        {image.init ? (
                          <StatusBadge
                            value="init"
                            label={t("kube.imageInit")}
                            tone="violet"
                          />
                        ) : null}
                      </div>
                      <div className="grid gap-0.5">
                        <p className="text-[11px] tracking-wide text-muted-foreground">
                          {t("kube.imageCurrent")}
                        </p>
                        <p className="font-mono text-xs break-all">
                          {image.image || "—"}
                        </p>
                      </div>
                      <FormField
                        label={t("kube.image")}
                        htmlFor={`workload-image-${key}`}
                        required
                        error={fieldErrors[key]}
                      >
                        <Input
                          id={`workload-image-${key}`}
                          value={drafts[key] ?? ""}
                          onChange={(event) =>
                            setDraft(key, event.target.value)
                          }
                          placeholder={t("kube.imagePlaceholder")}
                          autoComplete="off"
                          spellCheck={false}
                          disabled={pending}
                          className="font-mono text-xs"
                        />
                      </FormField>
                      {image.versions.length > 0 ? (
                        <div className="grid gap-1.5">
                          <p className="text-[11px] font-medium tracking-wide text-muted-foreground">
                            {t("kube.imageVersions")}
                          </p>
                          <ul className="flex min-w-0 flex-wrap gap-1.5">
                            {image.versions.map((version) => (
                              <li
                                key={`${version.revision}-${version.image}`}
                                className="max-w-full min-w-0"
                              >
                                <Button
                                  type="button"
                                  variant="outline"
                                  className="h-8 max-w-full justify-start font-mono text-xs"
                                  title={version.image}
                                  disabled={pending}
                                  onClick={() => setDraft(key, version.image)}
                                >
                                  <span className="min-w-0 truncate">
                                    {version.image}
                                  </span>
                                </Button>
                              </li>
                            ))}
                          </ul>
                        </div>
                      ) : (
                        <p className="text-[11px] text-muted-foreground">
                          {t("kube.imageNoVersions")}
                        </p>
                      )}
                    </li>
                  )
                })}
              </ul>
            ) : null}
            {!loading &&
            !failed &&
            images !== null &&
            images.length > 0 &&
            changed.length === 0 ? (
              <p className="text-xs text-muted-foreground">
                {t("kube.imageUnchanged")}
              </p>
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
              {pending ? t("action.saving") : t("kube.setImage")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
