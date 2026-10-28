import { Check, ChevronDown, Plus, Search } from "lucide-react"
import { useEffect, useId, useState, type KeyboardEvent } from "react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/class-names"

export interface SearchableSelectOption {
  value: string
  label?: string
  description?: string
}

type SelectRow =
  | {
      kind: "option"
      id: string
      value: string
      label: string
      description?: string
    }
  | { kind: "custom"; id: string; value: string; label: string }

interface SearchableSelectProps {
  value: string
  options: SearchableSelectOption[]
  onChange: (value: string) => void
  placeholder: string
  searchPlaceholder: string
  emptyText: string
  /** 允许把输入的内容作为自定义值提交；只用于值这类需要自由输入的场景。 */
  allowCustom?: boolean
  /** allowCustom 时，自定义值选项显示的文案。 */
  customLabel?: (value: string) => string
  /** allowCustom 时，面板底部的补充说明。 */
  customHint?: string
  clearLabel?: string
  disabled?: boolean
  /** 校验失败：给触发器加 aria-invalid 与红框。 */
  invalid?: boolean
  id?: string
  className?: string
  ariaLabel?: string
  "aria-describedby"?: string
}

/**
 * 单选下拉框：支持搜索、键盘上下选择/回车确认/Esc 关闭。
 * 选项由调用方在父组件计算；allowCustom 时可以把未命中选项的输入作为自定义值提交。
 */
export function SearchableSelect({
  value,
  options,
  onChange,
  placeholder,
  searchPlaceholder,
  emptyText,
  allowCustom = false,
  customLabel,
  customHint,
  clearLabel,
  disabled = false,
  invalid = false,
  id,
  className,
  ariaLabel,
  "aria-describedby": ariaDescribedBy,
}: SearchableSelectProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")
  const [activeIndex, setActiveIndex] = useState(0)
  const listboxId = useId()

  const keyword = query.trim().toLowerCase()
  const filteredOptions =
    keyword === ""
      ? options
      : options.filter((option) => {
          const haystack =
            `${option.label ?? option.value} ${option.value} ${option.description ?? ""}`.toLowerCase()
          return haystack.includes(keyword)
        })

  const trimmedQuery = query.trim()
  // 只有输入内容不是任何一个选项时才提供自定义值，避免和已有选项重复。
  const customValue =
    allowCustom &&
    trimmedQuery !== "" &&
    !options.some(
      (option) =>
        option.value === trimmedQuery ||
        (option.label ?? option.value) === trimmedQuery
    )
      ? trimmedQuery
      : ""

  const rows: SelectRow[] = []
  if (customValue !== "") {
    rows.push({
      kind: "custom",
      id: `${listboxId}-row-0`,
      value: customValue,
      label: customLabel ? customLabel(customValue) : customValue,
    })
  }
  for (const option of filteredOptions) {
    rows.push({
      kind: "option",
      id: `${listboxId}-row-${rows.length}`,
      value: option.value,
      label: option.label ?? option.value,
      description: option.description,
    })
  }

  const safeActiveIndex =
    rows.length === 0 ? -1 : Math.min(Math.max(activeIndex, 0), rows.length - 1)
  const activeRow = safeActiveIndex >= 0 ? rows[safeActiveIndex] : undefined
  const activeRowId = activeRow?.id
  const selectedOption = options.find((option) => option.value === value)
  const displayValue =
    value === "" ? placeholder : (selectedOption?.label ?? value)

  useEffect(() => {
    if (!open || !activeRowId) {
      return
    }
    document.getElementById(activeRowId)?.scrollIntoView({ block: "nearest" })
  }, [activeRowId, open])

  function handleOpenChange(next: boolean) {
    setOpen(next)
    if (next) {
      const selectedIndex = options.findIndex(
        (option) => option.value === value
      )
      setActiveIndex(selectedIndex >= 0 ? selectedIndex : 0)
    } else {
      setQuery("")
      setActiveIndex(0)
    }
  }

  function commit(row: SelectRow) {
    // 自定义值按完整匹配的正则使用，提交前去掉首尾空白。
    onChange(row.kind === "custom" ? row.value.trim() : row.value)
    handleOpenChange(false)
  }

  function changeQuery(next: string) {
    setQuery(next)
    setActiveIndex(0)
  }

  function onSearchKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    // 中文输入法组合期间不要抢占方向键和回车。
    if (event.nativeEvent.isComposing) {
      return
    }
    if (event.key === "ArrowDown") {
      if (rows.length === 0) {
        return
      }
      event.preventDefault()
      setActiveIndex(Math.min(safeActiveIndex + 1, rows.length - 1))
      return
    }
    if (event.key === "ArrowUp") {
      if (rows.length === 0) {
        return
      }
      event.preventDefault()
      setActiveIndex(Math.max(safeActiveIndex - 1, 0))
      return
    }
    if (event.key === "Enter") {
      if (!activeRow) {
        return
      }
      event.preventDefault()
      commit(activeRow)
      return
    }
    if (event.key === "Escape") {
      event.preventDefault()
      handleOpenChange(false)
    }
  }

  function onTriggerKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault()
      handleOpenChange(true)
    }
  }

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button
          id={id}
          type="button"
          variant="outline"
          role="combobox"
          data-slot="select-trigger"
          aria-expanded={open}
          aria-invalid={invalid || undefined}
          aria-controls={open ? listboxId : undefined}
          aria-label={ariaLabel}
          aria-describedby={ariaDescribedBy}
          aria-haspopup="listbox"
          disabled={disabled}
          title={value === "" ? undefined : displayValue}
          onKeyDown={onTriggerKeyDown}
          className={cn(
            "h-8 w-full min-w-0 justify-between gap-2 font-normal",
            className
          )}
        >
          <span
            className={cn(
              "min-w-0 flex-1 truncate text-left",
              value === "" && "text-muted-foreground"
            )}
          >
            {displayValue}
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
          <Search
            className="size-4 shrink-0 text-muted-foreground"
            aria-hidden
          />
          <Input
            autoFocus
            value={query}
            onChange={(event) => changeQuery(event.target.value)}
            onKeyDown={onSearchKeyDown}
            placeholder={searchPlaceholder}
            aria-label={searchPlaceholder}
            aria-autocomplete="list"
            aria-controls={listboxId}
            aria-activedescendant={activeRowId}
            autoComplete="off"
            spellCheck={false}
            className="border-0 bg-transparent px-0 shadow-none focus-visible:ring-0"
          />
        </div>
        <div
          id={listboxId}
          role="listbox"
          aria-label={ariaLabel ?? placeholder}
          className="max-h-64 overflow-y-auto overscroll-contain p-1"
        >
          {rows.length === 0 ? (
            <p className="px-3 py-6 text-center text-sm text-muted-foreground">
              {emptyText}
            </p>
          ) : (
            rows.map((row, index) => {
              const isSelected = row.kind === "option" && row.value === value
              return (
                <button
                  key={row.id}
                  id={row.id}
                  type="button"
                  role="option"
                  tabIndex={-1}
                  aria-selected={isSelected}
                  data-active={index === safeActiveIndex}
                  title={row.value}
                  onPointerMove={() => setActiveIndex(index)}
                  onClick={() => commit(row)}
                  className={cn(
                    "flex w-full min-w-0 items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted",
                    index === safeActiveIndex &&
                      "bg-accent text-accent-foreground",
                    row.kind === "custom" && "text-primary-ink"
                  )}
                >
                  {row.kind === "custom" ? (
                    <Plus className="size-4 shrink-0" aria-hidden />
                  ) : null}
                  <span className="min-w-0 flex-1">
                    <span className="block truncate">{row.label}</span>
                    {row.kind === "option" && row.description ? (
                      <span className="block truncate text-xs text-muted-foreground">
                        {row.description}
                      </span>
                    ) : null}
                  </span>
                  {isSelected ? (
                    <Check
                      className="size-4 shrink-0 text-primary-ink"
                      aria-hidden
                    />
                  ) : null}
                </button>
              )
            })
          )}
        </div>
        {allowCustom && customHint ? (
          <p className="border-t px-3 py-2 text-xs text-muted-foreground">
            {customHint}
          </p>
        ) : null}
        {clearLabel && value !== "" ? (
          <div className="border-t p-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="w-full justify-start"
              onClick={() => {
                onChange("")
                handleOpenChange(false)
              }}
            >
              {clearLabel}
            </Button>
          </div>
        ) : null}
      </PopoverContent>
    </Popover>
  )
}
