import { ArrowLeft, PlugZap, RefreshCw, Search } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, Navigate, useParams } from "react-router-dom"

import {
  getCluster,
  listKubernetesNamespaces,
  testCluster,
} from "@/api/cluster"
import type { ClusterItem, ClusterProbe } from "@/api/cluster"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { FilterChips, type FilterChip } from "@/components/filter-chips"
import { FilterToggle } from "@/components/filter-toggle"
import { PageHeader } from "@/components/page-header"
import { SearchableSelect } from "@/components/searchable-select"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { failureMessage, formErrorMessage } from "@/lib/form-error"

import { KubernetesConfigMapsTab } from "../configmaps-tab"
import { KubernetesIngressesTab } from "../ingresses-tab"
import { KubernetesNodesTab } from "../nodes-tab"
import { KubernetesPersistentVolumeClaimsTab } from "../persistent-volume-claims-tab"
import { KubernetesPersistentVolumesTab } from "../persistent-volumes-tab"
import { KubernetesPodsTab } from "../pods-tab"
import { ClusterProbeDialog } from "../probe-dialog"
import { KubernetesSecretsTab } from "../secrets-tab"
import { KubernetesServicesTab } from "../services-tab"
import { KubernetesWorkloadsTab } from "../workloads-tab"
import { ClusterSummary } from "./cluster-summary"
import { ClusterOverviewTab } from "./overview-tab"
import { ClusterResourceNav, ClusterResourcePicker } from "./resource-nav"
import {
  clusterSectionScope,
  clusterSectionSupportsName,
  findClusterNavEntry,
  readClusterNavLabel,
  type ClusterNavCounts,
  type ClusterSection,
} from "./sections"
import { useClusterDetailQuery } from "./use-cluster-detail-query"
import { useClusterOverview } from "./use-cluster-overview"
import { useClusterResourceCounts } from "./use-resource-counts"

const FILTER_PANEL_ID = "cluster-resource-filters"

/**
 * 集群详情：左侧按 Kubernetes 资源类型分组的导航 + 右侧资源面板。
 * 节点 / PV 是集群级资源，其余资源跟随命名空间切换，全部状态写进 URL
 *（tab / kind / namespace / name / page / pageSize）。
 */
export function ClusterDetailPage() {
  const { t } = useTranslation()
  const { uid = "" } = useParams()
  const session = useSession()
  const query = useClusterDetailQuery()
  const [cluster, setCluster] = useState<ClusterItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [notFound, setNotFound] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [reloadKey, setReloadKey] = useState(0)
  const [refreshKey, setRefreshKey] = useState(0)
  const [testPending, setTestPending] = useState(false)
  const [probeDialog, setProbeDialog] = useState<{
    probe: ClusterProbe | null
    error: string
  } | null>(null)
  const [namespaces, setNamespaces] = useState<string[]>([])
  const [namespacesLoading, setNamespacesLoading] = useState(true)
  const [namespacesFailed, setNamespacesFailed] = useState(false)
  const [namespacesAttempt, setNamespacesAttempt] = useState(0)
  const [filtersOpen, setFiltersOpen] = useState(false)
  // 连接状态整块卡片默认收起，由页头的集群名称控制显隐。
  const [summaryOpen, setSummaryOpen] = useState(false)
  const [draftState, setDraftState] = useState<{
    signature: string
    name: string
  } | null>(null)

  const canWrite = session.can("cluster:write")
  // 测试连通性只读集群元数据，服务端按 cluster:read 校验。
  const canTest = session.can("cluster:read")

  const section = query.section
  const kind = query.kind
  const scope = clusterSectionScope(section, kind)
  const entry = findClusterNavEntry(section, kind)
  const supportsNameFilter = clusterSectionSupportsName(section)
  const overview = useClusterOverview({
    clusterUid: uid,
    namespace: query.namespace,
    refreshKey,
  })
  const resourceCounts = useClusterResourceCounts({
    clusterUid: uid,
    namespace: query.namespace,
    refreshKey,
  })

  useEffect(() => {
    if (!uid) {
      return
    }
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      setNotFound(false)
      try {
        const result = await getCluster(uid)
        if (!active) {
          return
        }
        setCluster(result.cluster)
        setLoadError("")
      } catch (err) {
        if (!active) {
          return
        }
        setCluster(null)
        setFailed(true)
        setNotFound(err instanceof ApiError && err.status === 404)
        setLoadError(failureMessage(err, ""))
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
  }, [uid, reloadKey])

  useEffect(() => {
    if (!uid) {
      return
    }
    let active = true
    async function loadNamespaces() {
      setNamespacesLoading(true)
      setNamespacesFailed(false)
      try {
        const result = await listKubernetesNamespaces(uid)
        if (!active) {
          return
        }
        setNamespaces(result.items ?? [])
      } catch {
        if (active) {
          setNamespaces([])
          setNamespacesFailed(true)
        }
      } finally {
        if (active) {
          setNamespacesLoading(false)
        }
      }
    }
    void loadNamespaces()
    return () => {
      active = false
    }
  }, [uid, namespacesAttempt])

  const navCounts = useMemo<ClusterNavCounts>(() => {
    const counts: ClusterNavCounts = { ...resourceCounts }
    const data = overview.data
    if (!data) {
      return counts
    }
    if (!data.failed.nodes && data.nodes) {
      counts.nodes = data.nodes.total
    }
    if (!data.failed.pods && data.pods) {
      counts.pods = data.pods.total
    }
    if (!data.failed.ingresses && data.ingresses) {
      counts.ingresses = data.ingresses.total
    }
    if (!data.failed.workloads && data.workloads) {
      counts.workloads = data.workloads.total
      for (const row of data.workloads.rows) {
        counts[`workloads/${row.kind}`] = row.total
      }
    }
    return counts
  }, [overview.data, resourceCounts])

  // 命名空间下拉：第一项「全部命名空间」用空值表示，其余按名称搜索选择。
  const namespaceOptions = useMemo(
    () => [
      { value: "", label: t("kube.allNamespaces") },
      ...namespaces.map((item) => ({ value: item, label: item })),
    ],
    [namespaces, t]
  )

  if (!session.loading && !session.can("cluster:read")) {
    return <Navigate to="/forbidden" replace />
  }

  const applied = {
    section,
    kind,
    namespace: query.namespace,
    name: query.name,
  }
  const appliedSignature = JSON.stringify(applied)
  const draftName =
    draftState?.signature === appliedSignature ? draftState.name : applied.name
  const filtersDirty = draftName !== applied.name
  const appliedCount = applied.name ? 1 : 0
  const filterChips: FilterChip[] = applied.name
    ? [
        {
          key: "name",
          label: `${t("kube.name")}: ${applied.name}`,
          onRemove: () => query.update({ name: "" }),
        },
      ]
    : []

  function selectSection(nextSection: ClusterSection, nextKind?: string) {
    if (nextSection === section && (nextKind ?? "") === kind) {
      return
    }
    query.update({ section: nextSection, kind: nextKind ?? "" })
    setDraftState(null)
  }

  function applyFilters() {
    query.update({ name: draftName })
    setDraftState(null)
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "" })
    setDraftState(null)
  }

  async function runTest() {
    setTestPending(true)
    try {
      const result = await testCluster(uid)
      setCluster(result.cluster)
      setProbeDialog({ probe: result.probe, error: "" })
    } catch (err) {
      setProbeDialog({
        probe: null,
        error: formErrorMessage(err, t("cluster.testFailed")),
      })
      // 探测失败时后端也会刷新最近检测时间与最近错误，尽力同步一次元数据。
      try {
        const refreshed = await getCluster(uid)
        setCluster(refreshed.cluster)
      } catch {
        // 保留当前元数据即可。
      }
    } finally {
      setTestPending(false)
    }
  }

  function renderResourcePanel() {
    switch (section) {
      case "overview":
        return (
          <ClusterOverviewTab
            overview={overview}
            namespace={query.namespace}
            onOpenWorkloadKind={(nextKind) =>
              selectSection("workloads", nextKind)
            }
          />
        )
      case "nodes":
        return (
          <KubernetesNodesTab
            clusterUid={uid}
            name={applied.name}
            canWrite={canWrite}
            refreshKey={refreshKey}
          />
        )
      case "workloads":
        return (
          <KubernetesWorkloadsTab
            clusterUid={uid}
            namespace={applied.namespace}
            kind={kind}
            name={applied.name}
            canWrite={canWrite}
            refreshKey={refreshKey}
          />
        )
      case "pods":
        return (
          <KubernetesPodsTab
            clusterUid={uid}
            namespace={applied.namespace}
            name={applied.name}
            canWrite={canWrite}
            page={query.page}
            pageSize={query.pageSize}
            onPageChange={(page) => query.update({ page })}
            refreshKey={refreshKey}
          />
        )
      case "ingresses":
        return (
          <KubernetesIngressesTab
            clusterUid={uid}
            namespace={applied.namespace}
            name={applied.name}
            page={query.page}
            pageSize={query.pageSize}
            onPageChange={(page) => query.update({ page })}
            refreshKey={refreshKey}
          />
        )
      case "services":
        return (
          <KubernetesServicesTab
            clusterUid={uid}
            namespace={applied.namespace}
            name={applied.name}
            page={query.page}
            pageSize={query.pageSize}
            onPageChange={(page) => query.update({ page })}
            refreshKey={refreshKey}
          />
        )
      case "configmaps":
        return (
          <KubernetesConfigMapsTab
            clusterUid={uid}
            namespace={applied.namespace}
            name={applied.name}
            page={query.page}
            pageSize={query.pageSize}
            onPageChange={(page) => query.update({ page })}
            refreshKey={refreshKey}
          />
        )
      case "secrets":
        return (
          <KubernetesSecretsTab
            clusterUid={uid}
            namespace={applied.namespace}
            name={applied.name}
            page={query.page}
            pageSize={query.pageSize}
            onPageChange={(page) => query.update({ page })}
            refreshKey={refreshKey}
          />
        )
      case "pvs":
        return (
          <KubernetesPersistentVolumesTab
            clusterUid={uid}
            name={applied.name}
            page={query.page}
            pageSize={query.pageSize}
            onPageChange={(page) => query.update({ page })}
            refreshKey={refreshKey}
          />
        )
      case "pvcs":
        return (
          <KubernetesPersistentVolumeClaimsTab
            clusterUid={uid}
            namespace={applied.namespace}
            name={applied.name}
            page={query.page}
            pageSize={query.pageSize}
            onPageChange={(page) => query.update({ page })}
            refreshKey={refreshKey}
          />
        )
      default:
        return null
    }
  }

  const sectionLabel = entry ? readClusterNavLabel(t, entry.label) : ""
  const SectionIcon = entry?.icon
  const ready = !loading && !failed && cluster !== null

  return (
    <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-4">
      <PageHeader
        title={cluster?.name ?? t("cluster.detailTitle")}
        onTitleClick={ready ? () => setSummaryOpen((open) => !open) : undefined}
        titleExpanded={summaryOpen}
        titleToggleLabel={
          summaryOpen ? t("cluster.hideSummary") : t("cluster.showSummary")
        }
      >
        <div className="flex flex-wrap items-center justify-end gap-2">
          {canTest ? (
            <Button
              variant="outline"
              disabled={testPending || !cluster}
              onClick={() => void runTest()}
            >
              <PlugZap aria-hidden="true" data-icon="inline-start" />
              {testPending ? t("cluster.testing") : t("cluster.testConnection")}
            </Button>
          ) : null}
          <Button variant="outline" asChild>
            <Link to="/clusters">
              <ArrowLeft aria-hidden="true" data-icon="inline-start" />
              {t("cluster.backToClusters")}
            </Link>
          </Button>
        </div>
      </PageHeader>
      {loading ? (
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      ) : null}
      {!loading && notFound ? (
        <div className="flex items-center gap-2">
          <p className="text-sm text-muted-foreground">
            {t("cluster.notFound")}
          </p>
          <Button variant="outline" asChild>
            <Link to="/clusters">{t("cluster.backToClusters")}</Link>
          </Button>
        </div>
      ) : null}
      {!loading && !notFound && failed ? (
        <div className="flex items-center gap-2">
          <p className="text-sm text-destructive">
            {loadError || t("state.failed")}
          </p>
          <Button
            variant="outline"
            onClick={() => setReloadKey((value) => value + 1)}
          >
            {t("action.retry")}
          </Button>
        </div>
      ) : null}
      {ready && summaryOpen ? <ClusterSummary cluster={cluster} /> : null}
      {ready ? (
        <div className="grid min-h-0 min-w-0 flex-1 gap-4 overflow-hidden lg:grid-cols-[15.5rem_minmax(0,1fr)]">
          <aside className="hidden min-h-0 min-w-0 flex-col overflow-hidden rounded-xl border border-border/60 bg-card/55 lg:flex">
            <div className="min-h-0 flex-1 overflow-y-auto p-2">
              <ClusterResourceNav
                section={section}
                kind={kind}
                counts={navCounts}
                onSelect={selectSection}
              />
            </div>
          </aside>
          <div className="flex min-h-0 min-w-0 flex-col gap-3">
            <div className="flex shrink-0 flex-wrap items-center justify-between gap-2 rounded-xl border border-border/60 bg-card/60 px-3 py-2">
              <div className="flex min-w-0 items-center gap-2 lg:flex-1">
                <ClusterResourcePicker
                  section={section}
                  kind={kind}
                  onSelect={selectSection}
                  className="lg:hidden"
                />
                <div className="hidden min-w-0 items-center gap-2 lg:flex">
                  {SectionIcon ? (
                    <SectionIcon
                      aria-hidden="true"
                      className="size-4 shrink-0 text-primary-ink"
                    />
                  ) : null}
                  <h2 className="truncate font-heading text-sm font-medium">
                    {sectionLabel}
                  </h2>
                </div>
                <Badge
                  variant="outline"
                  className="shrink-0 font-normal text-muted-foreground"
                >
                  {scope === "cluster"
                    ? t("kube.scopeCluster")
                    : t("kube.scopeNamespaced")}
                </Badge>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                {scope === "namespaced" ? (
                  <SearchableSelect
                    className="w-44 shrink-0"
                    ariaLabel={t("kube.namespace")}
                    value={applied.namespace}
                    options={namespaceOptions}
                    placeholder={t("kube.allNamespaces")}
                    searchPlaceholder={t("kube.namespaceSearch")}
                    emptyText={t("kube.namespacesEmpty")}
                    disabled={namespacesLoading}
                    onChange={(value) => query.update({ namespace: value })}
                  />
                ) : null}
                {scope === "namespaced" && namespacesFailed ? (
                  <div className="flex items-center gap-1">
                    <span className="text-[11px] text-destructive">
                      {t("kube.namespacesFailed")}
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setNamespacesAttempt((value) => value + 1)}
                    >
                      {t("action.retry")}
                    </Button>
                  </div>
                ) : null}
                {supportsNameFilter ? (
                  <FilterToggle
                    open={filtersOpen}
                    count={appliedCount}
                    controls={FILTER_PANEL_ID}
                    onClick={() => setFiltersOpen((open) => !open)}
                  />
                ) : null}
                <Button
                  variant="outline"
                  size="icon"
                  aria-label={t("action.refresh")}
                  onClick={() => setRefreshKey((value) => value + 1)}
                >
                  <RefreshCw aria-hidden="true" />
                </Button>
              </div>
            </div>
            {supportsNameFilter && filterChips.length > 0 ? (
              <FilterChips items={filterChips} onClearAll={clearFilters} />
            ) : null}
            {supportsNameFilter ? (
              <form
                id={FILTER_PANEL_ID}
                className={`flex min-w-0 shrink-0 flex-nowrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3 ${filtersOpen ? "" : "hidden"}`}
                onSubmit={(event) => {
                  event.preventDefault()
                  applyFilters()
                }}
              >
                <div className="min-w-0 flex-1 basis-48">
                  <Input
                    id="cluster-resource-name-filter"
                    value={draftName}
                    placeholder={t("kube.namePlaceholder")}
                    aria-label={t("kube.name")}
                    onChange={(event) =>
                      setDraftState({
                        signature: appliedSignature,
                        name: event.target.value,
                      })
                    }
                  />
                </div>
                <div className="flex shrink-0 flex-nowrap items-center gap-2">
                  {filtersDirty ? (
                    <span
                      className="text-xs whitespace-nowrap text-amber-600 dark:text-amber-400"
                      role="status"
                    >
                      {t("filter.unapplied")}
                    </span>
                  ) : null}
                  {appliedCount > 0 ? (
                    <Button
                      type="button"
                      variant="ghost"
                      onClick={clearFilters}
                    >
                      {t("filter.clear")}
                    </Button>
                  ) : null}
                  <Button type="submit" variant="outline">
                    <Search aria-hidden="true" data-icon="inline-start" />
                    {t("filter.apply")}
                  </Button>
                </div>
              </form>
            ) : null}
            <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
              {renderResourcePanel()}
            </div>
          </div>
        </div>
      ) : null}
      <ClusterProbeDialog
        open={probeDialog !== null}
        name={cluster?.name ?? ""}
        probe={probeDialog?.probe ?? null}
        error={probeDialog?.error ?? ""}
        onOpenChange={(open) => {
          if (!open) {
            setProbeDialog(null)
          }
        }}
      />
    </section>
  )
}
