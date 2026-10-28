import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { ApiError } from "@/api/http"
import { changeEmail, sendSelfChangeEmailCode } from "@/api/self"
import { CaptchaField } from "@/components/captcha-field"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { useCaptcha } from "@/hooks/use-captcha"
import {
  RETRY_AFTER_SECONDS_KEY,
  useSendCodeCooldown,
} from "@/hooks/use-send-code-cooldown"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { readFormError } from "@/lib/form-error"

interface ChangeEmailDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentEmail: string
  /** 邮箱更新成功后通知父组件重新拉取资料。 */
  onChanged: () => void
}

/**
 * 修改邮箱：先给新邮箱发验证码，验证通过后才写入。
 * 邮箱是 OAuth 登录的身份锚点，所以后端要求新地址必须被验证过，
 * 前端不能省掉发码这一步，也不能把未验证的地址直接提交。
 */
export function ChangeEmailDialog({
  open,
  onOpenChange,
  currentEmail,
  onChanged,
}: ChangeEmailDialogProps) {
  const { t } = useTranslation()
  const captcha = useCaptcha()
  const cooldown = useSendCodeCooldown("change-email")
  const [email, setEmail] = useState("")
  const [code, setCode] = useState("")
  const [captchaCode, setCaptchaCode] = useState("")
  const [sending, setSending] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})

  function validateEmail(): string | null {
    const trimmed = requireTrimmed(email, t("form.required"))
    if (trimmed) {
      return trimmed
    }
    return email.includes("@") ? null : t("form.invalidEmail")
  }

  async function onSendCode() {
    if (sending) {
      return
    }
    const errors = collectFieldErrors([
      ["email", validateEmail()],
      ["captchaCode", requireTrimmed(captchaCode, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setError("")
      return
    }
    if (!captcha.captchaId) {
      setError(t("form.captchaUnavailable"))
      return
    }
    setSending(true)
    setError("")
    try {
      await sendSelfChangeEmailCode({
        email: email.trim(),
        captchaId: captcha.captchaId,
        captchaCode,
      })
      cooldown.start()
      toast.success(t("profile.changeEmailSent"))
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
      if (err instanceof ApiError) {
        const retryAfter = Number(err.metadata[RETRY_AFTER_SECONDS_KEY] ?? 0)
        if (retryAfter > 0) {
          cooldown.start(retryAfter)
        }
      }
      if (failure.fields.captchaCode) {
        setCaptchaCode("")
        void captcha.refresh()
      }
    } finally {
      setSending(false)
    }
  }

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (submitting) {
      return
    }
    const errors = collectFieldErrors([
      ["email", validateEmail()],
      ["code", requireTrimmed(code, t("form.required"))],
      ["captchaCode", requireTrimmed(captchaCode, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setError("")
      return
    }
    if (!captcha.captchaId) {
      setError(t("form.captchaUnavailable"))
      return
    }
    setSubmitting(true)
    setError("")
    try {
      await changeEmail({
        email: email.trim(),
        code,
        captchaId: captcha.captchaId,
        captchaCode,
      })
      toast.success(t("profile.changeEmailChanged"))
      onChanged()
      onOpenChange(false)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
      if (failure.fields.captchaCode) {
        setCaptchaCode("")
        void captcha.refresh()
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader className="shrink-0">
          <DialogTitle>{t("profile.changeEmailTitle")}</DialogTitle>
          <DialogDescription>
            {t("profile.changeEmailDescription", { email: currentEmail })}
          </DialogDescription>
        </DialogHeader>
        <form
          id="change-email-form"
          className="grid min-h-0 gap-4 overflow-y-auto"
          noValidate
          onSubmit={onSubmit}
        >
          <FormField
            label={t("profile.changeEmailNew")}
            htmlFor="change-email"
            required
            error={fieldErrors.email}
          >
            <Input
              id="change-email"
              type="email"
              autoComplete="email"
              value={email}
              placeholder={currentEmail}
              aria-invalid={fieldErrors.email ? true : undefined}
              onChange={(event) => setEmail(event.target.value)}
            />
          </FormField>
          <CaptchaField
            value={captchaCode}
            onChange={setCaptchaCode}
            captcha={captcha}
            error={fieldErrors.captchaCode}
          />
          <FormField
            label={t("profile.changeEmailCode")}
            htmlFor="change-email-code"
            required
            error={fieldErrors.code}
          >
            <div className="flex items-center gap-2">
              <Input
                id="change-email-code"
                className="min-w-0 flex-1"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={code}
                placeholder={t("login.codePlaceholder")}
                aria-invalid={fieldErrors.code ? true : undefined}
                onChange={(event) => setCode(event.target.value)}
              />
              <Button
                type="button"
                variant="outline"
                className="shrink-0"
                disabled={sending || cooldown.remaining > 0 || !captcha.captchaId}
                onClick={() => void onSendCode()}
              >
                {cooldown.remaining > 0
                  ? t("profile.changeEmailResendIn", {
                      seconds: cooldown.remaining,
                    })
                  : sending
                    ? t("profile.changeEmailSending")
                    : t("profile.changeEmailSend")}
              </Button>
            </div>
          </FormField>
          {error ? (
            <p role="alert" className="text-sm text-destructive">
              {error}
            </p>
          ) : null}
        </form>
        <DialogFooter className="shrink-0">
          <Button type="submit" form="change-email-form" disabled={submitting}>
            {submitting ? t("action.saving") : t("action.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
