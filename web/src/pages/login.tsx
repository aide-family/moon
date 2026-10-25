import { Loader2 } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { emailLogin, listOAuth2Providers, sendEmailLoginCode } from "@/api/auth"
import type { LoginResult, OAuth2ProviderItem } from "@/api/auth/types"
import { ApiError } from "@/api/http"
import {
  newOAuthState,
  oauthRedirectUri,
  storeOAuthRedirect,
} from "@/auth/oauth-redirect"
import { CaptchaField } from "@/components/captcha-field"
import { FormField } from "@/components/form-field"
import { GiteeIcon } from "@/components/gitee-icon"
import { GitHubIcon } from "@/components/github-icon"
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

export function LoginPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const next = readNextPath(params.get("next"))
  const captcha = useCaptcha()
  // 每次进入登录页生成一次性 state，第三方授权回调凭它取回本轮现场。
  const [oauthState] = useState(newOAuthState)
  const [providers, setProviders] = useState<OAuth2ProviderItem[]>([])
  const [email, setEmail] = useState("")
  const [code, setCode] = useState("")
  const [captchaCode, setCaptchaCode] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [sendingCode, setSendingCode] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [oauthPending, setOauthPending] = useState("")
  const cooldown = useSendCodeCooldown("login")
  const pending = sendingCode || submitting

  useEffect(() => {
    let active = true
    async function loadProviders() {
      try {
        const result = await listOAuth2Providers({
          redirectUri: oauthRedirectUri(),
          state: oauthState,
        })
        if (active) {
          setProviders(result.providers ?? [])
        }
      } catch {
        // 第三方登录是附加入口：未配置或接口不可用时只隐藏入口，不阻塞邮箱验证码登录。
        if (active) {
          setProviders([])
        }
      }
    }
    void loadProviders()
    return () => {
      active = false
    }
  }, [oauthState])

  function startOAuth(provider: OAuth2ProviderItem) {
    if (oauthPending) {
      return
    }
    setOauthPending(provider.provider)
    storeOAuthRedirect({
      state: oauthState,
      provider: provider.provider,
      redirectUri: oauthRedirectUri(),
      next,
    })
    window.location.assign(provider.authorizeUrl)
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
      await sendEmailLoginCode({
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
      const result = await emailLogin({
        email,
        code,
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
    <AuthShell title={t("login.welcome")} subtitle={t("login.welcomeSub")}>
      <form noValidate className="grid gap-4" onSubmit={onSubmit}>
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
          {submitting ? t("login.loggingIn") : t("action.login")}
        </Button>
        {error ? (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        ) : null}
      </form>
      {providers.length > 0 ? (
        <div className="grid gap-3">
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span aria-hidden="true" className="h-px flex-1 bg-border" />
            <span>{t("login.oauthDivider")}</span>
            <span aria-hidden="true" className="h-px flex-1 bg-border" />
          </div>
          <div className="grid gap-2 sm:grid-cols-2">
            {providers.map((provider) => (
              <Button
                key={provider.provider}
                type="button"
                variant="outline"
                className="w-full"
                disabled={oauthPending !== ""}
                aria-busy={oauthPending === provider.provider || undefined}
                onClick={() => startOAuth(provider)}
              >
                {oauthPending === provider.provider ? (
                  <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                ) : provider.provider === "GITHUB" ? (
                  <GitHubIcon className="size-4" />
                ) : provider.provider === "GITEE" ? (
                  <GiteeIcon className="size-4" />
                ) : null}
                <span>{provider.name || provider.provider}</span>
              </Button>
            ))}
          </div>
        </div>
      ) : null}
      <div className="flex justify-end">
        <Link
          className="text-sm text-primary underline-offset-4 hover:underline"
          to={
            next === "/namespaces"
              ? "/register"
              : `/register?next=${encodeURIComponent(next)}`
          }
        >
          {t("login.noAccount")}
        </Link>
      </div>
    </AuthShell>
  )
}
