import { Search } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useSearchParams } from "react-router-dom"

import { listApplications, listApplicationDeployments } from "@/api/application"
import { listClusters } from "@/api/cluster"
import type {
  ApplicationDeploymentItem,
  ApplicationItem,
  DeploymentStatus,
} from "@/api/application/types"
import type { ClusterItem } from "@/api/cluster/types"
import { useSession } from "@/auth/session"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsCell, TableActionsHead } from "@/components/table-heads"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
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
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import {
  APPLICATION_CATEGORY_MAX,
  APPLICATION_CATEGORY_PRESETS,
  applicationCategoryLabel,
} from "@/lib/application-category"
import { formatUnixDateTime } from "@/lib/date-time"
import { enumLabel } from "@/lib/enum-label"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

import { ApplicationDialog } from "./app-dialog"

const ALL = "ALL"
const STATUSES: DeploymentStatus[] = ["PENDING", "DEPLOYING", "RUNNING", "PARTIAL", "FAILED", "UNINSTALLING", "UNINSTALLED"]

/** 部署组内各集群的当前版本（去重），用于列表的「集群当前版本」列。 */
function clusterCurrentVersions(item: ApplicationDeploymentItem): string[] {
  const versions = new Set<string>()
  for (const cluster of item.clusters ?? []) {
    if (cluster.currentVersion) {
      versions.add(cluster.currentVersion)
    }
  }
  return [...versions]
}

export function AppStorePage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const [params, setParams] = useSearchParams()
  const tab = params.get("tab") === "deployments" ? "deployments" : "apps"
  const [dialogOpen, setDialogOpen] = useState(false)

  if (!session.loading && !session.can("application:read") && !session.can("release:read")) {
    return <Navigate to="/forbidden" replace />
  }

  function setTab(next: string) {
    const query = new URLSearchParams(params)
    query.set("tab", next)
    query.delete("page")
    query.delete("name")
    query.delete("status")
    setParams(query)
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("appStore.title")} description={t("appStore.description")}>
        {tab === "apps" && session.can("application:write") ? (
          <Button onClick={() => setDialogOpen(true)}>{t("appStore.newApp")}</Button>
        ) : null}
      </PageHeader>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="inline-flex items-center rounded-lg border border-border/70 bg-muted/40 p-1">
          <Button size="sm" variant={tab === "apps" ? "secondary" : "ghost"} onClick={() => setTab("apps")}>{t("appStore.apps")}</Button>
          <Button size="sm" variant={tab === "deployments" ? "secondary" : "ghost"} onClick={() => setTab("deployments")}>{t("appStore.deployments")}</Button>
        </div>
        <p className="text-sm text-muted-foreground">{tab === "apps" ? t("appStore.appsHint") : t("appStore.deploymentsHint")}</p>
      </div>
      {tab === "apps" ? <AppsView navigate={navigate} /> : <DeploymentsView navigate={navigate} />}
      <ApplicationDialog
        key={dialogOpen ? "app-create" : "app-create-closed"}
        open={dialogOpen}
        application={null}
        onOpenChange={setDialogOpen}
        onSaved={(item) => {
          setDialogOpen(false)
          navigate(`/appstore/apps/${item.uid}`)
        }}
      />
    </section>
  )
}

function AppsView({ navigate }: { navigate: (path: string) => void }) {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const { items, total, loading, failed, error, reload } = usePagedList<ApplicationItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${query.type}`,
    fetch: (page, pageSize) => listApplications({ page, pageSize, name: query.name, category: query.type }),
  })
  return (
    <>
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative w-full max-w-[500px]">
          <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input className="pl-9" value={query.name} placeholder={t("appStore.searchApps")} onChange={(event) => query.update({ name: event.target.value })} />
        </div>
        <Input
          className="w-44"
          value={query.type}
          list="app-category-filter"
          maxLength={APPLICATION_CATEGORY_MAX}
          placeholder={t("appStore.categoryFilterPlaceholder")}
          aria-label={t("appStore.categoryFilter")}
          onChange={(event) => query.update({ type: event.target.value })}
        />
        <datalist id="app-category-filter">
          {APPLICATION_CATEGORY_PRESETS.map((preset) => (
            <option key={preset} value={preset}>{t(`appStore.categories.${preset}`)}</option>
          ))}
        </datalist>
      </div>
      <ListFrame total={total} page={query.page} pageSize={query.pageSize} showPager={!loading && !failed && items.length > 0} onPrev={() => query.update({ page: Math.max(1, query.page - 1) })} onNext={() => query.update({ page: query.page + 1 })}>
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2"><p className="text-sm text-destructive">{error || t("state.failed")}</p><Button variant="outline" onClick={() => reload()}>{t("action.retry")}</Button></div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <div className="grid gap-2"><p className="text-sm text-muted-foreground">{t("appStore.emptyApps")}</p><p className="text-sm text-muted-foreground">{t("appStore.emptyAppsHint")}</p></div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader><TableRow><TableHead>{t("appStore.appName")}</TableHead><TableHead>{t("appStore.category")}</TableHead><TableHead className={TABLE_NOWRAP}>{t("appStore.versions")}</TableHead><TableHead>{t("appStore.latestVersion")}</TableHead><TableHead className={TABLE_NOWRAP}>{t("action.createdAt")}</TableHead><TableActionsHead>{t("action.actions")}</TableActionsHead></TableRow></TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell><div className="grid gap-1"><button type="button" className="text-left font-medium underline-offset-4 hover:underline" onClick={() => navigate(`/appstore/apps/${item.uid}`)}>{item.name}</button>{item.displayName ? <span className="text-xs text-muted-foreground">{item.displayName}</span> : null}</div></TableCell>
                  <TableCell><Badge variant="outline">{applicationCategoryLabel(t, item.category)}</Badge></TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{item.versionCount}</TableCell>
                  <TableCell className="font-mono text-xs">{item.latestVersion || "—"}</TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{formatUnixDateTime(item.createdAt) || "—"}</TableCell>
                  <TableActionsCell><RowActions nowrap detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => navigate(`/appstore/apps/${item.uid}`)} actions={[]} /></TableActionsCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
    </>
  )
}

function DeploymentsView({ navigate }: { navigate: (path: string) => void }) {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const status = STATUSES.includes(query.status as DeploymentStatus) ? (query.status as DeploymentStatus) : ""
  const [clusters, setClusters] = useState<ClusterItem[]>([])
  useEffect(() => {
    let active = true
    void listClusters({ page: 1, pageSize: 100 })
      .then((result) => {
        if (active) setClusters(result.items)
      })
      .catch(() => undefined)
    return () => {
      active = false
    }
  }, [])
  const { items, total, loading, failed, error, reload } = usePagedList<ApplicationDeploymentItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${status}|${query.clusterUid}`,
    fetch: (page, pageSize) => listApplicationDeployments({ page, pageSize, name: query.name, status, clusterUid: query.clusterUid }),
  })
  return (
    <>
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative w-full max-w-[500px]">
          <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input className="pl-9" value={query.name} placeholder={t("appStore.searchDeployments")} onChange={(event) => query.update({ name: event.target.value })} />
        </div>
        <Select value={status || ALL} onValueChange={(value) => query.update({ status: value === ALL ? "" : value })}>
          <SelectTrigger className="w-44"><SelectValue>{status ? enumLabel(t, status) : t("appStore.allStatuses")}</SelectValue></SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>{t("appStore.allStatuses")}</SelectItem>
            {STATUSES.map((item) => <SelectItem key={item} value={item}>{enumLabel(t, item)}</SelectItem>)}
          </SelectContent>
        </Select>
        <Select value={query.clusterUid || ALL} onValueChange={(value) => query.update({ clusterUid: value === ALL ? "" : value })}>
          <SelectTrigger className="w-44"><SelectValue>{clusters.find((item) => item.uid === query.clusterUid)?.name || t("appStore.allClusters")}</SelectValue></SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>{t("appStore.allClusters")}</SelectItem>
            {clusters.map((item) => <SelectItem key={item.uid} value={item.uid}>{item.name}</SelectItem>)}
          </SelectContent>
        </Select>
      </div>
      <ListFrame total={total} page={query.page} pageSize={query.pageSize} showPager={!loading && !failed && items.length > 0} onPrev={() => query.update({ page: Math.max(1, query.page - 1) })} onNext={() => query.update({ page: query.page + 1 })}>
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2"><p className="text-sm text-destructive">{error || t("state.failed")}</p><Button variant="outline" onClick={() => reload()}>{t("action.retry")}</Button></div>
        ) : null}
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("appStore.emptyDeployments")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader><TableRow><TableHead>{t("appStore.deploymentName")}</TableHead><TableHead>{t("appStore.applicationVersion")}</TableHead><TableHead>{t("appStore.targetClusters")}</TableHead><TableHead>{t("appStore.clusterCurrentVersion")}</TableHead><TableHead className={TABLE_NOWRAP}>{t("appStore.status")}</TableHead><TableHead className={TABLE_NOWRAP}>{t("action.updatedAt")}</TableHead><TableActionsHead>{t("action.actions")}</TableActionsHead></TableRow></TableHeader>
            <TableBody>
              {items.map((item) => {
                const currentVersions = clusterCurrentVersions(item)
                const allInSync = item.clusters.length > 0 && item.clusters.every((cluster) => cluster.inSync)
                return (
                  <TableRow key={item.uid}>
                    <TableCell><button type="button" className="text-left font-medium underline-offset-4 hover:underline" onClick={() => navigate(`/appstore/deployments/${item.uid}`)}>{item.name}</button><p className="text-xs text-muted-foreground">{item.targetNamespace}</p></TableCell>
                    <TableCell><div className="grid gap-1"><span>{item.applicationName}</span><span className="font-mono text-xs text-muted-foreground">{item.version}</span></div></TableCell>
                    <TableCell><div className="flex flex-wrap gap-1">{item.clusters.map((cluster) => <Badge key={cluster.uid} variant="outline">{cluster.name}</Badge>)}</div></TableCell>
                    <TableCell>
                      <div className="grid gap-1">
                        <span className="font-mono text-xs">
                          {currentVersions.length === 0 ? "—" : currentVersions.slice(0, 2).join(", ")}
                          {currentVersions.length > 2 ? ` +${currentVersions.length - 2}` : ""}
                        </span>
                        {currentVersions.length > 0 ? (
                          <StatusBadge
                            value={allInSync ? "IN_SYNC" : "OUT_OF_SYNC"}
                            label={allInSync ? t("appStore.inSyncYes") : t("appStore.inSyncNo")}
                          />
                        ) : null}
                      </div>
                    </TableCell>
                    <TableCell><StatusBadge value={item.status} /></TableCell>
                    <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{formatUnixDateTime(item.createdAt) || "—"}</TableCell>
                    <TableActionsCell><RowActions nowrap detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => navigate(`/appstore/deployments/${item.uid}`)} actions={[]} /></TableActionsCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
    </>
  )
}
