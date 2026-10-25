import { LayoutTemplate, Search, Trash2 } from "lucide-react"
import { useMemo, useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import type { QuerySnapshot } from "@/lib/query-store"

import {
  QueryStatementRow,
  handleRowKeyDown,
} from "@/pages/datasource/components/query-statement-row"
import { statementPreview } from "@/pages/datasource/query-dialect"

export interface TemplateItem {
  id: string
  name: string
  statement: string
  /** 选中后要合并进当前查询的参数；内置模板只带语句，自定义模板带完整快照。 */
  patch: Partial<QuerySnapshot>
  builtin: boolean
}

interface QueryTemplatePopoverProps {
  items: TemplateItem[]
  pending: boolean
  canSave: boolean
  onFill: (patch: Partial<QuerySnapshot>) => void
  onRequestSave: () => void
  onRestoreDefault: () => void
  onDeleteTemplate: (id: string) => void
}

// 模板浮层：搜索 + 分组（常用 / 我的模板）+ 保存入口。
// 选中模板只填入编辑器（不自动执行），执行仍由用户按下运行按钮或 Enter 决定。
export function QueryTemplatePopover({
  items,
  pending,
  canSave,
  onFill,
  onRequestSave,
  onRestoreDefault,
  onDeleteTemplate,
}: QueryTemplatePopoverProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [keyword, setKeyword] = useState("")

  const filtered = useMemo(() => {
    const needle = keyword.trim().toLowerCase()
    if (!needle) {
      return items
    }
    return items.filter(
      (item) =>
        item.name.toLowerCase().includes(needle) ||
        item.statement.toLowerCase().includes(needle)
    )
  }, [items, keyword])

  const builtin = filtered.filter((item) => item.builtin)
  const custom = filtered.filter((item) => !item.builtin)

  function choose(patch: Partial<QuerySnapshot>) {
    onFill(patch)
    setOpen(false)
  }

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        if (!next) {
          setKeyword("")
        }
      }}
    >
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          aria-label={t("datasource.query.templates")}
          aria-haspopup="dialog"
        >
          <LayoutTemplate />
          <span className="hidden sm:inline">{t("datasource.query.templates")}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[24rem] max-w-[calc(100vw-2rem)] p-0">
        <div className="border-b p-2">
          <div className="relative">
            <Search className="pointer-events-none absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              autoFocus
              value={keyword}
              className="pl-7"
              placeholder={t("datasource.query.searchTemplates")}
              aria-label={t("datasource.query.searchTemplates")}
              onChange={(event) => setKeyword(event.target.value)}
            />
          </div>
        </div>
        <div
          role="list"
          aria-label={t("datasource.query.templates")}
          className="max-h-80 overflow-y-auto p-1"
          onKeyDown={handleRowKeyDown}
        >
          {filtered.length === 0 ? (
            <p className="px-2 py-6 text-center text-sm text-muted-foreground">
              {t("datasource.query.noTemplateMatch")}
            </p>
          ) : null}
          {builtin.length > 0 ? (
            <TemplateGroup label={t("datasource.query.groupBuiltin")}>
              {builtin.map((item) => (
                <QueryStatementRow
                  key={item.id}
                  title={item.name}
                  statement={statementPreview(item.statement)}
                  disabled={pending}
                  onPrimary={() => choose(item.patch)}
                />
              ))}
            </TemplateGroup>
          ) : null}
          {custom.length > 0 ? (
            <TemplateGroup label={t("datasource.query.groupCustom")}>
              {custom.map((item) => (
                <QueryStatementRow
                  key={item.id}
                  title={item.name}
                  statement={statementPreview(item.statement)}
                  disabled={pending}
                  onPrimary={() => choose(item.patch)}
                  actions={
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-xs"
                      aria-label={t("datasource.query.deleteTemplate")}
                      onClick={() => onDeleteTemplate(item.id)}
                    >
                      <Trash2 />
                    </Button>
                  }
                />
              ))}
            </TemplateGroup>
          ) : null}
        </div>
        <div className="flex items-center justify-between gap-2 border-t p-1">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={!canSave}
            title={canSave ? undefined : t("datasource.query.saveTemplateEmpty")}
            onClick={() => {
              setOpen(false)
              onRequestSave()
            }}
          >
            {t("datasource.query.saveTemplate")}
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => {
              setOpen(false)
              onRestoreDefault()
            }}
          >
            {t("datasource.query.restoreDefault")}
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}

function TemplateGroup({
  label,
  children,
}: {
  label: string
  children: ReactNode
}) {
  return (
    <div role="group" aria-label={label} className="py-1">
      <p className="px-2 pt-1 pb-1 text-xs font-medium text-muted-foreground">
        {label}
      </p>
      {children}
    </div>
  )
}
