import { useState } from "react"
import { useTranslation } from "react-i18next"

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
import { checkField } from "@/lib/form"

const DESCRIPTION_MAX = 2048

interface ApplicationComponentDialogProps {
  open: boolean
  mode: "create" | "edit"
  initialName?: string
  initialDescription?: string
  pending?: boolean
  error?: string
  onOpenChange: (open: boolean) => void
  onSave: (name: string, description: string) => void
}

/** 新增/编辑服务（组件）：服务名 + 支持 Markdown 的备注。 */
export function ApplicationComponentDialog({
  open,
  mode,
  initialName = "",
  initialDescription = "",
  pending = false,
  error,
  onOpenChange,
  onSave,
}: ApplicationComponentDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(initialName)
  const [description, setDescription] = useState(initialDescription)
  const [fieldError, setFieldError] = useState("")
  const [descriptionError, setDescriptionError] = useState("")

  function save() {
    const nameError = checkField({
      label: t("appStore.resourceForm.component"),
      value: name,
      required: true,
      max: 63,
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
      hint: t("appStore.resourceForm.componentHelp"),
    })
    if (nameError) {
      setFieldError(nameError)
      return
    }
    const nextDescriptionError = checkField({
      label: t("appStore.services.descriptionLabel"),
      value: description,
      max: DESCRIPTION_MAX,
    })
    if (nextDescriptionError) {
      setDescriptionError(nextDescriptionError)
      return
    }
    setFieldError("")
    setDescriptionError("")
    onSave(name.trim(), description.trim())
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {mode === "edit" ? t("appStore.services.renameTitle") : t("appStore.services.createTitle")}
          </DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-4">
          <p className="text-sm text-muted-foreground">
            {mode === "edit" ? t("appStore.services.renameHint") : t("appStore.services.createHint")}
          </p>
          <FormField
            label={t("appStore.resourceForm.component")}
            htmlFor="component-name"
            required
            error={fieldError || error}
            hint={t("appStore.resourceForm.componentHelp")}
          >
            <Input
              id="component-name"
              className="font-mono"
              value={name}
              placeholder="api"
              onChange={(event) => setName(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault()
                  save()
                }
              }}
            />
          </FormField>
          <FormField
            label={t("appStore.services.descriptionLabel")}
            htmlFor="component-description"
            error={descriptionError}
            hint={t("appStore.services.descriptionHelp")}
            counter={{ value: description.length, max: DESCRIPTION_MAX }}
          >
            <Textarea
              id="component-description"
              className="min-h-24"
              value={description}
              maxLength={DESCRIPTION_MAX}
              placeholder={t("appStore.services.descriptionPlaceholder")}
              onChange={(event) => setDescription(event.target.value)}
            />
          </FormField>
        </DialogBody>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button disabled={pending} onClick={save}>{pending ? t("action.saving") : t("action.save")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
