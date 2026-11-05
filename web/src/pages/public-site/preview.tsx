import {
  ArrowLeft,
  Eye,
  LockKeyhole,
  LogIn,
  RefreshCw,
  TriangleAlert,
} from "lucide-react"
import { useEffect, useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Link, useParams } from "react-router-dom"

import {
  fetchSiteHomeDraft,
  fetchSitePageDraft,
  SitePreviewError,
} from "@/api/site/preview"
import { fetchPublicSiteNav } from "@/api/site/public"
import type {
  PublicSitePage,
  SiteHome,
  SiteNavItem,
  SitePageDetail,
} from "@/api/site/types"
import { SiteShell } from "@/components/site-shell"
import { Button } from "@/components/ui/button"
import { useSiteConfig } from "@/hooks/use-site-config"
import { formatDateTime } from "@/lib/date-time"

import { SiteArticle } from "./article"
import { DocumentSkeleton } from "./document"
import { EMPTY_HOME, SiteHomeSections } from "./home/sections"
import { HomeSkeleton } from "./home/skeleton"

type PreviewState =
  | "loading"
  | "ready"
  | "unauthenticated"
  | "forbidden"
  | "notFound"
  | "failed"

/** 预览页头部：补齐草稿的 document.title / description，并阻止搜索引擎索引。 */
function usePreviewHead(detail: SitePageDetail | null) {
  useEffect(() => {
    if (!detail) {
      return
    }
    const previousTitle = document.title
    const title =
      detail.draft.seoTitle.trim() ||
      detail.draft.title.trim() ||
      detail.item.slug
    const description =
      detail.draft.seoDescription.trim() || detail.draft.summary.trim()
    document.title = title

    // 记录被改写的 meta，卸载时恢复原值；本次新建的标签直接移除。
    const previous = new Map<string, string | null>()
    const created: HTMLMetaElement[] = []
    function setMeta(name: string, content: string) {
      const existing = document.head.querySelector<HTMLMetaElement>(
        `meta[name="${name}"]`
      )
      previous.set(name, existing?.getAttribute("content") ?? null)
      if (existing) {
        existing.setAttribute("content", content)
        return
      }
      const tag = document.createElement("meta")
      tag.name = name
      tag.content = content
      document.head.appendChild(tag)
      created.push(tag)
    }

    setMeta("robots", "noindex")
    if (description) {
      setMeta("description", description)
    }

    return () => {
      document.title = previousTitle
      for (const tag of created) {
        tag.remove()
      }
      for (const [name, content] of previous) {
        if (content === null) {
          continue
        }
        document.head
          .querySelector(`meta[name="${name}"]`)
          ?.setAttribute("content", content)
      }
    }
  }, [detail])
}

function PreviewBanner({ uid }: { uid: string }) {
  const { t } = useTranslation()
  return (
    <div role="note" className="border-b border-warning-border bg-warning-soft">
      <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center gap-x-3 gap-y-2 px-4 py-2.5 sm:px-6">
        <span className="flex min-w-0 items-center gap-2 text-sm text-warning">
          <Eye aria-hidden="true" className="size-4 shrink-0" />
          <span className="font-medium break-words">
            {t("sitePages.preview.banner")}
          </span>
        </span>
        <Button asChild variant="outline" size="sm" className="ml-auto">
          <Link to={`/system/site/pages?uid=${encodeURIComponent(uid)}`}>
            <ArrowLeft aria-hidden="true" data-icon="inline-start" />
            {t("sitePages.preview.backToEditor")}
          </Link>
        </Button>
      </div>
    </div>
  )
}

function StatusPanel({
  icon,
  title,
  body,
  tone = "status",
  children,
}: {
  icon: ReactNode
  title: string
  body: string
  tone?: "status" | "alert"
  children?: ReactNode
}) {
  return (
    <div
      role={tone === "alert" ? "alert" : undefined}
      className="mx-auto flex w-full max-w-md flex-1 flex-col items-center justify-center gap-3 px-4 py-16 text-center"
    >
      <span
        aria-hidden="true"
        className="grid size-10 place-items-center rounded-full bg-muted text-muted-foreground"
      >
        {icon}
      </span>
      <p className="font-heading text-xl font-semibold break-words">{title}</p>
      <p className="text-sm break-words text-muted-foreground">{body}</p>
      {children}
    </div>
  )
}

/**
 * 官网页面预览：读取后台草稿，套用公开官网的页头页脚与真实版式。
 *
 * - 首页（slug=home）用固定首页版式，文案按草稿语言读 `/v1/system/site/home`；
 * - 其它页面复用 `/p/:slug` 的文档版式，数据全部来自草稿；
 * - 未登录（401）时留在本页提示「请先登录」，不跳走登录页。
 */
export function SitePreviewPage() {
  const { t } = useTranslation()
  const { uid = "" } = useParams()
  const { config, failed: configFailed, reload: reloadConfig } = useSiteConfig()
  const [attempt, setAttempt] = useState(0)
  // 请求结果与它对应的 requestKey 一起保存：key 不匹配就视为仍在加载，
  // 这样重试 / 重新打开时不需要在 effect 里同步 setState。
  const [settled, setSettled] = useState<{
    key: string
    state: Exclude<PreviewState, "loading">
    detail: SitePageDetail | null
  } | null>(null)
  const requestKey = `${uid}|${attempt}`

  useEffect(() => {
    if (!uid) {
      return
    }
    let active = true
    void fetchSitePageDraft(uid)
      .then((page) => {
        if (!active) {
          return
        }
        setSettled({ key: requestKey, state: "ready", detail: page })
      })
      .catch((err: unknown) => {
        if (!active) {
          return
        }
        const status = err instanceof SitePreviewError ? err.status : 0
        const nextState: Exclude<PreviewState, "loading"> =
          status === 401
            ? "unauthenticated"
            : status === 403
              ? "forbidden"
              : status === 404
                ? "notFound"
                : "failed"
        setSettled({ key: requestKey, state: nextState, detail: null })
      })
    return () => {
      active = false
    }
  }, [requestKey, uid])

  const loading = uid !== "" && settled?.key !== requestKey
  const state: PreviewState = uid === "" ? "notFound" : loading ? "loading" : settled?.state ?? "failed"
  const detail = state === "ready" ? settled?.detail ?? null : null

  // 首页文案与导航是附加内容：任一失败都降级为默认文案 / 空导航，不让整页失败。
  // 同样用 key 标记「本轮是否已经落地」，请求期间先显示首页骨架。
  const homeKey =
    detail && detail.item.slug === "home"
      ? `${detail.item.uid}|${detail.item.locale}`
      : ""
  const [homeSettled, setHomeSettled] = useState<{
    key: string
    copy: SiteHome | null
    navItems: SiteNavItem[]
  } | null>(null)

  useEffect(() => {
    if (!detail || homeKey === "") {
      return
    }
    let active = true
    const contentLocale = detail.item.locale === "en" ? "en" : "zh"
    void Promise.allSettled([
      fetchSiteHomeDraft(contentLocale),
      fetchPublicSiteNav(contentLocale),
    ]).then(([homeResult, navResult]) => {
      if (!active) {
        return
      }
      setHomeSettled({
        key: homeKey,
        copy: homeResult.status === "fulfilled" ? homeResult.value : null,
        navItems: navResult.status === "fulfilled" ? navResult.value : [],
      })
    })
    return () => {
      active = false
    }
  }, [detail, homeKey])

  usePreviewHead(detail)

  const isHome = detail?.item.slug === "home"
  const homeReady = homeKey !== "" && homeSettled?.key === homeKey
  const homeCopy = homeReady ? homeSettled.copy : null
  const navItems = homeReady ? homeSettled.navItems : []
  const loginHref = `/login?next=${encodeURIComponent(`/site-preview/${uid}`)}`
  const previewConfig =
    config && isHome
      ? {
          ...config,
          home: homeCopy?.home ?? EMPTY_HOME,
          features: homeCopy?.features ?? [],
        }
      : null
  const previewPage: PublicSitePage | null = detail
    ? {
        slug: detail.item.slug,
        locale: detail.item.locale,
        title: detail.draft.title,
        summary: detail.draft.summary,
        bodyMd: detail.draft.bodyMd,
        coverImageUrl: detail.draft.coverImageUrl,
        seoTitle: detail.draft.seoTitle,
        seoDescription: detail.draft.seoDescription,
        publishedAt: "",
        localeFallback: false,
      }
    : null
  const contentLocaleLabel = t(
    detail?.item.locale === "en"
      ? "sitePages.field.localeEn"
      : "sitePages.field.localeZh"
  )

  return (
    <SiteShell>
      <PreviewBanner uid={uid} />

      {state === "loading" ? <DocumentSkeleton /> : null}

      {state === "unauthenticated" ? (
        <StatusPanel
          icon={<LockKeyhole className="size-5" />}
          title={t("sitePages.preview.loginTitle")}
          body={t("sitePages.preview.loginBody")}
        >
          <Button asChild className="mt-2">
            <Link to={loginHref}>
              <LogIn aria-hidden="true" data-icon="inline-start" />
              {t("site.login")}
            </Link>
          </Button>
        </StatusPanel>
      ) : null}

      {state === "forbidden" ? (
        <StatusPanel
          icon={<LockKeyhole className="size-5" />}
          title={t("sitePages.preview.forbiddenTitle")}
          body={t("sitePages.preview.forbiddenBody")}
          tone="alert"
        />
      ) : null}

      {state === "notFound" ? (
        <StatusPanel
          icon={<TriangleAlert className="size-5" />}
          title={t("error.sitePageNotFound")}
          body={t("site.pageNotFoundDescription")}
          tone="alert"
        >
          <Button asChild variant="outline" className="mt-2">
            <Link to="/">{t("site.home")}</Link>
          </Button>
        </StatusPanel>
      ) : null}

      {state === "failed" ? (
        <StatusPanel
          icon={<TriangleAlert className="size-5" />}
          title={t("site.loadFailedTitle")}
          body={t("site.loadFailedHint")}
          tone="alert"
        >
          <Button
            variant="outline"
            className="mt-2"
            onClick={() => setAttempt((value) => value + 1)}
          >
            <RefreshCw aria-hidden="true" data-icon="inline-start" />
            {t("site.retry")}
          </Button>
        </StatusPanel>
      ) : null}

      {state === "ready" && detail && isHome ? (
        configFailed ? (
          <StatusPanel
            icon={<TriangleAlert className="size-5" />}
            title={t("site.loadFailedTitle")}
            body={t("site.loadFailedHint")}
            tone="alert"
          >
            <Button variant="outline" className="mt-2" onClick={reloadConfig}>
              <RefreshCw aria-hidden="true" data-icon="inline-start" />
              {t("site.retry")}
            </Button>
          </StatusPanel>
        ) : previewConfig && homeReady ? (
          <SiteHomeSections
            config={previewConfig}
            page={previewPage}
            navItems={navItems}
          />
        ) : (
          <HomeSkeleton />
        )
      ) : null}

      {state === "ready" && detail && !isHome ? (
        <article className="mx-auto w-full max-w-3xl px-4 py-12 sm:px-6 sm:py-16">
          <SiteArticle
            title={detail.draft.title.trim() || detail.item.slug}
            meta={
              <p className="mt-3 flex flex-wrap items-center gap-x-2 font-mono text-xs text-muted-foreground tabular-nums">
                <span>{contentLocaleLabel}</span>
                <span aria-hidden="true">·</span>
                <span>
                  {t("sitePages.updatedAt", {
                    time: formatDateTime(detail.item.updatedAt),
                  })}
                </span>
              </p>
            }
            coverImageUrl={detail.draft.coverImageUrl}
            bodyMd={detail.draft.bodyMd}
            emptyText={t("site.emptyBody")}
          />
        </article>
      ) : null}
    </SiteShell>
  )
}
