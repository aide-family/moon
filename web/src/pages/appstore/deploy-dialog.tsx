import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { createApplicationDeployment } from "@/api/application"
import type { ApplicationDeploymentItem } from "@/api/application/types"
import { listClusters } from "@/api/cluster"
import type { ClusterItem } from "@/api/cluster/types"
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
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { checkField } from "@/lib/form"
import { readFormError } from "@/lib/form-error"

interface DeployDialogProps {
  open: boolean
  applicationUid: string
  versionUid: string
  version: string
  hasClusterScoped?: boolean
  onOpenChange: (open: boolean) => void
  onCreated: (deployment: ApplicationDeploymentItem) => void
}

/** 多集群部署对话框：一次创建部署组 + 每集群实例与操作。 */
export function DeployDialog({ open, applicationUid, versionUid, version, hasClusterScoped, onOpenChange, onCreated }: DeployDialogProps) {
  const { t } = useTranslation()
  const [clusters, setClusters] = useState<ClusterItem[]>([])
  const [selected, setSelected] = useState<string[]>([])
  const [name, setName] = useState("")
  const [namespace, setNamespace] = useState("default")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (!open) return
    let active = true
    void listClusters({ page: 1, pageSize: 100 })
      .then((result) => {
        if (!active) return
        setClusters(result.items)
        if (result.items[0]) {
          setSelected((current) => current.length > 0 ? current : [result.items[0].uid])
        }
        setName((current) => current || `app-${version.replace(/\./g, "-")}`)
      })
      .catch(() => undefined)
    return () => {
      active = false
    }
    // 只在对话框打开/版本变化时加载集群并填默认值。
  }, [open, version])

  async function submit() {
    const next: Record<string, string> = {}
    if (selected.length === 0 || selected.length > 20) {
      next.clusterUids = t("appStore.clusterSelectionInvalid")
    }
    // Kubernetes 的 Deployment / Namespace 名都是 DNS-1123 label。
    const dnsLabel = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
    const nameError = checkField({
      label: t("appStore.deploymentName"),
      value: name,
      required: true,
      max: 63,
      pattern: dnsLabel,
      hint: t("appStore.deploymentNameHelp"),
    })
    if (nameError) {
      next.name = nameError
    }
    const namespaceError = checkField({
      label: t("appStore.targetNamespace"),
      value: namespace,
      required: true,
      max: 63,
      pattern: dnsLabel,
      hint: t("appStore.targetNamespaceHelp"),
    })
    if (namespaceError) {
      next.namespace = namespaceError
    }
    setFieldErrors(next)
    if (Object.keys(next).length > 0) {
      setError("")
      return
    }
    setSaving(true)
    setError("")
    try {
      const result = await createApplicationDeployment({
        applicationUid,
        versionUid,
        name: name.trim(),
        targetNamespace: namespace.trim(),
        clusterUids: selected,
      })
      onCreated(result.deployment)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("appStore.deployTitle", { version })}</DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-4">
          <FormField label={t("appStore.targetClusters")} required hint={t("appStore.clusterSelectionHint", { count: selected.length })} description={t("appStore.multiClusterHint")} error={fieldErrors.clusterUids}>
            <div className="grid max-h-56 gap-2 overflow-auto rounded-lg border border-border/60 p-2">
              {clusters.map((cluster) => {
                const checked = selected.includes(cluster.uid)
                return (
                  <label key={cluster.uid} className="flex cursor-pointer items-center gap-3 rounded-md px-2 py-1.5 hover:bg-muted">
                    <Checkbox checked={checked} onCheckedChange={(value) => setSelected((current) => value === true ? [...current, cluster.uid] : current.filter((uid) => uid !== cluster.uid))} />
                    <span className="grid min-w-0 flex-1 gap-0.5">
                      <span className="text-sm font-medium">{cluster.name}</span>
                      <span className="truncate font-mono text-xs text-muted-foreground">{cluster.server} · {cluster.version || "—"}</span>
                    </span>
                  </label>
                )
              })}
              {clusters.length === 0 ? <p className="p-2 text-sm text-muted-foreground">{t("appStore.noClusters")}</p> : null}
            </div>
          </FormField>
          {hasClusterScoped ? <p className="rounded-lg border border-warning-border bg-warning-soft p-2 text-xs text-warning">{t("appStore.clusterScopedWarning")}</p> : null}
          <FormField label={t("appStore.deploymentName")} htmlFor="deploy-name" required description={t("appStore.deploymentNameHelp")} error={fieldErrors.name}>
            <Input id="deploy-name" className="font-mono" value={name} placeholder="app-prod" onChange={(event) => setName(event.target.value)} />
          </FormField>
          <FormField label={t("appStore.targetNamespace")} htmlFor="deploy-namespace" required description={t("appStore.targetNamespaceHelp")} error={fieldErrors.namespace}>
            <Input id="deploy-namespace" className="font-mono" value={namespace} placeholder="default" onChange={(event) => setNamespace(event.target.value)} />
          </FormField>
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
        </DialogBody>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button disabled={saving} onClick={() => void submit()}>{saving ? t("action.saving") : t("appStore.deployAction")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
