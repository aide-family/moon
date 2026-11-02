import { ArrowDown, ArrowUp, Plus, Trash2 } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { createBoard, updateBoard } from "@/api/board"
import type {
  BoardItem,
  BoardPanelInput,
  SaveBoardParams,
} from "@/api/board/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  collectFieldErrors,
  hasFieldErrors,
  maxLength,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { readFormError } from "@/lib/form-error"
import { GRAFANA_URL_MAX_BYTES, validateGrafanaURL } from "@/lib/grafana-url"

// 校验上限与后端 proto / biz 保持一致；前端只做即时提示，服务端仍是最终判定。
const NAME_MAX = 64
const DESCRIPTION_MAX = 1024
const MAX_PANELS = 50
const PANEL_TITLE_MAX = 128
const PANEL_WIDTH_MIN = 1
const PANEL_WIDTH_MAX = 12
const PANEL_HEIGHT_MIN = 120
const PANEL_HEIGHT_MAX = 1600

/** 面板草稿：宽高用字符串保存，允许输入过程中为空。 */
interface PanelDraft {
  /** 仅用于 React key；不提交。 */
  key: string
  /** 编辑已有面板时由服务端返回；空表示新增。 */
  uid?: string
  title: string
  description: string
  grafanaUrl: string
  width: string
  height: string
}

let panelKeySeed = 0

function nextPanelKey(): string {
  panelKeySeed += 1
  return `board-panel-${panelKeySeed}`
}

function emptyPanel(): PanelDraft {
  return {
    key: nextPanelKey(),
    title: "",
    description: "",
    grafanaUrl: "",
    width: "6",
    height: "320",
  }
}

function toDraft(panel: BoardItem["panels"][number]): PanelDraft {
  return {
    key: nextPanelKey(),
    uid: panel.uid,
    title: panel.title,
    description: panel.description,
    grafanaUrl: panel.grafanaUrl,
    width: String(panel.width),
    height: String(panel.height),
  }
}

/** 整数并在闭区间内：空串、小数、越界都算非法。 */
function isIntegerInRange(value: string, min: number, max: number): boolean {
  const trimmed = value.trim()
  if (!/^\d+$/.test(trimmed)) {
    return false
  }
  const parsed = Number(trimmed)
  return parsed >= min && parsed <= max
}

interface BoardDialogProps {
  open: boolean
  /** null 表示新建；否则编辑这块大盘（必须带全量面板）。 */
  board: BoardItem | null
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}

/**
 * 新建 / 编辑大盘。父组件每次打开换 key 重新挂载，初始值直接来自当前编辑对象，
 * 无需在 effect 里重置表单。面板顺序就是提交顺序，服务端按数组下标重写 sort。
 */
export function BoardDialog({
  open,
  board,
  onOpenChange,
  onSaved,
}: BoardDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(() => board?.name ?? "")
  const [description, setDescription] = useState(() => board?.description ?? "")
  const [panels, setPanels] = useState<PanelDraft[]>(() =>
    (board?.panels ?? []).map(toDraft)
  )
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState(false)

  /** 面板字段错误 key：panel.<index>.<field>，与提交时的下标一一对应。 */
  function panelField(index: number, field: string): string {
    return `panel.${index}.${field}`
  }

  function urlIssueText(index: number): string | undefined {
    const panel = panels[index]
    if (!panel || panel.grafanaUrl.trim() === "") {
      return undefined
    }
    const issue = validateGrafanaURL(panel.grafanaUrl)
    return issue ? t(`board.form.urlIssue.${issue}`) : undefined
  }

  function urlError(index: number): string | undefined {
    const panel = panels[index]
    if (!panel) {
      return undefined
    }
    // 有输入时以实时校验为准，避免提交失败后旧错误一直挂在字段下面。
    if (panel.grafanaUrl.trim() !== "") {
      return urlIssueText(index)
    }
    return fieldErrors[panelField(index, "url")]
  }

  function validate(): FieldErrors {
    const entries: Array<[string, string | null]> = [
      ["name", requireTrimmed(name, t("board.errors.nameRequired"))],
      ["name", maxLength(name, NAME_MAX, t("board.errors.nameTooLong"))],
      [
        "description",
        maxLength(
          description,
          DESCRIPTION_MAX,
          t("board.errors.descriptionTooLong")
        ),
      ],
      [
        "panels",
        panels.length > MAX_PANELS ? t("board.errors.panelsTooMany") : null,
      ],
    ]
    panels.forEach((panel, index) => {
      entries.push([
        panelField(index, "title"),
        requireTrimmed(panel.title, t("board.errors.panelTitleRequired")),
      ])
      entries.push([
        panelField(index, "title"),
        maxLength(
          panel.title,
          PANEL_TITLE_MAX,
          t("board.errors.panelTitleTooLong")
        ),
      ])
      entries.push([
        panelField(index, "description"),
        maxLength(
          panel.description,
          DESCRIPTION_MAX,
          t("board.errors.panelDescriptionTooLong")
        ),
      ])
      const issue = validateGrafanaURL(panel.grafanaUrl)
      entries.push([
        panelField(index, "url"),
        issue ? t(`board.form.urlIssue.${issue}`) : null,
      ])
      entries.push([
        panelField(index, "width"),
        isIntegerInRange(panel.width, PANEL_WIDTH_MIN, PANEL_WIDTH_MAX)
          ? null
          : t("board.errors.widthInvalid"),
      ])
      entries.push([
        panelField(index, "height"),
        isIntegerInRange(panel.height, PANEL_HEIGHT_MIN, PANEL_HEIGHT_MAX)
          ? null
          : t("board.errors.heightInvalid"),
      ])
    })
    return collectFieldErrors(entries)
  }

  function patchPanel(index: number, patch: Partial<PanelDraft>) {
    setPanels((current) =>
      current.map((panel, position) =>
        position === index ? { ...panel, ...patch } : panel
      )
    )
  }

  function movePanel(index: number, offset: -1 | 1) {
    setPanels((current) => {
      const target = index + offset
      if (target < 0 || target >= current.length) {
        return current
      }
      const next = [...current]
      ;[next[index], next[target]] = [next[target], next[index]]
      return next
    })
  }

  function removePanel(index: number) {
    setPanels((current) => current.filter((_, position) => position !== index))
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = validate()
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    const payload: SaveBoardParams = {
      name: name.trim(),
      description: description.trim(),
      panels: panels.map<BoardPanelInput>((panel) => ({
        ...(panel.uid ? { uid: panel.uid } : {}),
        title: panel.title.trim(),
        description: panel.description.trim(),
        grafanaUrl: panel.grafanaUrl.trim(),
        width: Number(panel.width.trim()),
        height: Number(panel.height.trim()),
      })),
    }
    setPending(true)
    setFormError("")
    setFieldErrors({})
    try {
      if (board) {
        await updateBoard({ ...payload, uid: board.uid })
      } else {
        await createBoard(payload)
      }
      onSaved()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setFormError(failure.form)
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {board ? t("board.form.editTitle") : t("board.form.createTitle")}
          </DialogTitle>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSave}
        >
          <DialogBody className="grid gap-4">
            <FormField
              label={t("board.name")}
              htmlFor="board-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="board-name"
                value={name}
                maxLength={NAME_MAX}
                placeholder={t("board.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("board.descriptionLabel")}
              htmlFor="board-description"
              error={fieldErrors.description}
            >
              <Textarea
                id="board-description"
                value={description}
                className="min-h-16"
                placeholder={t("board.descriptionPlaceholder")}
                onChange={(event) => setDescription(event.target.value)}
              />
            </FormField>
            <div className="grid gap-3">
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div className="min-w-0">
                  <p className="text-sm font-medium">
                    {t("board.form.panels")}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {t("board.form.panelsHint", {
                      count: panels.length,
                      max: MAX_PANELS,
                    })}
                  </p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={pending || panels.length >= MAX_PANELS}
                  onClick={() =>
                    setPanels((current) => [...current, emptyPanel()])
                  }
                >
                  <Plus aria-hidden="true" data-icon="inline-start" />
                  {t("board.form.addPanel")}
                </Button>
              </div>
              {fieldErrors.panels ? (
                <p role="alert" className="text-sm text-destructive">
                  {fieldErrors.panels}
                </p>
              ) : null}
              {panels.length === 0 ? (
                <p className="rounded-lg border border-dashed border-border/70 px-3 py-6 text-center text-sm text-muted-foreground">
                  {t("board.form.noPanels")}
                </p>
              ) : null}
              <div className="grid gap-3">
                {panels.map((panel, index) => (
                  <div
                    key={panel.key}
                    className="grid gap-3 rounded-lg border border-border/70 bg-muted/20 p-3"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <span className="text-xs font-medium text-muted-foreground">
                        {t("board.form.panelIndex", { index: index + 1 })}
                      </span>
                      <div className="flex items-center gap-1">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={t("board.form.moveUp")}
                          title={t("board.form.moveUp")}
                          disabled={index === 0}
                          onClick={() => movePanel(index, -1)}
                        >
                          <ArrowUp aria-hidden="true" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={t("board.form.moveDown")}
                          title={t("board.form.moveDown")}
                          disabled={index === panels.length - 1}
                          onClick={() => movePanel(index, 1)}
                        >
                          <ArrowDown aria-hidden="true" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={t("board.form.removePanel")}
                          title={t("board.form.removePanel")}
                          onClick={() => removePanel(index)}
                        >
                          <Trash2 aria-hidden="true" />
                        </Button>
                      </div>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-2">
                      <FormField
                        className="sm:col-span-2"
                        label={t("board.form.panelTitle")}
                        htmlFor={`board-panel-title-${panel.key}`}
                        required
                        error={fieldErrors[panelField(index, "title")]}
                      >
                        <Input
                          id={`board-panel-title-${panel.key}`}
                          value={panel.title}
                          maxLength={PANEL_TITLE_MAX}
                          placeholder={t("board.form.panelTitlePlaceholder")}
                          onChange={(event) =>
                            patchPanel(index, { title: event.target.value })
                          }
                        />
                      </FormField>
                      <FormField
                        className="sm:col-span-2"
                        label={t("board.form.panelUrl")}
                        htmlFor={`board-panel-url-${panel.key}`}
                        hint={t("board.form.urlHint")}
                        error={urlError(index)}
                      >
                        <Input
                          id={`board-panel-url-${panel.key}`}
                          value={panel.grafanaUrl}
                          placeholder={t("board.form.panelUrlPlaceholder")}
                          aria-invalid={urlError(index) ? true : undefined}
                          onChange={(event) =>
                            patchPanel(index, {
                              grafanaUrl: event.target.value,
                            })
                          }
                        />
                      </FormField>
                      <FormField
                        className="sm:col-span-2"
                        label={t("board.form.panelDescription")}
                        htmlFor={`board-panel-description-${panel.key}`}
                        error={fieldErrors[panelField(index, "description")]}
                      >
                        <Textarea
                          id={`board-panel-description-${panel.key}`}
                          value={panel.description}
                          className="min-h-16"
                          placeholder={t(
                            "board.form.panelDescriptionPlaceholder"
                          )}
                          onChange={(event) =>
                            patchPanel(index, {
                              description: event.target.value,
                            })
                          }
                        />
                      </FormField>
                      <FormField
                        label={t("board.form.panelWidth")}
                        htmlFor={`board-panel-width-${panel.key}`}
                        required
                        error={fieldErrors[panelField(index, "width")]}
                      >
                        <Input
                          id={`board-panel-width-${panel.key}`}
                          type="number"
                          inputMode="numeric"
                          min={PANEL_WIDTH_MIN}
                          max={PANEL_WIDTH_MAX}
                          value={panel.width}
                          onChange={(event) =>
                            patchPanel(index, { width: event.target.value })
                          }
                        />
                      </FormField>
                      <FormField
                        label={t("board.form.panelHeight")}
                        htmlFor={`board-panel-height-${panel.key}`}
                        required
                        error={fieldErrors[panelField(index, "height")]}
                      >
                        <Input
                          id={`board-panel-height-${panel.key}`}
                          type="number"
                          inputMode="numeric"
                          min={PANEL_HEIGHT_MIN}
                          max={PANEL_HEIGHT_MAX}
                          value={panel.height}
                          onChange={(event) =>
                            patchPanel(index, { height: event.target.value })
                          }
                        />
                      </FormField>
                    </div>
                  </div>
                ))}
              </div>
              <p className="text-xs text-muted-foreground">
                {t("board.form.limitsHint", {
                  url: GRAFANA_URL_MAX_BYTES,
                  title: PANEL_TITLE_MAX,
                  heightMin: PANEL_HEIGHT_MIN,
                  heightMax: PANEL_HEIGHT_MAX,
                })}
              </p>
            </div>
          </DialogBody>
          {formError ? (
            <p className="text-sm text-destructive">{formError}</p>
          ) : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => onOpenChange(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button type="submit" disabled={pending}>
              {pending ? t("action.saving") : t("action.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
