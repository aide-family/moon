import { Search } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"

import {
  deleteInspectionPlan,
  listInspectionPlans,
  listInspectionTemplates,
  runInspectionPlan,
  updateInspectionPlanStatus,
} from "@/api/inspection"
import type {
  InspectionPlan,
  InspectionPlanStatus,
  InspectionTemplate,
} from "@/api/inspection/types"
import { listNotificationGroups } from "@/api/notification-group"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import type { MultiSelectOption } from "@/components/multi-select"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
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
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { InspectionPlanDetail } from "./inspection-plan-detail"
import { InspectionPlanDialog } from "./inspection-plan-dialog"

const ALL_STATUS = "ALL"

// 计划只有启用/暂停两种状态，URL 里的其他值按不筛选处理。
function normalizePlanStatus(value: string): "" | InspectionPlanStatus {
  return value === "ACTIVE" || value === "PAUSED" ? value : ""
}

export function InspectionPlanPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const session = useSession()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({
    name: query.name,
    status: normalizePlanStatus(query.status),
  })
  const [templates, setTemplates] = useState<InspectionTemplate[]>([])
  const [templatesLoaded, setTemplatesLoaded] = useState(false)
  const [groups, setGroups] = useState<MultiSelectOption[]>([])
  const [editorOpen, setEditorOpen] = useState(false)
  const [editorSession, setEditorSession] = useState(0)
  const [editing, setEditing] = useState<InspectionPlan | null>(null)
  const [confirm, setConfirm] = useState<InspectionPlan | null>(null)
  const [confirmPending, setConfirmPending] = useState(false)
  const [confirmError, setConfirmError] = useState("")
  const [runningUid, setRunningUid] = useState("")
  const [detailPlan, setDetailPlan] = useState<InspectionPlan | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [noTemplateOpen, setNoTemplateOpen] = useState(false)

  const canWrite = session.can("inspection:write")
  const canDelete = session.can("inspection:delete")
  const canRun = session.can("inspection:run")
  const status = normalizePlanStatus(query.status)

  const { items, total, loading, failed, error, reload } =
    usePagedList<InspectionPlan>({
      page: query.page,
      pageSize: query.pageSize,
      filterKey: `${session.namespaceUid}|${query.name}|${status}`,
      fetch: (page, pageSize) =>
        listInspectionPlans({ page, pageSize, name: query.name, status }),
    })

  useEffect(() => {
    let active = true
    listInspectionTemplates({ page: 1, pageSize: 100 })
      .then((result) => {
        if (active) {
          setTemplates(result.items)
          setTemplatesLoaded(true)
        }
      })
      .catch(() => {
        if (active) {
          setTemplates([])
          setTemplatesLoaded(true)
        }
      })
    listNotificationGroups({ page: 1, pageSize: 100 })
      .then((result) => {
        if (active) {
          setGroups(result.items.map((item) => ({ value: item.uid, label: item.name })))
        }
      })
      .catch(() => {
        if (active) {
          setGroups([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  const templateName = (uid: string) =>
    templates.find((template) => template.uid === uid)?.name ?? "—"

  function openCreate() {
    if (!templatesLoaded) {
      return
    }
    if (templates.length === 0) {
      setNoTemplateOpen(true)
      return
    }
    setEditing(null)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }

  function openEdit(plan: InspectionPlan) {
    setEditing(plan)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }

  async function runPlan(plan: InspectionPlan) {
    if (runningUid) {
      return
    }
    setRunningUid(plan.uid)
    try {
      await runInspectionPlan(plan.uid)
      toast.success(t("inspection.started"))
      reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setRunningUid("")
    }
  }

  async function toggleStatus(plan: InspectionPlan) {
    try {
      await updateInspectionPlanStatus(
        plan.uid,
        plan.status === "ACTIVE" ? "PAUSED" : "ACTIVE"
      )
      reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  async function confirmDelete() {
    if (!confirm) {
      return
    }
    setConfirmPending(true)
    setConfirmError("")
    try {
      await deleteInspectionPlan(confirm.uid)
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setConfirmPending(false)
    }
  }

  function rowActions(plan: InspectionPlan): RowAction[] {
    const actions: RowAction[] = []
    if (canWrite) {
      actions.push({
        key: "edit",
        label: t("action.edit"),
        onSelect: () => openEdit(plan),
      })
    }
    if (canRun && plan.status === "ACTIVE") {
      actions.push({
        key: "run",
        label: t("inspection.run"),
        onSelect: () => void runPlan(plan),
      })
    }
    if (canWrite) {
      actions.push({
        key: "toggle",
        label: plan.status === "ACTIVE" ? t("inspection.pause") : t("inspection.resume"),
        onSelect: () => void toggleStatus(plan),
      })
    }
    if (canDelete) {
      actions.push({
        key: "delete",
        label: t("inspection.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirm(plan)
        },
      })
    }
    return actions
  }

  const filterChips: FilterChip[] = []
  if (query.name) {
    filterChips.push({
      key: "name",
      label: `${t("inspection.planName")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (status) {
    filterChips.push({
      key: "status",
      label: `${t("inspection.runStatus")}: ${
        status === "ACTIVE"
          ? t("inspection.statusActive")
          : t("inspection.statusPaused")
      }`,
      onRemove: () => query.update({ status: "" }),
    })
  }

  function applyFilters() {
    query.update({ name: draft.name.trim(), status: draft.status })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "", status: "" })
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("inspection.plans")}
        description={t("inspection.plansDesc")}
      >
        {canWrite ? (
          <Button onClick={openCreate} disabled={!templatesLoaded}>
            {t("inspection.newPlan")}
          </Button>
        ) : null}
      </PageHeader>
      <ListFilter
        id="inspection-plan-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={filterChips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={filterChips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="min-w-0 w-full max-w-[500px] flex-1 basis-56">
          <FieldLabel
            htmlFor="inspection-plan-name"
            className="text-xs text-muted-foreground"
          >
            {t("inspection.planName")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="inspection-plan-name"
              className="pl-9"
              value={draft.name}
              placeholder={t("inspection.planName")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
        <Field className="w-40 shrink-0">
          <FieldLabel
            htmlFor="inspection-plan-status"
            className="text-xs text-muted-foreground"
          >
            {t("inspection.runStatus")}
          </FieldLabel>
          <Select
            value={draft.status || ALL_STATUS}
            onValueChange={(value) =>
              update("status", normalizePlanStatus(value))
            }
          >
            <SelectTrigger id="inspection-plan-status" className="w-full">
              <SelectValue placeholder={t("inspection.all")}>
                <span className="flex items-center gap-2">
                  <EnumIcon value={draft.status} />
                  <span>
                    {draft.status === "ACTIVE"
                      ? t("inspection.statusActive")
                      : draft.status === "PAUSED"
                        ? t("inspection.statusPaused")
                        : t("inspection.all")}
                  </span>
                </span>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_STATUS}>{t("inspection.all")}</SelectItem>
              <SelectItem value="ACTIVE">
                <span className="flex items-center gap-2">
                  <EnumIcon value="ACTIVE" />
                  <span>{t("inspection.statusActive")}</span>
                </span>
              </SelectItem>
              <SelectItem value="PAUSED">
                <span className="flex items-center gap-2">
                  <EnumIcon value="PAUSED" />
                  <span>{t("inspection.statusPaused")}</span>
                </span>
              </SelectItem>
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
            <p className="text-sm text-destructive">{error || t("inspection.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>
              {t("inspection.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <div className="grid gap-2">
            <p className="text-sm text-muted-foreground">{t("inspection.empty")}</p>
            <p className="text-sm text-muted-foreground">{t("inspection.emptyDesc")}</p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("inspection.planName")}</TableHead>
                <TableHead>{t("inspection.cron")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("inspection.items")}</TableHead>
                <TableHead>{t("inspection.template")}</TableHead>
                <TableStatusHead>{t("inspection.runStatus")}</TableStatusHead>
                <TableHead className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}>
                  {t("inspection.nextRunLabel")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((plan) => (
                <TableRow key={plan.uid}>
                  <TableCell>
                    <div className="grid gap-1">
                      <span className="font-medium break-words">{plan.name}</span>
                      <span className="text-xs text-muted-foreground">
                        {plan.description || "—"}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="grid gap-1 font-mono text-xs">
                      <span>{plan.cron}</span>
                      <span className="text-muted-foreground">{plan.timezone}</span>
                    </div>
                  </TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>
                    {plan.items.length}
                  </TableCell>
                  <TableCell className="break-words text-sm">
                    {templateName(plan.templateUid)}
                    <div className="text-xs text-muted-foreground">
                      {(plan.outputFormats ?? []).join(" · ")}
                    </div>
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <StatusBadge
                      value={plan.status}
                      label={
                        plan.status === "ACTIVE"
                          ? t("inspection.statusActive")
                          : t("inspection.statusPaused")
                      }
                      tone={plan.status === "ACTIVE" ? "success" : "neutral"}
                    />
                  </TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>
                    {plan.status === "ACTIVE" ? formatDateTime(plan.nextRunAt) || "—" : "—"}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("inspection.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => {
                        setDetailPlan(plan)
                        setDetailOpen(true)
                      }}
                      actions={rowActions(plan)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <InspectionPlanDialog
        key={`${editing?.uid ?? "create"}-${editorSession}`}
        open={editorOpen}
        plan={editing}
        templates={templates}
        groups={groups}
        onOpenChange={setEditorOpen}
        onSaved={() => {
          setEditorOpen(false)
          toast.success(t("inspection.saved"))
          reload()
        }}
      />
      <InspectionPlanDetail
        open={detailOpen}
        plan={detailPlan}
        templateName={detailPlan ? templateName(detailPlan.templateUid) : "—"}
        groups={groups}
        onOpenChange={setDetailOpen}
      />
      <AlertDialog open={noTemplateOpen} onOpenChange={setNoTemplateOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("inspection.noTemplateTitle")}</AlertDialogTitle>
            <AlertDialogDescription>{t("inspection.noTemplateDesc")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("action.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                setNoTemplateOpen(false)
                navigate("/inspection-templates?create=1")
              }}
            >
              {t("inspection.noTemplateAction")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <ConfirmAction
        open={confirm !== null}
        title={t("inspection.delete")}
        description={t("inspection.deleteConfirm")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={confirmPending}
        error={confirmError}
        onConfirm={() => void confirmDelete()}
        onOpenChange={(open) => {
          if (!open) {
            setConfirm(null)
          }
        }}
      />
    </section>
  )
}
