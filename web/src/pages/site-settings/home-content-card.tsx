import {
  ArrowDown,
  ArrowUp,
  CircleAlert,
  Loader2,
  Plus,
  Save,
  Trash2,
} from "lucide-react"
import { useEffect, useState, type FormEvent } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { ApiError } from "@/api/http"
import { getSiteHome, updateSiteHome } from "@/api/site"
import type { SaveSiteHomeParams, SiteHomeContent } from "@/api/site/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { failureMessage } from "@/lib/form-error"

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

/** 每个字段的字数上限，用于输入框 maxLength 与本地提示。 */
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
  const [locale, setLocale] = useState<"zh" | "en">("zh")
  const [form, setForm] = useState<HomeForm>(emptyHomeForm)
  const [features, setFeatures] = useState<FeatureForm[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState("")

  useEffect(() => {
    if (!canRead) {
      return
    }
    let active = true
    setLoading(true)
    setFailed(false)
    void getSiteHome(locale)
      .then((data) => {
        if (!active) {
          return
        }
        setForm(toHomeForm(data.home))
        setFeatures(
          data.features.map((item) => ({
            iconKey: item.iconKey,
            title: item.title,
            description: item.description,
          }))
        )
      })
      .catch(() => {
        if (active) {
          setFailed(true)
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [canRead, locale, refreshKey])

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

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    if (!canWrite || saving) {
      return
    }
    if (
      (form.ctaSecondaryLabel.trim() === "") !==
      (form.ctaSecondaryHref.trim() === "")
    ) {
      setFormError(t("systemSite.home.errors.secondaryCtaPair"))
      return
    }
    setSaving(true)
    setFormError("")
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
      setForm(toHomeForm(saved.home))
      setFeatures(
        saved.features.map((item) => ({
          iconKey: item.iconKey,
          title: item.title,
          description: item.description,
        }))
      )
      toast.success(t("systemSite.home.saved"))
    } catch (err) {
      const raw = err instanceof ApiError ? err.raw : ""
      if (raw === "invalid secondary cta link") {
        setFormError(t("systemSite.home.errors.invalidCtaLink"))
      } else if (raw === "invalid hero image url") {
        setFormError(t("systemSite.home.errors.invalidHeroImage"))
      } else if (raw === "too many feature cards") {
        setFormError(t("systemSite.home.errors.tooManyFeatures"))
      } else {
        setFormError(failureMessage(err, t("state.failed")))
      }
    } finally {
      setSaving(false)
    }
  }

  function field(field: HomeField, labelKey: string, hintKey?: string) {
    const limit = FIELD_LIMITS[field]
    return (
      <FormField
        label={t(labelKey)}
        htmlFor={`home-${field}`}
        hint={hintKey ? t(hintKey) : undefined}
      >
        <Input
          id={`home-${field}`}
          value={form[field]}
          maxLength={limit}
          readOnly={!canWrite}
          onChange={(event) => update(field, event.target.value)}
        />
      </FormField>
    )
  }

  function textarea(field: HomeField, labelKey: string, rows = 2) {
    const limit = FIELD_LIMITS[field]
    return (
      <FormField label={t(labelKey)} htmlFor={`home-${field}`}>
        <textarea
          id={`home-${field}`}
          value={form[field]}
          rows={rows}
          maxLength={limit}
          readOnly={!canWrite}
          className="min-h-16 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs outline-none focus-visible:ring-3 focus-visible:ring-ring"
          onChange={(event) => update(field, event.target.value)}
        />
      </FormField>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("systemSite.home.title")}</CardTitle>
        <CardDescription>{t("systemSite.home.description")}</CardDescription>
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
                  onClick={() => setLocale(item)}
                >
                  {t(`systemSite.home.locale.${item}`)}
                </Button>
              ))}
            </div>

            <section className="grid gap-3">
              <h3 className="text-sm font-medium">
                {t("systemSite.home.sections.hero")}
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                {field("heroTitle", "systemSite.home.fields.heroTitle")}
                {field("heroSubtitle", "systemSite.home.fields.heroSubtitle")}
                {field("heroImageUrl", "systemSite.home.fields.heroImageUrl")}
                {field("ctaPrimaryLabel", "systemSite.home.fields.ctaPrimaryLabel")}
                {field("ctaSecondaryLabel", "systemSite.home.fields.ctaSecondaryLabel")}
                {field("ctaSecondaryHref", "systemSite.home.fields.ctaSecondaryHref")}
              </div>
            </section>

            <section className="grid gap-3">
              <h3 className="text-sm font-medium">
                {t("systemSite.home.sections.features")}
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                {field("featuresTitle", "systemSite.home.fields.featuresTitle")}
                {field("featuresSubtitle", "systemSite.home.fields.featuresSubtitle")}
              </div>
              {features.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  {t("systemSite.home.features.empty")}
                </p>
              ) : (
                <ul className="grid gap-3">
                  {features.map((item, index) => (
                    <li
                      key={index}
                      className="grid gap-3 rounded-lg border border-border/60 p-3 sm:grid-cols-[8rem_minmax(0,1fr)_auto]"
                    >
                      <FormField
                        label={t("systemSite.home.features.icon")}
                        htmlFor={`feature-icon-${index}`}
                      >
                        <select
                          id={`feature-icon-${index}`}
                          value={item.iconKey}
                          disabled={!canWrite}
                          className="h-8 w-full rounded-md border border-input bg-transparent px-2 text-sm"
                          onChange={(event) =>
                            updateFeature(index, { iconKey: event.target.value })
                          }
                        >
                          {FEATURE_ICONS.map((icon) => (
                            <option key={icon} value={icon}>
                              {t(`systemSite.home.icons.${icon}`)}
                            </option>
                          ))}
                        </select>
                      </FormField>
                      <div className="grid gap-2">
                        <FormField
                          label={t("systemSite.home.features.title")}
                          htmlFor={`feature-title-${index}`}
                        >
                          <Input
                            id={`feature-title-${index}`}
                            value={item.title}
                            maxLength={16}
                            readOnly={!canWrite}
                            onChange={(event) =>
                              updateFeature(index, { title: event.target.value })
                            }
                          />
                        </FormField>
                        <FormField
                          label={t("systemSite.home.features.description")}
                          htmlFor={`feature-desc-${index}`}
                        >
                          <Input
                            id={`feature-desc-${index}`}
                            value={item.description}
                            maxLength={48}
                            readOnly={!canWrite}
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
                          <ArrowUp />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          aria-label={t("systemSite.home.features.moveDown")}
                          disabled={!canWrite || index === features.length - 1}
                          onClick={() => moveFeature(index, 1)}
                        >
                          <ArrowDown />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          aria-label={t("systemSite.home.features.remove")}
                          disabled={!canWrite}
                          onClick={() =>
                            setFeatures((current) =>
                              current.filter((_, position) => position !== index)
                            )
                          }
                        >
                          <Trash2 />
                        </Button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
              <div className="flex items-center gap-3">
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
                  <Plus />
                  {t("systemSite.home.features.add")}
                </Button>
                <span className="text-xs text-muted-foreground">
                  {t("systemSite.home.features.limit", {
                    current: features.length,
                    max: MAX_FEATURES,
                  })}
                </span>
              </div>
            </section>

            <section className="grid gap-3">
              <h3 className="text-sm font-medium">
                {t("systemSite.home.sections.explainers")}
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                {field("howTitle", "systemSite.home.fields.howTitle")}
                {field("permTitle", "systemSite.home.fields.permTitle")}
                {textarea("howBody", "systemSite.home.fields.howBody")}
                {textarea("permBody", "systemSite.home.fields.permBody")}
                {field("deployTitle", "systemSite.home.fields.deployTitle")}
                {field("docTitle", "systemSite.home.fields.docTitle")}
                {textarea("deployBody", "systemSite.home.fields.deployBody")}
              </div>
            </section>

            <section className="grid gap-3">
              <h3 className="text-sm font-medium">
                {t("systemSite.home.sections.closing")}
              </h3>
              <div className="grid gap-3 sm:grid-cols-2">
                {field("closingTitle", "systemSite.home.fields.closingTitle")}
                {field("closingDescription", "systemSite.home.fields.closingDescription")}
                {field("closingButtonLabel", "systemSite.home.fields.closingButtonLabel")}
              </div>
            </section>

            {formError ? (
              <p role="alert" className="flex items-center gap-2 text-sm text-destructive">
                <CircleAlert className="size-4" aria-hidden="true" />
                {formError}
              </p>
            ) : null}

            {canWrite ? (
              <div className="flex justify-end">
                <Button type="submit" disabled={saving} aria-busy={saving || undefined}>
                  {saving ? (
                    <Loader2 className="size-4 animate-spin" aria-hidden="true" />
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
    </Card>
  )
}
