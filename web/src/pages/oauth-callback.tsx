import { Loader2 } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate, useSearchParams } from "react-router-dom"

import { oauth2Login } from "@/api/auth"
import type { LoginResult, OAuth2Provider, TokenPair } from "@/api/auth/types"
import { ApiError } from "@/api/http"
import { bindSelfOAuth } from "@/api/self"
import {
  clearOAuthRedirect,
  oauthCodeParam,
  oauthProviderName,
  readOAuthRedirect,
  type OAuthRedirect,
} from "@/auth/oauth-redirect"
import { OAuthEmailBinding } from "@/components/oauth-email-binding"
import { Button } from "@/components/ui/button"
import { enumLabel } from "@/lib/enum-label"
import { formErrorMessage } from "@/lib/form-error"
import { readNextPath } from "@/lib/redirect"
import { AuthShell } from "@/pages/auth-shell"

function storeSession(session: TokenPair) {
  localStorage.setItem("token", session.accessToken)
  localStorage.setItem("refreshToken", session.refreshToken)
}

// 回调失败现场。detail 是第三方返回的错误描述（例如 invalid_scope），用于排查授权配置。
// backTo 决定失败出口：绑定路径回个人资料，登录路径（缺省）回登录页。
type OAuthFailure = {
  message: string
  detail?: string
  backTo?: "login" | "profile"
}

// 授权码换取后需要补充邮箱的待绑定现场；redirect 用于绑定成功后清理本轮 OAuth 现场并回跳。
type OAuthEmailBindingState = {
  redirect: OAuthRedirect
  bindToken: string
  suggestedEmail: string
}

// 第三方授权回调页：校验本轮 state 现场，用授权码换会话，成功后回到 next。
// 第三方没有可用邮箱时留在本页补充并验证邮箱，验证通过后才签发会话。
export function OAuthCallbackPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const [failure, setFailure] = useState<OAuthFailure | null>(null)
  const [emailBinding, setEmailBinding] =
    useState<OAuthEmailBindingState | null>(null)
  // 非空表示本轮是「绑定到当前账号」而非登录，标题与处理文案随之切换。
  const [binding, setBinding] = useState<OAuth2Provider | "">("")
  const started = useRef(false)

  useEffect(() => {
    if (started.current) {
      return
    }
    started.current = true
    const state = params.get("state")
    const providerError = params.get("error")
    const providerErrorDescription = params.get("error_description")
    const redirect = readOAuthRedirect(state)
    // 钉钉把授权码放在 authCode，其余平台放在 code。
    const code = params.get(
      redirect ? oauthCodeParam(redirect.provider) : "code"
    )
    // 本轮是「绑定到当前账号」时的来源；历史现场缺省为登录，绑定值只由个人资料页写入。
    const bindProvider: OAuth2Provider | "" =
      redirect?.intent === "bind" ? redirect.provider : ""

    async function exchange() {
      if (bindProvider) {
        setBinding(bindProvider)
      }
      if (providerError) {
        if (redirect) {
          clearOAuthRedirect(redirect.state)
        }
        setFailure({
          message: t("oauthCallback.providerError"),
          detail: providerErrorDescription || providerError,
          backTo: bindProvider ? "profile" : undefined,
        })
        return
      }
      if (!code) {
        if (redirect) {
          clearOAuthRedirect(redirect.state)
        }
        setFailure({
          message: t("oauthCallback.missingCode"),
          backTo: bindProvider ? "profile" : undefined,
        })
        return
      }
      if (!redirect) {
        setFailure({ message: t("oauthCallback.stateMismatch") })
        return
      }
      if (bindProvider) {
        // 绑定路径绝不调用 oauth2Login：第三方身份若已绑到他人，登录接口会给那个人
        // 签发会话，在已登录页面里等同于账号劫持。绑定必须走带鉴权的 bindSelfOAuth。
        if (!localStorage.getItem("token")) {
          clearOAuthRedirect(redirect.state)
          setFailure({
            message: t("profile.oauth.sessionExpired"),
            backTo: "login",
          })
          return
        }
        try {
          await bindSelfOAuth({
            provider: redirect.provider,
            code,
            redirectUri: redirect.redirectUri,
          })
          clearOAuthRedirect(redirect.state)
          navigate(`/profile?oauth=linked&provider=${redirect.provider}`, {
            replace: true,
          })
        } catch (err) {
          clearOAuthRedirect(redirect.state)
          if (err instanceof ApiError && err.status === 409) {
            navigate(`/profile?oauth=conflict&provider=${redirect.provider}`, {
              replace: true,
            })
            return
          }
          setFailure({
            message:
              formErrorMessage(err, t("profile.oauth.linkFailed")) ||
              t("profile.oauth.linkFailed"),
            // 保留服务端原文，便于排查授权配置；本地化文案只用于 message。
            detail:
              err instanceof ApiError
                ? err.raw
                : err instanceof Error
                  ? err.message
                  : undefined,
            backTo: "profile",
          })
        }
        return
      }
      try {
        const result = await oauth2Login({
          provider: redirect.provider,
          code,
          redirectUri: redirect.redirectUri,
        })
        if (result.session) {
          storeSession(result.session)
          clearOAuthRedirect(redirect.state)
          navigate(readNextPath(redirect.next), { replace: true })
          return
        }
        // 第三方没有可用邮箱：留在回调页补充并验证邮箱，验证通过前不签发会话。
        if (result.emailBindingRequired && result.bindToken) {
          setEmailBinding({
            redirect,
            bindToken: result.bindToken,
            suggestedEmail: result.suggestedEmail ?? "",
          })
          return
        }
        clearOAuthRedirect(redirect.state)
        setFailure({ message: t("error.oauthFailed") })
      } catch (err) {
        clearOAuthRedirect(redirect.state)
        setFailure({ message: formErrorMessage(err, t("error.oauthFailed")) })
      }
    }
    void exchange()
  }, [navigate, params, t])

  // 绑定并登录成功：保存会话、清理本轮 OAuth 现场，回到登录前要去的页面。
  function onSignedIn(result: LoginResult) {
    if (!emailBinding) {
      return
    }
    storeSession(result.session)
    clearOAuthRedirect(emailBinding.redirect.state)
    navigate(readNextPath(emailBinding.redirect.next), { replace: true })
  }

  return (
    <AuthShell
      title={
        binding
          ? t("oauthCallback.linkedTitle")
          : emailBinding
            ? t("oauthCallback.bindingTitle")
            : t("oauthCallback.title")
      }
      subtitle={
        emailBinding
          ? t("oauthCallback.bindingSubtitle", {
              provider: oauthProviderName(emailBinding.redirect.provider),
            })
          : t("oauthCallback.subtitle")
      }
    >
      {emailBinding ? (
        <OAuthEmailBinding
          bindToken={emailBinding.bindToken}
          suggestedEmail={emailBinding.suggestedEmail}
          onSignedIn={onSignedIn}
        />
      ) : failure === null ? (
        <p
          role="status"
          className="flex items-center gap-2 text-sm text-muted-foreground"
        >
          <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          {binding
            ? t("oauthCallback.linking", {
                provider: enumLabel(t, binding),
              })
            : t("oauthCallback.processing")}
        </p>
      ) : (
        <div className="grid gap-4">
          <p role="alert" className="text-sm text-destructive">
            {failure.message}
          </p>
          {failure.detail ? (
            <p className="text-xs break-all text-muted-foreground">
              {failure.detail}
            </p>
          ) : null}
          <Button asChild variant="outline" className="w-full">
            {failure.backTo === "profile" ? (
              <Link to="/profile">{t("oauthCallback.backToProfile")}</Link>
            ) : (
              <Link to="/login">{t("oauthCallback.backToLogin")}</Link>
            )}
          </Button>
        </div>
      )}
    </AuthShell>
  )
}
