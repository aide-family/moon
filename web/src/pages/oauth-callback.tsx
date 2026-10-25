import { useEffect, useRef, useState } from "react"
import { Loader2, MailCheck } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate, useSearchParams } from "react-router-dom"

import { oauth2Login } from "@/api/auth"
import type { LoginResult, OAuth2Provider } from "@/api/auth/types"
import { ApiError } from "@/api/http"
import {
  OAUTH_EMAIL_UNVERIFIED_MESSAGE,
  clearOAuthRedirect,
  oauthEmailSettingsUrl,
  oauthProviderName,
  readOAuthRedirect,
} from "@/auth/oauth-redirect"
import { Button } from "@/components/ui/button"
import { formErrorMessage } from "@/lib/form-error"
import { readNextPath } from "@/lib/redirect"
import { AuthShell } from "@/pages/auth-shell"

function storeSession(result: LoginResult) {
  localStorage.setItem("token", result.session.accessToken)
  localStorage.setItem("refreshToken", result.session.refreshToken)
}

// 回调失败现场。邮箱未验证单独一类，用来给出去平台验证邮箱的引导。
// detail 是第三方返回的错误描述（例如 invalid_scope），用于排查授权配置。
type OAuthFailure =
  | { kind: "emailUnverified"; provider: OAuth2Provider }
  | { kind: "message"; message: string; detail?: string }

// 第三方授权回调页：校验本轮 state 现场，用授权码换会话，成功后回到 next。
export function OAuthCallbackPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const [failure, setFailure] = useState<OAuthFailure | null>(null)
  const started = useRef(false)

  useEffect(() => {
    if (started.current) {
      return
    }
    started.current = true
    const state = params.get("state")
    const code = params.get("code")
    const providerError = params.get("error")
    const providerErrorDescription = params.get("error_description")
    const redirect = readOAuthRedirect(state)

    async function exchange() {
      if (providerError) {
        if (redirect) {
          clearOAuthRedirect(redirect.state)
        }
        setFailure({
          kind: "message",
          message: t("oauthCallback.providerError"),
          detail: providerErrorDescription || providerError,
        })
        return
      }
      if (!code) {
        if (redirect) {
          clearOAuthRedirect(redirect.state)
        }
        setFailure({ kind: "message", message: t("oauthCallback.missingCode") })
        return
      }
      if (!redirect) {
        setFailure({ kind: "message", message: t("oauthCallback.stateMismatch") })
        return
      }
      try {
        const result = await oauth2Login({
          provider: redirect.provider,
          code,
          redirectUri: redirect.redirectUri,
        })
        storeSession(result)
        clearOAuthRedirect(redirect.state)
        navigate(readNextPath(redirect.next), { replace: true })
      } catch (err) {
        clearOAuthRedirect(redirect.state)
        if (err instanceof ApiError && err.raw === OAUTH_EMAIL_UNVERIFIED_MESSAGE) {
          setFailure({ kind: "emailUnverified", provider: redirect.provider })
          return
        }
        setFailure({ kind: "message", message: formErrorMessage(err, t("error.oauthFailed")) })
      }
    }
    void exchange()
  }, [navigate, params, t])

  return (
    <AuthShell
      title={t("oauthCallback.title")}
      subtitle={
        failure?.kind === "emailUnverified"
          ? t("oauthCallback.emailUnverifiedSubtitle")
          : t("oauthCallback.subtitle")
      }
    >
      {failure === null ? (
        <p role="status" className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          {t("oauthCallback.processing")}
        </p>
      ) : failure.kind === "emailUnverified" ? (
        <OAuthEmailUnverified provider={failure.provider} />
      ) : (
        <div className="grid gap-4">
          <p role="alert" className="text-sm text-destructive">
            {failure.message}
          </p>
          {failure.detail ? (
            <p className="text-xs break-all text-muted-foreground">{failure.detail}</p>
          ) : null}
          <Button asChild variant="outline" className="w-full">
            <Link to="/login">{t("oauthCallback.backToLogin")}</Link>
          </Button>
        </div>
      )}
    </AuthShell>
  )
}

// 第三方邮箱未验证且邮箱已被本地账号占用：给出“去平台验证邮箱”和“重新登录”的出口，
// 而不是只显示一句错误。验证完成后重新发起第三方登录即可并入已有账号。
function OAuthEmailUnverified({ provider }: { provider: OAuth2Provider }) {
  const { t } = useTranslation()
  const providerName = oauthProviderName(provider)
  const settingsUrl = oauthEmailSettingsUrl(provider)

  return (
    <div className="grid gap-4">
      <div className="flex items-start gap-3 rounded-lg border bg-muted/40 p-3">
        <MailCheck className="mt-0.5 size-5 shrink-0 text-primary" aria-hidden="true" />
        <div className="grid gap-1">
          <p role="alert" className="text-sm font-medium">
            {t("oauthCallback.emailUnverifiedTitle")}
          </p>
          <p className="text-sm text-muted-foreground">
            {t("oauthCallback.emailUnverifiedDescription", { provider: providerName })}
          </p>
        </div>
      </div>
      <div className="grid gap-2">
        {settingsUrl ? (
          <Button asChild className="w-full">
            <a href={settingsUrl} target="_blank" rel="noreferrer">
              {t("oauthCallback.verifyEmail", { provider: providerName })}
            </a>
          </Button>
        ) : null}
        <Button asChild variant="outline" className="w-full">
          <Link to="/login">{t("oauthCallback.retryLogin")}</Link>
        </Button>
      </div>
      <p className="text-sm text-muted-foreground">
        {t("oauthCallback.emailUnverifiedHint", { provider: providerName })}
      </p>
    </div>
  )
}
