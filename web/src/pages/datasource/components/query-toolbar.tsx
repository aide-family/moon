import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import type { QueryLibrary } from "@/hooks/use-query-store"
import {
  normalizeSnapshot,
  type QueryDialect,
  type QuerySnapshot,
} from "@/lib/query-store"

import { QueryHistoryPopover } from "@/pages/datasource/components/query-history-popover"
import { QueryHistorySheet } from "@/pages/datasource/components/query-history-sheet"
import { relativeTime } from "@/pages/datasource/components/query-statement-row"
import {
  QueryTemplatePopover,
  type TemplateItem,
} from "@/pages/datasource/components/query-template-popover"
import { SaveTemplateDialog } from "@/pages/datasource/components/save-template-dialog"
import { queryStatement } from "@/pages/datasource/query-dialect"
import { BUILTIN_QUERY_TEMPLATES } from "@/pages/datasource/query-templates"

interface DatasourceQueryToolbarProps {
  dialect: QueryDialect
  library: QueryLibrary
  /** 当前编辑器里的参数，用于保存模板与判断是否可保存。 */
  snapshot: QuerySnapshot
  defaultSnapshot: QuerySnapshot
  lastUsedAt: number | null
  pending: boolean
  onFill: (snapshot: QuerySnapshot) => void
  onRun: (snapshot: QuerySnapshot) => void
}

// 所有数据源共用的一条查询工具条：上次使用提示 + 模板 + 历史。
// 落在编辑器 label 行右侧，不额外占用纵向空间。
export function DatasourceQueryToolbar({
  dialect,
  library,
  snapshot,
  defaultSnapshot,
  lastUsedAt,
  pending,
  onFill,
  onRun,
}: DatasourceQueryToolbarProps) {
  const { t } = useTranslation()
  const [sheetOpen, setSheetOpen] = useState(false)
  const [saveOpen, setSaveOpen] = useState(false)

  const items = useMemo<TemplateItem[]>(() => {
    const builtin = BUILTIN_QUERY_TEMPLATES[dialect].map((template) => ({
      id: template.id,
      name: t(template.nameKey),
      statement: queryStatement(dialect, normalizeSnapshot(template.snapshot)),
      patch: template.snapshot,
      builtin: true,
    }))
    const custom = library.templates.map((template) => ({
      id: template.id,
      name: template.name,
      statement: queryStatement(dialect, template.snapshot),
      patch: template.snapshot as Partial<QuerySnapshot>,
      builtin: false,
    }))
    return [...builtin, ...custom]
  }, [dialect, library.templates, t])

  const canSave = queryStatement(dialect, snapshot).length > 0

  return (
    <>
      <div className="ml-auto flex items-center gap-1">
        {lastUsedAt ? (
          <span className="hidden text-xs text-muted-foreground lg:inline">
            {t("datasource.query.lastUsed", {
              time: relativeTime(lastUsedAt, t),
            })}
          </span>
        ) : null}
        <QueryTemplatePopover
          items={items}
          pending={pending}
          canSave={canSave}
          onFill={(patch) => onFill(normalizeSnapshot({ ...snapshot, ...patch }))}
          onRequestSave={() => setSaveOpen(true)}
          onRestoreDefault={() => onFill(defaultSnapshot)}
          onDeleteTemplate={library.deleteTemplate}
        />
        <QueryHistoryPopover
          entries={library.history}
          pending={pending}
          onRun={onRun}
          onFill={onFill}
          onViewAll={() => setSheetOpen(true)}
        />
      </div>
      <QueryHistorySheet
        open={sheetOpen}
        onOpenChange={setSheetOpen}
        entries={library.history}
        pending={pending}
        onRun={onRun}
        onFill={onFill}
        onDelete={library.removeEntry}
        onClear={library.clear}
      />
      <SaveTemplateDialog
        open={saveOpen}
        onOpenChange={setSaveOpen}
        statement={queryStatement(dialect, snapshot)}
        onSubmit={({ name, description }) => {
          library.saveTemplate({ name, description, snapshot })
          toast.success(t("datasource.query.templateSaved"))
        }}
      />
    </>
  )
}
