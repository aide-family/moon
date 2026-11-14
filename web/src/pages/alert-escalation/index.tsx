import { Info, Search } from "lucide-react"
import { useEffect, useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { toast } from "sonner"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import {
  deleteAlertEscalationPolicy,
  listAlertEscalationPolicies,
  updateAlertEscalationPolicyStatus,
} from "@/api/alert-escalation"
import type { AlertEscalationPolicyItem } from "@/api/alert-escalation/types"
import { listAlertRules } from "@/api/alert-rule"
import type { AlertRuleItem } from "@/api/alert-rule/types"
import { selectDatasources } from "@/api/datasource"
import type { SelectItem as DatasourceOption } from "@/api/datasource/types"
import { ApiError } from "@/api/http"
import { selectNotificationGroups } from "@/api/notification-group"
import type { SelectItem as GroupOption } from "@/api/notification-group/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet, type DetailSection } from "@/components/detail-sheet"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { LabelTags } from "@/components/label-tags"
import { LevelBadge, StatusBadge } from "@/components/status-badge"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Field, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectGroup,
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
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { parseLabelPairs } from "@/lib/labels"
import { loadAllPagedItems, loadAllSelectItems } from "@/lib/select-pages"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { AlertEscalationPolicyDialog } from "./policy-dialog"

function toOptions<T extends { uid: string; name: string }>(items: T[]) {
  return items.map((item) => ({ value: item.uid, label: item.name }))
}

/** 升级策略列表：匹配条件、有序步骤与启停删除；新建/编辑在对话框内完成。 */
export function AlertEscalationPolicyPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({
    name: query.name,
    status: query.status,
  })
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [datasources, setDatasources] = useState<DatasourceOption[]>([])
  const [rules, setRules] = useState<AlertRuleItem[]>([])
  const [groups, setGroups] = useState<GroupOption[]>([])
  const [open, setOpen] = useState(false)
  const [editorSession, setEditorSession] = useState(0)
  const [editing, setEditing] = useState<AlertEscalationPolicyItem | null>(null)
  const [detail, setDetail] = useState<AlertEscalationPolicyItem | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [confirm, setConfirm] = useState<AlertEscalationPolicyItem | null>(null)
  const [pending, setPending] = useState(false)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("alert_escalation:write")
  const canDelete = session.can("alert_escalation:delete")
  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<AlertEscalationPolicyItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${query.status}`,
    fetch: (page, pageSize) =>
      listAlertEscalationPolicies({
        page,
        pageSize,
        name: query.name,
        status: query.status,
      }),
  })

  // 匹配条件与步骤通知组的选项互相独立，任一失败不影响其他几项。
  useEffect(() => {
    let active = true
    void Promise.allSettled([
      loadAllPagedItems((page, pageSize) =>
        listAlertDictionaries({ page, pageSize, type: "ALERT" })
      ),
      loadAllSelectItems(
        (lastUid) => selectDatasources({ limit: 100, lastUid }),
        100
      ),
      loadAllPagedItems((page, pageSize) => listAlertRules({ page, pageSize })),
      loadAllSelectItems(
        (lastUid) =>
          selectNotificationGroups({ limit: 100, lastUid, activeOnly: true }),
        100
      ),
    ]).then(([levelResult, datasourceResult, ruleResult, groupResult]) => {
      if (!active) {
        return
      }
      setLevels(levelResult.status === "fulfilled" ? levelResult.value : [])
      setDatasources(
        datasourceResult.status === "fulfilled" ? datasourceResult.value : []
      )
      setRules(ruleResult.status === "fulfilled" ? ruleResult.value : [])
      setGroups(groupResult.status === "fulfilled" ? groupResult.value : [])
    })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  // 删除或筛选收缩后页码可能超过最后一页：回到最后一页，避免空白列表与失效分页。
  useEffect(() => {
    if (loading || failed || query.page <= 1) {
      return
    }
    const lastPage = Math.max(1, Math.ceil(total / query.pageSize))
    if (
      lastPage !== query.page &&
      (items.length === 0 || query.page > lastPage)
    ) {
      query.update({ page: lastPage })
    }
    // query.update 每次渲染都是新函数，这里用页码与总数做真正的依赖。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, failed, items.length, total, query.page, query.pageSize])

  if (!session.loading && !session.can("alert_escalation:read")) {
    return <Navigate to="/forbidden" replace />
  }

  function applyFilters() {
    query.update({ name: draft.name.trim(), status: draft.status })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "", status: "" })
  }

  function openCreate() {
    setEditing(null)
    setEditorSession((value) => value + 1)
    setOpen(true)
  }

  function openEdit(item: AlertEscalationPolicyItem) {
    setEditing(item)
    setEditorSession((value) => value + 1)
    setOpen(true)
  }

  function openDetail(item: AlertEscalationPolicyItem) {
    setDetail(item)
    setDetailOpen(true)
  }

  async function toggleStatus(item: AlertEscalationPolicyItem) {
    const next = item.status === "ACTIVE" ? "DISABLED" : "ACTIVE"
    setPending(true)
    try {
      await updateAlertEscalationPolicyStatus(item.uid, next)
      reload()
    } catch (err) {
      // 403 由 http 层跳 /forbidden，404/5xx 已由 http 层全局 toast；这里补 400/409 与网络错误。
      if (
        err instanceof ApiError &&
        (err.status === 403 || err.status === 404 || err.status >= 500)
      ) {
        return
      }
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function runDelete() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await deleteAlertEscalationPolicy(confirm.uid)
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AlertEscalationPolicyItem): RowAction[] {
    const list: RowAction[] = []
    if (canWrite) {
      list.push({
        key: "edit",
        label: t("action.edit"),
        onSelect: () => openEdit(item),
      })
      list.push(
        item.status === "ACTIVE"
          ? {
              key: "disable",
              label: t("action.disable"),
              onSelect: () => void toggleStatus(item),
            }
          : {
              key: "enable",
              label: t("action.enable"),
              onSelect: () => void toggleStatus(item),
            }
      )
    }
    if (canDelete) {
      list.push({
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirm(item)
        },
      })
    }
    return list
  }

  function matcherSummary(item: AlertEscalationPolicyItem): string {
    const labelCount = parseLabelPairs(item.labels).filter(
      (pair) => pair.key.trim() !== ""
    ).length
    const levelCount = item.levelUids?.length ?? 0
    const datasourceCount = item.datasourceUids?.length ?? 0
    const ruleCount = item.ruleUids?.length ?? 0
    if (
      labelCount === 0 &&
      levelCount === 0 &&
      datasourceCount === 0 &&
      ruleCount === 0
    ) {
      return t("escalation.allEvents")
    }
    return t("escalation.matchSummary", {
      labels: labelCount,
      levels: levelCount,
      datasources: datasourceCount,
      rules: ruleCount,
    })
  }

  function levelTags(uids: string[]): ReactNode {
    if (uids.length === 0) {
      return null
    }
    return (
      <span className="flex flex-wrap gap-1.5">
        {uids.map((uid) => {
          const level = levels.find((entry) => entry.uid === uid)
          return (
            <LevelBadge
              key={uid}
              name={level?.name ?? uid}
              color={level?.color}
            />
          )
        })}
      </span>
    )
  }

  function nameTags(
    uids: string[],
    dictionary: Array<{ uid: string; name: string }>
  ): ReactNode {
    if (uids.length === 0) {
      return null
    }
    return (
      <span className="flex flex-wrap gap-1.5">
        {uids.map((uid) => (
          <Badge
            key={uid}
            variant="outline"
            className="h-auto max-w-full rounded-md px-2 py-0.5 font-normal break-all whitespace-normal"
          >
            {dictionary.find((entry) => entry.uid === uid)?.name ?? uid}
          </Badge>
        ))}
      </span>
    )
  }

  function groupNames(uids: string[]): string {
    if (uids.length === 0) {
      return "—"
    }
    return uids
      .map((uid) => groups.find((group) => group.uid === uid)?.name ?? uid)
      .join(t("escalation.groupSeparator"))
  }

  function stepSections(item: AlertEscalationPolicyItem): DetailSection[] {
    return [
      {
        title: t("escalation.matchConditions"),
        description: matcherSummary(item),
        layout: "inline",
        fields: [
          {
            label: t("escalation.labels"),
            value: <LabelTags labels={item.labels} />,
          },
          {
            label: t("escalation.levels"),
            value: levelTags(item.levelUids ?? []),
          },
          {
            label: t("escalation.datasources"),
            value: nameTags(item.datasourceUids ?? [], datasources),
          },
          {
            label: t("escalation.rules"),
            value: nameTags(item.ruleUids ?? [], rules),
          },
        ],
      },
      {
        title: t("escalation.steps"),
        description: t("escalation.stepCount", {
          count: item.steps?.length ?? 0,
        }),
        content:
          (item.steps?.length ?? 0) === 0 ? (
            <p className="text-sm text-muted-foreground">—</p>
          ) : (
            <ol className="grid gap-2">
              {item.steps.map((step) => (
                <li
                  key={step.uid}
                  className="grid gap-1 rounded-lg border border-border/70 bg-muted/30 px-3 py-2"
                >
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className="text-sm font-medium">
                      {t("escalation.stepIndex", {
                        index: step.stepOrder + 1,
                      })}
                    </span>
                    <span className="font-mono text-xs text-muted-foreground tabular-nums">
                      {t("escalation.delayValue", {
                        minutes: step.delayMinutes,
                      })}
                    </span>
                  </div>
                  <p className="text-xs break-words text-muted-foreground">
                    {groupNames(step.notificationGroupUids ?? [])}
                  </p>
                </li>
              ))}
            </ol>
          ),
      },
      {
        title: t("escalation.basicInfo"),
        layout: "inline",
        fields: [
          {
            label: t("escalation.status"),
            value: <StatusBadge value={item.status} />,
          },
          {
            label: t("escalation.createdBy"),
            value: item.createdBy || "—",
          },
          {
            label: t("escalation.updatedAt"),
            value: formatDateTime(item.updatedAt) || "—",
          },
          {
            label: t("escalation.createdAt"),
            value: formatDateTime(item.createdAt) || "—",
          },
        ],
      },
      ...(item.description
        ? [
            {
              title: t("escalation.descriptionLabel"),
              content: (
                <p className="text-sm leading-relaxed break-words whitespace-pre-wrap">
                  {item.description}
                </p>
              ),
            },
          ]
        : []),
    ]
  }

  const filterChips: FilterChip[] = []
  if (query.name) {
    filterChips.push({
      key: "name",
      label: `${t("escalation.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (query.status) {
    filterChips.push({
      key: "status",
      label: `${t("escalation.statusFilter")}: ${
        query.status === "ACTIVE"
          ? t("escalation.filterActive")
          : t("escalation.filterDisabled")
      }`,
      onRemove: () => query.update({ status: "" }),
    })
  }

  return (
    <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("escalation.title")}
        description={t("escalation.description")}
      >
        {canWrite ? (
          <Button onClick={openCreate}>{t("escalation.createAction")}</Button>
        ) : null}
      </PageHeader>
      <div
        role="note"
        className="flex items-start gap-2 rounded-lg border border-info-border bg-info-soft px-3 py-2 text-sm text-info"
      >
        <Info aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
        <p className="min-w-0 break-words">{t("escalation.skipHint")}</p>
      </div>
      <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-4">
        <ListFilter
          id="escalation-filters"
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
              htmlFor="escalation-name"
              className="text-xs text-muted-foreground"
            >
              {t("escalation.name")}
            </FieldLabel>
            <div className="relative">
              <Search
                aria-hidden="true"
                className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
              />
              <Input
                id="escalation-name"
                className="pl-9"
                value={draft.name}
                placeholder={t("escalation.searchPlaceholder")}
                onChange={(event) => update("name", event.target.value)}
              />
            </div>
          </Field>
          <Field className="w-40 shrink-0">
            <FieldLabel
              htmlFor="escalation-status"
              className="text-xs text-muted-foreground"
            >
              {t("escalation.statusFilter")}
            </FieldLabel>
            <Select
              value={draft.status || "ALL"}
              onValueChange={(value) =>
                update("status", value === "ALL" ? "" : value)
              }
            >
              <SelectTrigger id="escalation-status" className="w-full">
                <SelectValue placeholder={t("escalation.allStatuses")} />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="ALL">
                    {t("escalation.allStatuses")}
                  </SelectItem>
                  <SelectItem value="ACTIVE">
                    <span className="flex items-center gap-2">
                      <EnumIcon value="ACTIVE" />
                      <span>{t("escalation.filterActive")}</span>
                    </span>
                  </SelectItem>
                  <SelectItem value="DISABLED">
                    <span className="flex items-center gap-2">
                      <EnumIcon value="DISABLED" />
                      <span>{t("escalation.filterDisabled")}</span>
                    </span>
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </ListFilter>
        <ListFrame
          total={total}
          page={query.page}
          pageSize={query.pageSize}
          showPager={!loading && !failed && total > 0}
          onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
          onNext={() => query.update({ page: query.page + 1 })}
        >
          {loading ? (
            <p className="text-sm text-muted-foreground">
              {t("state.loading")}
            </p>
          ) : null}
          {!loading && failed ? (
            <div className="flex items-center gap-2">
              <p className="text-sm text-destructive">
                {loadError || t("state.failed")}
              </p>
              <Button variant="outline" onClick={() => reload()}>
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {!loading && !failed && items.length === 0 ? (
            <div className="grid gap-1">
              <p className="text-sm text-muted-foreground">
                {t("escalation.empty")}
              </p>
              <p className="text-sm text-muted-foreground">
                {t("escalation.emptyHint")}
              </p>
            </div>
          ) : null}
          {!loading && !failed && items.length > 0 ? (
            <Table className={TABLE_MIN_WIDTH.default}>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("escalation.name")}</TableHead>
                  <TableHead>{t("escalation.matchConditions")}</TableHead>
                  <TableHead
                    className={`${TABLE_COLUMN_WIDTH.level} ${TABLE_NOWRAP}`}
                  >
                    {t("escalation.stepCountLabel")}
                  </TableHead>
                  <TableStatusHead>{t("escalation.status")}</TableStatusHead>
                  <TableHead
                    className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                  >
                    {t("escalation.updatedAt")}
                  </TableHead>
                  <TableActionsHead>{t("action.actions")}</TableActionsHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell className="font-medium break-words">
                      {item.name}
                    </TableCell>
                    <TableCell className="max-w-md text-sm text-muted-foreground">
                      <span className="break-words">
                        {matcherSummary(item)}
                      </span>
                    </TableCell>
                    <TableCell
                      className={`font-mono text-xs tabular-nums ${TABLE_COLUMN_WIDTH.level} ${TABLE_NOWRAP}`}
                    >
                      {item.steps?.length ?? 0}
                    </TableCell>
                    <TableCell className={TABLE_NOWRAP}>
                      <StatusBadge value={item.status} />
                    </TableCell>
                    <TableCell
                      className={`font-mono text-xs tabular-nums ${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                    >
                      {formatDateTime(item.updatedAt) || "—"}
                    </TableCell>
                    <TableCell>
                      <RowActions
                        nowrap
                        detailLabel={t("action.detail")}
                        moreLabel={t("action.more")}
                        onDetail={() => openDetail(item)}
                        actions={actions(item)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
        </ListFrame>
      </div>
      <AlertEscalationPolicyDialog
        key={`${editing?.uid ?? "create"}-${editorSession}`}
        open={open}
        policy={editing}
        levels={toOptions(levels)}
        datasources={toOptions(datasources)}
        rules={toOptions(rules)}
        groups={toOptions(groups)}
        onOpenChange={setOpen}
        onSaved={() => {
          setOpen(false)
          reload()
        }}
      />
      <DetailSheet
        open={detailOpen}
        size="40vw"
        title={detail?.name ?? ""}
        description={detail ? <StatusBadge value={detail.status} /> : undefined}
        sections={detail ? stepSections(detail) : []}
        onOpenChange={setDetailOpen}
      />
      <ConfirmAction
        open={confirm !== null}
        title={t("escalation.deleteTitle")}
        description={t("escalation.deleteDescription")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runDelete()}
        onOpenChange={(value) => {
          if (!value) {
            setConfirm(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
