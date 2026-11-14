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
  InspectionTemplate,
} from "@/api/inspection/types"
import { listNotificationGroups } from "@/api/notification-group"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { FilterChips, type FilterChip } from "@/components/filter-chips"
import { FilterToggle } from "@/components/filter-toggle"
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

export function InspectionPlanPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const session = useSession()
  const query = useListQuery()
  const [draftName, setDraftName] = useState(query.name)
  const [draftStatus, setDraftStatus] = useState(
    query.status === "ACTIVE" || query.status === "PAUSED" ? query.status : ""
  )
  const [filtersOpen, setFiltersOpen] = useState(false)
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
  const status = query.status === "ACTIVE" || query.status === "PAUSED" ? query.status : ""

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
      onRemove: () => {
        setDraftName("")
        query.update({ name: "" })
      },
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
      onRemove: () => {
        setDraftStatus("")
        query.update({ status: "" })
      },
    })
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("inspection.plans")}
        description={t("inspection.plansDesc")}
      >
        <div className="flex flex-wrap items-center justify-end gap-2">
          <FilterToggle
            open={filtersOpen}
            count={filterChips.length}
            controls="inspection-plan-filters"
            onClick={() => setFiltersOpen((open) => !open)}
          />
          {canWrite ? (
            <Button onClick={openCreate} disabled={!templatesLoaded}>
              {t("inspection.newPlan")}
            </Button>
          ) : null}
        </div>
      </PageHeader>
      <div
        className={`min-w-0 flex-col gap-2 ${filtersOpen || filterChips.length > 0 ? "flex" : "hidden"}`}
      >
        {filterChips.length > 0 ? (
          <FilterChips
            items={filterChips}
            onClearAll={() => {
              setDraftName("")
              setDraftStatus("")
              query.update({ name: "", status: "" })
            }}
          />
        ) : null}
        <form
          id="inspection-plan-filters"
          className={`flex min-w-0 flex-wrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3 sm:flex-nowrap sm:p-4 ${filtersOpen ? "" : "hidden"}`}
          onSubmit={(event) => {
            event.preventDefault()
            query.update({ name: draftName.trim(), status: draftStatus })
          }}
        >
          <div className="min-w-0 flex-1 basis-48">
            <Input
              value={draftName}
              placeholder={t("inspection.planName")}
              aria-label={t("inspection.planName")}
              onChange={(event) => setDraftName(event.target.value)}
            />
          </div>
          <div className="w-full shrink-0 sm:w-40">
            <Select
              value={draftStatus || ALL_STATUS}
              onValueChange={(value) =>
                setDraftStatus(value === ALL_STATUS ? "" : value)
              }
            >
              <SelectTrigger className="w-full" aria-label={t("inspection.runStatus")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL_STATUS}>{t("inspection.all")}</SelectItem>
                <SelectItem value="ACTIVE">{t("inspection.statusActive")}</SelectItem>
                <SelectItem value="PAUSED">{t("inspection.statusPaused")}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button type="submit" variant="outline">
            <Search data-icon="inline-start" />
            {t("filter.apply")}
          </Button>
        </form>
      </div>
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
