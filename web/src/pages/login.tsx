import { KeyRound, Loader2, Mail } from "lucide-react"
import { useEffect, useRef, useState } from "react"
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

// 与 README 徽章和 git remote 一致的仓库地址；AuthShell 的仓库 icon 与 Star 摘要共用。
const repositoryURL = "https://github.com/aide-family/moon"

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
  // 登录方式默认第三方；切到邮箱后第三方按钮收起，两个入口共用同一套开关。
  const [emailMode, setEmailMode] = useState(false)
  // 切换后把焦点送进新出现的面板，键盘用户不用逆着 DOM 顺序找回来。
  const focusAfterSwitch = useRef<"email" | "oauth" | null>(null)
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
  // 没有配置第三方时邮箱是唯一入口，直接展开；有第三方时由用户选择登录方式。
  const emailPanelOpen = providersReady && !hasProviders ? true : emailMode
  const showMethodSwitch = !providersReady || hasProviders
  const showOAuthSection = !providersReady || hasProviders

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

  // 两个面板始终挂载（保留已填内容），切换后把焦点交给新面板的第一个控件。
  useEffect(() => {
    const target = focusAfterSwitch.current
    if (!target) {
      return
    }
    const node =
      target === "email"
        ? document.getElementById("email")
        : document.querySelector<HTMLButtonElement>(
            "#oauth-login-panel button[data-provider]"
          )
    if (!node) {
      return
    }
    focusAfterSwitch.current = null
    node.focus()
  }, [emailPanelOpen, providersReady])

  function switchLoginMethod() {
    const nextEmailMode = !emailPanelOpen
    focusAfterSwitch.current = nextEmailMode ? "email" : "oauth"
    setEmailMode(nextEmailMode)
  }

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
    <AuthShell
      title={t("login.welcome")}
      subtitle={t("login.welcomeSub")}
      repositoryURL={repositoryURL}
    >
      <div className="grid gap-6">
        {/* 主操作：第三方登录置顶，48px 整行按钮、品牌图标块与清晰的外链提示。
            列表未就绪时用同高骨架占位，避免邮箱表单先展开再收起的层级闪动。 */}
        {showOAuthSection ? (
          <div
            id="oauth-login-panel"
            role="group"
            aria-label={t("login.oauthGroupLabel")}
            className={cn("grid gap-3", emailPanelOpen && "hidden")}
          >
            {providersReady ? (
              providers.map((provider) => (
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
              ))
            ) : (
              <div className="grid gap-3" aria-hidden="true">
                <Skeleton className="h-12 rounded-xl" />
                <Skeleton className="h-12 rounded-xl" />
              </div>
            )}
          </div>
        ) : null}

        {/* 次操作：邮箱验证码登录。两个面板都保持挂载，切换时只隐藏其中一个，
            已填邮箱与验证码不会因切换而丢失；没有第三方配置时它是唯一入口。 */}
        <div
          id="email-login-panel"
          className={cn("grid gap-4", !emailPanelOpen && "hidden")}
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
        </div>

        {/* 登录方式开关：无论从第三方切到邮箱还是反向切换，都复用同一条分隔线与按钮。 */}
        {showMethodSwitch ? (
          <div className="grid gap-3">
            <div
              aria-hidden="true"
              className="flex items-center gap-3 text-xs text-muted-foreground"
            >
              <span className="h-px flex-1 bg-border" />
              <span>{t("login.oauthOr")}</span>
              <span className="h-px flex-1 bg-border" />
            </div>
            <Button
              type="button"
              variant="ghost"
              aria-controls={
                emailPanelOpen ? "oauth-login-panel" : "email-login-panel"
              }
              className="h-10 w-full justify-center gap-2 rounded-lg font-normal text-muted-foreground hover:text-foreground"
              onClick={switchLoginMethod}
            >
              {emailPanelOpen ? (
                <KeyRound className="size-4" aria-hidden="true" />
              ) : (
                <Mail className="size-4" aria-hidden="true" />
              )}
              <span>
                {emailPanelOpen
                  ? t("login.oauthLoginShow")
                  : t("login.emailLoginShow")}
              </span>
            </Button>
          </div>
        ) : null}

        {/* 底部只留注册入口；GitHub 仓库与 Star 已收到右上角同一行。 */}
        <div className="flex justify-center">
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
