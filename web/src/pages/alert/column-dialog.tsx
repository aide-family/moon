import { ArrowDown, ArrowUp, Plus, Trash2 } from "lucide-react"
import { useMemo, useState } from "react"
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
import { FIELD_COLUMNS, fieldLabel } from "@/pages/alert/columns"

interface AlertColumnDialogProps {
  open: boolean
  columns: AlertColumnItem[]
  pending?: boolean
  error?: string
  onSave: (columns: AlertColumnItem[]) => void
  onOpenChange: (open: boolean) => void
}

const NONE = "__none__"

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

  const used = useMemo(
    () => new Set(draft.map((item) => `${item.type}:${item.key}`)),
    [draft]
  )
  const remainingFields = FIELD_COLUMNS.filter(
    (item) => !used.has(`FIELD:${item.key}`)
  )

  function move(index: number, delta: number) {
    const next = [...draft]
    const target = index + delta
    if (target < 0 || target >= next.length) {
      return
    }
    ;[next[index], next[target]] = [next[target], next[index]]
    setDraft(next)
  }

  function addField() {
    if (!fieldKey || used.has(`FIELD:${fieldKey}`)) {
      return
    }
    const nextUsed = new Set([...used, `FIELD:${fieldKey}`])
    const nextKey = FIELD_COLUMNS.find(
      (item) => !nextUsed.has(`FIELD:${item.key}`)
    )?.key
    setDraft([...draft, { key: fieldKey, type: "FIELD", label: "" }])
    setFieldKey(nextKey ?? "")
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
        <div className="grid min-h-0 flex-1 content-start gap-2 overflow-y-auto pr-1">
          {draft.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {t("alertEvent.columnsEmpty")}
            </p>
          ) : null}
          {draft.map((column, index) => (
            <div
              key={`${column.type}:${column.key}`}
              className="flex items-center gap-2 rounded-lg border p-2"
            >
              <span className="min-w-0 flex-1 truncate font-mono text-xs">
                {column.type === "FIELD"
                  ? fieldLabel(t, column.key)
                  : column.key}
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
                disabled={index === 0}
                aria-label={t("alertEvent.columnMoveUp")}
                onClick={() => move(index, -1)}
              >
                <ArrowUp aria-hidden />
              </Button>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                disabled={index === draft.length - 1}
                aria-label={t("alertEvent.columnMoveDown")}
                onClick={() => move(index, 1)}
              >
                <ArrowDown aria-hidden />
              </Button>
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
          ))}
        </div>
        <div className="grid shrink-0 gap-3 border-t pt-3">
          <div className="flex flex-wrap items-end gap-2">
            <FormField label={t("alertEvent.addFieldColumn")}>
              <Select value={fieldKey} onValueChange={setFieldKey}>
                <SelectTrigger className="w-48" aria-label={t("alertEvent.addFieldColumn")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {remainingFields.length === 0 ? (
                    <SelectItem value={NONE} disabled>
                      {t("alertEvent.columnsAllFields")}
                    </SelectItem>
                  ) : (
                    remainingFields.map((item) => (
                      <SelectItem key={item.key} value={item.key}>
                        {t(item.labelKey)}
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
            </FormField>
            <Button
              type="button"
              variant="outline"
              disabled={remainingFields.length === 0 || used.has(`FIELD:${fieldKey}`)}
              onClick={addField}
            >
              <Plus aria-hidden />
              {t("action.add")}
            </Button>
          </div>
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
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
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
