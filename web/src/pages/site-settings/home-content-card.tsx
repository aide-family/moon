import { ChevronRight, CircleAlert, Eye, Loader2, Save } from "lucide-react"
import { useEffect, useMemo, useState, type FormEvent } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { ApiError } from "@/api/http"
import { getSiteHome, updateSiteHome } from "@/api/site"
import type {
  SaveSiteHomeParams,
  SiteConfig,
  SiteHomeContent,
} from "@/api/site/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
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
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { RenderGuard } from "@/components/render-guard"
import { useSiteConfig } from "@/hooks/use-site-config"
import { failureMessage } from "@/lib/form-error"
import { SiteHomeSections } from "@/pages/public-site/home/sections"

/** 首页文案字段。版式固定，这里只有短文本；长度上限与后端校验保持一致。 */
type HomeField = keyof Omit<
  SiteHomeContent,
  "locale" | "updatedBy" | "updatedAt"
>

type HomeForm = Record<HomeField, string>

interface FeatureForm {
  iconKey: string
  title: string
  description: string
}

type FieldErrors = Partial<Record<HomeField, string>> & {
  features?: Record<number, { title?: string; description?: string }>
}

/** 能力卡片图标白名单，与后端 siteFeatureIcons 一一对应。 */
const FEATURE_ICONS = [
  "activity",
  "bell",
  "shield",
  "database",
  "globe",
  "boxes",
  "terminal",
  "gauge",
  "server",
  "users",
  "git-branch",
  "layers",
] as const

const MAX_FEATURES = 6

/** 每个字段的字数上限，用于输入框 maxLength、本地校验与字数提示。 */
const FIELD_LIMITS: Partial<Record<HomeField, number>> = {
  heroTitle: 16,
  heroSubtitle: 72,
  heroImageUrl: 512,
  ctaPrimaryLabel: 12,
  ctaSecondaryLabel: 12,
  ctaSecondaryHref: 128,
  featuresTitle: 16,
  featuresSubtitle: 48,
  howTitle: 16,
  howBody: 80,
  permTitle: 16,
  permBody: 80,
  deployTitle: 16,
  deployBody: 80,
  docTitle: 16,
  closingTitle: 20,
  closingDescription: 48,
  closingButtonLabel: 12,
}

/** 与后端一致：这些词属于品牌红线，保存前就拦下来。 */
const BANNED_WORDS = [
  "赋能",
  "闭环",
  "一站式",
  "极致",
  "颠覆",
  "军工级",
  "金融级",
  "毫秒级",
  "海量",
  "深度打通",
  "重新定义",
]

const emptyHomeForm: HomeForm = {
  heroTitle: "",
  heroSubtitle: "",
  heroImageUrl: "",
  ctaPrimaryLabel: "",
  ctaSecondaryLabel: "",
  ctaSecondaryHref: "",
  featuresTitle: "",
  featuresSubtitle: "",
  howTitle: "",
  howBody: "",
  permTitle: "",
  permBody: "",
  deployTitle: "",
  deployBody: "",
  docTitle: "",
  closingTitle: "",
  closingDescription: "",
  closingButtonLabel: "",
}

/** 文案分组：与官网首页的区块顺序一致，折叠时也能看清结构。 */
const GROUPS: Array<{
  key: string
  labelKey: string
  fields: HomeField[]
  textareas?: HomeField[]
}> = [
  {
    key: "hero",
    labelKey: "systemSite.home.sections.hero",
    fields: [
      "heroTitle",
      "heroSubtitle",
      "heroImageUrl",
      "ctaPrimaryLabel",
      "ctaSecondaryLabel",
      "ctaSecondaryHref",
    ],
  },
  {
    key: "features",
    labelKey: "systemSite.home.sections.features",
    fields: ["featuresTitle", "featuresSubtitle"],
  },
  {
    key: "explainers",
    labelKey: "systemSite.home.sections.explainers",
    fields: [
      "howTitle",
      "howBody",
      "permTitle",
      "permBody",
      "deployTitle",
      "deployBody",
    ],
    textareas: ["howBody", "permBody", "deployBody"],
  },
  {
    key: "closing",
    labelKey: "systemSite.home.sections.closing",
    fields: [
      "docTitle",
      "closingTitle",
      "closingDescription",
      "closingButtonLabel",
    ],
  },
]

function toHomeForm(content: SiteHomeContent): HomeForm {
  return {
    heroTitle: content.heroTitle,
    heroSubtitle: content.heroSubtitle,
    heroImageUrl: content.heroImageUrl,
    ctaPrimaryLabel: content.ctaPrimaryLabel,
    ctaSecondaryLabel: content.ctaSecondaryLabel,
    ctaSecondaryHref: content.ctaSecondaryHref,
    featuresTitle: content.featuresTitle,
    featuresSubtitle: content.featuresSubtitle,
    howTitle: content.howTitle,
    howBody: content.howBody,
    permTitle: content.permTitle,
    permBody: content.permBody,
    deployTitle: content.deployTitle,
    deployBody: content.deployBody,
    docTitle: content.docTitle,
    closingTitle: content.closingTitle,
    closingDescription: content.closingDescription,
    closingButtonLabel: content.closingButtonLabel,
  }
}

function mapFeatures(
  items: Array<{ iconKey: string; title: string; description: string }>
): FeatureForm[] {
  return items.map((item) => ({
    iconKey: item.iconKey,
    title: item.title,
    description: item.description,
  }))
}

function HomeSkeleton({ label }: { label: string }) {
  return (
    <div className="grid gap-3" aria-busy="true" aria-label={label}>
      <Skeleton className="h-10 rounded-lg" />
      <div className="grid gap-3 sm:grid-cols-2">
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
      </div>
      <Skeleton className="h-40 rounded-lg" />
    </div>
  )
}

export function HomeContentCard({
  canRead,
  canWrite,
  refreshKey,
}: {
  canRead: boolean
  canWrite: boolean
  refreshKey: number
}) {
  const { t } = useTranslation()
  const { config } = useSiteConfig()
  const [locale, setLocale] = useState<"zh" | "en">("zh")
  const [form, setForm] = useState<HomeForm>(emptyHomeForm)
  const [features, setFeatures] = useState<FeatureForm[]>([])
  const [savedForm, setSavedForm] = useState<HomeForm>(emptyHomeForm)
  const [savedFeatures, setSavedFeatures] = useState<FeatureForm[]>([])
  const [savedAt, setSavedAt] = useState("")
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [previewOpen, setPreviewOpen] = useState(false)
  const [pendingLocale, setPendingLocale] = useState<"zh" | "en" | null>(null)

  const dirty = useMemo(
    () =>
      JSON.stringify(form) !== JSON.stringify(savedForm) ||
      JSON.stringify(features) !== JSON.stringify(savedFeatures),
    [features, form, savedFeatures, savedForm]
  )

  useEffect(() => {
    if (!canRead) {
      return
    }
    let active = true
    void (async () => {
      setLoading(true)
      setFailed(false)
      try {
        const data = await getSiteHome(locale)
        if (!active) {
          return
        }
        const nextForm = toHomeForm(data.home)
        setForm(nextForm)
        setSavedForm(nextForm)
        const nextFeatures = mapFeatures(data.features)
        setFeatures(nextFeatures)
        setSavedFeatures(nextFeatures)
        setSavedAt(data.home.updatedAt)
        setFieldErrors({})
        setFormError("")
      } catch {
        if (active) {
          setFailed(true)
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    })()
    return () => {
      active = false
    }
  }, [canRead, locale, refreshKey])

  // 有未保存改动时，关闭/刷新页面要提示一次，避免静默丢失。
  useEffect(() => {
    if (!dirty) {
      return
    }
    const handler = (event: BeforeUnloadEvent) => {
      event.preventDefault()
      event.returnValue = ""
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [dirty])

  function update(field: HomeField, value: string) {
    setForm((current) => ({ ...current, [field]: value }))
  }

  function updateFeature(index: number, patch: Partial<FeatureForm>) {
    setFeatures((current) =>
      current.map((item, position) =>
        position === index ? { ...item, ...patch } : item
      )
    )
  }

  function moveFeature(index: number, offset: number) {
    setFeatures((current) => {
      const target = index + offset
      if (target < 0 || target >= current.length) {
        return current
      }
      const next = [...current]
      const [item] = next.splice(index, 1)
      next.splice(target, 0, item)
      return next
    })
  }

  function switchLocale(next: "zh" | "en") {
    if (next === locale) {
      return
    }
    if (dirty) {
      setPendingLocale(next)
      return
    }
    setLocale(next)
  }

  /** 本地先拦一遍：长度、URL、CTA 成对、禁用词，错误落到具体字段。 */
  function validate(): FieldErrors {
    const errors: FieldErrors = {}
    for (const [field, limit] of Object.entries(FIELD_LIMITS) as Array<
      [HomeField, number]
    >) {
      if (form[field].trim().length > limit) {
        errors[field] = t("systemSite.home.errors.tooLong", { max: limit })
      }
    }
    if (form.heroImageUrl.trim()) {
      const value = form.heroImageUrl.trim()
      if (!/^https?:\/\/\S+$/i.test(value)) {
        errors.heroImageUrl = t("systemSite.home.errors.invalidHeroImage")
      }
    }
    const hasLabel = form.ctaSecondaryLabel.trim() !== ""
    const hasHref = form.ctaSecondaryHref.trim() !== ""
    if (hasLabel !== hasHref) {
      errors[hasLabel ? "ctaSecondaryHref" : "ctaSecondaryLabel"] = t(
        "systemSite.home.errors.secondaryCtaPair"
      )
    } else if (hasHref && !/^[/#]/.test(form.ctaSecondaryHref.trim())) {
      errors.ctaSecondaryHref = t("systemSite.home.errors.invalidCtaLink")
    }
    for (const [field, value] of Object.entries(form)) {
      if (BANNED_WORDS.some((word) => value.includes(word))) {
        errors[field as HomeField] = t(
          "systemSite.home.errors.disallowedWording"
        )
      }
    }
    features.forEach((item, index) => {
      const featureError: { title?: string; description?: string } = {}
      if (item.title.length > 16) {
        featureError.title = t("systemSite.home.errors.tooLong", { max: 16 })
      }
      if (item.description.length > 48) {
        featureError.description = t("systemSite.home.errors.tooLong", {
          max: 48,
        })
      }
      const banned = BANNED_WORDS.some(
        (word) => item.title.includes(word) || item.description.includes(word)
      )
      if (banned) {
        featureError.title = t("systemSite.home.errors.disallowedWording")
      }
      if (Object.keys(featureError).length > 0) {
        errors.features = { ...errors.features, [index]: featureError }
      }
    })
    return errors
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    if (!canWrite || saving) {
      return
    }
    const errors = validate()
    setFieldErrors(errors)
    setFormError("")
    if (Object.keys(errors).length > 0) {
      return
    }
    setSaving(true)
    const payload: SaveSiteHomeParams = {
      locale,
      home: form,
      features: features.map((item) => ({
        iconKey: item.iconKey,
        title: item.title,
        description: item.description,
      })),
    }
    try {
      const saved = await updateSiteHome(payload)
      const nextForm = toHomeForm(saved.home)
      const nextFeatures = mapFeatures(saved.features)
      setForm(nextForm)
      setSavedForm(nextForm)
      setFeatures(nextFeatures)
      setSavedFeatures(nextFeatures)
      setSavedAt(saved.home.updatedAt)
      toast.success(t("systemSite.home.saved"))
    } catch (err) {
      const raw = err instanceof ApiError ? err.raw : ""
      switch (raw) {
        case "invalid hero image url":
          setFieldErrors({
            heroImageUrl: t("systemSite.home.errors.invalidHeroImage"),
          })
          break
        case "invalid secondary cta link":
          setFieldErrors({
            ctaSecondaryHref: t("systemSite.home.errors.invalidCtaLink"),
          })
          break
        case "secondary cta label and link must be set together":
          setFieldErrors({
            ctaSecondaryLabel: t("systemSite.home.errors.secondaryCtaPair"),
          })
          break
        default:
          setFormError(failureMessage(err, t("state.failed")))
      }
    } finally {
      setSaving(false)
    }
  }

  /** 预览用当前未保存的表单数据，让运营改完立刻看到首页效果。 */
  const previewConfig = useMemo<SiteConfig>(() => {
    const base: SiteConfig = config ?? {
      site: {
        siteName: "",
        logoUrl: "",
        faviconUrl: "",
        contactEmail: "",
        repositoryUrl: "",
        copyrightText: "",
        filingIcpNumber: "",
        filingIcpUrl: "",
        filingPoliceNumber: "",
        filingPoliceUrl: "",
        filingExtraText: "",
        updatedBy: "",
        updatedAt: "",
      },
      loginMethods: {
        emailEnabled: true,
        oauth2Enabled: true,
        updatedBy: "",
        updatedAt: "",
      },
      home: {
        locale,
        updatedBy: "",
        updatedAt: "",
        ...emptyHomeForm,
      },
      features: [],
    }
    return {
      ...base,
      home: { ...base.home, locale, ...form },
      features: features.map((item, index) => ({
        uid: `draft-${index}`,
        iconKey: item.iconKey,
        title: item.title,
        description: item.description,
        sortOrder: index,
      })),
    }
  }, [config, features, form, locale])

  function renderField(field: HomeField, labelKey: string, hintKey?: string) {
    const limit = FIELD_LIMITS[field]
    return (
      <FormField
        label={t(labelKey)}
        htmlFor={`home-${field}`}
        hint={hintKey ? t(hintKey) : undefined}
        error={fieldErrors[field]}
      >
        <Input
          id={`home-${field}`}
          value={form[field]}
          maxLength={limit}
          readOnly={!canWrite}
          aria-invalid={fieldErrors[field] ? true : undefined}
          onChange={(event) => update(field, event.target.value)}
        />
      </FormField>
    )
  }

  function renderTextarea(field: HomeField, labelKey: string) {
    const limit = FIELD_LIMITS[field] ?? 255
    return (
      <FormField
        label={t(labelKey)}
        htmlFor={`home-${field}`}
        hint={t("systemSite.home.count", {
          current: form[field].length,
          max: limit,
        })}
        error={fieldErrors[field]}
      >
        <Textarea
          id={`home-${field}`}
          value={form[field]}
          maxLength={limit}
          rows={2}
          readOnly={!canWrite}
          aria-invalid={fieldErrors[field] ? true : undefined}
          onChange={(event) => update(field, event.target.value)}
        />
      </FormField>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="grid gap-1.5">
            <CardTitle>{t("systemSite.home.title")}</CardTitle>
            <CardDescription>
              {t("systemSite.home.description")}
            </CardDescription>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {dirty ? (
              <span className="text-xs text-warning">
                {t("systemSite.home.unsaved")}
              </span>
            ) : savedAt ? (
              <span className="text-xs text-muted-foreground">
                {t("systemSite.home.savedAt", {
                  time: new Date(savedAt).toLocaleString(),
                })}
              </span>
            ) : null}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setPreviewOpen(true)}
              disabled={!config && !dirty}
            >
              <Eye aria-hidden="true" />
              {t("systemSite.home.preview")}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-4">
        {!canRead ? (
          <p className="text-sm text-muted-foreground">
            {t("systemSite.noPermission")}
          </p>
        ) : loading ? (
          <HomeSkeleton label={t("state.loading")} />
        ) : failed ? (
          <div className="flex flex-col items-start gap-2">
            <p className="text-sm text-destructive">
              {t("systemSite.home.loadFailed")}
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setLocale((current) => current)}
            >
              {t("action.retry")}
            </Button>
          </div>
        ) : (
          <form noValidate className="grid gap-5" onSubmit={onSubmit}>
            <div
              role="group"
              aria-label={t("systemSite.home.localeLabel")}
              className="inline-flex w-fit items-center gap-0.5 rounded-lg border bg-muted/40 p-0.5"
            >
              {(["zh", "en"] as const).map((item) => (
                <Button
                  key={item}
                  type="button"
                  size="xs"
                  variant={locale === item ? "secondary" : "ghost"}
                  aria-pressed={locale === item}
                  onClick={() => switchLocale(item)}
                >
                  {t(`systemSite.home.locale.${item}`)}
                </Button>
              ))}
            </div>

            {GROUPS.map((group) => (
              <Collapsible key={group.key} defaultOpen>
                <CollapsibleTrigger asChild>
                  <Button
                    type="button"
                    variant="ghost"
                    className="group w-full justify-start px-2 text-sm font-medium"
                  >
                    <ChevronRight
                      aria-hidden="true"
                      className="transition-transform group-data-[state=open]:rotate-90"
                    />
                    {t(group.labelKey)}
                  </Button>
                </CollapsibleTrigger>
                <CollapsibleContent className="grid gap-3 px-2 pt-3 pb-1 sm:grid-cols-2">
                  {group.fields.map((field) => {
                    const labelKey = `systemSite.home.fields.${field}`
                    if (group.textareas?.includes(field)) {
                      return renderTextarea(field, labelKey)
                    }
                    const hintKey =
                      field === "heroImageUrl"
                        ? "systemSite.home.fields.heroImageUrlHint"
                        : field === "ctaSecondaryHref"
                          ? "systemSite.home.fields.ctaSecondaryHrefHint"
                          : field === "heroTitle"
                            ? "systemSite.home.fields.heroTitleHint"
                            : field === "ctaPrimaryLabel"
                              ? "systemSite.home.fields.ctaPrimaryHint"
                              : undefined
                    return (
                      <div
                        key={field}
                        className={
                          field === "heroSubtitle" ? "sm:col-span-2" : undefined
                        }
                      >
                        {renderField(field, labelKey, hintKey)}
                      </div>
                    )
                  })}
                </CollapsibleContent>
              </Collapsible>
            ))}

            <section className="grid gap-3 rounded-lg border border-border/60 p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h3 className="text-sm font-medium">
                  {t("systemSite.home.features.heading")}
                </h3>
                <span className="text-xs text-muted-foreground">
                  {t("systemSite.home.features.limit", {
                    current: features.length,
                    max: MAX_FEATURES,
                  })}
                </span>
              </div>
              {features.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  {t("systemSite.home.features.empty")}
                </p>
              ) : (
                <ul className="grid gap-3">
                  {features.map((item, index) => {
                    const featureError = fieldErrors.features?.[index]
                    return (
                      <li
                        key={index}
                        className="grid gap-3 rounded-lg border border-border/60 p-3 sm:grid-cols-[8rem_minmax(0,1fr)_auto]"
                      >
                        <FormField
                          label={t("systemSite.home.features.icon")}
                          htmlFor={`feature-icon-${index}`}
                        >
                          <Select
                            value={item.iconKey}
                            disabled={!canWrite}
                            onValueChange={(value) =>
                              updateFeature(index, { iconKey: value })
                            }
                          >
                            <SelectTrigger
                              id={`feature-icon-${index}`}
                              aria-label={t("systemSite.home.features.icon")}
                            >
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {FEATURE_ICONS.map((icon) => (
                                <SelectItem key={icon} value={icon}>
                                  {t(`systemSite.home.icons.${icon}`)}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </FormField>
                        <div className="grid gap-2">
                          <FormField
                            label={t("systemSite.home.features.title")}
                            htmlFor={`feature-title-${index}`}
                            error={featureError?.title}
                          >
                            <Input
                              id={`feature-title-${index}`}
                              value={item.title}
                              maxLength={16}
                              readOnly={!canWrite}
                              aria-invalid={
                                featureError?.title ? true : undefined
                              }
                              onChange={(event) =>
                                updateFeature(index, {
                                  title: event.target.value,
                                })
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.home.features.description")}
                            htmlFor={`feature-desc-${index}`}
                            hint={t("systemSite.home.count", {
                              current: item.description.length,
                              max: 48,
                            })}
                            error={featureError?.description}
                          >
                            <Textarea
                              id={`feature-desc-${index}`}
                              value={item.description}
                              maxLength={48}
                              rows={2}
                              readOnly={!canWrite}
                              aria-invalid={
                                featureError?.description ? true : undefined
                              }
                              onChange={(event) =>
                                updateFeature(index, {
                                  description: event.target.value,
                                })
                              }
                            />
                          </FormField>
                        </div>
                        <div className="flex items-start gap-1 sm:pt-6">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label={t("systemSite.home.features.moveUp")}
                            disabled={!canWrite || index === 0}
                            onClick={() => moveFeature(index, -1)}
                          >
                            <ChevronRight
                              className="size-4 -rotate-90"
                              aria-hidden="true"
                            />
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label={t("systemSite.home.features.moveDown")}
                            disabled={
                              !canWrite || index === features.length - 1
                            }
                            onClick={() => moveFeature(index, 1)}
                          >
                            <ChevronRight
                              className="size-4 rotate-90"
                              aria-hidden="true"
                            />
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label={t("systemSite.home.features.remove")}
                            disabled={!canWrite}
                            onClick={() =>
                              setFeatures((current) =>
                                current.filter(
                                  (_, position) => position !== index
                                )
                              )
                            }
                          >
                            <CircleAlert
                              className="size-4"
                              aria-hidden="true"
                            />
                          </Button>
                        </div>
                      </li>
                    )
                  })}
                </ul>
              )}
              <div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={!canWrite || features.length >= MAX_FEATURES}
                  onClick={() =>
                    setFeatures((current) => [
                      ...current,
                      { iconKey: "activity", title: "", description: "" },
                    ])
                  }
                >
                  {t("systemSite.home.features.add")}
                </Button>
              </div>
            </section>

            {formError ? (
              <p
                role="alert"
                className="flex items-center gap-2 text-sm text-destructive"
              >
                <CircleAlert className="size-4" aria-hidden="true" />
                {formError}
              </p>
            ) : null}

            {canWrite ? (
              <div className="flex justify-end">
                <Button
                  type="submit"
                  disabled={saving}
                  aria-busy={saving || undefined}
                >
                  {saving ? (
                    <Loader2
                      className="size-4 animate-spin"
                      aria-hidden="true"
                    />
                  ) : (
                    <Save />
                  )}
                  {t("systemSite.home.save")}
                </Button>
              </div>
            ) : null}
          </form>
        )}
      </CardContent>

      <Dialog open={previewOpen} onOpenChange={setPreviewOpen}>
        <DialogContent className="sm:max-w-5xl">
          <DialogHeader>
            <DialogTitle>{t("systemSite.home.previewTitle")}</DialogTitle>
            <DialogDescription>
              {t("systemSite.home.previewDescription")}
            </DialogDescription>
          </DialogHeader>
          <DialogBody>
            <RenderGuard
              message={t("systemSite.home.previewFailed")}
              retryLabel={t("action.retry")}
            >
              <SiteHomeSections
                config={previewConfig}
                page={null}
                navItems={[]}
              />
            </RenderGuard>
          </DialogBody>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setPreviewOpen(false)}
            >
              {t("systemSite.home.previewClose")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={pendingLocale !== null}
        onOpenChange={(open) => {
          if (!open) {
            setPendingLocale(null)
          }
        }}
      >
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t("systemSite.home.localeConfirmTitle")}</DialogTitle>
            <DialogDescription>
              {t("systemSite.home.localeConfirmBody")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setPendingLocale(null)}
            >
              {t("action.cancel")}
            </Button>
            <Button
              type="button"
              onClick={() => {
                if (pendingLocale) {
                  setLocale(pendingLocale)
                }
                setPendingLocale(null)
              }}
            >
              {t("systemSite.home.localeConfirmAction")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  )
}
