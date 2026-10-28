import { CircleCheck, PlugZap } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { createCluster, probeKubeconfig, updateCluster } from "@/api/cluster"
import type { ClusterItem, ClusterProbe } from "@/api/cluster"
import { CodeEditor } from "@/components/code-editor"
import { FormField } from "@/components/form-field"
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
  collectFieldErrors,
  hasFieldErrors,
  maxLength,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"

interface ClusterFormState {
  name: string
  kubeconfig: string
  remark: string
}

const emptyForm: ClusterFormState = { name: "", kubeconfig: "", remark: "" }

// 后端用同样的两条子串校验 kubeconfig，这里提前拦一次，避免无谓的请求。
function looksLikeKubeconfig(value: string): boolean {
  return value.includes("apiVersion") && value.includes("clusters")
}

interface ClusterKubeconfigDialogProps {
  open: boolean
  /** null 表示接入新集群；否则编辑该集群。 */
  cluster: ClusterItem | null
  onOpenChange: (open: boolean) => void
  onSaved: (cluster: ClusterItem) => void
}

/**
 * 接入 / 编辑集群的表单。kubeconfig 只在提交与「测试连接」时使用：
 * 编辑时留空表示保持原值，界面必须把这条规则明确展示出来。
 */
export function ClusterKubeconfigDialog({
  open,
  cluster,
  onOpenChange,
  onSaved,
}: ClusterKubeconfigDialogProps) {
  const { t } = useTranslation()
  // 每次打开由父组件换 key 重新挂载，初始值直接来自当前编辑对象，
  // 不需要在 effect 里重置表单。
  const [form, setForm] = useState<ClusterFormState>(() =>
    cluster
      ? { name: cluster.name, kubeconfig: "", remark: cluster.remark }
      : emptyForm
  )
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState(false)
  const [probe, setProbe] = useState<ClusterProbe | null>(null)
  const [probeError, setProbeError] = useState("")
  const [probePending, setProbePending] = useState(false)

  function validate(): FieldErrors {
    const trimmedKubeconfig = form.kubeconfig.trim()
    return collectFieldErrors([
      ["name", requireTrimmed(form.name, t("form.required"))],
      ["name", maxLength(form.name, 64, t("form.nameTooLong"))],
      [
        "kubeconfig",
        cluster
          ? null
          : requireTrimmed(form.kubeconfig, t("cluster.kubeconfigRequired")),
      ],
      [
        "kubeconfig",
        trimmedKubeconfig && !looksLikeKubeconfig(trimmedKubeconfig)
          ? t("cluster.kubeconfigInvalid")
          : null,
      ],
      ["remark", maxLength(form.remark, 256, t("cluster.remarkTooLong"))],
    ])
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = validate()
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setFormError("")
    setFieldErrors({})
    try {
      const name = form.name.trim()
      const kubeconfig = form.kubeconfig.trim()
      const remark = form.remark.trim()
      const result = cluster
        ? await updateCluster({ uid: cluster.uid, name, kubeconfig, remark })
        : await createCluster({ name, kubeconfig, remark })
      onSaved(result.cluster)
      onOpenChange(false)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setFormError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onProbe() {
    const kubeconfig = form.kubeconfig.trim()
    if (!kubeconfig) {
      setFieldErrors((current) => ({
        ...current,
        kubeconfig: t("cluster.kubeconfigRequired"),
      }))
      return
    }
    if (!looksLikeKubeconfig(kubeconfig)) {
      setFieldErrors((current) => ({
        ...current,
        kubeconfig: t("cluster.kubeconfigInvalid"),
      }))
      return
    }
    setFieldErrors((current) => {
      const next = { ...current }
      delete next.kubeconfig
      return next
    })
    setProbe(null)
    setProbeError("")
    setProbePending(true)
    try {
      const result = await probeKubeconfig(kubeconfig)
      setProbe(result.probe)
    } catch (err) {
      setProbeError(formErrorMessage(err, t("cluster.testFailed")))
    } finally {
      setProbePending(false)
    }
  }

  const editing = cluster !== null
  const canProbe = form.kubeconfig.trim() !== "" && !probePending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {editing ? t("cluster.editTitle") : t("cluster.createTitle")}
          </DialogTitle>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSave}
        >
          <DialogBody className="grid content-start gap-4 sm:grid-cols-2">
            <FormField
              label={t("cluster.name")}
              htmlFor="cluster-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="cluster-name"
                value={form.name}
                placeholder={t("cluster.namePlaceholder")}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    name: event.target.value,
                  }))
                }
              />
            </FormField>
            <FormField
              label={t("cluster.remark")}
              htmlFor="cluster-remark"
              error={fieldErrors.remark}
            >
              <Input
                id="cluster-remark"
                value={form.remark}
                placeholder={t("cluster.remarkPlaceholder")}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    remark: event.target.value,
                  }))
                }
              />
            </FormField>
            <FormField
              label={t("cluster.kubeconfig")}
              htmlFor="cluster-kubeconfig"
              required={!editing}
              hint={
                editing
                  ? t("cluster.kubeconfigEditHint")
                  : t("cluster.kubeconfigHint")
              }
              error={fieldErrors.kubeconfig}
              className="sm:col-span-full"
            >
              <CodeEditor
                id="cluster-kubeconfig"
                value={form.kubeconfig}
                onChange={(value) => {
                  setForm((current) => ({ ...current, kubeconfig: value }))
                  setProbe(null)
                  setProbeError("")
                }}
                placeholder={
                  editing
                    ? t("cluster.kubeconfigEditPlaceholder")
                    : t("cluster.kubeconfigPlaceholder")
                }
                ariaLabel={t("cluster.kubeconfig")}
              />
              {editing ? (
                <p className="text-xs text-muted-foreground">
                  {t("cluster.kubeconfigEditHint")}
                </p>
              ) : null}
            </FormField>
            {probe ? (
              <div className="grid gap-2 rounded-lg border border-emerald-500/30 bg-emerald-500/5 p-3 sm:col-span-full">
                <p className="flex items-center gap-2 text-sm font-medium">
                  <CircleCheck
                    aria-hidden="true"
                    className="size-4 text-emerald-600 dark:text-emerald-400"
                  />
                  {t("cluster.probeSuccess")}
                </p>
                <dl className="grid gap-x-4 gap-y-1 text-sm sm:grid-cols-[9rem_minmax(0,1fr)]">
                  <dt className="text-muted-foreground">
                    {t("cluster.server")}
                  </dt>
                  <dd className="min-w-0 font-mono text-xs break-all">
                    {probe.server || "—"}
                  </dd>
                  <dt className="text-muted-foreground">
                    {t("cluster.version")}
                  </dt>
                  <dd className="min-w-0 font-mono text-xs break-all">
                    {probe.version || "—"}
                  </dd>
                  <dt className="text-muted-foreground">
                    {t("cluster.nodeCount")}
                  </dt>
                  <dd className="font-mono text-xs tabular-nums">
                    {probe.nodeCount}
                  </dd>
                </dl>
              </div>
            ) : null}
            {probeError ? (
              <p className="text-sm break-words text-destructive sm:col-span-full">
                {probeError}
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
              disabled={!canProbe}
              onClick={() => void onProbe()}
            >
              <PlugZap aria-hidden="true" data-icon="inline-start" />
              {probePending
                ? t("cluster.testing")
                : t("cluster.testConnection")}
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => onOpenChange(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button type="submit" disabled={pending || probePending}>
              {pending ? t("action.saving") : t("action.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
