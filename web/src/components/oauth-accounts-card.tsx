import { KeyRound, Loader2 } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { listOAuth2Providers } from "@/api/auth"
import type { OAuth2Provider } from "@/api/auth/types"
import { ApiError } from "@/api/http"
import { listSelfOAuthLinks, unbindSelfOAuth } from "@/api/self"
import type { SelfOAuthLink } from "@/api/self/types"
import {
  newOAuthState,
  oauthRedirectUri,
  storeOAuthRedirect,
} from "@/auth/oauth-redirect"
import { ConfirmAction } from "@/components/confirm-action"
import { OAuthProviderMark } from "@/components/oauth-provider-mark"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/class-names"
import { formatDateTime } from "@/lib/date-time"
import { enumLabel } from "@/lib/enum-label"
import { failureMessage } from "@/lib/form-error"

/** 行模型：configured = 平台已配置该来源；link 有值 = 当前账号已绑定。 */
type ProviderRow = {
  /** 行标识：已绑定行用绑定 uid，未绑定行用来源，同一来源绑了多个身份也不会撞 key。 */
  key: string
  provider: OAuth2Provider
  link?: SelfOAuthLink
  configured: boolean
}

// 已知来源的展示顺序；未知来源（后端新增或来源已下线）追加在末尾。
const providerOrder: OAuth2Provider[] = [
  "GITHUB",
  "GITEE",
  "FEISHU",
  "DINGTALK",
]

/** open_id 只展示尾 4 位，完整值不下发到界面。 */
function maskOpenId(openId: string): string {
  if (!openId) {
    return ""
  }
  const tail = openId.slice(-4)
  return openId.length > 4 ? `…${tail}` : openId
}

/** 已绑定身份：优先第三方邮箱，其次掩码后的 open_id。 */
function linkAccount(link: SelfOAuthLink): string {
  return link.email || maskOpenId(link.openId)
}

function OAuthAccountsSkeleton() {
  return (
    <div className="flex items-center gap-3 py-3">
      <Skeleton className="size-9 shrink-0 rounded-lg" />
      <div className="grid min-w-0 flex-1 gap-2">
        <Skeleton className="h-4 w-24 max-w-full" />
        <Skeleton className="h-3 w-40 max-w-full" />
      </div>
      <Skeleton className="h-8 w-16 shrink-0" />
    </div>
  )
}

interface OAuthAccountRowProps {
  row: ProviderRow
  /** 本行正在取授权地址 / 跳转中。 */
  pending: boolean
  /** 其他行在跑或全局 pending。 */
  disabled: boolean
  onBind: (provider: OAuth2Provider) => void
  onUnbind: (link: SelfOAuthLink) => void
}

function OAuthAccountRow({
  row,
  pending,
  disabled,
  onBind,
  onUnbind,
}: OAuthAccountRowProps) {
  const { t } = useTranslation()
  const link = row.link
  const account = link ? linkAccount(link) : ""

  return (
    <li className="flex flex-wrap items-center gap-x-3 gap-y-2 py-3 sm:flex-nowrap">
      <OAuthProviderMark provider={row.provider} />
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-medium break-all">
            {enumLabel(t, row.provider)}
          </span>
          {link ? (
            <StatusBadge
              value="BOUND"
              label={t("profile.oauth.bound")}
              tone="success"
            />
          ) : (
            <StatusBadge
              value="UNBOUND"
              label={t("profile.oauth.notBound")}
              tone="neutral"
            />
          )}
        </div>
        {link ? (
          <p className="mt-1 text-xs break-all text-muted-foreground">
            {account || t("profile.oauth.accountUnknown")}
            {link.boundAt
              ? ` · ${t("profile.oauth.boundAt", {
                  time: formatDateTime(link.boundAt),
                })}`
              : null}
          </p>
        ) : null}
      </div>
      <div className="ml-auto flex shrink-0 items-center gap-2">
        {link ? (
          <Button
            variant="destructive"
            data-oauth-action={row.provider}
            disabled={disabled}
            onClick={() => onUnbind(link)}
          >
            {t("profile.oauth.unbind")}
          </Button>
        ) : (
          <Button
            variant="outline"
            data-oauth-action={row.provider}
            aria-busy={pending || undefined}
            disabled={disabled || pending || !row.configured}
            onClick={() => onBind(row.provider)}
          >
            {pending ? (
              <Loader2 className="size-4 animate-spin" aria-hidden="true" />
            ) : null}
            {pending ? t("profile.oauth.binding") : t("profile.oauth.bind")}
          </Button>
        )}
      </div>
    </li>
  )
}

/**
 * 个人资料页的「第三方账号」卡片。
 *
 * 行集合 = 已配置来源 ∪ 已绑定来源：只列已配置来源能避免死按钮，
 * 保留「已绑定但来源已下线」的行则让用户仍有解绑出口。
 * 列表 404 视为后端能力未部署，整块不渲染，页面保持原样。
 */
export function OAuthAccountsCard({ className }: { className?: string }) {
  const { t } = useTranslation()
  const [links, setLinks] = useState<SelfOAuthLink[] | null>(null)
  const [providers, setProviders] = useState<OAuth2Provider[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [forbidden, setForbidden] = useState(false)
  const [loadError, setLoadError] = useState("")
  // 后端未部署（列表 404）：整块不渲染，避免常驻一个「加载失败」。
  const [notDeployed, setNotDeployed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [binding, setBinding] = useState<OAuth2Provider | "">("")
  const [confirmTarget, setConfirmTarget] = useState<SelfOAuthLink | null>(null)
  const [unbinding, setUnbinding] = useState(false)
  const [unbindError, setUnbindError] = useState("")
  // 解绑成功后该行按钮已换成「绑定」，等重渲染后把焦点送回同一行的操作按钮。
  const focusAfterUnbind = useRef<OAuth2Provider | null>(null)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      setForbidden(false)
      setLoadError("")
      setNotDeployed(false)
      const redirectUri = oauthRedirectUri()
      // 配置来源与已绑来源互不阻塞：任一失败仍能渲染另一侧，或落到对应错误态。
      const [linksResult, providersResult] = await Promise.allSettled([
        listSelfOAuthLinks({ silent: true }),
        listOAuth2Providers({ redirectUri, state: newOAuthState() }),
      ])
      if (!active) {
        return
      }
      if (providersResult.status === "fulfilled") {
        setProviders(
          (providersResult.value.providers ?? []).map((item) => item.provider)
        )
      } else {
        setProviders([])
      }
      if (linksResult.status === "fulfilled") {
        setLinks(linksResult.value.items ?? [])
      } else if (
        linksResult.reason instanceof ApiError &&
        linksResult.reason.status === 404
      ) {
        setNotDeployed(true)
      } else if (
        linksResult.reason instanceof ApiError &&
        linksResult.reason.status === 403
      ) {
        setForbidden(true)
      } else {
        setFailed(true)
        setLoadError(failureMessage(linksResult.reason, ""))
      }
      setLoading(false)
    }
    void load()
    return () => {
      active = false
    }
  }, [attempt])

  useEffect(() => {
    const provider = focusAfterUnbind.current
    if (!provider) {
      return
    }
    focusAfterUnbind.current = null
    // 等对话框关闭后的焦点归位流程结束，再落到该行。
    const timer = window.setTimeout(() => {
      document
        .querySelector<HTMLButtonElement>(`[data-oauth-action="${provider}"]`)
        ?.focus()
    }, 0)
    return () => window.clearTimeout(timer)
  }, [links])

  const loadedLinks = links ?? []
  const providerSet = new Set<OAuth2Provider>([
    ...providers,
    ...loadedLinks.map((link) => link.provider),
  ])
  const orderedProviders = [
    ...providerOrder.filter((provider) => providerSet.has(provider)),
    ...[...providerSet].filter((provider) => !providerOrder.includes(provider)),
  ]
  // 同一来源允许绑多个第三方身份（后端按条存），因此每个绑定一条行，
  // 只有「已配置但未绑定」的来源才补一条未绑定行；否则第二条绑定会被吞掉。
  const rows: ProviderRow[] = orderedProviders.flatMap((provider) => {
    const configured = providers.includes(provider)
    const providerLinks = loadedLinks.filter(
      (link) => link.provider === provider
    )
    if (providerLinks.length === 0) {
      return [
        { key: `unlinked-${provider}`, provider, configured },
      ]
    }
    return providerLinks.map((link) => ({
      key: link.uid || `linked-${provider}-${link.openId}`,
      provider,
      link,
      configured,
    }))
  })

  if (notDeployed) {
    return null
  }

  /** 绑定：取带一次性 state 的授权地址，存现场后整页跳转第三方。 */
  async function onBind(provider: OAuth2Provider) {
    if (binding) {
      return
    }
    setBinding(provider)
    const providerName = enumLabel(t, provider)
    const redirectUri = oauthRedirectUri()
    const state = newOAuthState()
    try {
      const result = await listOAuth2Providers({ redirectUri, state })
      const authorizeUrl =
        result.providers?.find((item) => item.provider === provider)
          ?.authorizeUrl ?? ""
      if (!authorizeUrl) {
        toast.error(
          t("profile.oauth.authorizeUnavailable", { provider: providerName })
        )
        setBinding("")
        return
      }
      storeOAuthRedirect({
        state,
        provider,
        redirectUri,
        next: "/profile",
        intent: "bind",
      })
      // 与登录流程一致：整页跳转，回调页凭 state 取回本轮现场。
      window.location.assign(authorizeUrl)
    } catch {
      toast.error(
        t("profile.oauth.authorizeUnavailable", { provider: providerName })
      )
      setBinding("")
    }
  }

  function removeLink(link: SelfOAuthLink) {
    focusAfterUnbind.current = link.provider
    setLinks((prev) => (prev ?? []).filter((item) => item.uid !== link.uid))
    setConfirmTarget(null)
    setUnbindError("")
  }

  async function onConfirmUnbind() {
    const target = confirmTarget
    if (!target || unbinding) {
      return
    }
    const providerName = enumLabel(t, target.provider)
    setUnbinding(true)
    setUnbindError("")
    try {
      // 404 是「已经解绑」的正常分支，必须 silent，否则 http 层会先弹错误 toast。
      await unbindSelfOAuth(target.uid, { silent: true })
      removeLink(target)
      toast.success(
        t("profile.oauth.unlinkSuccess", { provider: providerName })
      )
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        removeLink(target)
        toast.success(
          t("profile.oauth.unlinkSuccess", { provider: providerName })
        )
      } else if (err instanceof ApiError && err.status === 409) {
        // 最后一个登录方式：保留对话框，让用户看到原因后再取消。
        setUnbindError(t("profile.oauth.unbindLastCredential"))
      } else {
        setUnbindError(t("profile.oauth.unlinkFailed"))
      }
    } finally {
      setUnbinding(false)
    }
  }

  const showSummary = !loading && !failed && !forbidden && rows.length > 0

  return (
    <Card className={cn("glass-card min-w-0", className)}>
      <CardHeader>
        <CardTitle>{t("profile.oauth.title")}</CardTitle>
        <CardDescription>{t("profile.oauth.description")}</CardDescription>
        {showSummary ? (
          <CardAction>
            <StatusBadge
              value="OAUTH_LINK_SUMMARY"
              label={t("profile.oauth.summary", {
                bound: rows.filter((row) => row.link).length,
                total: rows.length,
              })}
              tone="neutral"
            />
          </CardAction>
        ) : null}
      </CardHeader>
      <CardContent aria-busy={loading || undefined}>
        {loading ? (
          <div className="divide-y divide-border/60" aria-hidden="true">
            <OAuthAccountsSkeleton />
            <OAuthAccountsSkeleton />
            <OAuthAccountsSkeleton />
          </div>
        ) : failed ? (
          <div className="grid gap-3">
            <p role="alert" className="text-sm text-destructive">
              {loadError || t("state.failed")}
            </p>
            <Button
              variant="outline"
              className="justify-self-start"
              onClick={() => setAttempt((value) => value + 1)}
            >
              {t("action.retry")}
            </Button>
          </div>
        ) : forbidden ? (
          <div className="grid gap-3">
            <p role="alert" className="text-sm text-destructive">
              {t("profile.oauth.forbiddenHint")}
            </p>
            <Button
              variant="outline"
              className="justify-self-start"
              onClick={() => setAttempt((value) => value + 1)}
            >
              {t("action.retry")}
            </Button>
          </div>
        ) : rows.length === 0 ? (
          <div className="rounded-lg border border-dashed border-border/70 px-4 py-8 text-center">
            <KeyRound
              className="mx-auto size-5 text-muted-foreground/70"
              aria-hidden="true"
            />
            <p className="mt-2 text-sm font-medium">
              {t("profile.oauth.emptyTitle")}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              {t("profile.oauth.emptyHint")}
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-border/60">
            {rows.map((row) => (
              <OAuthAccountRow
                key={row.key}
                row={row}
                pending={binding === row.provider}
                disabled={binding !== "" && binding !== row.provider}
                onBind={(provider) => void onBind(provider)}
                onUnbind={(link) => {
                  setConfirmTarget(link)
                  setUnbindError("")
                }}
              />
            ))}
          </ul>
        )}
      </CardContent>
      <ConfirmAction
        open={confirmTarget !== null}
        title={
          confirmTarget
            ? t("profile.oauth.confirmTitle", {
                provider: enumLabel(t, confirmTarget.provider),
              })
            : ""
        }
        description={
          confirmTarget
            ? `${t("profile.oauth.confirmDescription")} ${t(
                "profile.oauth.accountLabel"
              )}: ${
                linkAccount(confirmTarget) || t("profile.oauth.accountUnknown")
              }`
            : ""
        }
        confirmLabel={t("profile.oauth.unbind")}
        pendingLabel={t("profile.oauth.unbinding")}
        cancelLabel={t("action.cancel")}
        pending={unbinding}
        error={unbindError || undefined}
        onConfirm={() => void onConfirmUnbind()}
        onOpenChange={(open) => {
          if (open || unbinding) {
            return
          }
          setConfirmTarget(null)
          setUnbindError("")
        }}
      />
    </Card>
  )
}
