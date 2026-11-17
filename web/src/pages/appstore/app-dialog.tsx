import { useState } from "react"
import { useTranslation } from "react-i18next"

import { createApplication, updateApplication } from "@/api/application"
import type {
  ApplicationCategory,
  ApplicationItem,
  SaveApplicationParams,
} from "@/api/application/types"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { checkField } from "@/lib/form"
import { readFormError } from "@/lib/form-error"

const NAME_MAX = 64
const DISPLAY_MAX = 128
const DESCRIPTION_MAX = 1024
const README_MAX = 65536

const CATEGORIES: ApplicationCategory[] = [
  "DATABASE",
  "MIDDLEWARE",
  "OBSERVABILITY",
  "TOOLING",
  "OTHER",
]

interface ApplicationDialogProps {
  open: boolean
  application: ApplicationItem | null
  onOpenChange: (open: boolean) => void
  onSaved: (item: ApplicationItem) => void
}

export function ApplicationDialog({
  open,
  application,
  onOpenChange,
  onSaved,
}: ApplicationDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(() => application?.name ?? "")
  const [displayName, setDisplayName] = useState(() => application?.displayName ?? "")
  const [category, setCategory] = useState<ApplicationCategory>(() => application?.category ?? "TOOLING")
  const [description, setDescription] = useState(() => application?.description ?? "")
  const [readme, setReadme] = useState(() => application?.readme ?? "")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")
  const [fieldError, setFieldError] = useState("")

  async function save() {
    const nameError = checkField({
      label: t("appStore.appName"),
      value: name,
      required: !application,
      max: 64,
    })
    if (nameError) {
      setFieldError(nameError)
      return
    }
    setSaving(true)
    setError("")
    setFieldError("")
    try {
      const body: SaveApplicationParams = {
        name: name.trim(),
        displayName: displayName.trim(),
        category,
        description: description.trim(),
        readme,
      }
      const result = application
        ? await updateApplication({ ...body, uid: application.uid })
        : await createApplication(body)
      onSaved(result.application)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      if (failure.fields.name) {
        setFieldError(failure.fields.name)
        setError("")
      } else {
        setFieldError("")
        setError(failure.form)
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {application ? t("appStore.editApp") : t("appStore.newApp")}
          </DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-4 sm:grid-cols-2">
          <FormField
            label={t("appStore.appName")}
            htmlFor="app-name"
            required={!application}
            error={fieldError}
            hint={application ? t("appStore.appNameImmutable") : undefined}
            description={t("appStore.appNameHelp")}
            counter={{ value: name.length, max: NAME_MAX }}
          >
            <Input
              id="app-name"
              className="font-mono"
              value={name}
              disabled={Boolean(application)}
              maxLength={NAME_MAX}
              placeholder="order-gateway"
              onChange={(event) => setName(event.target.value)}
            />
          </FormField>
          <FormField label={t("appStore.displayName")} htmlFor="app-display-name" description={t("appStore.displayNameHelp")}>
            <Input
              id="app-display-name"
              value={displayName}
              maxLength={DISPLAY_MAX}
              placeholder={t("appStore.displayNamePlaceholder")}
              onChange={(event) => setDisplayName(event.target.value)}
            />
          </FormField>
          <FormField label={t("appStore.category")} htmlFor="app-category" description={t("appStore.categoryHelp")}>
            <Select value={category} onValueChange={(value) => setCategory(value as ApplicationCategory)}>
              <SelectTrigger id="app-category"><SelectValue /></SelectTrigger>
              <SelectContent>
                {CATEGORIES.map((item) => (
                  <SelectItem key={item} value={item}>{t(`appStore.categories.${item}`)}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </FormField>
          <FormField label={t("appStore.descriptionLabel")} htmlFor="app-description" description={t("appStore.descriptionHelp")}>
            <Input
              id="app-description"
              value={description}
              maxLength={DESCRIPTION_MAX}
              placeholder={t("appStore.descriptionPlaceholder")}
              onChange={(event) => setDescription(event.target.value)}
            />
          </FormField>
          <FormField label={t("appStore.readme")} htmlFor="app-readme" className="sm:col-span-full" description={t("appStore.readmeHelp")}>
            <Textarea
              id="app-readme"
              className="min-h-32"
              value={readme}
              maxLength={README_MAX}
              placeholder={t("appStore.readmePlaceholder")}
              onChange={(event) => setReadme(event.target.value)}
            />
          </FormField>
          {error ? <p className="text-sm text-destructive sm:col-span-full">{error}</p> : null}
        </DialogBody>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button disabled={saving} onClick={() => void save()}>
            {saving ? t("action.saving") : t("action.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
