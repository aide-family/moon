import { useState } from "react"
import { useTranslation } from "react-i18next"

import { CodeBlock } from "@/components/code-block"
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
import { Textarea } from "@/components/ui/textarea"

interface SaveTemplateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  statement: string
  onSubmit: (input: { name: string; description: string }) => void
}

const MAX_NAME = 32

// 保存当前语句为模板：名称必填，说明可选，语句本身只读预览。
export function SaveTemplateDialog({
  open,
  onOpenChange,
  statement,
  onSubmit,
}: SaveTemplateDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [error, setError] = useState("")

  function close() {
    onOpenChange(false)
    setName("")
    setDescription("")
    setError("")
  }

  function submit() {
    const trimmed = name.trim()
    if (!trimmed) {
      setError(t("form.required"))
      return
    }
    onSubmit({ name: trimmed, description: description.trim() })
    close()
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) {
          close()
          return
        }
        onOpenChange(true)
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("datasource.query.saveTemplateTitle")}</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4">
          <FormField
            label={t("datasource.query.templateName")}
            htmlFor="query-template-name"
            required
            error={error}
          >
            <Input
              id="query-template-name"
              value={name}
              maxLength={MAX_NAME}
              placeholder={t("datasource.query.templateNamePlaceholder")}
              aria-invalid={error ? true : undefined}
              onChange={(event) => {
                setName(event.target.value)
                if (error) {
                  setError("")
                }
              }}
            />
          </FormField>
          <FormField
            label={t("datasource.query.templateDescription")}
            htmlFor="query-template-description"
            hint={t("datasource.query.templateDescriptionHint")}
          >
            <Textarea
              id="query-template-description"
              value={description}
              rows={2}
              placeholder={t("datasource.query.templateDescriptionPlaceholder")}
              onChange={(event) => setDescription(event.target.value)}
            />
          </FormField>
          <CodeBlock
            code={statement}
            language="text"
            label={t("datasource.query.statement")}
            copyable
            wrap
            maxHeightClass="max-h-32"
          />
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={close}>
            {t("action.cancel")}
          </Button>
          <Button type="button" onClick={submit}>
            {t("action.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
