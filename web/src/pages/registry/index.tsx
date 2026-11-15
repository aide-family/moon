import { Search } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { deleteRegistry, listRegistries, testRegistry } from "@/api/registry"
import type { RegistryDriver, RegistryItem } from "@/api/registry/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import type { FilterChip } from "@/components/filter-chips"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
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
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatUnixDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import {
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"
import { ServerErrorPage } from "@/pages/server-error"

import { RegistryDialog } from "./registry-dialog"

const ALL_DRIVERS = "ALL"
const DRIVER_VALUES = new Set(["HARBOR", "DOCKER_V2"])

function normalizeDriver(value: string): RegistryDriver | "" {
  return DRIVER_VALUES.has(value) ? (value as RegistryDriver) : ""
}

/** 镜像仓库连接列表：接入、测试、编辑、删除，并进入仓库详情。 */
export function RegistryPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [editorOpen, setEditorOpen] = useState(false)
  const [editorSession, setEditorSession] = useState(0)
  const [editing, setEditing] = useState<RegistryItem | null>(null)
  const [deleting, setDeleting] = useState<RegistryItem | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const [testingUid, setTestingUid] = useState("")
  const driver = normalizeDriver(query.type)
  const { draft, update, dirty, reset } = useFilterDraft({
    name: query.name,
    driver,
  })

  const { items, total, loading, failed, failedStatus, error, reload } =
    usePagedList<RegistryItem>({
      page: query.page,
      pageSize: query.pageSize,
      filterKey: `${session.namespaceUid}|${query.name}|${driver}`,
      fetch: (page, pageSize) =>
        listRegistries({ page, pageSize, name: query.name, driver }),
    })

  if (!session.loading && !session.can("registry:read")) {
    return <Navigate to="/forbidden" replace />
  }
  if (!loading && failed && failedStatus >= 500) {
    return (
      <ServerErrorPage
        fullPage={false}
        status={failedStatus === 503 ? 503 : 502}
        detail={error === t("error.badGatewayTitle") ? "" : error}
        onRetry={() => reload()}
      />
    )
  }

  const canWrite = session.can("registry:write")
  const canDelete = session.can("registry:delete")
  const chips: FilterChip[] = []
  if (query.name) {
    chips.push({
      key: "name",
      label: `${t("registry.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (driver) {
    chips.push({
      key: "driver",
      label: `${t("registry.driver")}: ${driver}`,
      onRemove: () => query.update({ type: "" }),
    })
  }

  function applyFilters() {
    query.update({ name: draft.name.trim(), type: draft.driver })
    setFiltersOpen(false)
  }
  function clearFilters() {
    query.update({ name: "", type: "" })
    reset()
  }
  function openCreate() {
    setEditing(null)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }
  function openEdit(item: RegistryItem) {
    setEditing(item)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }
  async function handleTest(item: RegistryItem) {
    setTestingUid(item.uid)
    try {
      const result = await testRegistry(item.uid)
      toast.success(t("registry.testSuccess"), {
        description: `${result.probe.endpoint} · ${result.probe.repositoryCount} repos`,
      })
      reload()
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) return
      toast.error(formErrorMessage(err, t("registry.testFailed")))
    } finally {
      setTestingUid("")
    }
  }
  async function handleDelete() {
    if (!deleting) return
    setDeletePending(true)
    setDeleteError("")
    try {
      await deleteRegistry(deleting.uid)
      setDeleting(null)
      toast.success(t("registry.deleted"))
      reload()
    } catch (err) {
      setDeleteError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeletePending(false)
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("registry.title")}
        description={t("registry.description")}
      >
        <Button variant="outline" onClick={() => navigate("/registries/syncs")}>
          {t("registry.syncTasks")}
        </Button>
        {canWrite ? <Button onClick={openCreate}>{t("registry.connect")}</Button> : null}
      </PageHeader>
      <ListFilter
        id="registry-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={chips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={chips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="min-w-0 w-full max-w-[500px] flex-1 basis-56">
          <FieldLabel htmlFor="registry-name-filter" className="text-xs text-muted-foreground">
            {t("registry.name")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="registry-name-filter"
              className="pl-9"
              value={draft.name}
              placeholder={t("registry.searchPlaceholder")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
        <Field className="w-44 shrink-0">
          <FieldLabel htmlFor="registry-driver-filter" className="text-xs text-muted-foreground">
            {t("registry.driver")}
          </FieldLabel>
          <Select
            value={draft.driver || ALL_DRIVERS}
            onValueChange={(value) =>
              update("driver", value === ALL_DRIVERS ? "" : (value as RegistryDriver))
            }
          >
            <SelectTrigger id="registry-driver-filter" className="w-full">
              <SelectValue>
                {draft.driver || t("registry.allDrivers")}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_DRIVERS}>{t("registry.allDrivers")}</SelectItem>
              <SelectItem value="HARBOR">HARBOR</SelectItem>
              <SelectItem value="DOCKER_V2">DOCKER_V2</SelectItem>
            </SelectContent>
          </Select>
        </Field>
      </ListFilter>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? (
          <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
        ) : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{error || t("state.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <div className="grid gap-2">
            <p className="text-sm text-muted-foreground">{t("registry.empty")}</p>
            <p className="text-sm text-muted-foreground">{t("registry.emptyHint")}</p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("registry.name")}</TableHead>
                <TableHead>{t("registry.driver")}</TableHead>
                <TableHead>{t("registry.endpoint")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("registry.lastChecked")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("registry.status")}</TableHead>
                <TableHead className="text-right">{t("action.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="font-medium break-words">
                    <button
                      type="button"
                      className="text-left underline-offset-4 hover:underline"
                      onClick={() => navigate(`/registries/${item.uid}`)}
                    >
                      {item.name}
                    </button>
                    {item.remark ? (
                      <p className="text-xs text-muted-foreground">{item.remark}</p>
                    ) : null}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{item.driver}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {item.endpoint}
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <div className="grid gap-1">
                      <span className="font-mono text-xs tabular-nums">
                        {formatUnixDateTime(item.lastCheckedAt) || t("registry.neverChecked")}
                      </span>
                      {item.lastError ? (
                        <span className="max-w-[18rem] truncate text-xs text-destructive" title={item.lastError}>
                          {item.lastError}
                        </span>
                      ) : null}
                    </div>
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    {item.lastError ? (
                      <Badge variant="destructive">{t("registry.unavailable")}</Badge>
                    ) : (
                      <Badge variant="secondary">{t("registry.available")}</Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => navigate(`/registries/${item.uid}`)}
                      actions={[
                        ...(canWrite
                          ? [
                              {
                                key: "test",
                                label: testingUid === item.uid ? t("registry.testing") : t("registry.testConnection"),
                                onSelect: () => void handleTest(item),
                              },
                              {
                                key: "edit",
                                label: t("action.edit"),
                                onSelect: () => openEdit(item),
                              },
                            ]
                          : []),
                        ...(canDelete
                          ? [
                              {
                                key: "delete",
                                label: t("action.delete"),
                                destructive: true,
                                onSelect: () => {
                                  setDeleteError("")
                                  setDeleting(item)
                                },
                              },
                            ]
                          : []),
                      ]}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <RegistryDialog
        key={`${editing?.uid ?? "create"}-${editorSession}`}
        open={editorOpen}
        registry={editing}
        onOpenChange={setEditorOpen}
        onSaved={() => {
          setEditorOpen(false)
          reload()
        }}
      />
      <ConfirmAction
        open={deleting !== null}
        title={t("registry.deleteTitle")}
        description={t("registry.deleteDescription")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={deletePending}
        error={deleteError}
        onOpenChange={(open) => {
          if (!open) setDeleting(null)
        }}
        onConfirm={() => void handleDelete()}
      />
    </section>
  )
}
