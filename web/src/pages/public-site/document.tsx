import { ArrowLeft, RefreshCw } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useParams } from "react-router-dom"

import error404Illustration from "@/assets/undraw/error-404.svg"
import { fetchPublicSitePage, SitePublicError } from "@/api/site/public"
import type { PublicSitePage } from "@/api/site/types"
import { ErrorPage } from "@/components/error-page"
import { SiteShell } from "@/components/site-shell"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useSiteLocale } from "@/hooks/use-site-locale"
import { formatDateTime } from "@/lib/date-time"

import { SiteArticle } from "./article"

type PageState = "loading" | "ready" | "failed" | "notFound"

export function DocumentSkeleton() {
  const { t } = useTranslation()
  return (
    <div
      className="mx-auto grid w-full max-w-3xl gap-4 px-4 py-12 sm:px-6 sm:py-16"
      aria-busy="true"
      aria-label={t("site.loading")}
    >
      <Skeleton className="h-4 w-32" />
      <Skeleton className="h-10 w-full max-w-xl" />
      <Skeleton className="h-4 w-40" />
      <Skeleton className="mt-4 h-4 w-full" />
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-3/4" />
    </div>
  )
}

/** 官网子页：标题、发布时间与正文；语言回退时在正文上方给出提示。 */
export function SiteDocumentPage() {
  const { t } = useTranslation()
  const { slug = "" } = useParams()
  const locale = useSiteLocale()
  const [page, setPage] = useState<PublicSitePage | null>(null)
  const [state, setState] = useState<PageState>("loading")
  const [settledKey, setSettledKey] = useState("")
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const requestKey = `${slug}|${locale}`

  useEffect(() => {
    let active = true
    if (!slug) {
      return () => {
        active = false
      }
    }
    void fetchPublicSitePage(slug, locale)
      .then((nextPage) => {
        if (!active) {
          return
        }
        setPage(nextPage)
        setSettledKey(requestKey)
        setState("ready")
      })
      .catch((err: unknown) => {
        if (!active) {
          return
        }
        setSettledKey(requestKey)
        if (err instanceof SitePublicError && err.status === 404) {
          setState("notFound")
          return
        }
        setError(
          err instanceof SitePublicError && err.status === 0
            ? t("site.loadFailedHint")
            : ""
        )
        setState("failed")
      })
    return () => {
      active = false
    }
  }, [attempt, requestKey, slug, locale, t])

  function retry() {
    setState("loading")
    setAttempt((value) => value + 1)
  }

  // 没有 slug 时直接按 404 处理，不经过请求；语言或 slug 变化后旧内容先让位给骨架。
  const loading =
    slug !== "" && (state === "loading" || settledKey !== requestKey)
  const notFound = slug === "" || (!loading && state === "notFound")
  const fallbackLanguage = t(page?.locale === "en" ? "lang.en" : "lang.zh")

  return (
    <SiteShell>
      {loading ? <DocumentSkeleton /> : null}
      {notFound ? (
        <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col px-4 py-8 sm:px-6">
          <ErrorPage
            status={404}
            label="Not Found"
            title={t("site.pageNotFoundTitle")}
            description={t("site.pageNotFoundDescription")}
            image={error404Illustration}
          >
            <Button asChild>
              <Link to="/">{t("site.home")}</Link>
            </Button>
          </ErrorPage>
        </div>
      ) : null}
      {!loading && state === "failed" ? (
        <div className="mx-auto flex w-full max-w-md flex-1 flex-col items-center justify-center gap-3 px-4 py-16 text-center">
          <p className="font-heading text-xl font-semibold">
            {t("site.loadFailedTitle")}
          </p>
          <p className="text-sm break-words text-muted-foreground">
            {error || t("site.loadFailedHint")}
          </p>
          <Button variant="outline" className="mt-2" onClick={retry}>
            <RefreshCw aria-hidden="true" data-icon="inline-start" />
            {t("site.retry")}
          </Button>
        </div>
      ) : null}
      {!loading && state === "ready" && page ? (
        <article className="mx-auto w-full max-w-5xl px-4 py-12 sm:px-6 sm:py-16">
          <SiteArticle
            title={page.title || page.slug}
            beforeTitle={
              <Button variant="ghost" size="sm" className="mb-6 -ml-2" asChild>
                <Link to="/">
                  <ArrowLeft aria-hidden="true" data-icon="inline-start" />
                  {t("site.home")}
                </Link>
              </Button>
            }
            meta={
              page.publishedAt ? (
                <p className="mt-3 font-mono text-xs text-muted-foreground tabular-nums">
                  {t("site.publishedAt", {
                    time: formatDateTime(page.publishedAt),
                  })}
                </p>
              ) : null
            }
            coverImageUrl={page.coverImageUrl}
            bodyMd={page.bodyMd}
            emptyText={t("site.emptyBody")}
          >
            {page.localeFallback ? (
              <p
                role="status"
                className="mt-6 rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-sm text-warning"
              >
                {t("site.localeFallback", { language: fallbackLanguage })}
              </p>
            ) : null}
          </SiteArticle>
        </article>
      ) : null}
    </SiteShell>
  )
}
