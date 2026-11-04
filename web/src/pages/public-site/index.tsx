import { RefreshCw } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { fetchPublicSiteNav, fetchPublicSitePage } from "@/api/site/public"
import type { PublicSitePage, SiteNavItem } from "@/api/site/types"
import { SiteShell } from "@/components/site-shell"
import { Button } from "@/components/ui/button"
import { useSiteConfig } from "@/hooks/use-site-config"
import { useSiteLocale } from "@/hooks/use-site-locale"

import { SiteHomeSections } from "./home/sections"
import { HomeSkeleton } from "./home/skeleton"

function LoadFailed({ onRetry }: { onRetry: () => void }) {
  const { t } = useTranslation()
  return (
    <div
      role="alert"
      className="site-container flex flex-col items-start gap-3 py-16"
    >
      <p className="text-section">{t("site.loadFailedTitle")}</p>
      <p className="max-w-[52ch] text-body break-words text-muted-foreground">
        {t("site.loadFailedHint")}
      </p>
      <Button variant="outline" className="mt-2" onClick={onRetry}>
        <RefreshCw aria-hidden="true" data-icon="inline-start" />
        {t("site.retry")}
      </Button>
    </div>
  )
}

/**
 * 官网首页：版式固定，内容来自 `GET /v1/site/config` 的 home / features。
 *
 * 状态划分：
 * - 主数据是站点配置（站点名、首页文案、能力卡片），失败时整页可重试；
 * - 产品说明正文与页面导航是附加内容，失败时只隐藏对应区块，不把整页拖进失败态，
 *   所以 `home` 页从未发布（404）也不再让首页 404；
 * - 已发布页面走 `/p/:slug`，语义不变。
 */
export function SiteHomePage() {
  const locale = useSiteLocale()
  const {
    config,
    loading: configLoading,
    failed,
    reload,
    locale: configLocale,
  } = useSiteConfig()
  const [supplementary, setSupplementary] = useState<{
    key: string
    page: PublicSitePage | null
    navItems: SiteNavItem[]
  } | null>(null)
  const [attempt, setAttempt] = useState(0)
  const requestKey = `${locale}|${attempt}`

  useEffect(() => {
    let active = true
    // 附加内容失败时只降级：正文缺失就不渲染摘要区，导航缺失就不渲染导航列。
    const pageRequest = fetchPublicSitePage("home", locale).catch(() => null)
    const navRequest = fetchPublicSiteNav(locale).catch(() => [])
    void Promise.all([pageRequest, navRequest]).then(([nextPage, nextNav]) => {
      if (!active) {
        return
      }
      setSupplementary({
        key: requestKey,
        page: nextPage,
        navItems: nextNav,
      })
    })
    return () => {
      active = false
    }
  }, [locale, requestKey])

  function retry() {
    reload()
    setAttempt((value) => value + 1)
  }

  // 请求 key 没落地的内容视为仍在加载：切换语言或重试后不显示上一轮的内容。
  const supplementaryReady = supplementary?.key === requestKey
  // 语言切换后，旧语言的首页文案不能冒充新语言：在配置落到当前语言前一直显示骨架。
  const staleConfig = config === null || configLocale !== locale
  const loading =
    (config === null && configLoading) ||
    (!failed && (staleConfig || !supplementaryReady))

  return (
    <SiteShell>
      {loading ? <HomeSkeleton /> : null}
      {!loading && failed ? <LoadFailed onRetry={retry} /> : null}
      {!loading && !failed && config ? (
        <SiteHomeSections
          config={config}
          page={supplementary?.page ?? null}
          navItems={supplementary?.navItems ?? []}
        />
      ) : null}
    </SiteShell>
  )
}
