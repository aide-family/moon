import { Loader2 } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"
import { toast } from "sonner"

import { bindOAuthEmail, sendOAuthBindEmailCode } from "@/api/auth"
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

/** 后端在待绑定凭证失效/过期时返回的原文。 */
const OAUTH_BINDING_EXPIRED_MESSAGE = "oauth binding is invalid or expired"
/** 后端在第三方账号已经被绑定时返回的原文。 */
const OAUTH_ACCOUNT_BOUND_MESSAGE = "oauth account is already bound"

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

type TerminalFailure = {
  title: string
  hint: string
}

interface OAuthEmailBindingProps {
  /** 授权码换取时拿到的待绑定凭证。 */
  bindToken: string
  /** 第三方返回的未验证邮箱，用于预填；平台没有返回时为空。 */
  suggestedEmail: string
  /** 绑定并登录成功：调用方保存会话、清理第三方现场并回跳。 */
  onSignedIn: (result: LoginResult) => void
}

/**
 * 第三方登录缺少可用邮箱时的补充邮箱表单：邮箱 + 图形验证码 + 邮箱验证码。
 * 验证通过才签发会话；凭证过期或账号已绑定时表单不可继续，给出返回登录的出口。
 */
export function OAuthEmailBinding({
  bindToken,
  suggestedEmail,
  onSignedIn,
}: OAuthEmailBindingProps) {
  const { t } = useTranslation()
  const captcha = useCaptcha()
  const cooldown = useSendCodeCooldown("oauth")
  const [email, setEmail] = useState(suggestedEmail)
  const [code, setCode] = useState("")
  const [captchaCode, setCaptchaCode] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [terminalFailure, setTerminalFailure] =
    useState<TerminalFailure | null>(null)
  const [sendingCode, setSendingCode] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const pending = sendingCode || submitting

  /** 待绑定凭证失效或账号已绑定时，表单无法继续，返回需要展示的业务错误。 */
  function terminalFailureOf(err: unknown): TerminalFailure | null {
    if (!(err instanceof ApiError)) {
      return null
    }
    if (err.raw === OAUTH_BINDING_EXPIRED_MESSAGE) {
      return {
        title: t("oauthCallback.bindingExpired"),
        hint: t("oauthCallback.bindingExpiredHint"),
      }
    }
    if (err.raw === OAUTH_ACCOUNT_BOUND_MESSAGE) {
      return {
        title: t("error.oauthAccountBound"),
        hint: t("oauthCallback.bindingExpiredHint"),
      }
    }
    return null
  }

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
      await sendOAuthBindEmailCode({
        bindToken,
        email,
        captchaId: captcha.captchaId,
        captchaCode,
      })
      cooldown.start()
      toast.success(t("login.codeSent"), {
        description: t("login.codeSentHint"),
      })
    } catch (err) {
      const terminal = terminalFailureOf(err)
      if (terminal) {
        setTerminalFailure(terminal)
        return
      }
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
      const result = await bindOAuthEmail({
        bindToken,
        email,
        code,
        captchaId: captcha.captchaId,
        captchaCode,
      })
      onSignedIn(result)
    } catch (err) {
      const terminal = terminalFailureOf(err)
      if (terminal) {
        setTerminalFailure(terminal)
        return
      }
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

  if (terminalFailure) {
    return (
      <div className="grid gap-4">
        <div className="grid gap-1">
          <p role="alert" className="text-sm font-medium">
            {terminalFailure.title}
          </p>
          <p className="text-sm text-muted-foreground">
            {terminalFailure.hint}
          </p>
        </div>
        <Button asChild variant="outline" className="w-full">
          <Link to="/login">{t("oauthCallback.backToLogin")}</Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="grid gap-4">
      <p className="text-sm text-muted-foreground">
        {t("oauthCallback.bindingDescription")}
      </p>
      <form noValidate className="grid gap-4" onSubmit={onSubmit}>
        <FormField
          label={t("login.emailLabel")}
          htmlFor="oauth-bind-email"
          hint={t("oauthCallback.bindingEmailHint")}
          required
          error={fieldErrors.email}
        >
          <Input
            id="oauth-bind-email"
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
          htmlFor="oauth-bind-code"
          hint={t("login.codeHint")}
          required
          error={fieldErrors.code}
        >
          <div className="flex gap-2">
            <Input
              id="oauth-bind-code"
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
          {submitting
            ? t("oauthCallback.bindingSubmitting")
            : t("oauthCallback.bindingSubmit")}
        </Button>
        {error ? (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        ) : null}
      </form>
    </div>
  )
}
