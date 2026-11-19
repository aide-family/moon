import { useState } from "react"
import { useTranslation } from "react-i18next"

import { duplicateApplicationVersion } from "@/api/application"
import type { ApplicationVersionItem } from "@/api/application/types"
import { ConfirmAction } from "@/components/confirm-action"
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
import { readFormError } from "@/lib/form-error"

interface ApplicationDuplicateVersionDialogProps {
  open: boolean
  applicationUid: string
  source: ApplicationVersionItem | null
  onOpenChange: (open: boolean) => void
  onDuplicated: (version: ApplicationVersionItem) => void
}

/** 复制版本：把已上架/已下架版本的内容复制成新的草稿版本。 */
export function ApplicationDuplicateVersionDialog({
  open,
  applicationUid,
  source,
  onOpenChange,
  onDuplicated,
}: ApplicationDuplicateVersionDialogProps) {
  const { t } = useTranslation()
  const [version, setVersion] = useState("")
  const [changelog, setChangelog] = useState("")
  const [versionError, setVersionError] = useState("")
  const [formError, setFormError] = useState("")
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pending, setPending] = useState(false)

  function submitForm() {
    const error = checkField({
      label: t("appStore.version"),
      value: version,
      required: true,
      max: 32,
    })
    if (error) {
      setVersionError(error)
      return
    }
    setVersionError("")
    setFormError("")
    setConfirmOpen(true)
  }

  async function duplicate() {
    if (!source) {
      return
    }
    setPending(true)
    setFormError("")
    try {
      const result = await duplicateApplicationVersion({
        applicationUid,
        uid: source.uid,
        version: version.trim(),
        changelog: changelog.trim(),
      })
      setConfirmOpen(false)
      setVersion("")
      setChangelog("")
      onDuplicated(result.version)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setConfirmOpen(false)
      setVersionError(failure.fields.version ?? "")
      setFormError(failure.form)
    } finally {
      setPending(false)
    }
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t("appStore.duplicateTitle", { version: source?.version ?? "" })}</DialogTitle>
          </DialogHeader>
          <DialogBody className="grid gap-4">
            <p className="text-sm text-muted-foreground">{t("appStore.duplicateHint")}</p>
            <FormField
              label={t("appStore.version")}
              htmlFor="duplicate-version"
              required
              error={versionError}
              hint={t("appStore.versionHelp")}
            >
              <Input
                id="duplicate-version"
                className="font-mono"
                value={version}
                placeholder={t("appStore.versionPlaceholder")}
                onChange={(event) => setVersion(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("appStore.changelog")}
              htmlFor="duplicate-changelog"
              hint={t("appStore.duplicateChangelogHelp")}
            >
              <Textarea
                id="duplicate-changelog"
                value={changelog}
                placeholder={t("appStore.changelogPlaceholder")}
                onChange={(event) => setChangelog(event.target.value)}
              />
            </FormField>
            {formError ? <p className="text-sm text-destructive">{formError}</p> : null}
          </DialogBody>
          <DialogFooter>
            <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
            <Button onClick={submitForm}>{t("appStore.copyVersion")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <ConfirmAction
        open={confirmOpen}
        title={t("appStore.duplicateConfirmTitle")}
        description={t("appStore.duplicateConfirmDescription", { source: source?.version ?? "", version: version.trim() })}
        confirmLabel={t("appStore.copyVersion")}
        pendingLabel={t("action.saving")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={formError}
        onOpenChange={setConfirmOpen}
        onConfirm={() => void duplicate()}
      />
    </>
  )
}
