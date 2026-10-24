import { ChevronsUpDown, GripVertical, Plus, Trash2 } from "lucide-react"
import { Fragment, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import type { AlertColumnItem } from "@/api/alert-page/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/class-names"
import { usePointerReorder } from "@/hooks/use-flip-reorder"
import { FIELD_COLUMNS, fieldLabel } from "@/pages/alert/columns"

interface AlertColumnDialogProps {
  open: boolean
  columns: AlertColumnItem[]
  pending?: boolean
  error?: string
  onSave: (columns: AlertColumnItem[]) => void
  onOpenChange: (open: boolean) => void
}

export function AlertColumnDialog({
  open,
  columns,
  pending = false,
  error = "",
  onSave,
  onOpenChange,
}: AlertColumnDialogProps) {
  const { t } = useTranslation()
  const [draft, setDraft] = useState<AlertColumnItem[]>(columns)
  const [fieldKeys, setFieldKeys] = useState<string[]>([])
  const [fieldPopoverOpen, setFieldPopoverOpen] = useState(false)
  const [labelKey, setLabelKey] = useState("")
  const [labelAlias, setLabelAlias] = useState("")
  const [localError, setLocalError] = useState("")
  // 拖动源脱离列表流，原位即时填补；目标位置只显示空占位。
  const {
    containerRef,
    dragKey,
    dropTarget: dropKey,
    dropPosition,
    dragSize,
    startDrag,
  } = usePointerReorder<HTMLDivElement>(
    draft.map((item) => `${item.type}:${item.key}`).join("|"),
    "y",
    reorder
  )

  const used = useMemo(
    () => new Set(draft.map((item) => `${item.type}:${item.key}`)),
    [draft]
  )
  const remainingFields = FIELD_COLUMNS.filter(
    (item) => !used.has(`FIELD:${item.key}`)
  )
  function toggleField(key: string, checked: boolean) {
    setFieldKeys((current) =>
      checked
        ? current.includes(key)
          ? current
          : [...current, key]
        : current.filter((item) => item !== key)
    )
  }

  function reorder(
    sourceKey: string,
    targetKey: string,
    position: "before" | "after"
  ) {
    const from = draft.findIndex(
      (item) => `${item.type}:${item.key}` === sourceKey
    )
    const target = draft.findIndex(
      (item) => `${item.type}:${item.key}` === targetKey
    )
    if (from < 0 || target < 0 || sourceKey === targetKey) {
      return
    }

    const next = [...draft]
    const [moved] = next.splice(from, 1)
    let insertionIndex = target + (position === "after" ? 1 : 0)
    if (from < insertionIndex) {
      insertionIndex -= 1
    }
    next.splice(insertionIndex, 0, moved)
    setDraft(next)
  }

  function addField() {
    const selected = new Set(fieldKeys)
    const fieldsToAdd = remainingFields.filter((item) => selected.has(item.key))
    if (fieldsToAdd.length === 0) {
      return
    }
    setDraft([
      ...draft,
      ...fieldsToAdd.map((item) => ({
        key: item.key,
        type: "FIELD" as const,
        label: "",
      })),
    ])
    setFieldKeys([])
    setFieldPopoverOpen(false)
    setLocalError("")
  }

  function addLabel() {
    const key = labelKey.trim()
    if (!key || !/^[A-Za-z_][A-Za-z0-9_]*$/.test(key)) {
      setLocalError(t("alertEvent.columnInvalidLabel"))
      return
    }
    if (used.has(`LABEL:${key}`)) {
      setLocalError(t("alertEvent.columnDuplicate"))
      return
    }
    setDraft([...draft, { key, type: "LABEL", label: labelAlias.trim() }])
    setLabelKey("")
    setLabelAlias("")
    setLocalError("")
  }

  function updateAlias(index: number, value: string) {
    const next = [...draft]
    next[index] = { ...next[index], label: value }
    setDraft(next)
  }

  function remove(index: number) {
    setDraft(draft.filter((_, current) => current !== index))
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[85vh] flex-col gap-4 overflow-hidden sm:max-w-2xl">
        <DialogHeader className="shrink-0">
          <DialogTitle>{t("alertEvent.columnsTitle")}</DialogTitle>
        </DialogHeader>
        <p className="shrink-0 text-sm text-muted-foreground">
          {t("alertEvent.columnsHint")}
        </p>
        <div
          ref={containerRef}
          className={cn(
            "relative grid min-h-0 flex-1 content-start gap-2 overflow-y-auto pr-1",
            dragKey && "select-none"
          )}
        >
          {draft.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {t("alertEvent.columnsEmpty")}
            </p>
          ) : null}
          {draft.map((column, index) => {
            const itemKey = `${column.type}:${column.key}`
            const label =
              column.type === "FIELD" ? fieldLabel(t, column.key) : column.key
            const row = (
              <div
                key={itemKey}
                data-flip-id={dragKey === itemKey ? undefined : itemKey}
                data-sort-key={itemKey}
                className={cn(
                  "flex flex-wrap items-center gap-2 rounded-lg border p-2 sm:flex-nowrap",
                  dropKey === itemKey &&
                    dragKey !== itemKey &&
                    "border-primary/60 ring-1 ring-primary/30",
                  dragKey === itemKey &&
                    "pointer-events-none absolute opacity-0"
                )}
                draggable={false}
              >
                <GripVertical
                  className="size-4 shrink-0 cursor-grab text-muted-foreground active:cursor-grabbing"
                  aria-hidden
                  style={{ touchAction: "none" }}
                  onPointerDown={(event) => startDrag(itemKey, event)}
                />
                <span className="min-w-0 flex-1 truncate font-mono text-xs">
                  {label}
                </span>
                <Input
                  className="max-w-[14rem]"
                  value={column.label}
                  placeholder={
                    column.type === "LABEL"
                      ? column.key
                      : t("alertEvent.columnAliasPlaceholder")
                  }
                  aria-label={t("alertEvent.columnAlias")}
                  onChange={(event) => updateAlias(index, event.target.value)}
                />
                <Button
                  type="button"
                  size="icon"
                  variant="ghost"
                  aria-label={t("action.delete")}
                  onClick={() => remove(index)}
                >
                  <Trash2 aria-hidden />
                </Button>
              </div>
            )
            const showDropSlot = dropKey === itemKey && dragKey !== itemKey
            const dropSlot = showDropSlot ? (
              <div
                key={`drop-slot:${itemKey}`}
                data-drop-slot
                data-drop-target={dropKey}
                data-drop-position={dropPosition}
                aria-hidden="true"
                className="w-full rounded-lg border border-dashed border-primary/60 bg-primary/5"
                style={{ height: dragSize.height || undefined }}
              />
            ) : null

            return (
              <Fragment key={itemKey}>
                {showDropSlot && dropPosition === "before" ? dropSlot : null}
                {row}
                {showDropSlot && dropPosition === "after" ? dropSlot : null}
              </Fragment>
            )
          })}
        </div>
        <div className="grid shrink-0 gap-3 border-t pt-3">
          {remainingFields.length > 0 ? (
            <div className="grid w-full grid-cols-[minmax(0,1fr)_auto] items-end gap-2">
              <FormField
                label={t("alertEvent.addFieldColumn")}
                className="min-w-0"
              >
                <Popover
                  open={fieldPopoverOpen}
                  onOpenChange={setFieldPopoverOpen}
                >
                  <PopoverTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      className="w-full justify-between font-normal"
                      aria-label={`${t("alertEvent.addFieldColumn")}: ${
                        fieldKeys.length > 0
                          ? t("alertEvent.columnFieldsSelected", {
                              count: fieldKeys.length,
                            })
                          : t("alertEvent.columnSelectFields")
                      }`}
                    >
                      <span
                        className={cn(
                          "min-w-0 flex-1 truncate text-left",
                          fieldKeys.length === 0 && "text-muted-foreground"
                        )}
                      >
                        {fieldKeys.length > 0
                          ? t("alertEvent.columnFieldsSelected", {
                              count: fieldKeys.length,
                            })
                          : t("alertEvent.columnSelectFields")}
                      </span>
                      <ChevronsUpDown data-icon="inline-end" aria-hidden />
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent
                    align="start"
                    className="w-[var(--radix-popover-trigger-width)] p-2"
                  >
                    <div className="grid max-h-60 gap-1 overflow-y-auto">
                      {remainingFields.map((item) => (
                        <div
                          key={item.key}
                          className="flex min-h-9 items-center gap-2 rounded-md px-2 hover:bg-accent"
                        >
                          <Checkbox
                            aria-label={fieldLabel(t, item.key)}
                            checked={fieldKeys.includes(item.key)}
                            onCheckedChange={(checked) =>
                              toggleField(item.key, checked === true)
                            }
                          />
                          <span className="min-w-0 truncate text-sm">
                            {fieldLabel(t, item.key)}
                          </span>
                        </div>
                      ))}
                    </div>
                  </PopoverContent>
                </Popover>
              </FormField>
              <Button
                type="button"
                variant="outline"
                className="shrink-0 self-end"
                disabled={fieldKeys.length === 0}
                onClick={addField}
              >
                <Plus data-icon="inline-start" aria-hidden />
                {t("action.add")}
              </Button>
            </div>
          ) : null}
          <div className="grid w-full grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-end gap-2">
            <FormField
              label={t("alertEvent.addLabelColumn")}
              className="min-w-0"
            >
              <Input
                className="w-full"
                value={labelKey}
                placeholder={t("alertEvent.columnLabelKeyPlaceholder")}
                aria-label={t("alertEvent.columnLabelKey")}
                onChange={(event) => setLabelKey(event.target.value)}
              />
            </FormField>
            <FormField label={t("alertEvent.columnAlias")} className="min-w-0">
              <Input
                className="w-full"
                value={labelAlias}
                placeholder={t("alertEvent.columnAliasPlaceholder")}
                aria-label={t("alertEvent.columnAlias")}
                onChange={(event) => setLabelAlias(event.target.value)}
              />
            </FormField>
            <Button
              type="button"
              variant="outline"
              className="shrink-0 self-end"
              onClick={addLabel}
            >
              <Plus data-icon="inline-start" aria-hidden />
              {t("action.add")}
            </Button>
          </div>
          {localError || error ? (
            <p className="text-sm text-destructive">{localError || error}</p>
          ) : null}
        </div>
        <DialogFooter className="shrink-0">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            {t("action.cancel")}
          </Button>
          <Button
            type="button"
            disabled={pending}
            onClick={() => onSave(draft)}
          >
            {t("action.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
