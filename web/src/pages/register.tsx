import { Loader2 } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { register, sendRegisterEmailCode } from "@/api/auth"
import type { LoginResult } from "@/api/auth/types"
import { ApiError } from "@/api/http"
import { CaptchaField } from "@/components/captcha-field"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useCaptcha } from "@/hooks/use-captcha"
import {
  RETRY_AFTER_SECONDS_KEY,
  SEND_CODE_COOLDOWN_SECONDS,
  useSendCodeCooldown,
} from "@/hooks/use-send-code-cooldown"
import { collectFieldErrors, hasFieldErrors, requireTrimmed } from "@/lib/form"
import { readFormError } from "@/lib/form-error"
import { readNextPath } from "@/lib/redirect"
import { USERNAME_MAX_LENGTH, usernameError } from "@/lib/username"
import { AuthShell } from "@/pages/auth-shell"

function storeSession(result: LoginResult) {
  localStorage.setItem("token", result.session.accessToken)
  localStorage.setItem("refreshToken", result.session.refreshToken)
}

/** 服务端节流时返回需要等待的秒数；其他错误返回 null。 */
function retryAfterSeconds(err: unknown): number | null {
  if (!(err instanceof ApiError) || err.raw !== "send code too frequently") {
    return null
  }
  const seconds = Number(err.metadata[RETRY_AFTER_SECONDS_KEY])
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return SEND_CODE_COOLDOWN_SECONDS
  }
  return Math.ceil(seconds)
}

export function RegisterPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const next = readNextPath(params.get("next"))
  const captcha = useCaptcha()
  const [email, setEmail] = useState(params.get("email") ?? "")
  const [name, setName] = useState("")
  const [code, setCode] = useState("")
  const [captchaCode, setCaptchaCode] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [sendingCode, setSendingCode] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const cooldown = useSendCodeCooldown("register")
  const pending = sendingCode || submitting

  async function onSendCode() {
    if (sendingCode || cooldown.remaining > 0) {
      return
    }
    const errors = collectFieldErrors([
      ["email", requireTrimmed(email, t("form.required"))],
      ["captchaCode", requireTrimmed(captchaCode, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setError("")
      return
    }
    if (!captcha.captchaId) {
      setError(t("form.captchaUnavailable"))
      setFieldErrors({})
      return
    }
    setSendingCode(true)
    setError("")
    setFieldErrors({})
    try {
      await sendRegisterEmailCode({
        email,
        captchaId: captcha.captchaId,
        captchaCode,
      })
      cooldown.start()
      toast.success(t("login.codeSent"), {
        description: t("login.codeSentHint"),
      })
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
      // 发码只校验图形验证码、不消费它：邮箱或节流错误时保留本题，只有答案错误才换一张。
      if (failure.fields.captchaCode) {
        setCaptchaCode("")
        void captcha.refresh()
      }
      // 服务端限制频次时也恢复倒计时，刷新页面后仍按剩余时间继续。
      const retryAfter = retryAfterSeconds(err)
      if (retryAfter !== null) {
        cooldown.start(retryAfter)
        toast.error(t("error.sendTooFrequent"), {
          description: t("login.resendIn", { seconds: retryAfter }),
        })
      } else if (failure.form) {
        toast.error(failure.form)
      }
    } finally {
      setSendingCode(false)
    }
  }

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (pending) {
      return
    }
    const errors = collectFieldErrors([
      ["email", requireTrimmed(email, t("form.required"))],
      ["code", requireTrimmed(code, t("form.required"))],
      ["captchaCode", requireTrimmed(captchaCode, t("form.required"))],
      ["name", usernameError(name, t("register.usernameInvalid"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setError("")
      return
    }
    if (!captcha.captchaId) {
      setError(t("form.captchaUnavailable"))
      setFieldErrors({})
      return
    }
    setSubmitting(true)
    setError("")
    setFieldErrors({})
    try {
      const result = await register({
        email,
        code,
        name: name.trim(),
        captchaId: captcha.captchaId,
        captchaCode,
      })
      storeSession(result)
      navigate(next)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
      // 邮箱码填错时图形验证码仍然有效，保留它让用户改正后直接重试；
      // 只有答案错误或过期才换一张。
      if (failure.fields.captchaCode) {
        setCaptchaCode("")
        void captcha.refresh()
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthShell
      title={t("register.welcome")}
      subtitle={t("register.welcomeSub")}
    >
      <form noValidate className="grid gap-4" onSubmit={onSubmit}>
        <FormField
          label={t("register.name")}
          htmlFor="name"
          hint={t("register.nameHint")}
          error={fieldErrors.name}
        >
          <Input
            id="name"
            value={name}
            placeholder={t("register.namePlaceholder")}
            autoComplete="username"
            maxLength={USERNAME_MAX_LENGTH}
            aria-invalid={fieldErrors.name ? true : undefined}
            onChange={(event) => setName(event.target.value)}
          />
        </FormField>
        <FormField
          label={t("login.emailLabel")}
          htmlFor="email"
          required
          error={fieldErrors.email}
        >
          <Input
            id="email"
            type="email"
            value={email}
            placeholder={t("login.emailPlaceholder")}
            autoComplete="email"
            aria-invalid={fieldErrors.email ? true : undefined}
            onChange={(event) => setEmail(event.target.value)}
            required
          />
        </FormField>
        <CaptchaField
          value={captchaCode}
          onChange={setCaptchaCode}
          captcha={captcha}
          error={fieldErrors.captchaCode}
        />
        <FormField
          label={t("login.codeLabel")}
          htmlFor="code"
          hint={t("login.codeHint")}
          required
          error={fieldErrors.code}
        >
          <div className="flex gap-2">
            <Input
              id="code"
              className="min-w-0 flex-1"
              value={code}
              placeholder={t("login.codePlaceholder")}
              autoComplete="one-time-code"
              aria-invalid={fieldErrors.code ? true : undefined}
              onChange={(event) => setCode(event.target.value)}
              required
            />
            <Button
              type="button"
              variant="outline"
              className="shrink-0"
              disabled={
                pending ||
                cooldown.remaining > 0 ||
                !email.trim() ||
                !captchaCode.trim()
              }
              aria-busy={sendingCode || undefined}
              onClick={() => void onSendCode()}
            >
              {sendingCode ? (
                <Loader2 className="size-4 animate-spin" aria-hidden="true" />
              ) : null}
              {sendingCode
                ? t("action.sendingCode")
                : cooldown.remaining > 0
                  ? t("login.resendIn", { seconds: cooldown.remaining })
                  : t("action.sendCode")}
            </Button>
          </div>
        </FormField>
        <Button
          className="w-full"
          type="submit"
          disabled={pending}
          aria-busy={submitting || undefined}
        >
          {submitting ? (
            <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          ) : null}
          {submitting ? t("action.registering") : t("action.register")}
        </Button>
        {error ? (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        ) : null}
      </form>
      <p className="text-sm text-muted-foreground">
        {t("register.platformHint")}
      </p>
      <div className="flex justify-end">
        <Link
          className="text-sm text-primary underline-offset-4 hover:underline"
          to={
            next === "/namespaces"
              ? "/login"
              : `/login?next=${encodeURIComponent(next)}`
          }
        >
          {t("register.hasAccount")}
        </Link>
      </div>
    </AuthShell>
  )
}
