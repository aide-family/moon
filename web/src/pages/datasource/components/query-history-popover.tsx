import { Copy, History, TextCursorInput } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import type { QueryHistoryEntry, QuerySnapshot } from "@/lib/query-store"

import {
  QueryStatementRow,
  handleRowKeyDown,
  relativeTime,
} from "@/pages/datasource/components/query-statement-row"
import { statementPreview } from "@/pages/datasource/query-dialect"

const RECENT_LIMIT = 8

interface QueryHistoryPopoverProps {
  entries: QueryHistoryEntry[]
  pending: boolean
  onRun: (snapshot: QuerySnapshot) => void
  onFill: (snapshot: QuerySnapshot) => void
  onViewAll: () => void
}

// 历史浮层只放最近几条：绝大多数场景点一下就重跑，
// 需要搜索/清理时再进全量抽屉。
export function QueryHistoryPopover({
  entries,
  pending,
  onRun,
  onFill,
  onViewAll,
}: QueryHistoryPopoverProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const recent = entries.slice(0, RECENT_LIMIT)

  function replay(entry: QueryHistoryEntry) {
    onRun(entry.snapshot)
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          aria-label={t("datasource.query.history")}
          aria-haspopup="dialog"
        >
          <History />
          <span className="hidden sm:inline">{t("datasource.query.history")}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[26rem] max-w-[calc(100vw-2rem)] p-0">
        <div
          role="list"
          aria-label={t("datasource.query.history")}
          className="max-h-80 overflow-y-auto p-1"
          onKeyDown={handleRowKeyDown}
        >
          {recent.length === 0 ? (
            <div className="px-2 py-6 text-center">
              <p className="text-sm text-muted-foreground">
                {t("datasource.query.noHistory")}
              </p>
              <p className="text-xs text-muted-foreground">
                {t("datasource.query.noHistoryHint")}
              </p>
            </div>
          ) : null}
          {recent.map((entry) => (
            <QueryStatementRow
              key={entry.id}
              title={statementPreview(entry.statement, 64)}
              tooltip={entry.error ? `${entry.statement}\n\n${entry.error}` : entry.statement}
              ok={entry.ok}
              disabled={pending}
              meta={`${relativeTime(entry.at, t)} · ${
                entry.ok ? t("datasource.query.succeeded") : t("datasource.query.failed")
              } · ${t("datasource.query.duration", { ms: entry.durationMs })}`}
              onPrimary={() => replay(entry)}
              actions={
                <>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    aria-label={t("datasource.query.fillOnly")}
                    onClick={() => {
                      onFill(entry.snapshot)
                      setOpen(false)
                    }}
                  >
                    <TextCursorInput />
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    aria-label={t("datasource.query.copyStatement")}
                    onClick={() => {
                      void copyStatement(entry.statement, t("datasource.query.copied"), t("action.copyFailed"))
                    }}
                  >
                    <Copy />
                  </Button>
                </>
              }
            />
          ))}
        </div>
        <div className="flex items-center justify-between gap-2 border-t p-1">
          <span className="px-2 text-xs text-muted-foreground">
            {t("datasource.query.localOnly")}
          </span>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={entries.length === 0}
            onClick={() => {
              setOpen(false)
              onViewAll()
            }}
          >
            {t("datasource.query.viewAll", { count: entries.length })}
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}

async function copyStatement(
  statement: string,
  copied: string,
  failed: string
) {
  try {
    await navigator.clipboard.writeText(statement)
    toast.success(copied)
  } catch {
    toast.error(failed)
  }
}
