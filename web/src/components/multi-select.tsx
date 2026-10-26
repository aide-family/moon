import { useMemo, useState } from "react"
import { Check, ChevronDown, Search } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/class-names"

export interface MultiSelectOption {
  value: string
  label: string
  description?: string
}

interface MultiSelectProps {
  values: string[]
  options: MultiSelectOption[]
  onChange: (values: string[]) => void
  placeholder: string
  searchPlaceholder: string
  emptyText: string
  allLabel: string
  clearLabel?: string
  /** 自定义触发器上的摘要；不传时展示前两个名称。 */
  selectedLabel?: (selected: number, total: number, labels: string[]) => string
  disabled?: boolean
  /** 校验失败：给触发器加 aria-invalid 与红框（FormField 已把字段标记为 invalid）。 */
  invalid?: boolean
  id?: string
  className?: string
  "aria-describedby"?: string
}

/**
 * 带搜索和全选的多选下拉框。选项加载在父组件完成，组件只负责选择与筛选。
 * 点击选项不会关闭面板，方便连续多选。
 */
export function MultiSelect({
  values,
  options,
  onChange,
  placeholder,
  searchPlaceholder,
  emptyText,
  allLabel,
  clearLabel,
  selectedLabel,
  disabled = false,
  invalid = false,
  id,
  className,
  "aria-describedby": ariaDescribedBy,
}: MultiSelectProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")

  const selected = useMemo(() => new Set(values), [values])
  const filtered = useMemo(() => {
    const keyword = query.trim().toLowerCase()
    if (keyword === "") {
      return options
    }
    return options.filter((option) => {
      const haystack = `${option.label} ${option.value} ${option.description ?? ""}`.toLowerCase()
      return haystack.includes(keyword)
    })
  }, [options, query])
  const filteredValues = filtered.map((option) => option.value)
  const allFilteredSelected =
    filteredValues.length > 0 && filteredValues.every((value) => selected.has(value))

  const selectedLabels = values
    .map((value) => options.find((option) => option.value === value)?.label ?? value)
    .filter(Boolean)
  const summary =
    values.length === 0
      ? placeholder
      : selectedLabel
        ? selectedLabel(values.length, options.length, selectedLabels)
        : selectedLabels.length <= 2
          ? selectedLabels.join(", ")
          : `${selectedLabels[0]} +${selectedLabels.length - 1}`

  function toggle(value: string) {
    if (selected.has(value)) {
      onChange(values.filter((item) => item !== value))
      return
    }
    onChange([...values, value])
  }

  function toggleAll() {
    if (allFilteredSelected) {
      const removing = new Set(filteredValues)
      onChange(values.filter((value) => !removing.has(value)))
      return
    }
    const next = new Set(values)
    for (const value of filteredValues) {
      next.add(value)
    }
    onChange([...next])
  }

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        if (!next) {
          setQuery("")
        }
      }}
    >
      <PopoverTrigger asChild>
        <Button
          id={id}
          type="button"
          variant="outline"
          role="combobox"
          data-slot="select-trigger"
          aria-expanded={open}
          aria-invalid={invalid || undefined}
          aria-describedby={ariaDescribedBy}
          disabled={disabled}
          className={cn(
            "h-8 w-full min-w-0 justify-between gap-2 font-normal",
            className
          )}
        >
          <span
            className={cn(
              "min-w-0 truncate",
              values.length === 0 && "text-muted-foreground"
            )}
          >
            {summary}
          </span>
          <ChevronDown className="size-4 shrink-0 opacity-60" aria-hidden />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        align="start"
        sideOffset={4}
        className="z-[60] w-[var(--radix-popover-trigger-width)] min-w-56 p-0"
      >
        <div className="flex items-center gap-2 border-b p-2">
          <Search className="size-4 shrink-0 text-muted-foreground" aria-hidden />
          <Input
            autoFocus
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={searchPlaceholder}
            aria-label={searchPlaceholder}
            className="border-0 bg-transparent px-0 shadow-none focus-visible:ring-0"
          />
        </div>
        <div className="flex items-center justify-between gap-2 border-b p-1">
          <button
            type="button"
            role="checkbox"
            aria-checked={allFilteredSelected}
            onClick={toggleAll}
            className="flex min-w-0 flex-1 items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted"
          >
            <span
              className={cn(
                "flex size-4 shrink-0 items-center justify-center rounded border",
                allFilteredSelected
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-input bg-background"
              )}
              aria-hidden
            >
              {allFilteredSelected ? <Check className="size-3" /> : null}
            </span>
            <span className="truncate font-medium">
              {query.trim() === ""
                ? allLabel
                : `${allLabel} (${filteredValues.length})`}
            </span>
          </button>
          {clearLabel && values.length > 0 ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => onChange([])}
            >
              {clearLabel}
            </Button>
          ) : null}
        </div>
        <div
          className="max-h-64 overflow-y-auto p-1"
          role="listbox"
          aria-multiselectable
        >
          {filtered.length === 0 ? (
            <p className="px-3 py-6 text-center text-sm text-muted-foreground">
              {emptyText}
            </p>
          ) : (
            filtered.map((option) => {
              const checked = selected.has(option.value)
              return (
                <button
                  key={option.value}
                  type="button"
                  role="option"
                  aria-selected={checked}
                  onClick={() => toggle(option.value)}
                  className="flex w-full min-w-0 items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted"
                >
                  <span
                    className={cn(
                      "flex size-4 shrink-0 items-center justify-center rounded border",
                      checked
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-input bg-background"
                    )}
                    aria-hidden
                  >
                    {checked ? <Check className="size-3" /> : null}
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block truncate">{option.label}</span>
                    {option.description ? (
                      <span className="block truncate text-xs text-muted-foreground">
                        {option.description}
                      </span>
                    ) : null}
                  </span>
                </button>
              )
            })
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}
