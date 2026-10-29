import { useState } from "react"
import { useTranslation } from "react-i18next"

import { updateSystemNamespaceStatus } from "@/api/system-namespace"
import type { SystemNamespaceItem } from "@/api/system-namespace/types"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { collectFieldErrors, hasFieldErrors } from "@/lib/form"
import { readFormError } from "@/lib/form-error"

const REASON_MIN = 1
const REASON_MAX = 500

/** 停用原因按字符数校验（emoji 不重复计数）。 */
function reasonLength(value: string): number {
  return Array.from(value).length
}

interface TenantStatusDialogProps {
  /** 要启停的租户；为空时关闭。 */
  tenant: SystemNamespaceItem | null
  /** disable 需要填写原因并邮件通知负责人；enable 直接确认。 */
  action: "disable" | "enable"
  onOpenChange: (open: boolean) => void
  onSaved: (uid: string) => void
}

/**
 * 租户启用 / 停用确认。停用必须填写 1-500 字原因，后端会邮件通知租户负责人；
 * 启用不要求原因，直接二次确认。
 */
export function TenantStatusDialog({
  tenant,
  action,
  onOpenChange,
  onSaved,
}: TenantStatusDialogProps) {
  const { t } = useTranslation()
  const [reason, setReason] = useState("")
  const [fieldError, setFieldError] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  // 每次打开都是一次新的启停操作：父组件用 key 重新挂载，状态从空原因开始。
  if (!tenant) {
    return null
  }

  const target = tenant
  const disabling = action === "disable"
  const length = reasonLength(reason.trim())

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (disabling) {
      const errors = collectFieldErrors([
        [
          "reason",
          length < REASON_MIN || length > REASON_MAX
            ? t("tenantAdmin.reasonInvalid", { max: REASON_MAX })
            : null,
        ],
      ])
      if (hasFieldErrors(errors)) {
        setFieldError(errors.reason ?? "")
        return
      }
    }
    setPending(true)
    setError("")
    setFieldError("")
    try {
      await updateSystemNamespaceStatus(target.uid, {
        status: disabling ? "DISABLED" : "ACTIVE",
        ...(disabling ? { reason: reason.trim() } : {}),
      })
      onSaved(target.uid)
      onOpenChange(false)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldError(failure.fields.reason ?? "")
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open={tenant !== null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>
            {disabling
              ? t("tenantAdmin.disableTitle", { name: tenant.name })
              : t("tenantAdmin.enableTitle", { name: tenant.name })}
          </DialogTitle>
          <DialogDescription>
            {disabling
              ? t("tenantAdmin.disableDescription")
              : t("tenantAdmin.enableDescription")}
          </DialogDescription>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSubmit}
        >
          <DialogBody className="grid gap-3">
            {disabling ? (
              <div className="grid min-w-0 gap-2">
                <label
                  htmlFor="tenant-disable-reason"
                  className="flex items-center gap-1 text-sm font-medium"
                >
                  <span aria-hidden="true" className="text-destructive">
                    *
                  </span>
                  {t("tenantAdmin.reason")}
                </label>
                <Textarea
                  id="tenant-disable-reason"
                  value={reason}
                  rows={4}
                  maxLength={REASON_MAX}
                  placeholder={t("tenantAdmin.reasonPlaceholder")}
                  aria-invalid={fieldError ? true : undefined}
                  aria-describedby={
                    fieldError ? "tenant-disable-reason-error" : undefined
                  }
                  onChange={(event) => setReason(event.target.value)}
                />
                <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
                  <span>{t("tenantAdmin.reasonHint", { max: REASON_MAX })}</span>
                  <span className="tabular-nums">
                    {t("tenantAdmin.reasonCount", {
                      count: length,
                      max: REASON_MAX,
                    })}
                  </span>
                </div>
                {fieldError ? (
                  <p
                    id="tenant-disable-reason-error"
                    role="alert"
                    className="text-sm text-destructive"
                  >
                    {fieldError}
                  </p>
                ) : null}
                <p className="rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-xs break-words text-warning">
                  {t("tenantAdmin.disableNotice", {
                    name: tenant.ownerName || tenant.ownerEmail || tenant.name,
                  })}
                </p>
              </div>
            ) : (
              <p className="text-sm break-words text-muted-foreground">
                {t("tenantAdmin.enableHint", { name: tenant.name })}
              </p>
            )}
          </DialogBody>
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button
              type="submit"
              variant={disabling ? "destructive" : "default"}
              disabled={pending || (disabling && length === 0)}
            >
              {pending
                ? t("action.saving")
                : disabling
                  ? t("tenantAdmin.disableConfirm")
                  : t("tenantAdmin.enableConfirm")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
