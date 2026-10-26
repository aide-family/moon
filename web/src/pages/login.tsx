import { ChevronDown, Loader2, Mail } from "lucide-react"
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
import { OAuthProviderButton } from "@/components/oauth-provider-button"
import { Button } from "@/components/ui/button"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { useCaptcha } from "@/hooks/use-captcha"
import {
  RETRY_AFTER_SECONDS_KEY,
  SEND_CODE_COOLDOWN_SECONDS,
  useSendCodeCooldown,
} from "@/hooks/use-send-code-cooldown"
import { cn } from "@/lib/class-names"
import type { FieldErrors } from "@/lib/form"
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

// 字段名 → 输入框 id；CaptchaField 的 id 与字段名不同，映射一次再聚焦。
const fieldOrder = ["email", "captchaCode", "code"]
const fieldIds: Record<string, string> = {
  email: "email",
  captchaCode: "captcha",
  code: "code",
}

/** 校验失败后把焦点移到第一个出错的输入框，键盘用户不用自己找。 */
function focusFirstError(errors: FieldErrors) {
  const field = fieldOrder.find((name) => errors[name])
  if (field) {
    document.getElementById(fieldIds[field])?.focus()
  }
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
  const [providersReady, setProvidersReady] = useState(false)
  // 邮箱验证码登录是次要入口：默认收起，用户主动展开后才显示表单。
  const [emailOpen, setEmailOpen] = useState(false)
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
  const hasProviders = providers.length > 0
  // 有第三方登录时邮箱入口默认收起；没有配置（含接口不可用）时它是唯一入口，直接展开。
  const emailPanelOpen = providersReady && !hasProviders ? true : emailOpen
  // 列表加载中也保留「或使用邮箱验证码登录」入口，慢网下不必等第三方列表就能登录。
  const showEmailToggle = !providersReady || hasProviders

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
        // 未配置或接口不可用时只隐藏第三方入口，不阻塞邮箱验证码登录。
        if (active) {
          setProviders([])
        }
      } finally {
        // 加载结束才决定层级：避免先展开邮箱表单、拿到第三方列表后又收起的闪动。
        if (active) {
          setProvidersReady(true)
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
      focusFirstError(errors)
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
      focusFirstError(failure.fields)
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
      focusFirstError(errors)
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
      // readNextPath 已过滤跨站目标；这里再兜一层，避免路由拦截把已登录用户卡在登录页。
      try {
        navigate(next)
      } catch {
        navigate("/", { replace: true })
      }
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
      focusFirstError(failure.fields)
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
      <div className="grid gap-6">
        {/* 主操作：第三方登录置顶，整行 44px 按钮、品牌色图标，间距优先。
            列表未就绪时用同高骨架占位，避免邮箱表单先展开再收起的层级闪动。 */}
        {providersReady ? (
          hasProviders ? (
            <div
              role="group"
              aria-label={t("login.oauthGroupLabel")}
              className="grid gap-2.5"
            >
              {providers.map((provider) => (
                <OAuthProviderButton
                  key={provider.provider}
                  provider={provider}
                  label={t("login.oauthContinue", {
                    provider: provider.name || provider.provider,
                  })}
                  pending={oauthPending === provider.provider}
                  disabled={oauthPending !== ""}
                  onClick={() => startOAuth(provider)}
                />
              ))}
            </div>
          ) : null
        ) : (
          <div className="grid gap-2.5" aria-hidden="true">
            <Skeleton className="h-11 rounded-lg" />
            <Skeleton className="h-11 rounded-lg" />
          </div>
        )}

        {/* 次操作：邮箱验证码登录。保留全部表单与校验，只降级为可收起的入口；
            第三方列表还在加载时入口照常可用，慢网下也不会只剩骨架、无法登录。 */}
        <Collapsible
          open={emailPanelOpen}
          onOpenChange={setEmailOpen}
          className="grid gap-3"
        >
          {showEmailToggle ? (
            <div className="flex items-center gap-3 text-xs text-muted-foreground">
              <span aria-hidden="true" className="h-px flex-1 bg-border" />
              <span>{t("login.oauthOr")}</span>
              <span aria-hidden="true" className="h-px flex-1 bg-border" />
            </div>
          ) : null}
          {showEmailToggle ? (
            <CollapsibleTrigger asChild>
              <Button
                variant="ghost"
                aria-controls="email-login-panel"
                className="h-10 w-full justify-center gap-2 font-normal text-muted-foreground hover:text-foreground"
              >
                <Mail className="size-4" aria-hidden="true" />
                <span>{t("login.emailLoginShow")}</span>
                <ChevronDown
                  aria-hidden="true"
                  className={cn(
                    "size-4 transition-transform duration-200 motion-reduce:transition-none",
                    emailOpen && "rotate-180"
                  )}
                />
              </Button>
            </CollapsibleTrigger>
          ) : null}
          {/* forceMount 保住已填的邮箱与验证码；收起时用 hidden 移出布局与无障碍树
              （Radix 在 forceMount 下不会自己隐藏内容，必须由受控状态补上）。 */}
          <CollapsibleContent
            forceMount
            hidden={!emailPanelOpen}
            id="email-login-panel"
            className={showEmailToggle ? "pt-1" : undefined}
          >
            <form noValidate className="grid gap-4" onSubmit={onSubmit}>
              <FormField
                label={t("login.emailLabel")}
                htmlFor="email"
                required
                error={fieldErrors.email}
              >
                <Input
                  id="email"
                  name="email"
                  type="email"
                  value={email}
                  placeholder={t("login.emailPlaceholder")}
                  autoComplete="email"
                  spellCheck={false}
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
                    name="code"
                    className="min-w-0 flex-1"
                    value={code}
                    placeholder={t("login.codePlaceholder")}
                    autoComplete="one-time-code"
                    spellCheck={false}
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
                      <Loader2
                        className="size-4 animate-spin"
                        aria-hidden="true"
                      />
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
          </CollapsibleContent>
        </Collapsible>

        <div className="flex justify-end">
          <Link
            className="text-sm text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
            to={
              next === "/"
                ? "/register"
                : `/register?next=${encodeURIComponent(next)}`
            }
          >
            {t("login.noAccount")}
          </Link>
        </div>
      </div>
    </AuthShell>
  )
}
