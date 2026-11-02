import { useState } from "react"
import { useTranslation } from "react-i18next"

import { createStatusPage, updateStatusPage } from "@/api/status-page"
import type {
  SaveStatusPageParams,
  StatusPageItem,
} from "@/api/status-page/types"
import { FormField } from "@/components/form-field"
import { MultiSelect, type MultiSelectOption } from "@/components/multi-select"
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

// 与后端 proto / biz 的校验上限保持一致；前端只做即时提示，服务端仍是最终判定。
const NAME_MAX = 64
const TITLE_MAX = 128
const DESCRIPTION_MAX = 1024
const MAX_GROUPS = 50

interface StatusPageDialogProps {
  open: boolean
  /** null 表示新建；否则编辑该状态页。 */
  page: StatusPageItem | null
  /** 可选的探测目标组（已按租户过滤）。 */
  groups: MultiSelectOption[]
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}

/**
 * 新建 / 编辑状态页。父组件每次打开换 key 重新挂载，初始值直接来自当前编辑对象。
 * 目标组至少一个，且只能从当前租户已加载的组里选择（后端会再次校验归属）。
 */
export function StatusPageDialog({
  open,
  page,
  groups,
  onOpenChange,
  onSaved,
}: StatusPageDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(() => page?.name ?? "")
  const [title, setTitle] = useState(() => page?.title ?? "")
  const [description, setDescription] = useState(() => page?.description ?? "")
  const [targetGroupUids, setTargetGroupUids] = useState<string[]>(
    () => page?.targetGroupUids ?? []
  )
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState(false)

  function selectionSummary(selected: number, total: number): string {
    return selected >= total && total > 0
      ? t("multiSelect.allSelected", { count: selected })
      : t("multiSelect.summary", { count: selected, total })
  }

  function validate(): FieldErrors {
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("statusPage.errors.nameRequired"))],
      ["name", maxLength(name, NAME_MAX, t("statusPage.errors.nameTooLong"))],
      ["title", requireTrimmed(title, t("statusPage.errors.titleRequired"))],
      [
        "title",
        maxLength(title, TITLE_MAX, t("statusPage.errors.titleTooLong")),
      ],
      [
        "description",
        maxLength(
          description,
          DESCRIPTION_MAX,
          t("statusPage.errors.descriptionTooLong")
        ),
      ],
      [
        "targetGroupUids",
        targetGroupUids.length === 0
          ? t("statusPage.errors.groupsRequired")
          : null,
      ],
      [
        "targetGroupUids",
        targetGroupUids.length > MAX_GROUPS
          ? t("statusPage.errors.groupsTooMany")
          : null,
      ],
    ])
    return errors
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = validate()
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    const payload: SaveStatusPageParams = {
      name: name.trim(),
      title: title.trim(),
      description: description.trim(),
      targetGroupUids,
    }
    setPending(true)
    setFormError("")
    setFieldErrors({})
    try {
      if (page) {
        await updateStatusPage({ ...payload, uid: page.uid })
      } else {
        await createStatusPage(payload)
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
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {page
              ? t("statusPage.form.editTitle")
              : t("statusPage.form.createTitle")}
          </DialogTitle>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSave}
        >
          <DialogBody className="grid gap-4">
            <FormField
              label={t("statusPage.name")}
              htmlFor="status-page-name"
              hint={t("statusPage.nameHint")}
              required
              error={fieldErrors.name}
            >
              <Input
                id="status-page-name"
                value={name}
                maxLength={NAME_MAX}
                placeholder={t("statusPage.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("statusPage.titleLabel")}
              htmlFor="status-page-title"
              hint={t("statusPage.titleHint")}
              required
              error={fieldErrors.title}
            >
              <Input
                id="status-page-title"
                value={title}
                maxLength={TITLE_MAX}
                placeholder={t("statusPage.titlePlaceholder")}
                onChange={(event) => setTitle(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("statusPage.descriptionLabel")}
              htmlFor="status-page-description"
              error={fieldErrors.description}
            >
              <Textarea
                id="status-page-description"
                value={description}
                className="min-h-20"
                placeholder={t("statusPage.descriptionPlaceholder")}
                onChange={(event) => setDescription(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("statusPage.targetGroups")}
              htmlFor="status-page-groups"
              hint={t("statusPage.targetGroupsHint")}
              required
              error={fieldErrors.targetGroupUids}
            >
              <MultiSelect
                id="status-page-groups"
                values={targetGroupUids}
                options={groups}
                onChange={setTargetGroupUids}
                placeholder={t("multiSelect.placeholder")}
                searchPlaceholder={t("multiSelect.searchPlaceholder")}
                emptyText={t("statusPage.noTargetGroups")}
                allLabel={t("multiSelect.all")}
                clearLabel={t("multiSelect.clear")}
                selectedLabel={selectionSummary}
                invalid={Boolean(fieldErrors.targetGroupUids)}
              />
            </FormField>
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
