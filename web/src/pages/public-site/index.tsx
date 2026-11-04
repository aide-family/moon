import { RefreshCw } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import error404Illustration from "@/assets/undraw/error-404.svg"
import {
  fetchPublicSiteNav,
  fetchPublicSitePage,
  SitePublicError,
} from "@/api/site/public"
import type { PublicSitePage, SiteNavItem } from "@/api/site/types"
import { ErrorPage } from "@/components/error-page"
import { SiteBody, excerptOf } from "@/components/site-body"
import { SiteShell, useSiteAccount } from "@/components/site-shell"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useSiteConfig } from "@/hooks/use-site-config"
import { useSiteLocale } from "@/hooks/use-site-locale"

type PageState = "loading" | "ready" | "failed" | "notFound"

function HomeSkeleton() {
  const { t } = useTranslation()
  return (
    <div
      className="mx-auto grid w-full max-w-6xl gap-4 px-4 py-12 sm:px-6 sm:py-16"
      aria-busy="true"
      aria-label={t("site.loading")}
    >
      <Skeleton className="h-4 w-28" />
      <Skeleton className="h-10 w-full max-w-xl" />
      <Skeleton className="h-4 w-full max-w-2xl" />
      <Skeleton className="h-4 w-2/3 max-w-xl" />
      <Skeleton className="mt-4 h-40 w-full max-w-3xl rounded-xl" />
      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <Skeleton className="h-24 rounded-xl" />
        <Skeleton className="h-24 rounded-xl" />
      </div>
    </div>
  )
}

function LoadFailed({
  error,
  onRetry,
}: {
  error: string
  onRetry: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="mx-auto flex w-full max-w-md flex-1 flex-col items-center justify-center gap-3 px-4 py-16 text-center">
      <p className="font-heading text-xl font-semibold">
        {t("site.loadFailedTitle")}
      </p>
      <p className="text-sm break-words text-muted-foreground">
        {error || t("site.loadFailedHint")}
      </p>
      <Button variant="outline" className="mt-2" onClick={onRetry}>
        <RefreshCw aria-hidden="true" data-icon="inline-start" />
        {t("site.retry")}
      </Button>
    </div>
  )
}

function SiteNavSection({ items }: { items: SiteNavItem[] }) {
  const { t } = useTranslation()
  if (items.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">{t("site.navEmpty")}</p>
    )
  }
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {items.map((item) => (
        <Link
          key={item.slug}
          to={`/p/${encodeURIComponent(item.slug)}`}
          className="group flex min-w-0 flex-col gap-2 rounded-xl border border-border/60 bg-card p-4 transition-colors hover:border-primary/40 hover:bg-muted/40 focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none"
        >
          <span className="min-w-0 font-heading text-base font-medium break-words group-hover:text-primary-ink">
            {item.title || item.slug}
          </span>
          <span className="truncate font-mono text-xs text-muted-foreground">
            /p/{item.slug}
          </span>
        </Link>
      ))}
    </div>
  )
}

function HomeContent({
  page,
  navItems,
}: {
  page: PublicSitePage
  navItems: SiteNavItem[]
}) {
  const { t } = useTranslation()
  const { config } = useSiteConfig()
  const { loggedIn, user, entering, enterAdmin, loginHref, registerHref } =
    useSiteAccount()
  const siteName = config?.site.siteName.trim() || t("site.brandFallback")
  const excerpt = excerptOf(page.bodyMd)
  const isEmpty = !page.summary.trim() && !excerpt.trim()

  return (
    <div className="mx-auto grid w-full max-w-6xl gap-12 px-4 py-12 sm:px-6 sm:py-16">
      <section className="max-w-3xl">
        <p className="font-mono text-xs tracking-[0.16em] text-muted-foreground uppercase">
          {siteName}
        </p>
        <h1 className="mt-2 font-heading text-3xl font-semibold tracking-tight break-words sm:text-4xl">
          {page.title || siteName}
        </h1>
        {page.summary.trim() ? (
          <p className="mt-3 text-base leading-7 break-words text-muted-foreground">
            {page.summary}
          </p>
        ) : null}
        {page.coverImageUrl ? (
          <img
            src={page.coverImageUrl}
            alt=""
            className="mt-6 w-full rounded-2xl border border-border/60 object-cover"
          />
        ) : null}
        {isEmpty ? (
          <div className="mt-8 rounded-xl border border-dashed border-border/70 px-4 py-8 text-center">
            <p className="font-medium">{t("site.emptyTitle")}</p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("site.emptyHint")}
            </p>
          </div>
        ) : excerpt ? (
          <SiteBody
            content={excerpt}
            emptyText={t("site.emptyBody")}
            className="mt-6"
          />
        ) : null}
      </section>

      <section aria-labelledby="site-nav-title" className="grid gap-4">
        <h2
          id="site-nav-title"
          className="font-heading text-xl font-semibold tracking-tight"
        >
          {t("site.navTitle")}
        </h2>
        <SiteNavSection items={navItems} />
      </section>

      <section className="rounded-2xl border border-border/60 bg-muted/30 px-5 py-8 sm:px-8">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div className="max-w-xl">
            <h2 className="font-heading text-xl font-semibold tracking-tight">
              {t("site.cta.title")}
            </h2>
            <p className="mt-2 text-sm leading-6 break-words text-muted-foreground">
              {t("site.cta.body")}
            </p>
            {loggedIn ? (
              <p className="mt-2 text-xs text-muted-foreground">
                {user?.email || user?.name || t("site.account")}
              </p>
            ) : null}
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {loggedIn ? (
              <Button disabled={entering} onClick={enterAdmin}>
                {entering ? t("site.tenant.entering") : t("site.enterAdmin")}
              </Button>
            ) : (
              <>
                <Button variant="outline" asChild>
                  <Link to={loginHref}>{t("site.login")}</Link>
                </Button>
                <Button asChild>
                  <Link to={registerHref}>{t("site.register")}</Link>
                </Button>
              </>
            )}
          </div>
        </div>
      </section>
    </div>
  )
}

/** 官网首页：已发布 home 页 + 公开导航 + 尾部 CTA。 */
export function SiteHomePage() {
  const { t } = useTranslation()
  const locale = useSiteLocale()
  const [page, setPage] = useState<PublicSitePage | null>(null)
  const [navItems, setNavItems] = useState<SiteNavItem[]>([])
  const [state, setState] = useState<PageState>("loading")
  // 已经结束请求的语言：切换语言后旧内容不再冒充新语言，先显示骨架。
  const [settledLocale, setSettledLocale] = useState("")
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let active = true
    // 导航是首页的附加内容：它失败时首页仍可展示，不把整页拖进失败态。
    const navRequest = fetchPublicSiteNav(locale).catch(() => [])
    void Promise.all([fetchPublicSitePage("home", locale), navRequest])
      .then(([nextPage, nextNav]) => {
        if (!active) {
          return
        }
        setPage(nextPage)
        setNavItems(nextNav)
        setSettledLocale(locale)
        setState("ready")
      })
      .catch((err: unknown) => {
        if (!active) {
          return
        }
        setSettledLocale(locale)
        if (err instanceof SitePublicError && err.status === 404) {
          setState("notFound")
          return
        }
        setState("failed")
        setError(
          err instanceof SitePublicError && err.status === 0
            ? t("site.loadFailedHint")
            : ""
        )
      })
    return () => {
      active = false
    }
  }, [locale, attempt, t])

  function retry() {
    setState("loading")
    setAttempt((value) => value + 1)
  }

  const loading = state === "loading" || settledLocale !== locale

  return (
    <SiteShell>
      {loading ? <HomeSkeleton /> : null}
      {!loading && state === "notFound" ? (
        <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col px-4 py-8 sm:px-6">
          <ErrorPage
            status={404}
            label="Not Found"
            title={t("site.homeNotFoundTitle")}
            description={t("site.homeNotFoundDescription")}
            image={error404Illustration}
          />
        </div>
      ) : null}
      {!loading && state === "failed" ? (
        <LoadFailed error={error} onRetry={retry} />
      ) : null}
      {!loading && state === "ready" && page ? (
        <HomeContent page={page} navItems={navItems} />
      ) : null}
    </SiteShell>
  )
}
