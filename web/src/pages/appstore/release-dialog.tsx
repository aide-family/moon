import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  createApplicationDeployment,
  listApplicationDeployments,
  listApplicationVersions,
  upgradeApplicationDeployment,
} from "@/api/application"
import type {
  ApplicationDeploymentItem,
  ApplicationVersionItem,
} from "@/api/application/types"
import { listClusters } from "@/api/cluster"
import type { ClusterItem } from "@/api/cluster/types"
import { FormField } from "@/components/form-field"
import { StatusBadge } from "@/components/status-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
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
import { hasClusterScopedResource } from "@/lib/application-version"
import { checkField } from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"

interface ApplicationReleaseDialogProps {
  open: boolean
  applicationUid: string
  /** 从某个版本行进入时预选该版本；仍可换成其它已发布版本。 */
  initialVersionUid?: string
  onOpenChange: (open: boolean) => void
  onReleased: (deployment: ApplicationDeploymentItem | null) => void
}

interface ClusterTarget {
  cluster: ClusterItem
  deployment: ApplicationDeploymentItem | null
}

/**
 * 统一「发布」入口：只列已发布版本，目标集群多选。
 * 目标集群上已有本应用的部署组时走升级，否则创建新部署组。
 */
export function ApplicationReleaseDialog({
  open,
  applicationUid,
  initialVersionUid,
  onOpenChange,
  onReleased,
}: ApplicationReleaseDialogProps) {
  const { t } = useTranslation()
  const [versions, setVersions] = useState<ApplicationVersionItem[]>([])
  const [versionUid, setVersionUid] = useState(initialVersionUid ?? "")
  const [targets, setTargets] = useState<ClusterTarget[]>([])
  const [selected, setSelected] = useState<string[]>([])
  const [name, setName] = useState("")
  const [namespace, setNamespace] = useState("default")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (!open) {
      return
    }
    let active = true
    void Promise.all([
      listApplicationVersions(applicationUid, { page: 1, pageSize: 200, status: "PUBLISHED" }),
      listClusters({ page: 1, pageSize: 200 }),
      listApplicationDeployments({ applicationUid, page: 1, pageSize: 200 }),
    ])
      .then(([versionResult, clusterResult, deploymentResult]) => {
        if (!active) return
        // 契约要求只列已上架版本；即使服务端忽略了 status 过滤，这里也不展示草稿/已下架。
        const published = versionResult.items.filter((item) => item.status === "PUBLISHED")
        setVersions(published)
        setVersionUid((current) => current || published[0]?.uid || "")
        const byCluster = new Map<string, ApplicationDeploymentItem>()
        for (const deployment of deploymentResult.items) {
          for (const cluster of deployment.clusters ?? []) {
            byCluster.set(cluster.uid, deployment)
          }
        }
        setTargets(
          clusterResult.items.map((cluster) => ({
            cluster,
            deployment: byCluster.get(cluster.uid) ?? null,
          })),
        )
        setName((current) => current || "app-release")
      })
      .catch((err) => {
        if (active) setError(formErrorMessage(err, t("state.failed")))
      })
    return () => {
      active = false
    }
    // 只在打开或应用变化时重新加载。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, applicationUid, initialVersionUid])

  const version = useMemo(
    () => versions.find((item) => item.uid === versionUid) ?? null,
    [versions, versionUid],
  )
  const selectedTargets = targets.filter((target) => selected.includes(target.cluster.uid))
  const newClusters = selectedTargets.filter((target) => target.deployment === null)
  const upgrades = selectedTargets.filter((target) => target.deployment !== null)
  const needsDeployment = newClusters.length > 0

  function toggleCluster(uid: string) {
    setSelected((current) =>
      current.includes(uid) ? current.filter((item) => item !== uid) : [...current, uid],
    )
  }

  async function submit() {
    const next: Record<string, string> = {}
    if (!versionUid) {
      next.versionUid = t("appStore.selectVersion")
    }
    if (selected.length === 0) {
      next.clusterUids = t("error.clusterSelectionInvalid")
    }
    const dnsLabel = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/
    if (needsDeployment) {
      const nameError = checkField({
        label: t("appStore.deploymentName"),
        value: name,
        required: true,
        max: 63,
        pattern: dnsLabel,
        hint: t("appStore.deploymentNameHelp"),
      })
      if (nameError) next.name = nameError
      const namespaceError = checkField({
        label: t("appStore.targetNamespace"),
        value: namespace,
        required: true,
        max: 63,
        pattern: dnsLabel,
        hint: t("appStore.targetNamespaceHelp"),
      })
      if (namespaceError) next.namespace = namespaceError
    }
    setFieldErrors(next)
    if (Object.keys(next).length > 0 || !version) {
      return
    }
    setSaving(true)
    setError("")
    let created: ApplicationDeploymentItem | null = null
    try {
      if (newClusters.length > 0) {
        const result = await createApplicationDeployment({
          applicationUid,
          versionUid: version.uid,
          name: name.trim(),
          targetNamespace: namespace.trim(),
          clusterUids: newClusters.map((target) => target.cluster.uid),
        })
        created = result.deployment
      }
      const failures: string[] = []
      const groups = new Map<string, string[]>()
      for (const target of upgrades) {
        const deployment = target.deployment
        if (!deployment) continue
        const list = groups.get(deployment.uid) ?? []
        list.push(target.cluster.uid)
        groups.set(deployment.uid, list)
      }
      for (const [deploymentUid, clusterUids] of groups) {
        try {
          const result = await upgradeApplicationDeployment(deploymentUid, version.uid, clusterUids)
          created = created ?? result.deployment
        } catch (err) {
          failures.push(formErrorMessage(err, t("state.failed")))
        }
      }
      if (failures.length > 0) {
        setError(failures.join("；"))
        toast.error(failures.join("；"))
        onReleased(created)
        return
      }
      onReleased(created)
      onOpenChange(false)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors((current) => ({ ...current, ...failure.fields }))
      setError(failure.form)
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("appStore.releaseTitle")}</DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-4">
          <FormField
            label={t("appStore.version")}
            htmlFor="release-version"
            required
            error={fieldErrors.versionUid}
            description={t("appStore.releaseVersionHelp")}
          >
            <Select value={versionUid} onValueChange={setVersionUid}>
              <SelectTrigger id="release-version">
                <SelectValue placeholder={t("appStore.selectVersion")} />
              </SelectTrigger>
              <SelectContent>
                {versions.map((item) => (
                  <SelectItem key={item.uid} value={item.uid}>
                    {item.version}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>
          {versions.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("appStore.noPublishedVersions")}</p>
          ) : null}

          <FormField
            label={t("appStore.targetClusters")}
            required
            error={fieldErrors.clusterUids}
            hint={t("appStore.clusterSelectionHint", { count: selected.length })}
            description={t("appStore.releaseClusterHint")}
          >
            <div className="grid max-h-64 gap-1 overflow-auto rounded-lg border border-border/60 p-2">
              {targets.map((target) => {
                const checked = selected.includes(target.cluster.uid)
                return (
                  <label key={target.cluster.uid} className="flex cursor-pointer items-center gap-3 rounded-md px-2 py-1.5 hover:bg-muted">
                    <Checkbox checked={checked} onCheckedChange={() => toggleCluster(target.cluster.uid)} />
                    <span className="grid min-w-0 flex-1 gap-0.5">
                      <span className="text-sm font-medium">{target.cluster.name}</span>
                      <span className="truncate font-mono text-xs text-muted-foreground">{target.cluster.server}</span>
                    </span>
                    {target.deployment ? (
                      <Badge variant="secondary" className="shrink-0">
                        {t("appStore.releaseModeUpgrade", { name: target.deployment.name })}
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="shrink-0">{t("appStore.releaseModeCreate")}</Badge>
                    )}
                  </label>
                )
              })}
              {targets.length === 0 ? <p className="p-2 text-sm text-muted-foreground">{t("appStore.noClusters")}</p> : null}
            </div>
          </FormField>

          <FormField
            label={t("appStore.deploymentName")}
            htmlFor="release-name"
            required={needsDeployment}
            error={fieldErrors.name}
            description={needsDeployment ? t("appStore.deploymentNameHelp") : t("appStore.releaseNoCreateHint")}
          >
            <Input id="release-name" className="font-mono" value={name} placeholder="app-prod" onChange={(event) => setName(event.target.value)} />
          </FormField>
          <FormField
            label={t("appStore.targetNamespace")}
            htmlFor="release-namespace"
            required={needsDeployment}
            error={fieldErrors.namespace}
            description={t("appStore.targetNamespaceHelp")}
          >
            <Input id="release-namespace" className="font-mono" value={namespace} placeholder="default" onChange={(event) => setNamespace(event.target.value)} />
          </FormField>

          <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
            <span>{t("appStore.releaseSummary", { upgrade: upgrades.length, create: newClusters.length })}</span>
            {version ? <StatusBadge value={version.status} /> : null}
            {version && hasClusterScopedResource(version) ? (
              <span className="rounded-md border border-warning-border bg-warning-soft px-2 py-0.5 text-warning">
                {t("appStore.clusterScopedWarning")}
              </span>
            ) : null}
          </div>
        </DialogBody>
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button disabled={saving || versions.length === 0} onClick={() => void submit()}>
            {saving ? t("action.saving") : t("appStore.releaseAction")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
