import { Plus, X } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

/** 一组待保存的标签键值。空键空值的行视为未填写的占位行。 */
export interface LabelRow {
  key: string
  value: string
}

export type LabelRowError =
  | "keyRequired"
  | "keyInvalid"
  | "valueInvalid"
  | "duplicateKey"

export type LabelFormError = "tooMany"

export interface LabelValidation {
  rowErrors: (LabelRowError | undefined)[]
  formError: LabelFormError | undefined
}

/** 与后端约束一致：键 1–63、值 0–63，只允许字母、数字、点、下划线和连字符。 */
const KEY_PATTERN = /^[A-Za-z0-9._-]{1,63}$/
const VALUE_PATTERN = /^[A-Za-z0-9._-]{0,63}$/

const MAX_LABELS = 32

const LABEL_ERROR_KEYS: Record<LabelRowError | LabelFormError, string> = {
  keyRequired: "label.keyRequired",
  keyInvalid: "label.keyInvalid",
  valueInvalid: "label.valueInvalid",
  duplicateKey: "label.duplicateKey",
  tooMany: "label.tooMany",
}

/** 校验错误码 → i18n key，编辑器和提交表单共用同一套文案。 */
export function labelErrorKey(error: LabelRowError | LabelFormError): string {
  return LABEL_ERROR_KEYS[error]
}

function isBlankRow(row: LabelRow): boolean {
  return row.key === "" && row.value === ""
}

/** 逐行校验：空键、非法字符、重复键分别报错；超过 32 对时给出表单级错误。 */
export function validateLabelRows(rows: LabelRow[]): LabelValidation {
  const seen = new Map<string, number>()
  for (const row of rows) {
    if (isBlankRow(row) || row.key === "") {
      continue
    }
    seen.set(row.key, (seen.get(row.key) ?? 0) + 1)
  }
  const rowErrors = rows.map((row): LabelRowError | undefined => {
    if (isBlankRow(row)) {
      return undefined
    }
    if (row.key === "") {
      return "keyRequired"
    }
    if (!KEY_PATTERN.test(row.key)) {
      return "keyInvalid"
    }
    if ((seen.get(row.key) ?? 0) > 1) {
      return "duplicateKey"
    }
    if (!VALUE_PATTERN.test(row.value)) {
      return "valueInvalid"
    }
    return undefined
  })
  const filled = rows.filter((row) => !isBlankRow(row)).length
  return {
    rowErrors,
    formError: filled > MAX_LABELS ? "tooMany" : undefined,
  }
}

/**
 * 把编辑行转成提交用的 JSON 文本。
 * 任何一行不合法或键为空时返回具体错误码，调用方据此阻止提交并显示原因。
 */
export function labelsToJSON(rows: LabelRow[]): {
  json: string
  labels: Record<string, string>
  error: LabelRowError | LabelFormError | undefined
} {
  const validation = validateLabelRows(rows)
  const firstRowError = validation.rowErrors.find(
    (error): error is LabelRowError => error !== undefined
  )
  const labels: Record<string, string> = {}
  for (const row of rows) {
    if (isBlankRow(row) || row.key === "") {
      continue
    }
    labels[row.key] = row.value
  }
  return {
    json: JSON.stringify(labels),
    labels,
    error: validation.formError ?? firstRowError,
  }
}

/** 解析机器返回的 labels JSON 文本；无法解析时按无标签处理，避免编辑表单打不开。 */
export function labelsFromJSON(raw: string): LabelRow[] {
  try {
    const parsed: unknown = JSON.parse(raw || "{}")
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return []
    }
    return Object.entries(parsed as Record<string, unknown>).map(
      ([key, value]) => ({
        key,
        value: typeof value === "string" ? value : String(value),
      })
    )
  } catch {
    return []
  }
}

/** 标签展示用：解析成键值对并按 key 排序；非法 JSON 返回 null。 */
export function readLabelEntries(
  raw: string
): Array<{ key: string; value: string }> | null {
  try {
    const parsed: unknown = JSON.parse(raw || "{}")
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return null
    }
    return Object.entries(parsed as Record<string, unknown>)
      .filter(([key]) => key !== "")
      .map(([key, value]) => ({
        key,
        value: typeof value === "string" ? value : String(value),
      }))
      .sort((left, right) => left.key.localeCompare(right.key))
  } catch {
    return null
  }
}

interface LabelEditorProps {
  /** 当前编辑的键值行；空数组时渲染一行占位输入。 */
  labels: LabelRow[]
  onChange: (labels: LabelRow[]) => void
  /** 提交时由父级传入的外部错误（例如后端 labels 字段校验失败）。 */
  error?: string
  disabled?: boolean
  /** 多个编辑器出现在同一表单时区分 input id。 */
  idPrefix?: string
}

/**
 * 键值标签编辑器：逐行编辑，支持增删，重复键与非法字符就地提示。
 * 只负责编辑和行内校验，提交前的整体校验由父级调用 labelsToJSON 完成。
 */
export function LabelEditor({
  labels,
  onChange,
  error,
  disabled = false,
  idPrefix = "label",
}: LabelEditorProps) {
  const { t } = useTranslation()
  const rows = labels.length > 0 ? labels : [{ key: "", value: "" }]
  const validation = validateLabelRows(rows)
  const filled = rows.filter((row) => !isBlankRow(row)).length
  const canAdd = filled < MAX_LABELS

  function updateRow(index: number, patch: Partial<LabelRow>) {
    onChange(rows.map((row, rowIndex) =>
      rowIndex === index ? { ...row, ...patch } : row
    ))
  }

  function removeRow(index: number) {
    onChange(rows.filter((_, rowIndex) => rowIndex !== index))
  }

  function addRow() {
    onChange([...rows, { key: "", value: "" }])
  }

  return (
    <div className="grid min-w-0 gap-2">
      <div className="grid min-w-0 gap-2">
        {rows.map((row, index) => {
          const rowError = validation.rowErrors[index]
          return (
            <div key={index} className="grid min-w-0 gap-1">
              <div className="flex min-w-0 items-center gap-2">
                <Input
                  id={`${idPrefix}-key-${index}`}
                  className="min-w-0 flex-1 font-mono text-xs"
                  value={row.key}
                  disabled={disabled}
                  aria-label={t("label.key")}
                  aria-invalid={rowError ? true : undefined}
                  placeholder={t("label.keyPlaceholder")}
                  autoComplete="off"
                  onChange={(event) =>
                    updateRow(index, { key: event.target.value })
                  }
                />
                <span aria-hidden="true" className="text-muted-foreground">
                  =
                </span>
                <Input
                  id={`${idPrefix}-value-${index}`}
                  className="min-w-0 flex-1 font-mono text-xs"
                  value={row.value}
                  disabled={disabled}
                  aria-label={t("label.value")}
                  aria-invalid={
                    rowError === "valueInvalid" ? true : undefined
                  }
                  placeholder={t("label.valuePlaceholder")}
                  autoComplete="off"
                  onChange={(event) =>
                    updateRow(index, { value: event.target.value })
                  }
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  disabled={disabled}
                  aria-label={t("label.remove")}
                  onClick={() => removeRow(index)}
                >
                  <X aria-hidden="true" />
                </Button>
              </div>
              {rowError ? (
                <p role="alert" className="text-xs text-destructive">
                  {t(labelErrorKey(rowError))}
                </p>
              ) : null}
            </div>
          )
        })}
      </div>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={disabled || !canAdd}
          onClick={addRow}
        >
          <Plus aria-hidden="true" data-icon="inline-start" />
          {t("label.add")}
        </Button>
        <span className="text-xs text-muted-foreground">
          {t("label.countHint", { count: filled, max: MAX_LABELS })}
        </span>
      </div>
      {validation.formError ? (
        <p role="alert" className="text-sm text-destructive">
          {t(labelErrorKey(validation.formError))}
        </p>
      ) : null}
      {error ? (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}
    </div>
  )
}
