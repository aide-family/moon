import { Search } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useParams, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import {
  getRegistry,
  listRegistryProjects,
  listRegistryRepositories,
  testRegistry,
} from "@/api/registry"
import type {
  RegistryItem,
  RegistryProjectItem,
  RegistryRepositoryItem,
} from "@/api/registry/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import type { FilterChip } from "@/components/filter-chips"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Field, FieldLabel } from "@/components/ui/field"
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
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatUnixDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

const ALL_PROJECTS = "ALL"

/** 仓库详情：项目筛选、镜像仓库列表，进入某个镜像的 Tag 列表。 */
export function RegistryDetailPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const { uid = "" } = useParams()
  const [params, setParams] = useSearchParams()
  const project = params.get("project") ?? ""
  const name = params.get("name") ?? ""
  const page = Math.max(1, Number(params.get("page")) || 1)
  const [registry, setRegistry] = useState<RegistryItem | null>(null)
  const [projects, setProjects] = useState<RegistryProjectItem[]>([])
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [testing, setTesting] = useState(false)
  const { draft, update, dirty, reset } = useFilterDraft({ project, name })
  const { items, total, loading, failed, error, reload } = usePagedList<RegistryRepositoryItem>({
    page,
    pageSize: 20,
    filterKey: `${uid}|${project}|${name}`,
    fetch: (pageNo, pageSize) =>
      listRegistryRepositories(uid, { project, name, page: pageNo, pageSize }),
  })

  useEffect(() => {
    if (!uid) return
    let active = true
    async function load() {
      try {
        const [registryResult, projectResult] = await Promise.all([
          getRegistry(uid),
          listRegistryProjects(uid),
        ])
        if (!active) return
        setRegistry(registryResult.registry)
        setProjects(projectResult.items)
      } catch (err) {
        if (err instanceof ApiError && (err.status === 403 || err.status === 404)) return
        toast.error(formErrorMessage(err, t("state.failed")))
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [uid, t])

  if (!session.loading && !session.can("registry:read")) {
    return <Navigate to="/forbidden" replace />
  }

  const chips: FilterChip[] = []
  if (project) {
    chips.push({
      key: "project",
      label: `${t("registry.projects")}: ${project}`,
      onRemove: () => updateParams({ project: "", page: 1 }),
    })
  }
  if (name) {
    chips.push({
      key: "name",
      label: `${t("registry.repository")}: ${name}`,
      onRemove: () => updateParams({ name: "", page: 1 }),
    })
  }

  function updateParams(next: { project?: string; name?: string; page?: number }) {
    const query = new URLSearchParams(params)
    if (next.project !== undefined) {
      if (next.project) query.set("project", next.project)
      else query.delete("project")
      query.set("page", "1")
    }
    if (next.name !== undefined) {
      if (next.name) query.set("name", next.name)
      else query.delete("name")
      query.set("page", "1")
    }
    if (next.page !== undefined) query.set("page", String(next.page))
    setParams(query)
  }

  async function handleTest() {
    if (!registry) return
    setTesting(true)
    try {
      const result = await testRegistry(registry.uid)
      toast.success(t("registry.testSuccess"), {
        description: `${result.probe.endpoint} · ${result.probe.repositoryCount} repos`,
      })
      setRegistry(result.registry)
    } catch (err) {
      toast.error(formErrorMessage(err, t("registry.testFailed")))
    } finally {
      setTesting(false)
    }
  }

  function applyFilters() {
    updateParams({ project: draft.project, name: draft.name.trim() })
    setFiltersOpen(false)
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={registry?.name ?? `/${uid}`}
        description={
          registry
            ? `${registry.endpoint} · ${registry.driver}${registry.username ? ` · ${registry.username}` : ""}`
            : t("state.loading")
        }
      >
        <Button variant="outline" onClick={() => navigate("/registries")}>
          {t("action.goBack")}
        </Button>
        <Button variant="outline" disabled={testing} onClick={() => void handleTest()}>
          {testing ? t("registry.testing") : t("registry.testConnection")}
        </Button>
        <Button variant="outline" onClick={() => navigate("/registries/syncs")}>
          {t("registry.syncTasks")}
        </Button>
      </PageHeader>
      {registry ? (
        <div className="flex flex-wrap items-center gap-3 rounded-xl border border-border/60 bg-card px-4 py-3 text-sm">
          <Badge variant="outline">{registry.driver}</Badge>
          <span className="font-mono text-xs break-all">{registry.endpoint}</span>
          {registry.lastError ? (
            <Badge variant="destructive">{t("registry.unavailable")}</Badge>
          ) : (
            <Badge variant="secondary">{t("registry.available")}</Badge>
          )}
          <span className="text-muted-foreground">
            {t("registry.lastChecked")}:{" "}
            {formatUnixDateTime(registry.lastCheckedAt) || t("registry.neverChecked")}
          </span>
          {registry.remark ? <span className="text-muted-foreground">{registry.remark}</span> : null}
        </div>
      ) : null}
      <ListFilter
        id="registry-detail-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={chips}
        onClearAll={() => {
          updateParams({ project: "", name: "", page: 1 })
          reset()
        }}
        onClear={() => {
          updateParams({ project: "", name: "", page: 1 })
          reset()
        }}
        appliedCount={chips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="w-56 shrink-0">
          <FieldLabel htmlFor="registry-project-filter" className="text-xs text-muted-foreground">
            {t("registry.projects")}
          </FieldLabel>
          <Select
            value={draft.project || ALL_PROJECTS}
            onValueChange={(value) =>
              update("project", value === ALL_PROJECTS ? "" : value)
            }
          >
            <SelectTrigger id="registry-project-filter" className="w-full">
              <SelectValue>{draft.project || t("registry.projectPlaceholder")}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_PROJECTS}>{t("registry.projectPlaceholder")}</SelectItem>
              {projects.map((item) => (
                <SelectItem key={item.name} value={item.name}>
                  {item.name} ({item.repositoryCount})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>
        <Field className="min-w-0 w-full max-w-[500px] flex-1 basis-56">
          <FieldLabel htmlFor="registry-repo-filter" className="text-xs text-muted-foreground">
            {t("registry.repository")}
          </FieldLabel>
          <div className="relative">
            <Search aria-hidden="true" className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="registry-repo-filter"
              className="pl-9"
              value={draft.name}
              placeholder={t("registry.searchRepository")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
      </ListFilter>
      <ListFrame
        total={total}
        page={page}
        pageSize={20}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => updateParams({ page: Math.max(1, page - 1) })}
        onNext={() => updateParams({ page: page + 1 })}
      >
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{error || t("state.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>{t("action.retry")}</Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("registry.emptyHint")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("registry.repository")}</TableHead>
                <TableHead>{t("registry.projects")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("registry.tagCount")}</TableHead>
                <TableHead className={`${TABLE_NOWRAP} text-right`}>{t("action.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.name}>
                  <TableCell className="font-mono text-xs break-all">{item.name}</TableCell>
                  <TableCell>{item.project || "—"}</TableCell>
                  <TableCell className={`font-mono text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {item.tagCount > 0 ? item.tagCount : "—"}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="link"
                      size="sm"
                      onClick={() =>
                        navigate(
                          `/registries/${uid}/repository?project=${encodeURIComponent(item.project)}&repository=${encodeURIComponent(item.name)}`
                        )
                      }
                    >
                      {t("registry.viewTags")}
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
    </section>
  )
}
