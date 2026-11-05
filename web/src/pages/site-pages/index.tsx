import { CircleAlert, Plus, RefreshCw, Search } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useSearchParams } from "react-router-dom"

import {
  getSitePage,
  listSitePages,
} from "@/api/site/pages"
import type { SitePageDetail, SitePageItem } from "@/api/site/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { FilterChips, type FilterChip } from "@/components/filter-chips"
import { FilterToggle } from "@/components/filter-toggle"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { usePagedList } from "@/hooks/use-paged-list"
import { cn } from "@/lib/class-names"
import { formatDateTime } from "@/lib/date-time"

import { PageEditor, type PageEditorHandle } from "./editor"

// Radix Select 不允许空字符串选项：用哨兵值表示「不限」。
const ALL = "ALL"

function boundedInt(
  value: string | null,
  fallback: number,
  min: number,
  max: number
): number {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed < min || parsed > max) {
    return fallback
  }
  return parsed
}

function localeLabel(
  t: (key: string) => string,
  locale: string
): string {
  return locale === "en"
    ? t("sitePages.field.localeEn")
    : t("sitePages.field.localeZh")
}

function stateLabel(
  t: (key: string, options?: Record<string, unknown>) => string,
  value: string
): string {
  return t(`sitePages.state.${value}`, { defaultValue: value })
}

function stateTone(value: string): "success" | "warning" | "danger" {
  if (value === "PUBLISHED") {
    return "success"
  }
  if (value === "ARCHIVED") {
    return "danger"
  }
  return "warning"
}

function ListItem({
  item,
  active,
  onSelect,
}: {
  item: SitePageItem
  active: boolean
  onSelect: () => void
}) {
  const { t } = useTranslation()
  return (
    <button
      type="button"
      aria-current={active ? "true" : undefined}
      onClick={onSelect}
      className={cn(
        "flex w-full min-w-0 flex-col gap-1.5 rounded-lg border border-transparent px-3 py-2.5 text-left transition-colors outline-none focus-visible:ring-3 focus-visible:ring-ring",
        active ? "border-primary/30 bg-primary-soft" : "hover:bg-muted/60"
      )}
    >
      <span className="flex min-w-0 items-start justify-between gap-2">
        <span className="min-w-0 font-medium break-words">
          {item.title || item.slug}
        </span>
        <StatusBadge
          value={item.state}
          label={stateLabel(t, item.state)}
          tone={stateTone(item.state)}
          className="shrink-0"
        />
      </span>
      <span className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
        <span className="font-mono">/p/{item.slug}</span>
        <span aria-hidden="true">·</span>
        <span>{localeLabel(t, item.locale)}</span>
        {item.slug === "home" ? (
          <StatusBadge
            value="HOME"
            label={t("sitePages.homeBadge")}
            tone="primary"
            className="px-1.5 py-0"
          />
        ) : null}
        {item.hasUnpublishedChanges ? (
          <span className="inline-flex items-center gap-1 text-warning">
            <span
              aria-hidden="true"
              className="size-1.5 rounded-full bg-warning"
            />
            {t("sitePages.unpublishedChanges")}
          </span>
        ) : null}
      </span>
      <span className="font-mono text-[11px] text-muted-foreground tabular-nums">
        {item.updatedAt
          ? t("sitePages.updatedAt", {
              time: formatDateTime(item.updatedAt),
            })
          : ""}
      </span>
    </button>
  )
}

/** 后台官网页面管理：左侧列表 + 右侧编辑器（草稿 / 发布 / 导航）。 */
export function SitePagesPage() {
  const { t } = useTranslation()
  const session = useSession()
  const canRead = session.canSystem("system:site:read")
  const canWrite = session.canSystem("system:site:write")
  const [searchParams, setSearchParams] = useSearchParams()

  const page = boundedInt(searchParams.get("page"), 1, 1, 1_000_000)
  const pageSize = boundedInt(searchParams.get("pageSize"), 20, 1, 100)
  const keyword = searchParams.get("keyword") ?? ""
  const rawState = searchParams.get("state") ?? ""
  const stateFilter =
    rawState === "DRAFT" ||
    rawState === "PUBLISHED" ||
    rawState === "ARCHIVED"
      ? rawState
      : ""
  const rawLocale = searchParams.get("locale") ?? ""
  const localeFilter =
    rawLocale === "zh" || rawLocale === "en" ? rawLocale : ""

  const [draftKeyword, setDraftKeyword] = useState(keyword)
  const [filtersOpen, setFiltersOpen] = useState(false)
  // 预览页的「返回编辑」带 uid 回来，直接选中该页面继续编辑。
  const initialUid = searchParams.get("uid") ?? ""
  const [mode, setMode] = useState<"none" | "edit" | "create">(
    initialUid ? "edit" : "none"
  )
  const [selectedUid, setSelectedUid] = useState(initialUid)
  const [detail, setDetail] = useState<SitePageDetail | null>(null)
  const [failedUid, setFailedUid] = useState("")
  const [reloadToken, setReloadToken] = useState(0)
  const [pendingSelect, setPendingSelect] = useState<
    { kind: "select"; uid: string } | { kind: "create" } | null
  >(null)
  const editorRef = useRef<PageEditorHandle | null>(null)

  const { items, total, loading, failed, reload } = usePagedList<SitePageItem>(
    {
      page,
      pageSize,
      filterKey: `${keyword}|${stateFilter}|${localeFilter}`,
      fetch: (nextPage, nextPageSize) =>
        listSitePages({
          page: nextPage,
          pageSize: nextPageSize,
          keyword,
          state: stateFilter,
          locale: localeFilter,
        }),
    }
  )

  useEffect(() => {
    if (mode !== "edit" || !selectedUid) {
      return
    }
    let active = true
    void getSitePage(selectedUid)
      .then((result) => {
        if (!active) {
          return
        }
        setDetail(result.page)
        setFailedUid("")
      })
      .catch(() => {
        if (!active) {
          return
        }
        setDetail(null)
        setFailedUid(selectedUid)
      })
    return () => {
      active = false
    }
  }, [mode, selectedUid, reloadToken])

  if (!session.loading && !canRead) {
    return <Navigate to="/forbidden" replace />
  }

  function updateQuery(next: {
    page?: number
    keyword?: string
    state?: string
    locale?: string
  }) {
    const query = new URLSearchParams(searchParams)
    if (next.page !== undefined) {
      query.set("page", String(next.page))
    }
    for (const key of ["keyword", "state", "locale"] as const) {
      const value = next[key]
      if (value === undefined) {
        continue
      }
      if (value === "") {
        query.delete(key)
      } else {
        query.set(key, value)
      }
      query.set("page", "1")
    }
    query.set("pageSize", String(pageSize))
    setSearchParams(query)
  }

  function guardDirty(): boolean {
    return editorRef.current?.isDirty() ?? false
  }

  /** 预览页「返回编辑」带来的 uid 只用一次：开始操作后从地址栏移除，避免刷新回到旧页面。 */
  function clearDeepLinkUid() {
    if (!searchParams.has("uid")) {
      return
    }
    setSearchParams(
      (previous) => {
        const next = new URLSearchParams(previous)
        next.delete("uid")
        return next
      },
      { replace: true }
    )
  }

  function startCreate() {
    clearDeepLinkUid()
    setMode("create")
    setSelectedUid("")
    setDetail(null)
    setFailedUid("")
  }

  function applySelect(uid: string) {
    clearDeepLinkUid()
    setMode("edit")
    setSelectedUid(uid)
    setFailedUid("")
  }

  function requestCreate() {
    if (guardDirty()) {
      setPendingSelect({ kind: "create" })
      return
    }
    startCreate()
  }

  function requestSelect(uid: string) {
    if (uid === selectedUid && mode === "edit") {
      return
    }
    if (guardDirty()) {
      setPendingSelect({ kind: "select", uid })
      return
    }
    applySelect(uid)
  }

  function confirmDiscard() {
    const target = pendingSelect
    setPendingSelect(null)
    if (!target) {
      return
    }
    if (target.kind === "create") {
      startCreate()
      return
    }
    applySelect(target.uid)
  }

  const currentDetail =
    detail && detail.item.uid === selectedUid ? detail : null
  const detailLoading =
    mode === "edit" &&
    selectedUid !== "" &&
    currentDetail === null &&
    failedUid !== selectedUid
  const detailFailed = mode === "edit" && failedUid === selectedUid

  const chips: FilterChip[] = []
  if (keyword) {
    chips.push({
      key: "keyword",
      label: keyword,
      onRemove: () => {
        setDraftKeyword("")
        updateQuery({ keyword: "" })
      },
    })
  }
  if (stateFilter) {
    chips.push({
      key: "state",
      label: stateLabel(t, stateFilter),
      onRemove: () => updateQuery({ state: "" }),
    })
  }
  if (localeFilter) {
    chips.push({
      key: "locale",
      label: localeLabel(t, localeFilter),
      onRemove: () => updateQuery({ locale: "" }),
    })
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto lg:overflow-hidden">
      <PageHeader
        title={t("sitePages.title")}
        description={t("sitePages.description")}
      >
        <div className="flex flex-wrap items-center gap-2">
          <Button type="button" variant="outline" size="sm" onClick={reload}>
            <RefreshCw aria-hidden="true" data-icon="inline-start" />
            {t("action.refresh")}
          </Button>
          {canWrite ? (
            <Button type="button" size="sm" onClick={requestCreate}>
              <Plus aria-hidden="true" data-icon="inline-start" />
              {t("sitePages.create")}
            </Button>
          ) : null}
        </div>
      </PageHeader>

      <div className="flex min-h-0 flex-1 flex-col gap-4 lg:grid lg:grid-cols-[minmax(0,21rem)_minmax(0,1fr)]">
        <aside className="flex min-h-0 min-w-0 flex-col gap-3">
          <div className="glass-card shrink-0 rounded-xl p-3">
            <form
              className="flex gap-2"
              onSubmit={(event) => {
                event.preventDefault()
                updateQuery({ keyword: draftKeyword.trim() })
              }}
            >
              <Input
                type="search"
                value={draftKeyword}
                aria-label={t("sitePages.searchPlaceholder")}
                placeholder={t("sitePages.searchPlaceholder")}
                className="min-w-0 flex-1"
                onChange={(event) => setDraftKeyword(event.target.value)}
              />
              <Button
                type="submit"
                variant="outline"
                size="icon"
                aria-label={t("action.search")}
              >
                <Search aria-hidden="true" />
              </Button>
            </form>
            <div className="mt-2 flex items-center justify-between gap-2">
              <FilterToggle
                open={filtersOpen}
                count={chips.length}
                controls="site-pages-filters"
                onClick={() => setFiltersOpen((value) => !value)}
              />
            </div>
            <div
              id="site-pages-filters"
              hidden={!filtersOpen}
              className="mt-2 grid gap-2 sm:grid-cols-2 lg:grid-cols-1"
            >
              <Select
                value={stateFilter || ALL}
                onValueChange={(value) =>
                  updateQuery({ state: value === ALL ? "" : value })
                }
              >
                <SelectTrigger
                  className="w-full"
                  aria-label={t("sitePages.stateFilter")}
                >
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ALL}>
                    {t("sitePages.allStates")}
                  </SelectItem>
                  <SelectItem value="DRAFT">
                    {t("sitePages.state.DRAFT")}
                  </SelectItem>
                  <SelectItem value="PUBLISHED">
                    {t("sitePages.state.PUBLISHED")}
                  </SelectItem>
                  <SelectItem value="ARCHIVED">
                    {t("sitePages.state.ARCHIVED")}
                  </SelectItem>
                </SelectContent>
              </Select>
              <Select
                value={localeFilter || ALL}
                onValueChange={(value) =>
                  updateQuery({ locale: value === ALL ? "" : value })
                }
              >
                <SelectTrigger
                  className="w-full"
                  aria-label={t("sitePages.localeFilter")}
                >
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ALL}>
                    {t("sitePages.allLocales")}
                  </SelectItem>
                  <SelectItem value="zh">
                    {t("sitePages.field.localeZh")}
                  </SelectItem>
                  <SelectItem value="en">
                    {t("sitePages.field.localeEn")}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            {chips.length > 0 ? (
              <FilterChips
                className="mt-2 flex flex-wrap items-center gap-1.5"
                items={chips}
                onClearAll={() => {
                  setDraftKeyword("")
                  updateQuery({ keyword: "", state: "", locale: "" })
                }}
              />
            ) : null}
          </div>

          <ListFrame
            total={total}
            page={page}
            pageSize={pageSize}
            showPager
            onPrev={() => updateQuery({ page: Math.max(1, page - 1) })}
            onNext={() => updateQuery({ page: page + 1 })}
          >
            {loading && items.length === 0 ? (
              <div className="grid gap-2">
                <Skeleton className="h-16 rounded-lg" />
                <Skeleton className="h-16 rounded-lg" />
                <Skeleton className="h-16 rounded-lg" />
              </div>
            ) : failed ? (
              <div className="flex flex-1 flex-col items-center justify-center gap-3 py-8 text-center">
                <CircleAlert
                  aria-hidden="true"
                  className="size-5 text-destructive"
                />
                <p className="text-sm">{t("sitePages.loadFailed")}</p>
                <Button variant="outline" size="sm" onClick={reload}>
                  {t("action.retry")}
                </Button>
              </div>
            ) : items.length === 0 ? (
              <div className="flex flex-1 flex-col items-center justify-center gap-1 py-8 text-center">
                <p className="text-sm font-medium">
                  {chips.length > 0
                    ? t("sitePages.noMatch")
                    : t("sitePages.empty")}
                </p>
                <p className="text-xs text-muted-foreground">
                  {t("sitePages.emptyHint")}
                </p>
              </div>
            ) : (
              <div
                role="list"
                aria-label={t("sitePages.listLabel")}
                className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto"
              >
                {items.map((item) => (
                  <ListItem
                    key={item.uid}
                    item={item}
                    active={mode === "edit" && item.uid === selectedUid}
                    onSelect={() => requestSelect(item.uid)}
                  />
                ))}
              </div>
            )}
          </ListFrame>
        </aside>

        {mode === "create" ? (
          <PageEditor
            ref={editorRef}
            detail={null}
            canWrite={canWrite}
            onSaved={() => undefined}
            onCreated={(created) => {
              setMode("edit")
              setSelectedUid(created.item.uid)
              setDetail(created)
              setFailedUid("")
              reload()
            }}
            onDeleted={() => undefined}
            onReload={() => setReloadToken((value) => value + 1)}
          />
        ) : mode === "edit" && detailLoading ? (
          <section className="flex min-h-0 min-w-0 flex-col rounded-xl border border-border/60 bg-card p-4 lg:h-full lg:overflow-hidden">
            <p className="text-sm text-muted-foreground">
              {t("sitePages.detailLoading")}
            </p>
            <div className="mt-4 grid gap-3" aria-hidden="true">
              <Skeleton className="h-8 w-1/2" />
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-40 w-full" />
            </div>
          </section>
        ) : detailFailed ? (
          <section className="flex min-h-0 min-w-0 flex-col items-center justify-center gap-3 rounded-xl border border-border/60 bg-card p-6 text-center lg:h-full lg:overflow-hidden">
            <CircleAlert
              aria-hidden="true"
              className="size-5 text-destructive"
            />
            <p className="text-sm">{t("sitePages.detailLoadFailed")}</p>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setFailedUid("")
                setReloadToken((value) => value + 1)
              }}
            >
              {t("action.retry")}
            </Button>
          </section>
        ) : currentDetail ? (
          <PageEditor
            key={`${currentDetail.item.uid}:${currentDetail.item.updatedAt}`}
            ref={editorRef}
            detail={currentDetail}
            canWrite={canWrite}
            onSaved={(saved) => {
              setDetail(saved)
              setFailedUid("")
              reload()
            }}
            onCreated={() => undefined}
            onDeleted={() => {
              setMode("none")
              setSelectedUid("")
              setDetail(null)
              setFailedUid("")
              reload()
            }}
            onReload={() => {
              setDetail(null)
              setFailedUid("")
              setReloadToken((value) => value + 1)
            }}
          />
        ) : (
          <section className="flex min-h-0 min-w-0 flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border/70 bg-card/60 p-6 text-center lg:h-full lg:overflow-hidden">
            <p className="text-sm font-medium">{t("sitePages.selectHint")}</p>
          </section>
        )}
      </div>

      <ConfirmAction
        open={pendingSelect !== null}
        title={t("sitePages.dirtyTitle")}
        description={t("sitePages.dirtyBody")}
        confirmLabel={t("sitePages.dirtyDiscard")}
        cancelLabel={t("sitePages.dirtyStay")}
        onConfirm={confirmDiscard}
        onOpenChange={(open) => {
          if (!open) {
            setPendingSelect(null)
          }
        }}
      />
    </div>
  )
}
