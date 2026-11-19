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
import { checkField } from "@/lib/form"

interface ApplicationComponentDialogProps {
  open: boolean
  mode: "create" | "rename"
  initialName?: string
  pending?: boolean
  error?: string
  onOpenChange: (open: boolean) => void
  onSave: (name: string) => void
}

/** 新增/重命名服务（组件）。组件名即资源的 component 字段。 */
export function ApplicationComponentDialog({
  open,
  mode,
  initialName = "",
  pending = false,
  error,
  onOpenChange,
  onSave,
}: ApplicationComponentDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(initialName)
  const [fieldError, setFieldError] = useState("")

  function save() {
    const next = checkField({
      label: t("appStore.resourceForm.component"),
      value: name,
      required: true,
      max: 63,
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
      hint: t("appStore.resourceForm.componentHelp"),
    })
    if (next) {
      setFieldError(next)
      return
    }
    setFieldError("")
    onSave(name.trim())
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {mode === "rename" ? t("appStore.services.renameTitle") : t("appStore.services.createTitle")}
          </DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-4">
          <p className="text-sm text-muted-foreground">
            {mode === "rename" ? t("appStore.services.renameHint") : t("appStore.services.createHint")}
          </p>
          <FormField
            label={t("appStore.resourceForm.component")}
            htmlFor="component-name"
            required
            error={fieldError || error}
            description={t("appStore.resourceForm.componentHelp")}
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
        </DialogBody>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button disabled={pending} onClick={save}>{pending ? t("action.saving") : t("action.save")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
