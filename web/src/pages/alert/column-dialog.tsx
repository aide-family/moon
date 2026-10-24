import { GripVertical, Plus, Trash2 } from "lucide-react"
import { Fragment, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import type { AlertColumnItem } from "@/api/alert-page/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
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
  const [fieldKey, setFieldKey] = useState(
    FIELD_COLUMNS.find(
      (item) =>
        !columns.some(
          (column) => column.type === "FIELD" && column.key === item.key
        )
    )?.key ?? ""
  )
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
  // 用户选中的字段被添加或被移出后，自动回落到第一个仍可添加的字段。
  const effectiveFieldKey =
    fieldKey && remainingFields.some((item) => item.key === fieldKey)
      ? fieldKey
      : (remainingFields[0]?.key ?? "")

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
    if (!effectiveFieldKey || used.has(`FIELD:${effectiveFieldKey}`)) {
      return
    }
    setDraft([...draft, { key: effectiveFieldKey, type: "FIELD", label: "" }])
    setFieldKey("")
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
            <div className="flex flex-wrap items-end gap-2">
              <FormField label={t("alertEvent.addFieldColumn")}>
                <Select value={effectiveFieldKey} onValueChange={setFieldKey}>
                  <SelectTrigger
                    className="w-48"
                    aria-label={t("alertEvent.addFieldColumn")}
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {remainingFields.map((item) => (
                      <SelectItem key={item.key} value={item.key}>
                        {t(item.labelKey)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
              <Button
                type="button"
                variant="outline"
                disabled={!effectiveFieldKey}
                onClick={addField}
              >
                <Plus aria-hidden />
                {t("action.add")}
              </Button>
            </div>
          ) : null}
          <div className="flex flex-wrap items-end gap-2">
            <FormField label={t("alertEvent.addLabelColumn")}>
              <Input
                className="w-48"
                value={labelKey}
                placeholder={t("alertEvent.columnLabelKeyPlaceholder")}
                aria-label={t("alertEvent.columnLabelKey")}
                onChange={(event) => setLabelKey(event.target.value)}
              />
            </FormField>
            <FormField label={t("alertEvent.columnAlias")}>
              <Input
                className="w-48"
                value={labelAlias}
                placeholder={t("alertEvent.columnAliasPlaceholder")}
                aria-label={t("alertEvent.columnAlias")}
                onChange={(event) => setLabelAlias(event.target.value)}
              />
            </FormField>
            <Button type="button" variant="outline" onClick={addLabel}>
              <Plus aria-hidden />
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
