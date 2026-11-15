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
import { ConfirmAction } from "@/components/confirm-action"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
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
  const [actionPending, setActionPending] = useState(false)
  const [uninstallOpen, setUninstallOpen] = useState(false)

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
        await upgradeApplicationDeployment(uid, selectedVersion, [])
      } else {
        await rollbackApplicationDeployment(uid, selectedVersion, [])
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

  async function redeploy() {
    try {
      await redeployApplicationDeployment(uid, [])
      toast.success(t("appStore.operationSubmitted"))
      await reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  async function uninstall() {
    setActionPending(true)
    try {
      await uninstallApplicationDeployment(uid, [], true, false)
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
        description={`${deployment.applicationName} ${deployment.version} · ${deployment.targetNamespace} · ${deployment.status}`}
      >
        <Button variant="outline" onClick={() => navigate("/appstore?tab=deployments")}>{t("action.goBack")}</Button>
        {session.can("release:write") ? <Button variant="outline" onClick={() => void redeploy()}>{t("appStore.redeploy")}</Button> : null}
        {session.can("release:write") ? <Button variant="outline" onClick={() => { setSelectedVersion(""); setActionOpen("upgrade") }}>{t("appStore.upgrade")}</Button> : null}
        {session.can("release:write") ? <Button variant="outline" onClick={() => { setSelectedVersion(""); setActionOpen("rollback") }}>{t("appStore.rollback")}</Button> : null}
        {session.can("release:delete") ? <Button variant="destructive" onClick={() => setUninstallOpen(true)}>{t("appStore.uninstall")}</Button> : null}
      </PageHeader>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <div className="rounded-xl border border-border/60 bg-card">
        <div className="border-b border-border px-4 py-3"><h2 className="font-heading text-base font-semibold">{t("appStore.targetClusters")}</h2></div>
        <Table className={TABLE_MIN_WIDTH.default}>
          <TableHeader><TableRow><TableHead>{t("appStore.cluster")}</TableHead><TableHead>{t("appStore.status")}</TableHead><TableHead>{t("appStore.version")}</TableHead><TableHead>{t("appStore.targetNamespace")}</TableHead><TableHead>{t("appStore.message")}</TableHead></TableRow></TableHeader>
          <TableBody>
            {deployment.clusters.map((cluster) => (
              <TableRow key={cluster.uid}>
                <TableCell>{cluster.name}</TableCell>
                <TableCell><Badge variant={cluster.status === "RUNNING" ? "secondary" : cluster.status === "FAILED" ? "destructive" : "outline"}>{cluster.status}</Badge></TableCell>
                <TableCell className="font-mono text-xs">{cluster.version || "—"}</TableCell>
                <TableCell className="font-mono text-xs">{deployment.targetNamespace}</TableCell>
                <TableCell className="text-xs text-muted-foreground">{cluster.message || "—"}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <div className="rounded-xl border border-border/60 bg-card">
        <div className="border-b border-border px-4 py-3"><h2 className="font-heading text-base font-semibold">{t("appStore.operations")}</h2></div>
        <Table className={TABLE_MIN_WIDTH.wide}>
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
      <Dialog open={actionOpen !== null} onOpenChange={(open) => { if (!open) setActionOpen(null) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>{actionOpen === "upgrade" ? t("appStore.upgrade") : t("appStore.rollback")}</DialogTitle></DialogHeader>
          <DialogBody className="grid gap-4">
            <FormField label={t("appStore.version")} htmlFor="target-version" required>
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
      <ConfirmAction
        open={uninstallOpen}
        title={t("appStore.uninstallTitle")}
        description={t("appStore.uninstallDescription")}
        confirmLabel={t("appStore.uninstall")}
        cancelLabel={t("action.cancel")}
        pending={actionPending}
        onOpenChange={setUninstallOpen}
        onConfirm={() => void uninstall()}
      />
    </section>
  )
}
