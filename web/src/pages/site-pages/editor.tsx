import {
  ChevronDown,
  ExternalLink,
  Loader2,
  RefreshCw,
} from "lucide-react"
import {
  useEffect,
  useImperativeHandle,
  useState,
  type Ref,
} from "react"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { ApiError } from "@/api/http"
import {
  createSitePage,
  deleteSitePage,
  getSitePage,
  publishSitePage,
  unpublishSitePage,
  updateSitePage,
  updateSitePageNav,
} from "@/api/site/pages"
import type {
  CreateSitePageParams,
  SitePageDetail,
} from "@/api/site/types"
import { ConfirmAction } from "@/components/confirm-action"
import { FormField } from "@/components/form-field"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/class-names"
import { formatDateTime } from "@/lib/date-time"
import {
  collectFieldErrors,
  hasFieldErrors,
  maxLength,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { readFormError } from "@/lib/form-error"

const TITLE_MAX = 128
const SLUG_MAX = 64
const SUMMARY_MAX = 300
const BODY_MAX = 100000
const SEO_TITLE_MAX = 128
const SEO_DESCRIPTION_MAX = 320
const COVER_MAX = 512
const SLUG_PATTERN = /^[a-z0-9]+(-[a-z0-9]+)*$/

export interface PageEditorHandle {
  isDirty: () => boolean
}

interface PageForm {
  title: string
  slug: string
  locale: string
  summary: string
  bodyMd: string
  coverImageUrl: string
  seoTitle: string
  seoDescription: string
  showInNav: boolean
  sortOrder: string
}

const emptyForm: PageForm = {
  title: "",
  slug: "",
  locale: "zh",
  summary: "",
  bodyMd: "",
  coverImageUrl: "",
  seoTitle: "",
  seoDescription: "",
  showInNav: false,
  sortOrder: "0",
}

function formFromDetail(detail: SitePageDetail): PageForm {
  return {
    title: detail.draft.title,
    slug: detail.item.slug,
    locale: detail.item.locale,
    summary: detail.draft.summary,
    bodyMd: detail.draft.bodyMd,
    coverImageUrl: detail.draft.coverImageUrl,
    seoTitle: detail.draft.seoTitle,
    seoDescription: detail.draft.seoDescription,
    showInNav: detail.item.showInNav,
    sortOrder: String(detail.item.sortOrder ?? 0),
  }
}

function isHTTPURL(value: string): boolean {
  try {
    const url = new URL(value)
    return (
      (url.protocol === "http:" || url.protocol === "https:") && url.host !== ""
    )
  } catch {
    return false
  }
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

function draftChanged(form: PageForm, detail: SitePageDetail): boolean {
  return (
    form.title !== detail.draft.title ||
    form.summary !== detail.draft.summary ||
    form.bodyMd !== detail.draft.bodyMd ||
    form.coverImageUrl !== detail.draft.coverImageUrl ||
    form.seoTitle !== detail.draft.seoTitle ||
    form.seoDescription !== detail.draft.seoDescription
  )
}

function navChanged(form: PageForm, detail: SitePageDetail): boolean {
  return (
    form.showInNav !== detail.item.showInNav ||
    form.sortOrder.trim() !== String(detail.item.sortOrder ?? 0)
  )
}

interface PageEditorProps {
  ref?: Ref<PageEditorHandle>
  /** null 表示新建。 */
  detail: SitePageDetail | null
  canWrite: boolean
  onSaved: (detail: SitePageDetail) => void
  onCreated: (detail: SitePageDetail) => void
  onDeleted: () => void
  onReload: () => void
}

/**
 * 草稿乐观锁冲突：先拉一次最新详情，拿到最新 updatedAt 后再让用户选择
 * 「重新加载」（放弃本地修改）或「保留我的修改并覆盖保存」。
 */
type ConflictState =
  | { phase: "loading" }
  | { phase: "ready"; latest: SitePageDetail }
  | { phase: "failed" }

/** 官网页面编辑器：草稿保存、导航、发布/下架/删除、预览与 409 冲突处理。 */
export function PageEditor({
  ref,
  detail,
  canWrite,
  onSaved,
  onCreated,
  onDeleted,
  onReload,
}: PageEditorProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const creating = detail === null
  const [form, setForm] = useState<PageForm>(() =>
    detail ? formFromDetail(detail) : emptyForm
  )
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState<
    "" | "save" | "publish" | "unpublish" | "delete"
  >("")
  const [confirm, setConfirm] = useState<
    "" | "publish" | "unpublish" | "delete"
  >("")
  const [confirmError, setConfirmError] = useState("")
  const [conflict, setConflict] = useState<ConflictState | null>(null)
  const [conflictPending, setConflictPending] = useState<"" | "overwrite">("")
  const [conflictError, setConflictError] = useState("")
  const [seoOpen, setSeoOpen] = useState(false)
  // 站内链接离开时的拦截目标；浏览器刷新/关闭由 beforeunload 处理。
  const [leaveTo, setLeaveTo] = useState<string | null>(null)

  const isHome = detail?.item.slug === "home"
  const dirty = creating
    ? Object.entries(form).some(([key, value]) => {
        if (key === "sortOrder") {
          return value.trim() !== "0"
        }
        if (key === "showInNav") {
          return value === true
        }
        return typeof value === "string" && value.trim() !== ""
      })
    : detail
      ? draftChanged(form, detail) || navChanged(form, detail)
      : false

  useImperativeHandle(ref, () => ({ isDirty: () => dirty }), [dirty])

  // 浏览器刷新/关闭与站内切换由父组件拦截；这里只处理浏览器原生离开。
  useEffect(() => {
    if (!dirty) {
      return
    }
    function onBeforeUnload(event: BeforeUnloadEvent) {
      event.preventDefault()
      event.returnValue = ""
    }
    window.addEventListener("beforeunload", onBeforeUnload)
    return () => window.removeEventListener("beforeunload", onBeforeUnload)
  }, [dirty])

  // AppShell 的侧栏与页内链接都是普通 <a>：捕获阶段拦下站内跳转，先问一次是否放弃修改。
  useEffect(() => {
    if (!dirty) {
      return
    }
    function onClick(event: MouseEvent) {
      if (
        event.defaultPrevented ||
        event.button !== 0 ||
        event.metaKey ||
        event.ctrlKey ||
        event.shiftKey ||
        event.altKey
      ) {
        return
      }
      const target = event.target
      if (!(target instanceof Element)) {
        return
      }
      const anchor = target.closest("a[href]")
      if (!(anchor instanceof HTMLAnchorElement)) {
        return
      }
      if (anchor.target && anchor.target !== "_self") {
        return
      }
      const href = anchor.getAttribute("href") ?? ""
      if (!href || href.startsWith("#")) {
        return
      }
      let url: URL
      try {
        url = new URL(href, window.location.origin)
      } catch {
        return
      }
      if (url.origin !== window.location.origin) {
        return
      }
      if (
        url.pathname === window.location.pathname &&
        url.search === window.location.search
      ) {
        return
      }
      event.preventDefault()
      event.stopPropagation()
      setLeaveTo(`${url.pathname}${url.search}${url.hash}`)
    }
    document.addEventListener("click", onClick, true)
    return () => document.removeEventListener("click", onClick, true)
  }, [dirty])

  function update<K extends keyof PageForm>(key: K, value: PageForm[K]) {
    setForm((current) => ({ ...current, [key]: value }))
  }

  function validate(): FieldErrors {
    const sortOrder = Number(form.sortOrder)
    return collectFieldErrors([
      ["title", requireTrimmed(form.title, t("form.required"))],
      ["title", maxLength(form.title, TITLE_MAX, t("sitePages.field.tooLong"))],
      ["slug", creating ? requireTrimmed(form.slug, t("form.required")) : null],
      [
        "slug",
        creating &&
        form.slug.trim() !== "" &&
        !SLUG_PATTERN.test(form.slug.trim())
          ? t("error.sitePageSlugInvalid")
          : null,
      ],
      [
        "slug",
        creating && form.slug.trim().length > SLUG_MAX
          ? t("sitePages.field.tooLong")
          : null,
      ],
      ["summary", maxLength(form.summary, SUMMARY_MAX, t("sitePages.field.tooLong"))],
      [
        "bodyMd",
        Array.from(form.bodyMd).length > BODY_MAX ? t("sitePages.field.tooLong") : null,
      ],
      [
        "coverImageUrl",
        form.coverImageUrl.trim() !== "" &&
        !isHTTPURL(form.coverImageUrl.trim())
          ? t("error.sitePageCoverInvalid")
          : null,
      ],
      [
        "coverImageUrl",
        form.coverImageUrl.trim().length > COVER_MAX
          ? t("sitePages.field.tooLong")
          : null,
      ],
      ["seoTitle", maxLength(form.seoTitle, SEO_TITLE_MAX, t("sitePages.field.tooLong"))],
      [
        "seoDescription",
        maxLength(
          form.seoDescription,
          SEO_DESCRIPTION_MAX,
          t("sitePages.field.tooLong")
        ),
      ],
      [
        "sortOrder",
        !/^\d+$/.test(form.sortOrder.trim()) ||
        sortOrder < 0 ||
        sortOrder > 999
          ? t("error.sitePageSortOrderInvalid")
          : null,
      ],
    ])
  }

  function isModifiedConflict(err: unknown): boolean {
    // 只有乐观锁冲突走「重新加载 / 覆盖保存」流程；slug/排序冲突等 409 仍按普通表单错误展示。
    return (
      err instanceof ApiError &&
      err.status === 409 &&
      err.raw === "site page was modified"
    )
  }

  function errorMessage(err: unknown, fallback: string): string {
    return err instanceof Error && err.message ? err.message : fallback
  }

  /** 收到乐观锁冲突后先拉最新详情，拿到最新 updatedAt 再打开确认框。 */
  async function openConflict() {
    if (!detail) {
      return
    }
    setConflict({ phase: "loading" })
    setConflictError("")
    try {
      const result = await getSitePage(detail.item.uid)
      setConflict({ phase: "ready", latest: result.page })
    } catch (err) {
      setConflict({ phase: "failed" })
      setConflictError(errorMessage(err, t("sitePages.conflict.loadFailed")))
    }
  }

  function handleApiError(err: unknown, fallback: string) {
    setFormError("")
    if (isModifiedConflict(err)) {
      // 乐观锁冲突：本地修改全部保留在表单里，由确认框决定重载还是覆盖。
      void openConflict()
      return
    }
    if (
      err instanceof ApiError &&
      (err.status === 404 || err.status >= 500)
    ) {
      // readFormError 把 404/5xx 留给页面级错误页；编辑器里没有错误页，直接展示原因。
      setFieldErrors({})
      setFormError(err.message || fallback)
      return
    }
    const failure = readFormError(err, fallback)
    setFieldErrors(failure.fields)
    setFormError(failure.form || Object.values(failure.fields)[0] || "")
  }

  function contentPayload() {
    return {
      title: form.title.trim(),
      summary: form.summary.trim(),
      bodyMd: form.bodyMd,
      coverImageUrl: form.coverImageUrl.trim(),
      seoTitle: form.seoTitle.trim(),
      seoDescription: form.seoDescription.trim(),
    }
  }

  /** 保存草稿与导航设置；返回服务端最新详情。 */
  async function persistDraft(): Promise<SitePageDetail | null> {
    if (!detail) {
      return null
    }
    let current = detail
    if (navChanged(form, detail)) {
      const result = await updateSitePageNav({
        uid: current.item.uid,
        showInNav: form.showInNav,
        sortOrder: Number(form.sortOrder.trim() || 0),
      })
      current = result.page
    }
    if (draftChanged(form, detail) || navChanged(form, detail)) {
      const payload = contentPayload()
      const result = await updateSitePage({
        uid: current.item.uid,
        ...payload,
        updatedAt: current.item.updatedAt,
      })
      current = result.page
    }
    return current
  }

  /** 打开完整官网预览：新标签读取草稿快照，未保存的改动不会出现，所以先要求保存。 */
  function openPreview() {
    if (!detail || creating || dirty) {
      return
    }
    window.open(
      `/site-preview/${encodeURIComponent(detail.item.uid)}`,
      "_blank",
      "noopener,noreferrer"
    )
  }

  /** 冲突处理 A：放弃本地修改，重新从服务端加载最新草稿。 */
  function reloadConflict() {
    if (conflict?.phase !== "ready" || conflictPending !== "") {
      return
    }
    setConflict(null)
    setConflictError("")
    toast.success(t("sitePages.toast.reloaded"))
    onReload()
  }

  /** 冲突处理 B：用刚取到的最新 updatedAt 提交本地修改。 */
  async function overwriteConflict() {
    if (!detail || conflict?.phase !== "ready" || conflictPending !== "") {
      return
    }
    const latestUpdatedAt = conflict.latest.item.updatedAt
    setConflictPending("overwrite")
    setConflictError("")
    try {
      const result = await updateSitePage({
        uid: detail.item.uid,
        ...contentPayload(),
        updatedAt: latestUpdatedAt,
      })
      setConflict(null)
      toast.success(t("sitePages.toast.overwritten"))
      onSaved(result.page)
    } catch (err) {
      if (isModifiedConflict(err)) {
        // 读取之后又被保存过：刷新 updatedAt，让用户基于最新版本再确认一次。
        try {
          const latest = await getSitePage(detail.item.uid)
          setConflict({ phase: "ready", latest: latest.page })
          setConflictError(t("sitePages.conflict.overwriteConflict"))
        } catch (reloadErr) {
          setConflict({ phase: "failed" })
          setConflictError(
            errorMessage(reloadErr, t("sitePages.conflict.loadFailed"))
          )
        }
        return
      }
      // 覆盖失败不静默：保留对话框并展示服务端原因，用户可以重试或取消。
      setConflictError(errorMessage(err, t("state.failed")))
    } finally {
      setConflictPending("")
    }
  }

  async function onCreate() {
    if (!canWrite || pending !== "") {
      return
    }
    const errors = validate()
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setFormError("")
      return
    }
    const payload: CreateSitePageParams = {
      slug: form.slug.trim(),
      locale: form.locale,
      ...contentPayload(),
    }
    setPending("save")
    setFieldErrors({})
    setFormError("")
    try {
      const result = await createSitePage(payload)
      toast.success(t("sitePages.toast.created"))
      onCreated(result.page)
    } catch (err) {
      handleApiError(err, t("state.failed"))
    } finally {
      setPending("")
    }
  }

  async function onSave() {
    if (!canWrite || !detail || pending !== "") {
      return
    }
    const errors = validate()
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setFormError("")
      return
    }
    setPending("save")
    setFieldErrors({})
    setFormError("")
    try {
      const next = await persistDraft()
      if (!next) {
        return
      }
      toast.success(
        navChanged(form, detail)
          ? t("sitePages.toast.navSaved")
          : t("sitePages.toast.saved")
      )
      onSaved(next)
    } catch (err) {
      handleApiError(err, t("state.failed"))
    } finally {
      setPending("")
    }
  }

  async function runConfirm() {
    if (!detail || pending !== "" || confirm === "") {
      return
    }
    setPending(confirm)
    setConfirmError("")
    try {
      if (confirm === "publish") {
        // 发布前先把编辑器里未保存的修改落成草稿，避免“看起来发布了旧内容”。
        if (dirty) {
          const saved = await persistDraft()
          if (!saved) {
            return
          }
        }
        const result = await publishSitePage(detail.item.uid)
        toast.success(t("sitePages.toast.published"))
        onSaved(result.page)
      } else if (confirm === "unpublish") {
        const result = await unpublishSitePage(detail.item.uid)
        toast.success(t("sitePages.toast.unpublished"))
        onSaved(result.page)
      } else if (confirm === "delete") {
        await deleteSitePage(detail.item.uid)
        toast.success(t("sitePages.toast.deleted"))
        onDeleted()
      }
      setConfirm("")
    } catch (err) {
      if (isModifiedConflict(err)) {
        setConfirm("")
        void openConflict()
        return
      }
      if (
        err instanceof ApiError &&
        (err.status === 404 || err.status >= 500)
      ) {
        setConfirmError(err.message || t("state.failed"))
        return
      }
      const failure = readFormError(err, t("state.failed"))
      setConfirmError(
        failure.form || Object.values(failure.fields)[0] || t("state.failed")
      )
    } finally {
      setPending("")
    }
  }

  const publishedAt = detail?.item.publishedAt ?? ""
  const headerTitle = creating
    ? t("sitePages.createTitle")
    : form.title.trim() || detail.item.slug || t("sitePages.editTitle")

  return (
    <section className="flex min-h-0 min-w-0 flex-col rounded-xl border border-border/60 bg-card lg:h-full lg:overflow-hidden">
      <div className="shrink-0 border-b border-border/60 p-4">
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="font-heading text-lg font-semibold tracking-tight break-words">
            {headerTitle}
          </h2>
          {detail ? (
            <>
              <StatusBadge
                value={detail.item.state}
                label={stateLabel(t, detail.item.state)}
                tone={stateTone(detail.item.state)}
              />
              {isHome ? (
                <StatusBadge
                  value="HOME"
                  label={t("sitePages.homeBadge")}
                  tone="primary"
                />
              ) : null}
              {detail.item.hasUnpublishedChanges ? (
                <StatusBadge
                  value="UNPUBLISHED_CHANGES"
                  label={t("sitePages.unpublishedChanges")}
                  tone="warning"
                />
              ) : null}
            </>
          ) : null}
        </div>
        <p className="mt-1 text-xs break-words text-muted-foreground">
          {detail
            ? [
                `/p/${detail.item.slug}`,
                localeLabel(t, detail.item.locale),
                detail.item.updatedAt
                  ? t("sitePages.updatedAt", {
                      time: formatDateTime(detail.item.updatedAt),
                    })
                  : "",
                publishedAt
                  ? t("sitePages.publishedAt", {
                      time: formatDateTime(publishedAt),
                    })
                  : t("sitePages.neverPublished"),
              ]
                .filter(Boolean)
                .join(" · ")
            : t("sitePages.field.slugHint")}
        </p>
      </div>

      <div className="min-h-0 p-4 lg:flex-1 lg:overflow-y-auto">
        {!canWrite ? (
          <p className="mb-4 text-sm text-muted-foreground">
            {t("sitePages.readOnly")}
          </p>
        ) : null}
        {isHome ? (
          <p className="mb-4 text-sm text-muted-foreground">
            {t("sitePages.homeLocked")}
          </p>
        ) : null}

        <div className="grid gap-4">
          <FormField
            label={t("sitePages.field.title")}
            htmlFor="site-page-title"
            required
            error={fieldErrors.title}
          >
            <Input
              id="site-page-title"
              value={form.title}
              maxLength={TITLE_MAX}
              readOnly={!canWrite}
              placeholder={t("sitePages.field.titlePlaceholder")}
              aria-invalid={fieldErrors.title ? true : undefined}
              onChange={(event) => update("title", event.target.value)}
            />
          </FormField>

          <div className="grid gap-4 sm:grid-cols-2">
            <FormField
              label={t("sitePages.field.slug")}
              htmlFor="site-page-slug"
              required={creating}
              hint={
                creating
                  ? t("sitePages.field.slugHint")
                  : t("sitePages.field.slugLockedHint")
              }
              error={fieldErrors.slug}
            >
              <Input
                id="site-page-slug"
                value={form.slug}
                maxLength={SLUG_MAX}
                readOnly={!creating || !canWrite}
                disabled={!creating}
                placeholder={t("sitePages.field.slugPlaceholder")}
                aria-invalid={fieldErrors.slug ? true : undefined}
                onChange={(event) => update("slug", event.target.value)}
              />
            </FormField>
            <FormField
              label={t("sitePages.field.locale")}
              htmlFor="site-page-locale"
              hint={t("sitePages.field.localeHint")}
              error={fieldErrors.locale}
            >
              {creating ? (
                <Select
                  value={form.locale}
                  disabled={!canWrite}
                  onValueChange={(value) => update("locale", value)}
                >
                  <SelectTrigger id="site-page-locale" className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="zh">
                      {t("sitePages.field.localeZh")}
                    </SelectItem>
                    <SelectItem value="en">
                      {t("sitePages.field.localeEn")}
                    </SelectItem>
                  </SelectContent>
                </Select>
              ) : (
                <Input
                  id="site-page-locale"
                  value={localeLabel(t, form.locale)}
                  readOnly
                  disabled
                />
              )}
            </FormField>
          </div>

          <FormField
            label={t("sitePages.field.summary")}
            htmlFor="site-page-summary"
            error={fieldErrors.summary}
          >
            <Textarea
              id="site-page-summary"
              value={form.summary}
              maxLength={SUMMARY_MAX}
              readOnly={!canWrite}
              placeholder={t("sitePages.field.summaryPlaceholder")}
              className="min-h-16"
              aria-invalid={fieldErrors.summary ? true : undefined}
              onChange={(event) => update("summary", event.target.value)}
            />
          </FormField>

          <FormField
            label={t("sitePages.field.body")}
            htmlFor="site-page-body"
            hint={t("sitePages.field.bodyPlaceholder")}
            error={fieldErrors.bodyMd}
          >
            <div className="relative">
              <Textarea
                id="site-page-body"
                value={form.bodyMd}
                readOnly={!canWrite}
                placeholder={t("sitePages.field.bodyPlaceholder")}
                className="min-h-64 resize-y pr-3 pb-8 font-mono text-sm leading-6"
                aria-invalid={fieldErrors.bodyMd ? true : undefined}
                onChange={(event) => update("bodyMd", event.target.value)}
              />
              <span className="pointer-events-none absolute right-2 bottom-2 rounded bg-background/90 px-1.5 py-0.5 font-mono text-xs text-muted-foreground tabular-nums">
                {t("sitePages.field.bodyCount", {
                  count: Array.from(form.bodyMd).length,
                })}
              </span>
            </div>
          </FormField>

          <FormField
            label={t("sitePages.field.cover")}
            htmlFor="site-page-cover"
            hint={t("sitePages.field.coverHint")}
            error={fieldErrors.coverImageUrl}
          >
            <Input
              id="site-page-cover"
              value={form.coverImageUrl}
              maxLength={COVER_MAX}
              readOnly={!canWrite}
              placeholder={t("sitePages.field.coverPlaceholder")}
              aria-invalid={fieldErrors.coverImageUrl ? true : undefined}
              onChange={(event) =>
                update("coverImageUrl", event.target.value)
              }
            />
          </FormField>

          <Collapsible open={seoOpen} onOpenChange={setSeoOpen}>
            <CollapsibleTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="w-fit"
                aria-expanded={seoOpen}
              >
                {t("sitePages.field.seo")}
                <ChevronDown
                  aria-hidden="true"
                  data-icon="inline-end"
                  className={cn(
                    "transition-transform",
                    seoOpen ? "rotate-180" : undefined
                  )}
                />
              </Button>
            </CollapsibleTrigger>
            <CollapsibleContent className="grid gap-4 pt-2">
              <FormField
                label={t("sitePages.field.seoTitle")}
                htmlFor="site-page-seo-title"
                error={fieldErrors.seoTitle}
              >
                <Input
                  id="site-page-seo-title"
                  value={form.seoTitle}
                  maxLength={SEO_TITLE_MAX}
                  readOnly={!canWrite}
                  placeholder={t("sitePages.field.seoTitlePlaceholder")}
                  aria-invalid={fieldErrors.seoTitle ? true : undefined}
                  onChange={(event) => update("seoTitle", event.target.value)}
                />
              </FormField>
              <FormField
                label={t("sitePages.field.seoDescription")}
                htmlFor="site-page-seo-description"
                error={fieldErrors.seoDescription}
              >
                <Textarea
                  id="site-page-seo-description"
                  value={form.seoDescription}
                  maxLength={SEO_DESCRIPTION_MAX}
                  readOnly={!canWrite}
                  placeholder={t(
                    "sitePages.field.seoDescriptionPlaceholder"
                  )}
                  className="min-h-16"
                  aria-invalid={
                    fieldErrors.seoDescription ? true : undefined
                  }
                  onChange={(event) =>
                    update("seoDescription", event.target.value)
                  }
                />
              </FormField>
            </CollapsibleContent>
          </Collapsible>

          {detail ? (
            <div className="grid gap-3 rounded-lg border border-border/60 p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm font-medium">
                  {t("sitePages.field.navTitle")}
                </p>
                <span className="text-xs text-muted-foreground">
                  {t("sitePages.field.sortOrderHint")}
                </span>
              </div>
              <label className="flex items-start gap-2.5 text-sm">
                <Checkbox
                  checked={form.showInNav}
                  disabled={!canWrite}
                  onCheckedChange={(value) =>
                    update("showInNav", value === true)
                  }
                />
                <span>{t("sitePages.field.showInNav")}</span>
              </label>
              <FormField
                label={t("sitePages.field.sortOrder")}
                htmlFor="site-page-sort-order"
                error={fieldErrors.sortOrder}
                className="max-w-40"
              >
                <Input
                  id="site-page-sort-order"
                  type="number"
                  min={0}
                  max={999}
                  value={form.sortOrder}
                  readOnly={!canWrite}
                  aria-invalid={fieldErrors.sortOrder ? true : undefined}
                  onChange={(event) =>
                    update("sortOrder", event.target.value)
                  }
                />
              </FormField>
            </div>
          ) : null}

          {formError ? (
            <p role="alert" className="text-sm text-destructive">
              {formError}
            </p>
          ) : null}
        </div>
      </div>

      <div className="flex shrink-0 flex-wrap items-center gap-2 border-t border-border/60 p-3">
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={!detail || creating || dirty}
          title={
            !detail || creating || dirty
              ? t("sitePages.preview.saveHint")
              : undefined
          }
          onClick={openPreview}
        >
          <ExternalLink aria-hidden="true" data-icon="inline-start" />
          {t("sitePages.action.preview")}
        </Button>
        {canWrite ? (
          <Button
            type="button"
            size="sm"
            disabled={pending !== "" || (!creating && !dirty)}
            onClick={() => void (creating ? onCreate() : onSave())}
          >
            {pending === "save" ? (
              <Loader2 className="animate-spin" aria-hidden="true" />
            ) : null}
            {creating
              ? pending === "save"
                ? t("sitePages.action.creating")
                : t("sitePages.action.create")
              : pending === "save"
                ? t("sitePages.action.saving")
                : t("sitePages.action.saveDraft")}
          </Button>
        ) : null}
        {canWrite && detail ? (
          <>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={pending !== ""}
              onClick={() => {
                setConfirmError("")
                setConfirm("publish")
              }}
            >
              {t("sitePages.action.publish")}
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={
                pending !== "" ||
                isHome ||
                detail.item.state !== "PUBLISHED"
              }
              title={isHome ? t("sitePages.homeLocked") : undefined}
              onClick={() => {
                setConfirmError("")
                setConfirm("unpublish")
              }}
            >
              {t("sitePages.action.unpublish")}
            </Button>
            <Button
              type="button"
              variant="destructive"
              size="sm"
              className="ml-auto"
              disabled={pending !== "" || isHome}
              title={isHome ? t("sitePages.homeLocked") : undefined}
              onClick={() => {
                setConfirmError("")
                setConfirm("delete")
              }}
            >
              {t("sitePages.action.delete")}
            </Button>
          </>
        ) : null}
      </div>

      <Dialog
        open={conflict !== null}
        onOpenChange={(open) => {
          if (!open && conflictPending === "") {
            setConflict(null)
            setConflictError("")
          }
        }}
      >
        <DialogContent
          className="sm:max-w-xl"
          showCloseButton={conflictPending === ""}
        >
          <DialogHeader>
            <DialogTitle>{t("sitePages.conflict.title")}</DialogTitle>
            <DialogDescription>
              {t("sitePages.conflict.body")}
            </DialogDescription>
          </DialogHeader>
          <DialogBody className="grid content-start gap-3">
            {conflict?.phase === "loading" ? (
              <p className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                {t("sitePages.detailLoading")}
              </p>
            ) : null}
            {conflict?.phase === "failed" ? (
              <div className="grid content-start gap-3">
                <p className="text-sm text-muted-foreground">
                  {t("sitePages.conflict.loadFailed")}
                </p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="w-fit"
                  onClick={() => void openConflict()}
                >
                  <RefreshCw aria-hidden="true" data-icon="inline-start" />
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {conflict?.phase === "ready" ? (
              <p className="text-sm text-muted-foreground">
                {t("sitePages.conflict.latestAt", {
                  time: formatDateTime(conflict.latest.item.updatedAt),
                })}
              </p>
            ) : null}
          </DialogBody>
          {conflictError ? (
            <p role="alert" className="text-sm text-destructive">
              {conflictError}
            </p>
          ) : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={conflict?.phase !== "ready" || conflictPending !== ""}
              onClick={reloadConflict}
            >
              {t("sitePages.conflict.reload")}
            </Button>
            <Button
              type="button"
              disabled={conflict?.phase !== "ready" || conflictPending !== ""}
              onClick={() => void overwriteConflict()}
            >
              {conflictPending === "overwrite" ? (
                <Loader2 className="animate-spin" aria-hidden="true" />
              ) : null}
              {conflictPending === "overwrite"
                ? t("sitePages.conflict.overwriting")
                : t("sitePages.conflict.overwrite")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmAction
        open={leaveTo !== null}
        title={t("sitePages.dirtyTitle")}
        description={t("sitePages.dirtyBody")}
        confirmLabel={t("sitePages.dirtyDiscard")}
        cancelLabel={t("sitePages.dirtyStay")}
        onConfirm={() => {
          const target = leaveTo
          setLeaveTo(null)
          if (target) {
            navigate(target)
          }
        }}
        onOpenChange={(open) => {
          if (!open) {
            setLeaveTo(null)
          }
        }}
      />

      <ConfirmAction
        open={confirm !== ""}
        title={
          confirm === "publish"
            ? t("sitePages.confirm.publishTitle")
            : confirm === "unpublish"
              ? t("sitePages.confirm.unpublishTitle")
              : t("sitePages.confirm.deleteTitle")
        }
        description={
          confirm === "publish"
            ? t("sitePages.confirm.publishBody")
            : confirm === "unpublish"
              ? t("sitePages.confirm.unpublishBody")
              : t("sitePages.confirm.deleteBody")
        }
        confirmLabel={
          confirm === "publish"
            ? t("sitePages.confirm.publishAction")
            : confirm === "unpublish"
              ? t("sitePages.confirm.unpublishAction")
              : t("sitePages.confirm.deleteAction")
        }
        pendingLabel={
          pending === "publish"
            ? t("sitePages.action.publishing")
            : pending === "unpublish"
              ? t("sitePages.action.unpublishing")
              : t("sitePages.action.deleting")
        }
        cancelLabel={t("action.cancel")}
        pending={
          pending === "publish" ||
          pending === "unpublish" ||
          pending === "delete"
        }
        error={confirmError}
        onConfirm={() => void runConfirm()}
        onOpenChange={(open) => {
          if (!open && pending === "") {
            setConfirm("")
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
