import { Copy, Search, TextCursorInput, Trash2 } from "lucide-react"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { ConfirmAction } from "@/components/confirm-action"
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
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import type { QueryHistoryEntry, QuerySnapshot } from "@/lib/query-store"

import {
  QueryStatementRow,
  handleRowKeyDown,
  relativeTime,
} from "@/pages/datasource/components/query-statement-row"
import { statementPreview } from "@/pages/datasource/query-dialect"

type StatusFilter = "all" | "ok" | "failed"

interface QueryHistorySheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  entries: QueryHistoryEntry[]
  pending: boolean
  onRun: (snapshot: QuerySnapshot) => void
  onFill: (snapshot: QuerySnapshot) => void
  onDelete: (id: string) => void
  onClear: () => void
}

// 全量历史抽屉：只在需要搜索、筛选或清理时才进来，
// 与详情/元数据的右侧 Sheet 保持同一套语言。
export function QueryHistorySheet({
  open,
  onOpenChange,
  entries,
  pending,
  onRun,
  onFill,
  onDelete,
  onClear,
}: QueryHistorySheetProps) {
  const { t } = useTranslation()
  const [keyword, setKeyword] = useState("")
  const [status, setStatus] = useState<StatusFilter>("all")
  const [confirmOpen, setConfirmOpen] = useState(false)

  const filtered = useMemo(() => {
    const needle = keyword.trim().toLowerCase()
    return entries.filter((entry) => {
      if (status === "ok" && !entry.ok) {
        return false
      }
      if (status === "failed" && entry.ok) {
        return false
      }
      if (!needle) {
        return true
      }
      return (
        entry.statement.toLowerCase().includes(needle) ||
        entry.error.toLowerCase().includes(needle)
      )
    })
  }, [entries, keyword, status])

  const groups = useMemo(
    () =>
      groupByDay(filtered, {
        today: t("datasource.query.today"),
        yesterday: t("datasource.query.yesterday"),
        earlier: t("datasource.query.earlier"),
      }),
    [filtered, t]
  )

  return (
    <Sheet
      open={open}
      onOpenChange={(next) => {
        onOpenChange(next)
        if (!next) {
          setKeyword("")
          setStatus("all")
        }
      }}
    >
      <SheetContent
        side="right"
        className="gap-0 p-0 data-[side=right]:sm:w-[40vw] data-[side=right]:sm:max-w-[40vw]"
      >
        <SheetHeader className="border-b">
          <SheetTitle>{t("datasource.query.history")}</SheetTitle>
          <SheetDescription>
            {t("datasource.query.historyDescription")}
          </SheetDescription>
        </SheetHeader>
        <div className="flex shrink-0 flex-wrap items-center gap-2 border-b p-3">
          <div className="relative min-w-48 flex-1">
            <Search className="pointer-events-none absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keyword}
              className="pl-7"
              placeholder={t("datasource.query.searchStatements")}
              aria-label={t("datasource.query.searchStatements")}
              onChange={(event) => setKeyword(event.target.value)}
            />
          </div>
          <Select
            value={status}
            onValueChange={(value) => setStatus(value as StatusFilter)}
          >
            <SelectTrigger className="w-32 shrink-0" aria-label={t("datasource.query.filter")}>
              <SelectValue placeholder={t("datasource.query.filter")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t("datasource.query.filterAll")}</SelectItem>
              <SelectItem value="ok">{t("datasource.query.succeeded")}</SelectItem>
              <SelectItem value="failed">{t("datasource.query.failed")}</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div
          role="list"
          aria-label={t("datasource.query.history")}
          className="min-h-0 flex-1 overflow-y-auto p-2"
          onKeyDown={handleRowKeyDown}
        >
          {entries.length === 0 ? (
            <EmptyState
              title={t("datasource.query.noHistory")}
              hint={t("datasource.query.noHistoryHint")}
            />
          ) : null}
          {entries.length > 0 && filtered.length === 0 ? (
            <EmptyState
              title={t("datasource.query.noStatementMatch")}
              hint={t("datasource.query.noStatementMatchHint")}
            />
          ) : null}
          {groups.map((group) => (
            <div key={group.label} className="py-1">
              <p className="px-2 pt-1 pb-1 text-xs font-medium text-muted-foreground">
                {group.label}
              </p>
              {group.entries.map((entry) => (
                <QueryStatementRow
                  key={entry.id}
                  title={statementPreview(entry.statement, 64)}
                  tooltip={
                    entry.error
                      ? `${entry.statement}\n\n${entry.error}`
                      : entry.statement
                  }
                  ok={entry.ok}
                  disabled={pending}
                  meta={`${relativeTime(entry.at, t)} · ${
                    entry.ok
                      ? t("datasource.query.succeeded")
                      : t("datasource.query.failed")
                  } · ${t("datasource.query.duration", { ms: entry.durationMs })}`}
                  onPrimary={() => {
                    onRun(entry.snapshot)
                    onOpenChange(false)
                  }}
                  actions={
                    <>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-xs"
                        aria-label={t("datasource.query.fillOnly")}
                        onClick={() => {
                          onFill(entry.snapshot)
                          onOpenChange(false)
                        }}
                      >
                        <TextCursorInput />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-xs"
                        aria-label={t("datasource.query.copyStatement")}
                        onClick={() => void copyStatement(entry.statement, t)}
                      >
                        <Copy />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-xs"
                        aria-label={t("action.delete")}
                        onClick={() => onDelete(entry.id)}
                      >
                        <Trash2 />
                      </Button>
                    </>
                  }
                />
              ))}
            </div>
          ))}
        </div>
        <div className="flex shrink-0 items-center justify-between gap-2 border-t p-3">
          <span className="text-xs text-muted-foreground">
            {t("datasource.query.localOnly")}
          </span>
          <Button
            type="button"
            variant="destructive"
            size="sm"
            disabled={entries.length === 0}
            onClick={() => setConfirmOpen(true)}
          >
            {t("datasource.query.clearHistory")}
          </Button>
        </div>
      </SheetContent>
      <ConfirmAction
        open={confirmOpen}
        title={t("datasource.query.clearHistoryTitle")}
        description={t("datasource.query.clearHistoryDescription")}
        confirmLabel={t("datasource.query.clearHistory")}
        cancelLabel={t("action.cancel")}
        onConfirm={() => {
          onClear()
          setConfirmOpen(false)
        }}
        onOpenChange={setConfirmOpen}
      />
    </Sheet>
  )
}

function EmptyState({ title, hint }: { title: string; hint: string }) {
  return (
    <div className="flex flex-col items-center gap-1 px-4 py-12 text-center">
      <p className="text-sm text-muted-foreground">{title}</p>
      <p className="text-xs text-muted-foreground">{hint}</p>
    </div>
  )
}

interface DayGroup {
  label: string
  entries: QueryHistoryEntry[]
}

function groupByDay(
  entries: QueryHistoryEntry[],
  labels: { today: string; yesterday: string; earlier: string }
): DayGroup[] {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const yesterday = new Date(today)
  yesterday.setDate(today.getDate() - 1)
  const todayAt = today.getTime()
  const yesterdayAt = yesterday.getTime()

  const groups: DayGroup[] = []
  const push = (label: string, entry: QueryHistoryEntry) => {
    const last = groups.at(-1)
    if (last && last.label === label) {
      last.entries.push(entry)
      return
    }
    groups.push({ label, entries: [entry] })
  }
  for (const entry of entries) {
    if (entry.at >= todayAt) {
      push(labels.today, entry)
    } else if (entry.at >= yesterdayAt) {
      push(labels.yesterday, entry)
    } else {
      push(labels.earlier, entry)
    }
  }
  return groups
}

async function copyStatement(
  statement: string,
  t: (key: string) => string
) {
  try {
    await navigator.clipboard.writeText(statement)
    toast.success(t("datasource.query.copied"))
  } catch {
    toast.error(t("action.copyFailed"))
  }
}
