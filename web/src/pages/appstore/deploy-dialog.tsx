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
import { formErrorMessage } from "@/lib/form-error"

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

  useEffect(() => {
    if (!open) return
    let active = true
    void listClusters({ page: 1, pageSize: 100 })
      .then((result) => {
        if (!active) return
        setClusters(result.items)
        if (selected.length === 0 && result.items[0]) {
          setSelected([result.items[0].uid])
        }
        if (!name) {
          setName(`app-${version.replace(/\./g, "-")}`)
        }
      })
      .catch(() => undefined)
    return () => {
      active = false
    }
  }, [open, selected.length, name, version])

  async function submit() {
    if (selected.length === 0 || selected.length > 20 || !name.trim() || !namespace.trim()) {
      setError(t("appStore.clusterSelectionInvalid"))
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
      setError(formErrorMessage(err, t("state.failed")))
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
          <FormField label={t("appStore.targetClusters")} required hint={t("appStore.clusterSelectionHint", { count: selected.length })}>
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
          <p className="text-xs text-muted-foreground">{t("appStore.multiClusterHint")}</p>
          {hasClusterScoped ? <p className="rounded-lg border border-warning-border bg-warning-soft p-2 text-xs text-warning">{t("appStore.clusterScopedWarning")}</p> : null}
          <FormField label={t("appStore.deploymentName")} htmlFor="deploy-name" required>
            <Input id="deploy-name" className="font-mono" value={name} onChange={(event) => setName(event.target.value)} />
          </FormField>
          <FormField label={t("appStore.targetNamespace")} htmlFor="deploy-namespace" required>
            <Input id="deploy-namespace" className="font-mono" value={namespace} onChange={(event) => setNamespace(event.target.value)} />
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
