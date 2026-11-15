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
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { formErrorMessage } from "@/lib/form-error"

interface DeployDialogProps {
  open: boolean
  applicationUid: string
  versionUid: string
  version: string
  onOpenChange: (open: boolean) => void
  onCreated: (deployment: ApplicationDeploymentItem) => void
}

/** M36 单集群部署对话框；M37 再扩展为多选。 */
export function DeployDialog({ open, applicationUid, versionUid, version, onOpenChange, onCreated }: DeployDialogProps) {
  const { t } = useTranslation()
  const [clusters, setClusters] = useState<ClusterItem[]>([])
  const [clusterUid, setClusterUid] = useState("")
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
        if (!clusterUid && result.items[0]) {
          setClusterUid(result.items[0].uid)
        }
        if (!name) {
          setName(`app-${version.replace(/\./g, "-")}`)
        }
      })
      .catch(() => undefined)
    return () => {
      active = false
    }
  }, [open, clusterUid, name, version])

  async function submit() {
    if (!clusterUid || !name.trim() || !namespace.trim()) {
      setError(t("form.required"))
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
        clusterUids: [clusterUid],
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
          <FormField label={t("appStore.targetCluster")} htmlFor="deploy-cluster" required>
            <Select value={clusterUid} onValueChange={setClusterUid}>
              <SelectTrigger id="deploy-cluster"><SelectValue placeholder={t("appStore.selectCluster")} /></SelectTrigger>
              <SelectContent>
                {clusters.map((cluster) => <SelectItem key={cluster.uid} value={cluster.uid}>{cluster.name}</SelectItem>)}
              </SelectContent>
            </Select>
          </FormField>
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
