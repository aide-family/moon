import { Search } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import { listAgents } from "@/api/agent"
import type { AgentItem } from "@/api/agent/types"
import { listSSHCommandAudits } from "@/api/ssh-command"
import type { SSHCommandAuditItem } from "@/api/ssh-command/types"
import { CodeBlock } from "@/components/code-block"
import { DateRangePicker } from "@/components/date-range-picker"
import { DetailSheet } from "@/components/detail-sheet"
import { DetailCopyText } from "@/components/detail-values"
import { FilterChips, type FilterChip } from "@/components/filter-chips"
import { FilterToggle } from "@/components/filter-toggle"
import { ListFrame } from "@/components/list-frame"
import { RowActions } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead } from "@/components/table-heads"
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
import { usePagedList } from "@/hooks/use-paged-list"
import { dayStartSeconds, nextDayStartSeconds } from "@/lib/date-range"
import { formatDateTime } from "@/lib/date-time"
import { enumLabel } from "@/lib/enum-label"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

const FILTER_PANEL_ID = "ssh-audit-filters"
const AUDIT_STATUSES = ["DONE", "FAILED"] as const
const FILTER_KEYS = [
  "keyword",
  "status",
  "agentUid",
  "operatorUid",
  "from",
  "to",
] as const

interface AuditFilters {
  keyword: string
  status: string
  agentUid: string
  operatorUid: string
  from: string
  to: string
}

function boundedInt(
  value: string | null,
  fallback: number,
  min: number,
  max: number
): number {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed < min || parsed > max) {
    return fallback
  }
  return parsed
}

function operatorName(audit: SSHCommandAuditItem): string {
  return audit.operatorNickname || audit.operatorName || audit.operatorUid
}

/**
 * 统一执行审计：只读列表 + 详情抽屉。
 * 只查 kind=EXECUTE，筛选与分页条件都写进 URL，标签页可回放。
 */
export function SSHCommandAudits() {
  const { t } = useTranslation()
  const [params, setParams] = useSearchParams()
  const page = boundedInt(params.get("page"), 1, 1, 1_000_000)
  const pageSize = boundedInt(params.get("pageSize"), 20, 1, 100)
  const appliedFilters: AuditFilters = {
    keyword: params.get("keyword") ?? "",
    status: params.get("status") ?? "",
    agentUid: params.get("agentUid") ?? "",
    operatorUid: params.get("operatorUid") ?? "",
    from: params.get("from") ?? "",
    to: params.get("to") ?? "",
  }
  const appliedSignature = JSON.stringify(appliedFilters)
  const [draftState, setDraftState] = useState<{
    signature: string
    filters: AuditFilters
  } | null>(null)
  const filters =
    draftState?.signature === appliedSignature
      ? draftState.filters
      : appliedFilters
  const filtersDirty = JSON.stringify(filters) !== appliedSignature
  const appliedCount = Object.values(appliedFilters).filter(Boolean).length
  const [filtersOpen, setFiltersOpen] = useState(appliedCount === 0)
  const [agents, setAgents] = useState<AgentItem[]>([])
  const [detail, setDetail] = useState<SSHCommandAuditItem | null>(null)

  useEffect(() => {
    let active = true
    listAgents({ page: 1, pageSize: 100 })
      .then((result) => {
        if (active) {
          setAgents(result.items ?? [])
        }
      })
      .catch(() => {
        // 机器筛选是辅助条件，列表本身仍可正常加载。
        if (active) {
          setAgents([])
        }
      })
    return () => {
      active = false
    }
  }, [])

  const { items, total, loading, failed, error, reload } =
    usePagedList<SSHCommandAuditItem>({
      page,
      pageSize,
      filterKey: appliedSignature,
      fetch: (nextPage, nextPageSize) =>
        listSSHCommandAudits({
          page: nextPage,
          pageSize: nextPageSize,
          kind: "EXECUTE",
          keyword: appliedFilters.keyword || undefined,
          status: appliedFilters.status || undefined,
          agentUid: appliedFilters.agentUid || undefined,
          operatorUid: appliedFilters.operatorUid || undefined,
          startTime: dayStartSeconds(appliedFilters.from),
          endTime: nextDayStartSeconds(appliedFilters.to),
        }),
    })

  const agentNameByUid = useMemo(
    () => new Map(agents.map((agent) => [agent.uid, agent.name])),
    [agents]
  )

  function updateDraft(key: keyof AuditFilters, value: string) {
    setDraftState({
      signature: appliedSignature,
      filters: { ...filters, [key]: value },
    })
  }

  function patchFilters(patch: AuditFilters, keys: readonly string[]) {
    const next = new URLSearchParams(params)
    for (const key of keys) {
      const value = patch[key as keyof AuditFilters]
      if (value) {
        next.set(key, value)
      } else {
        next.delete(key)
      }
    }
    next.set("page", "1")
    setParams(next)
  }

  function applyFilters() {
    patchFilters(filters, FILTER_KEYS)
    setDraftState(null)
    setFiltersOpen(false)
  }

  function removeFilters(keys: string[]) {
    const next = new URLSearchParams(params)
    for (const key of keys) {
      next.delete(key)
    }
    next.set("page", "1")
    setParams(next)
    setDraftState(null)
  }

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(params)
    next.set("page", String(nextPage))
    setParams(next)
  }

  const machineLabel = (uid: string) => agentNameByUid.get(uid) || uid

  const filterChips: FilterChip[] = []
  if (appliedFilters.keyword) {
    filterChips.push({
      key: "keyword",
      label: `${t("ssh.auditKeyword")}: ${appliedFilters.keyword}`,
      onRemove: () => removeFilters(["keyword"]),
    })
  }
  if (appliedFilters.status) {
    filterChips.push({
      key: "status",
      label: `${t("ssh.status")}: ${enumLabel(t, appliedFilters.status)}`,
      onRemove: () => removeFilters(["status"]),
    })
  }
  if (appliedFilters.agentUid) {
    filterChips.push({
      key: "agentUid",
      label: `${t("ssh.machine")}: ${machineLabel(appliedFilters.agentUid)}`,
      onRemove: () => removeFilters(["agentUid"]),
    })
  }
  if (appliedFilters.operatorUid) {
    filterChips.push({
      key: "operatorUid",
      label: `${t("ssh.operator")}: ${appliedFilters.operatorUid}`,
      onRemove: () => removeFilters(["operatorUid"]),
    })
  }
  if (appliedFilters.from || appliedFilters.to) {
    filterChips.push({
      key: "range",
      label: `${t("filter.timeRange")}: ${appliedFilters.from || "…"} → ${appliedFilters.to || "…"}`,
      onRemove: () => removeFilters(["from", "to"]),
    })
  }

  const detailAgent = detail
    ? detail.agentName || agentNameByUid.get(detail.agentUid) || detail.agentUid
    : ""

  return (
    <>
      <div className="flex min-w-0 flex-col gap-2">
        <div className="flex flex-wrap items-center justify-end gap-2">
          <FilterToggle
            open={filtersOpen}
            count={appliedCount}
            controls={FILTER_PANEL_ID}
            onClick={() => setFiltersOpen((open) => !open)}
          />
        </div>
        {filterChips.length > 0 ? (
          <FilterChips items={filterChips} onClearAll={() => removeFilters([...FILTER_KEYS])} />
        ) : null}
        <form
          id={FILTER_PANEL_ID}
          onSubmit={(event) => {
            event.preventDefault()
            applyFilters()
          }}
          className={`flex min-w-0 flex-wrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3 sm:p-4 ${filtersOpen ? "" : "hidden"}`}
        >
          <Field className="min-w-0 flex-1 basis-56">
            <FieldLabel
              htmlFor="ssh-audit-keyword"
              className="text-xs text-muted-foreground"
            >
              {t("ssh.auditKeyword")}
            </FieldLabel>
            <Input
              id="ssh-audit-keyword"
              value={filters.keyword}
              placeholder={t("ssh.auditKeywordPlaceholder")}
              onChange={(event) => updateDraft("keyword", event.target.value)}
            />
          </Field>
          <Field className="w-40 shrink-0">
            <FieldLabel
              htmlFor="ssh-audit-status"
              className="text-xs text-muted-foreground"
            >
              {t("ssh.status")}
            </FieldLabel>
            <Select
              value={filters.status || "ALL"}
              onValueChange={(value) =>
                updateDraft("status", value === "ALL" ? "" : value)
              }
            >
              <SelectTrigger id="ssh-audit-status" className="w-full">
                <SelectValue placeholder={t("ssh.allStatus")} />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="ALL">{t("ssh.allStatus")}</SelectItem>
                  {AUDIT_STATUSES.map((item) => (
                    <SelectItem key={item} value={item}>
                      {enumLabel(t, item)}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <Field className="w-48 shrink-0">
            <FieldLabel
              htmlFor="ssh-audit-agent"
              className="text-xs text-muted-foreground"
            >
              {t("ssh.machine")}
            </FieldLabel>
            <Select
              value={filters.agentUid || "ALL"}
              onValueChange={(value) =>
                updateDraft("agentUid", value === "ALL" ? "" : value)
              }
            >
              <SelectTrigger id="ssh-audit-agent" className="w-full">
                <SelectValue placeholder={t("ssh.allMachines")} />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="ALL">{t("ssh.allMachines")}</SelectItem>
                  {agents.map((agent) => (
                    <SelectItem key={agent.uid} value={agent.uid}>
                      {agent.name}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <Field className="w-44 shrink-0">
            <FieldLabel
              htmlFor="ssh-audit-operator"
              className="text-xs text-muted-foreground"
            >
              {t("ssh.operator")}
            </FieldLabel>
            <Input
              id="ssh-audit-operator"
              value={filters.operatorUid}
              placeholder={t("ssh.operatorUidPlaceholder")}
              onChange={(event) =>
                updateDraft("operatorUid", event.target.value)
              }
            />
          </Field>
          <Field className="w-64 shrink-0">
            <FieldLabel
              htmlFor="ssh-audit-range"
              className="text-xs text-muted-foreground"
            >
              {t("filter.timeRange")}
            </FieldLabel>
            {/* 默认最多 30 天，且不能选未来日期。 */}
            <DateRangePicker
              id="ssh-audit-range"
              from={filters.from}
              to={filters.to}
              onChange={(range) =>
                setDraftState({
                  signature: appliedSignature,
                  filters: { ...filters, from: range.from, to: range.to },
                })
              }
            />
          </Field>
          <div className="flex items-center gap-2">
            {filtersDirty ? (
              <span className="text-xs text-warning" role="status">
                {t("filter.unapplied")}
              </span>
            ) : null}
            {appliedCount > 0 ? (
              <Button
                type="button"
                variant="ghost"
                onClick={() => removeFilters([...FILTER_KEYS])}
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
      </div>
      <ListFrame
        total={total}
        page={page}
        pageSize={pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => goToPage(Math.max(1, page - 1))}
        onNext={() => goToPage(page + 1)}
      >
        {loading ? (
          <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
        ) : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">
              {error || t("state.failed")}
            </p>
            <Button variant="outline" onClick={() => reload()}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("ssh.auditEmpty")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("ssh.auditTime")}
                </TableHead>
                <TableHead>{t("ssh.command")}</TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("ssh.machine")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("ssh.operator")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("ssh.status")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("ssh.exitCode")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((audit) => (
                <TableRow key={audit.uid}>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatDateTime(audit.createdAt) || "—"}
                  </TableCell>
                  <TableCell className="max-w-md text-xs break-words whitespace-normal">
                    <span className="font-mono">
                      {audit.command || audit.commandName || "—"}
                    </span>
                  </TableCell>
                  <TableCell className="text-xs break-words whitespace-normal">
                    {audit.agentName || machineLabel(audit.agentUid) || "—"}
                  </TableCell>
                  <TableCell className="text-xs break-words whitespace-normal">
                    {operatorName(audit) || "—"}
                  </TableCell>
                  <TableCell>
                    <StatusBadge value={audit.status} />
                  </TableCell>
                  <TableCell
                    className={`font-mono text-xs tabular-nums ${TABLE_NOWRAP}`}
                  >
                    <span
                      className={
                        audit.exitCode === 0 ? undefined : "text-destructive"
                      }
                    >
                      {audit.exitCode}
                    </span>
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(audit)}
                      actions={[]}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <DetailSheet
        open={detail !== null}
        title={detailAgent || t("ssh.auditDetailTitle")}
        size="wide"
        description={
          detail ? (
            <>
              <StatusBadge value={detail.status} />
              <span className="text-xs tabular-nums">
                {formatDateTime(detail.createdAt) || "—"}
              </span>
            </>
          ) : null
        }
        sections={
          detail
            ? [
                {
                  title: t("ssh.auditInfo"),
                  fields: [
                    {
                      label: t("ssh.status"),
                      value: <StatusBadge value={detail.status} />,
                    },
                    {
                      label: t("ssh.exitCode"),
                      value: (
                        <span
                          className={
                            detail.exitCode === 0
                              ? "font-mono text-xs tabular-nums"
                              : "font-mono text-xs text-destructive tabular-nums"
                          }
                        >
                          {detail.exitCode}
                        </span>
                      ),
                    },
                    {
                      label: t("ssh.machine"),
                      value: (
                        <DetailCopyText
                          value={detailAgent}
                          label={t("ssh.machine")}
                        />
                      ),
                    },
                    {
                      label: t("ssh.operator"),
                      value: (
                        <DetailCopyText
                          value={operatorName(detail)}
                          label={t("ssh.operator")}
                          mono={false}
                        />
                      ),
                    },
                    {
                      label: t("ssh.commandName"),
                      value: detail.commandName || null,
                    },
                    {
                      label: t("ssh.auditTime"),
                      value: formatDateTime(detail.createdAt) || "—",
                    },
                  ],
                },
                {
                  title: t("ssh.command"),
                  content: (
                    <CodeBlock
                      variant="plain"
                      code={detail.command}
                      language="bash"
                      label={t("ssh.command")}
                      copyable
                      wrap
                      placeholder={t("state.empty")}
                    />
                  ),
                },
                {
                  title: t("ssh.output"),
                  description: detail.truncated
                    ? t("ssh.outputTruncated")
                    : undefined,
                  content: (
                    <CodeBlock
                      code={detail.output}
                      language="text"
                      label={t("ssh.output")}
                      copyable
                      wrap
                      maxHeightClass="max-h-[50vh]"
                      placeholder={t("ssh.noOutput")}
                    />
                  ),
                },
                ...(detail.detail
                  ? [
                      {
                        title: t("ssh.errorDetail"),
                        content: (
                          <p className="rounded-md border border-destructive/20 bg-destructive/5 p-3 text-sm break-words text-destructive">
                            {detail.detail}
                          </p>
                        ),
                      },
                    ]
                  : []),
              ]
            : []
        }
        onOpenChange={(open) => {
          if (!open) {
            setDetail(null)
          }
        }}
      />
    </>
  )
}
