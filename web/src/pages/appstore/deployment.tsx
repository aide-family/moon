import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useParams, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import {
  getApplicationDeployment,
  listApplicationDeploymentOperations,
  listApplicationReleases,
  listApplicationVersions,
  redeployApplicationDeployment,
  rollbackApplicationDeployment,
  uninstallApplicationDeployment,
  upgradeApplicationDeployment,
} from "@/api/application"
import type {
  ApplicationDeploymentItem,
  ApplicationOperationItem,
  ApplicationReleaseItem,
  ApplicationVersionItem,
} from "@/api/application/types"
import { useSession } from "@/auth/session"
import { FormField } from "@/components/form-field"
import { HeaderActionsMenu } from "@/components/header-actions-menu"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsCell, TableActionsHead } from "@/components/table-heads"
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { formatUnixDateTime } from "@/lib/date-time"
import { enumLabel } from "@/lib/enum-label"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

const TABS = ["clusters", "releases", "operations"] as const
type DeploymentTab = (typeof TABS)[number]

export function ApplicationDeploymentPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const { uid = "" } = useParams()
  const [params, setParams] = useSearchParams()
  const tabParam = params.get("tab")
  const tab: DeploymentTab = TABS.includes(tabParam as DeploymentTab) ? (tabParam as DeploymentTab) : "clusters"
  const [deployment, setDeployment] = useState<ApplicationDeploymentItem | null>(null)
  const [operations, setOperations] = useState<ApplicationOperationItem[]>([])
  const [releases, setReleases] = useState<ApplicationReleaseItem[]>([])
  const [versions, setVersions] = useState<ApplicationVersionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [actionOpen, setActionOpen] = useState<"upgrade" | "rollback" | null>(null)
  const [selectedVersion, setSelectedVersion] = useState("")
  const [selectedClusters, setSelectedClusters] = useState<string[]>([])
  const [actionPending, setActionPending] = useState(false)
  const [actionError, setActionError] = useState("")
  const [uninstallOpen, setUninstallOpen] = useState(false)
  const [keepPVC, setKeepPVC] = useState(true)
  const [deleteClusterScoped, setDeleteClusterScoped] = useState(false)

  const upgradeCandidates = useMemo(
    () => versions.filter((item) => item.status === "PUBLISHED"),
    [versions],
  )
  const rollbackCandidates = useMemo(
    () => versions.filter((item) => item.status === "PUBLISHED" || item.status === "DEPRECATED"),
    [versions],
  )
  const candidates = actionOpen === "upgrade" ? upgradeCandidates : rollbackCandidates

  async function reload() {
    setError("")
    try {
      const deploymentResult = await getApplicationDeployment(uid)
      const applicationUid = deploymentResult.deployment.applicationUid
      const [operationsResult, releasesResult, versionsResult] = await Promise.all([
        listApplicationDeploymentOperations(uid, { page: 1, pageSize: 200 }),
        listApplicationReleases({ applicationUid, page: 1, pageSize: 200 }),
        listApplicationVersions(applicationUid, { page: 1, pageSize: 200 }),
      ])
      setDeployment(deploymentResult.deployment)
      setOperations(operationsResult.items)
      setReleases(releasesResult.items.filter((item) => item.deploymentUid === uid))
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
        const deploymentResult = await getApplicationDeployment(uid)
        const applicationUid = deploymentResult.deployment.applicationUid
        const [operationsResult, releasesResult, versionsResult] = await Promise.all([
          listApplicationDeploymentOperations(uid, { page: 1, pageSize: 200 }),
          listApplicationReleases({ applicationUid, page: 1, pageSize: 200 }),
          listApplicationVersions(applicationUid, { page: 1, pageSize: 200 }),
        ])
        if (!active) return
        setDeployment(deploymentResult.deployment)
        setOperations(operationsResult.items)
        setReleases(releasesResult.items.filter((item) => item.deploymentUid === uid))
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
    return (
      <section className="flex min-h-0 flex-1 flex-col gap-6">
        <PageHeader title={t("appStore.deploymentName")} description={loading ? t("state.loading") : error} />
      </section>
    )
  }

  function setTab(next: string) {
    const query = new URLSearchParams(params)
    query.set("tab", next)
    setParams(query)
  }

  async function runAction() {
    setActionPending(true)
    setActionError("")
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
      setActionError(formErrorMessage(err, t("state.failed")))
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
    setActionError("")
    try {
      await uninstallApplicationDeployment(uid, selectedClusters, keepPVC, deleteClusterScoped)
      setUninstallOpen(false)
      toast.success(t("appStore.operationSubmitted"))
      await reload()
    } catch (err) {
      setActionError(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  const selectedClusterCount = selectedClusters.length
  const outOfSync = deployment.clusters.filter((cluster) => cluster.inSync === false)

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={deployment.name}
        description={`${deployment.applicationName} ${deployment.version} · ${deployment.targetNamespace} · ${enumLabel(t, deployment.status)}${selectedClusterCount > 0 ? ` · ${t("appStore.selectedClusters", { count: selectedClusterCount })}` : ""}`}
        titleAccessory={<StatusBadge value={deployment.status} />}
      >
        <Button variant="outline" onClick={() => navigate("/appstore?tab=deployments")}>{t("action.goBack")}</Button>
        {session.can("release:write") ? <Button variant="outline" onClick={() => void redeploy()}>{t("appStore.redeploy")}</Button> : null}
        {session.can("release:write") ? <Button onClick={() => { setSelectedVersion(""); setActionError(""); setActionOpen("upgrade") }}>{t("appStore.upgrade")}</Button> : null}
        {session.can("release:write") || session.can("release:delete") ? (
          <HeaderActionsMenu
            label={t("action.more")}
            actions={[
              ...(session.can("release:write") ? [{ key: "rollback", label: t("appStore.rollback"), onSelect: () => { setSelectedVersion(""); setActionError(""); setActionOpen("rollback") } }] : []),
              ...(session.can("release:delete") ? [{ key: "uninstall", label: t("appStore.uninstall"), destructive: true, onSelect: () => { setActionError(""); setUninstallOpen(true) } }] : []),
            ]}
          />
        ) : null}
      </PageHeader>

      <div className="grid gap-2">
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        {outOfSync.length > 0 ? (
          <p className="rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-sm text-warning">
            {t("appStore.outOfSyncHint", { count: outOfSync.length })}
          </p>
        ) : null}
      </div>

      <Tabs value={tab} onValueChange={setTab} className="min-h-0 flex-1">
        <TabsList className="shrink-0">
          <TabsTrigger value="clusters">{t("appStore.tabs.clusters")}</TabsTrigger>
          <TabsTrigger value="releases">{t("appStore.tabs.releases")}</TabsTrigger>
          <TabsTrigger value="operations">{t("appStore.operations")}</TabsTrigger>
        </TabsList>

        <TabsContent value="clusters" className="min-h-0 flex-1 overflow-y-auto pb-2">
          <div className="grid gap-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm text-muted-foreground">{t("appStore.clustersHint")}</p>
              <span className="text-xs text-muted-foreground">
                {selectedClusterCount === 0 ? t("appStore.allClustersHint") : t("appStore.selectedClusters", { count: selectedClusterCount })}
              </span>
            </div>
            <div className="overflow-hidden rounded-xl border border-border/60 bg-card">
              <Table className={TABLE_MIN_WIDTH.wide}>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-10">
                      <Checkbox
                        checked={selectedClusters.length === deployment.clusters.length && deployment.clusters.length > 0}
                        onCheckedChange={(value) => setSelectedClusters(value === true ? deployment.clusters.map((cluster) => cluster.uid) : [])}
                      />
                    </TableHead>
                    <TableHead>{t("appStore.cluster")}</TableHead>
                    <TableHead>{t("appStore.desiredVersion")}</TableHead>
                    <TableHead>{t("appStore.currentVersion")}</TableHead>
                    <TableHead>{t("appStore.inSync")}</TableHead>
                    <TableHead>{t("appStore.status")}</TableHead>
                    <TableHead>{t("appStore.message")}</TableHead>
                    <TableActionsHead>{t("action.actions")}</TableActionsHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {deployment.clusters.map((cluster) => (
                    <TableRow key={cluster.uid}>
                      <TableCell>
                        <Checkbox
                          checked={selectedClusters.includes(cluster.uid)}
                          onCheckedChange={(value) =>
                            setSelectedClusters((current) =>
                              value === true ? [...current, cluster.uid] : current.filter((item) => item !== cluster.uid),
                            )
                          }
                        />
                      </TableCell>
                      <TableCell>{cluster.name}</TableCell>
                      <TableCell className="font-mono text-xs">{cluster.version || deployment.version}</TableCell>
                      <TableCell className="font-mono text-xs">{cluster.currentVersion || "—"}</TableCell>
                      <TableCell>
                        <StatusBadge
                          value={cluster.inSync ? "IN_SYNC" : "OUT_OF_SYNC"}
                          label={cluster.inSync ? t("appStore.inSyncYes") : t("appStore.inSyncNo")}
                        />
                      </TableCell>
                      <TableCell><StatusBadge value={cluster.status} /></TableCell>
                      <TableCell className="max-w-[18rem] text-xs text-muted-foreground break-words">{cluster.message || "—"}</TableCell>
                      <TableActionsCell>
                        <RowActions
                          nowrap
                          detailLabel={t("appStore.retryCluster")}
                          moreLabel={t("action.more")}
                          onDetail={() => void redeploy([cluster.uid])}
                          detailDisabled={actionPending || cluster.status === "UNINSTALLED"}
                          actions={[]}
                        />
                      </TableActionsCell>
                    </TableRow>
                  ))}
                  {deployment.clusters.length === 0 ? (
                    <TableRow><TableCell colSpan={8} className="text-sm text-muted-foreground">{t("appStore.noClusters")}</TableCell></TableRow>
                  ) : null}
                </TableBody>
              </Table>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="releases" className="min-h-0 flex-1 overflow-y-auto pb-2">
          <div className="overflow-hidden rounded-xl border border-border/60 bg-card">
            <Table className={TABLE_MIN_WIDTH.default}>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("appStore.cluster")}</TableHead>
                  <TableHead>{t("appStore.version")}</TableHead>
                  <TableHead>{t("appStore.status")}</TableHead>
                  <TableHead>{t("appStore.targetNamespace")}</TableHead>
                  <TableHead className={TABLE_NOWRAP}>{t("appStore.deployedAt")}</TableHead>
                  <TableHead>{t("appStore.message")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {releases.map((release) => (
                  <TableRow key={release.uid}>
                    <TableCell>{release.clusterName || "—"}</TableCell>
                    <TableCell className="font-mono text-xs">{release.version || "—"}</TableCell>
                    <TableCell><StatusBadge value={release.status} /></TableCell>
                    <TableCell className="font-mono text-xs">{release.targetNamespace || "—"}</TableCell>
                    <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{formatUnixDateTime(release.deployedAt) || "—"}</TableCell>
                    <TableCell className="max-w-[20rem] text-xs text-muted-foreground break-words">{release.message || "—"}</TableCell>
                  </TableRow>
                ))}
                {releases.length === 0 ? (
                  <TableRow><TableCell colSpan={6} className="text-sm text-muted-foreground">{t("appStore.emptyReleases")}</TableCell></TableRow>
                ) : null}
              </TableBody>
            </Table>
          </div>
        </TabsContent>

        <TabsContent value="operations" className="min-h-0 flex-1 overflow-y-auto pb-2">
          <div className="overflow-hidden rounded-xl border border-border/60 bg-card">
            <Table className={TABLE_MIN_WIDTH.default}>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("appStore.operationAction")}</TableHead>
                  <TableHead>{t("appStore.status")}</TableHead>
                  <TableHead>{t("appStore.cluster")}</TableHead>
                  <TableHead>{t("appStore.message")}</TableHead>
                  <TableHead className={TABLE_NOWRAP}>{t("action.createdAt")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {operations.map((operation) => (
                  <TableRow key={operation.uid}>
                    <TableCell>{enumLabel(t, operation.action)}</TableCell>
                    <TableCell><StatusBadge value={operation.status} /></TableCell>
                    <TableCell>{operation.clusterName || "—"}</TableCell>
                    <TableCell className="max-w-[22rem] text-xs text-muted-foreground break-words">{operation.message || "—"}</TableCell>
                    <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{formatUnixDateTime(operation.createdAt) || "—"}</TableCell>
                  </TableRow>
                ))}
                {operations.length === 0 ? (
                  <TableRow><TableCell colSpan={5} className="text-sm text-muted-foreground">{t("appStore.emptyOperations")}</TableCell></TableRow>
                ) : null}
              </TableBody>
            </Table>
          </div>
        </TabsContent>
      </Tabs>

      <Dialog open={actionOpen !== null} onOpenChange={(open) => { if (!open) setActionOpen(null) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{actionOpen === "upgrade" ? t("appStore.upgrade") : t("appStore.rollback")}</DialogTitle>
          </DialogHeader>
          <DialogBody className="grid gap-4">
            <FormField
              label={t("appStore.version")}
              htmlFor="target-version"
              required
              description={actionOpen === "upgrade" ? t("appStore.upgradeVersionHelp") : t("appStore.rollbackVersionHelp")}
            >
              <Select value={selectedVersion} onValueChange={setSelectedVersion}>
                <SelectTrigger id="target-version"><SelectValue placeholder={t("appStore.selectVersion")} /></SelectTrigger>
                <SelectContent>
                  {candidates.map((item) => (
                    <SelectItem key={item.uid} value={item.uid}>
                      {`${item.version} · ${enumLabel(t, item.status)}`}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <p className="text-xs text-muted-foreground">
              {t("appStore.actionClustersHint", {
                count: selectedClusterCount,
                all: selectedClusterCount === 0 ? t("appStore.allClustersLabel") : "",
              })}
            </p>
          </DialogBody>
          {actionError ? <p className="text-sm text-destructive">{actionError}</p> : null}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setActionOpen(null)}>{t("action.cancel")}</Button>
            <Button disabled={actionPending || !selectedVersion} onClick={() => void runAction()}>
              {actionPending ? t("action.saving") : t("action.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={uninstallOpen} onOpenChange={setUninstallOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>{t("appStore.uninstallTitle")}</DialogTitle></DialogHeader>
          <DialogBody className="grid gap-3 text-sm">
            <p className="text-muted-foreground">{t("appStore.uninstallDescription")}</p>
            <p className="text-xs text-muted-foreground">
              {selectedClusterCount === 0 ? t("appStore.allClustersHint") : t("appStore.selectedClusters", { count: selectedClusterCount })}
            </p>
            <label className="flex items-center gap-3 rounded-md border border-border/60 p-3">
              <Checkbox checked={keepPVC} onCheckedChange={(value) => setKeepPVC(value === true)} />
              <span>{t("appStore.keepPVC")}</span>
            </label>
            <label className="flex items-center gap-3 rounded-md border border-border/60 p-3">
              <Checkbox checked={deleteClusterScoped} onCheckedChange={(value) => setDeleteClusterScoped(value === true)} />
              <span className="grid gap-1">
                <span>{t("appStore.deleteClusterScoped")}</span>
                <span className="text-xs text-muted-foreground">{t("appStore.deleteClusterScopedHint")}</span>
              </span>
            </label>
          </DialogBody>
          {actionError ? <p className="text-sm text-destructive">{actionError}</p> : null}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setUninstallOpen(false)}>{t("action.cancel")}</Button>
            <Button variant="destructive" disabled={actionPending} onClick={() => void uninstall()}>
              {actionPending ? t("action.saving") : t("appStore.uninstall")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  )
}
