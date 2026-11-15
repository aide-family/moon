import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import {
  getApplicationDeployment,
  listApplicationDeploymentOperations,
  listApplicationVersions,
  redeployApplicationDeployment,
  rollbackApplicationDeployment,
  uninstallApplicationDeployment,
  upgradeApplicationDeployment,
} from "@/api/application"
import type {
  ApplicationDeploymentItem,
  ApplicationOperationItem,
  ApplicationVersionItem,
} from "@/api/application/types"
import { useSession } from "@/auth/session"
import { HeaderActionsMenu } from "@/components/header-actions-menu"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { TableActionsCell, TableActionsHead } from "@/components/table-heads"
import { Badge } from "@/components/ui/badge"
import { Checkbox } from "@/components/ui/checkbox"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { FormField } from "@/components/form-field"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatUnixDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

export function ApplicationDeploymentPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const { uid = "" } = useParams()
  const [deployment, setDeployment] = useState<ApplicationDeploymentItem | null>(null)
  const [operations, setOperations] = useState<ApplicationOperationItem[]>([])
  const [versions, setVersions] = useState<ApplicationVersionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [actionOpen, setActionOpen] = useState<"upgrade" | "rollback" | null>(null)
  const [selectedVersion, setSelectedVersion] = useState("")
  const [selectedClusters, setSelectedClusters] = useState<string[]>([])
  const [actionPending, setActionPending] = useState(false)
  const [uninstallOpen, setUninstallOpen] = useState(false)
  const [keepPVC, setKeepPVC] = useState(true)
  const [deleteClusterScoped, setDeleteClusterScoped] = useState(false)

  async function reload() {
    try {
      const [deploymentResult, operationsResult] = await Promise.all([
        getApplicationDeployment(uid),
        listApplicationDeploymentOperations(uid, { page: 1, pageSize: 100 }),
      ])
      setDeployment(deploymentResult.deployment)
      setOperations(operationsResult.items)
      const versionsResult = await listApplicationVersions(deploymentResult.deployment.applicationUid, { page: 1, pageSize: 100 })
      setVersions(versionsResult.items)
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let active = true
    async function load() {
      try {
        const [deploymentResult, operationsResult] = await Promise.all([
          getApplicationDeployment(uid),
          listApplicationDeploymentOperations(uid, { page: 1, pageSize: 100 }),
        ])
        const versionsResult = await listApplicationVersions(deploymentResult.deployment.applicationUid, { page: 1, pageSize: 100 })
        if (!active) return
        setDeployment(deploymentResult.deployment)
        setOperations(operationsResult.items)
        setVersions(versionsResult.items)
      } catch (err) {
        if (active) setError(formErrorMessage(err, t("state.failed")))
      } finally {
        if (active) setLoading(false)
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [uid, t])

  if (!session.loading && !session.can("release:read")) {
    return <Navigate to="/forbidden" replace />
  }
  if (!deployment) {
    return <section className="flex min-h-0 flex-1 flex-col gap-6"><PageHeader title={t("appStore.deploymentName")} description={loading ? t("state.loading") : error} /></section>
  }

  async function runAction() {
    setActionPending(true)
    try {
      if (actionOpen === "upgrade") {
        await upgradeApplicationDeployment(uid, selectedVersion, selectedClusters)
      } else {
        await rollbackApplicationDeployment(uid, selectedVersion, selectedClusters)
      }
      setActionOpen(null)
      toast.success(t("appStore.operationSubmitted"))
      await reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  async function redeploy(clusterUIDs = selectedClusters) {
    try {
      await redeployApplicationDeployment(uid, clusterUIDs)
      toast.success(t("appStore.operationSubmitted"))
      await reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  async function uninstall() {
    setActionPending(true)
    try {
      await uninstallApplicationDeployment(uid, selectedClusters, keepPVC, deleteClusterScoped)
      setUninstallOpen(false)
      toast.success(t("appStore.operationSubmitted"))
      await reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  const candidates = versions.filter((item) => item.status === "PUBLISHED" || item.status === "DEPRECATED")

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={deployment.name}
        description={`${deployment.applicationName} ${deployment.version} · ${deployment.targetNamespace} · ${deployment.status}${selectedClusters.length > 0 ? ` · ${t("appStore.selectedClusters", { count: selectedClusters.length })}` : ""}`}
      >
        <Button variant="outline" onClick={() => navigate("/appstore?tab=deployments")}>{t("action.goBack")}</Button>
        {session.can("release:write") ? <Button variant="outline" onClick={() => void redeploy()}>{t("appStore.redeploy")}</Button> : null}
        {session.can("release:write") ? <Button onClick={() => { setSelectedVersion(""); setActionOpen("upgrade") }}>{t("appStore.upgrade")}</Button> : null}
        {session.can("release:write") || session.can("release:delete") ? (
          <HeaderActionsMenu
            label={t("action.more")}
            actions={[
              ...(session.can("release:write") ? [{ key: "rollback", label: t("appStore.rollback"), onSelect: () => { setSelectedVersion(""); setActionOpen("rollback") } }] : []),
              ...(session.can("release:delete") ? [{ key: "uninstall", label: t("appStore.uninstall"), destructive: true, onSelect: () => setUninstallOpen(true) }] : []),
            ]}
          />
        ) : null}
      </PageHeader>
      <div className="min-h-0 flex-1 overflow-y-auto">
      <div className="grid gap-6 pb-2">
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <div className="rounded-xl border border-border/60 bg-card">
        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-3">
          <h2 className="font-heading text-base font-semibold">{t("appStore.targetClusters")}</h2>
          <span className="text-xs text-muted-foreground">{selectedClusters.length === 0 ? t("appStore.allClustersHint") : t("appStore.selectedClusters", { count: selectedClusters.length })}</span>
        </div>
        <Table className={TABLE_MIN_WIDTH.wide}>
          <TableHeader><TableRow><TableHead className="w-10"><Checkbox checked={selectedClusters.length === deployment.clusters.length && deployment.clusters.length > 0} onCheckedChange={(value) => setSelectedClusters(value === true ? deployment.clusters.map((cluster) => cluster.uid) : [])} /></TableHead><TableHead>{t("appStore.cluster")}</TableHead><TableHead>{t("appStore.status")}</TableHead><TableHead>{t("appStore.version")}</TableHead><TableHead>{t("appStore.targetNamespace")}</TableHead><TableHead>{t("appStore.message")}</TableHead><TableActionsHead>{t("action.actions")}</TableActionsHead></TableRow></TableHeader>
          <TableBody>
            {deployment.clusters.map((cluster) => (
              <TableRow key={cluster.uid}>
                <TableCell><Checkbox checked={selectedClusters.includes(cluster.uid)} onCheckedChange={(value) => setSelectedClusters((current) => value === true ? [...current, cluster.uid] : current.filter((uid) => uid !== cluster.uid))} /></TableCell>
                <TableCell>{cluster.name}</TableCell>
                <TableCell><Badge variant={cluster.status === "RUNNING" ? "secondary" : cluster.status === "FAILED" ? "destructive" : "outline"}>{cluster.status}</Badge></TableCell>
                <TableCell className="font-mono text-xs">{cluster.version || "—"}</TableCell>
                <TableCell className="font-mono text-xs">{deployment.targetNamespace}</TableCell>
                <TableCell className="text-xs text-muted-foreground">{cluster.message || "—"}</TableCell>
                <TableActionsCell><RowActions nowrap detailLabel={t("appStore.retryCluster")} moreLabel={t("action.more")} onDetail={() => void redeploy([cluster.uid])} detailDisabled={actionPending || cluster.status === "UNINSTALLED"} actions={[]} /></TableActionsCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <div className="rounded-xl border border-border/60 bg-card">
        <div className="border-b border-border px-4 py-3"><h2 className="font-heading text-base font-semibold">{t("appStore.operations")}</h2></div>
        <Table className={TABLE_MIN_WIDTH.default}>
          <TableHeader><TableRow><TableHead>{t("appStore.operationAction")}</TableHead><TableHead>{t("appStore.status")}</TableHead><TableHead>{t("appStore.cluster")}</TableHead><TableHead>{t("appStore.message")}</TableHead><TableHead className={TABLE_NOWRAP}>{t("action.createdAt")}</TableHead></TableRow></TableHeader>
          <TableBody>
            {operations.map((operation) => (
              <TableRow key={operation.uid}>
                <TableCell>{operation.action}</TableCell>
                <TableCell><Badge variant={operation.status === "SUCCEEDED" ? "secondary" : operation.status === "FAILED" ? "destructive" : "outline"}>{operation.status}</Badge></TableCell>
                <TableCell>{operation.clusterName || "—"}</TableCell>
                <TableCell className="text-xs text-muted-foreground">{operation.message || "—"}</TableCell>
                <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{formatUnixDateTime(operation.createdAt) || "—"}</TableCell>
              </TableRow>
            ))}
            {operations.length === 0 ? <TableRow><TableCell colSpan={5} className="text-sm text-muted-foreground">{t("appStore.emptyOperations")}</TableCell></TableRow> : null}
          </TableBody>
        </Table>
      </div>
      </div>
      </div>
      <Dialog open={actionOpen !== null} onOpenChange={(open) => { if (!open) setActionOpen(null) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>{actionOpen === "upgrade" ? t("appStore.upgrade") : t("appStore.rollback")}</DialogTitle></DialogHeader>
          <DialogBody className="grid gap-4">
            <FormField label={t("appStore.version")} htmlFor="target-version" required description={t("appStore.selectVersionHelp")}>
              <Select value={selectedVersion} onValueChange={setSelectedVersion}>
                <SelectTrigger id="target-version"><SelectValue placeholder={t("appStore.selectVersion")} /></SelectTrigger>
                <SelectContent>{candidates.map((item) => <SelectItem key={item.uid} value={item.uid}>{item.version} ({item.status})</SelectItem>)}</SelectContent>
              </Select>
            </FormField>
          </DialogBody>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setActionOpen(null)}>{t("action.cancel")}</Button>
            <Button disabled={actionPending || !selectedVersion} onClick={() => void runAction()}>{actionPending ? t("action.saving") : t("action.confirm")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={uninstallOpen} onOpenChange={setUninstallOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>{t("appStore.uninstallTitle")}</DialogTitle></DialogHeader>
          <DialogBody className="grid gap-3 text-sm">
            <p className="text-muted-foreground">{t("appStore.uninstallDescription")}</p>
            <p className="text-xs text-muted-foreground">{selectedClusters.length === 0 ? t("appStore.allClustersHint") : t("appStore.selectedClusters", { count: selectedClusters.length })}</p>
            <label className="flex items-center gap-3 rounded-md border border-border/60 p-3">
              <Checkbox checked={keepPVC} onCheckedChange={(value) => setKeepPVC(value === true)} />
              <span>{t("appStore.keepPVC")}</span>
            </label>
            <label className="flex items-center gap-3 rounded-md border border-border/60 p-3">
              <Checkbox checked={deleteClusterScoped} onCheckedChange={(value) => setDeleteClusterScoped(value === true)} />
              <span className="grid gap-1"><span>{t("appStore.deleteClusterScoped")}</span><span className="text-xs text-muted-foreground">{t("appStore.deleteClusterScopedHint")}</span></span>
            </label>
            {selectedClusters.length > 0 ? <p className="text-xs text-muted-foreground">{t("appStore.selectedClusters", { count: selectedClusters.length })}</p> : null}
          </DialogBody>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setUninstallOpen(false)}>{t("action.cancel")}</Button>
            <Button variant="destructive" disabled={actionPending} onClick={() => void uninstall()}>{actionPending ? t("action.saving") : t("appStore.uninstall")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  )
}
