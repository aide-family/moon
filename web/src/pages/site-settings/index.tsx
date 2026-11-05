import { CircleAlert, Loader2, RefreshCw, Save } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { toast } from "sonner"

import { ApiError } from "@/api/http"
import {
  getLoginMethods,
  getSiteSettings,
  updateLoginMethods,
  updateSiteSettings,
} from "@/api/site"
import type {
  LoginMethods,
  SaveLoginMethodsParams,
  SaveSiteSettingsParams,
  SiteSettings,
} from "@/api/site/types"
import { useSession } from "@/auth/session"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { isEmail } from "@/lib/email"

import { EmailSettingsCard } from "./email-settings-card"
import { HomeContentCard } from "./home-content-card"
import { OAuthProvidersCard } from "./oauth-providers-card"

/** 设置页区块顺序；左侧快捷导航与滚动高亮共用这一份。 */
const SECTIONS = [
  { key: "site", labelKey: "systemSite.nav.site" },
  { key: "login", labelKey: "systemSite.nav.login" },
  { key: "email", labelKey: "systemSite.nav.email" },
  { key: "oauth", labelKey: "systemSite.nav.oauth" },
  { key: "home", labelKey: "systemSite.nav.home" },
] as const

type SectionKey = (typeof SECTIONS)[number]["key"]

type SiteField = keyof SaveSiteSettingsParams
type SiteFieldErrors = Partial<Record<SiteField, string>>

const emptySiteForm: SaveSiteSettingsParams = {
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
}

const emptyLoginForm: SaveLoginMethodsParams = {
  emailEnabled: true,
  oauth2Enabled: true,
}

// 后端校验失败只返回一条英文原文，没有字段名；按原文把错误放回对应输入框。
const messageFields: Record<string, SiteField> = {
  "contact email is invalid": "contactEmail",
  "logo url is invalid": "logoUrl",
  "favicon url is invalid": "faviconUrl",
  "repository url is invalid": "repositoryUrl",
  "filing icp url is invalid": "filingIcpUrl",
  "filing police url is invalid": "filingPoliceUrl",
}

function toSiteForm(site: SiteSettings): SaveSiteSettingsParams {
  return {
    siteName: site.siteName,
    logoUrl: site.logoUrl,
    faviconUrl: site.faviconUrl,
    contactEmail: site.contactEmail,
    repositoryUrl: site.repositoryUrl,
    copyrightText: site.copyrightText,
    filingIcpNumber: site.filingIcpNumber,
    filingIcpUrl: site.filingIcpUrl,
    filingPoliceNumber: site.filingPoliceNumber,
    filingPoliceUrl: site.filingPoliceUrl,
    filingExtraText: site.filingExtraText,
  }
}

function toLoginForm(methods: LoginMethods): SaveLoginMethodsParams {
  return {
    emailEnabled: methods.emailEnabled,
    oauth2Enabled: methods.oauth2Enabled,
  }
}

function trimSiteForm(form: SaveSiteSettingsParams): SaveSiteSettingsParams {
  return {
    siteName: form.siteName.trim(),
    logoUrl: form.logoUrl.trim(),
    faviconUrl: form.faviconUrl.trim(),
    contactEmail: form.contactEmail.trim(),
    repositoryUrl: form.repositoryUrl.trim(),
    copyrightText: form.copyrightText.trim(),
    filingIcpNumber: form.filingIcpNumber.trim(),
    filingIcpUrl: form.filingIcpUrl.trim(),
    filingPoliceNumber: form.filingPoliceNumber.trim(),
    filingPoliceUrl: form.filingPoliceUrl.trim(),
    filingExtraText: form.filingExtraText.trim(),
  }
}

/** 与后端保持一致：只接受带 host 的 http(s) 地址。 */
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

function readErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof ApiError) {
    return err.message || fallback
  }
  if (err instanceof Error && err.message) {
    return err.message
  }
  return fallback
}

/** 官网页脚的实时预览：只拼当前非空的行，空行由文案说明不会展示。 */
function FooterPreview({ form }: { form: SaveSiteSettingsParams }) {
  const { t } = useTranslation()
  const year = new Date().getFullYear()
  // 与官网页脚同一规则：填了就是完整一行，留空才拼默认文案。
  const copyrightLine =
    form.copyrightText.trim() ||
    t("site.footer.copyright", {
      year,
      name: t("systemSite.preview.copyrightFallback"),
    })
  const lines = [
    copyrightLine,
    form.filingIcpNumber.trim(),
    form.filingPoliceNumber.trim(),
    form.filingExtraText.trim(),
  ].filter((line) => line !== "")
  const hasFiling =
    form.filingIcpNumber.trim() !== "" ||
    form.filingPoliceNumber.trim() !== "" ||
    form.filingExtraText.trim() !== ""
  return (
    <div className="grid gap-1 rounded-lg border border-border/60 bg-muted/20 px-3 py-2 sm:col-span-full">
      <p className="text-xs text-muted-foreground">
        {t("systemSite.preview.title")}
      </p>
      <p className="text-sm break-words">
        {lines.map((line, index) => (
          <span key={`${line}-${index}`}>
            {index > 0 ? (
              <span aria-hidden="true" className="text-muted-foreground">
                {" · "}
              </span>
            ) : null}
            {line}
          </span>
        ))}
      </p>
      {hasFiling ? null : (
        <p className="text-xs text-muted-foreground">
          {t("systemSite.preview.empty")}
        </p>
      )}
    </div>
  )
}

export function SiteSettingsPage() {
  const { t } = useTranslation()
  const session = useSession()
  const canRead = session.canSystem("system:site:read")
  const canWrite = session.canSystem("system:site:write")
  const canLoginRead = session.canSystem("system:login:read")
  const canLoginWrite = session.canSystem("system:login:write")

  const [site, setSite] = useState<SiteSettings | null>(null)
  const [siteForm, setSiteForm] = useState(emptySiteForm)
  const [loginForm, setLoginForm] = useState(emptyLoginForm)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [savingSite, setSavingSite] = useState(false)
  const [savingLogin, setSavingLogin] = useState(false)
  const [siteFieldErrors, setSiteFieldErrors] = useState<SiteFieldErrors>({})
  const [siteFormError, setSiteFormError] = useState("")
  const [loginError, setLoginError] = useState("")

  // 左侧快捷导航：长表单页按区块跳转，滚动时高亮当前区块。
  const scrollRef = useRef<HTMLDivElement>(null)
  const [activeSection, setActiveSection] = useState<SectionKey>("site")
  useEffect(() => {
    const root = scrollRef.current
    if (!root) {
      return
    }
    const sync = () => {
      const top = root.getBoundingClientRect().top
      let current: SectionKey = SECTIONS[0].key
      for (const section of SECTIONS) {
        const element = document.getElementById(`site-settings-${section.key}`)
        if (element && element.getBoundingClientRect().top - top <= 28) {
          current = section.key
        }
      }
      setActiveSection(current)
    }
    sync()
    root.addEventListener("scroll", sync, { passive: true })
    return () => root.removeEventListener("scroll", sync)
  }, [session.loading, site])

  function jumpTo(key: SectionKey) {
    const root = scrollRef.current
    const element = document.getElementById(`site-settings-${key}`)
    if (!root || !element) {
      return
    }
    // 只滚这个内容容器：scrollIntoView 会连外层固定视口一起滚，导致整页上移、
    // 标题区被推出屏幕——这正是「点左侧导航后顶部标题消失」的原因。
    const offset =
      element.getBoundingClientRect().top -
      root.getBoundingClientRect().top +
      root.scrollTop
    root.scrollTo({ top: offset, behavior: "smooth" })
  }

  useEffect(() => {
    if (session.loading || !canRead) {
      return
    }
    let active = true
    const loginRequest = canLoginRead
      ? getLoginMethods()
      : Promise.resolve<{ loginMethods: LoginMethods } | null>(null)
    void Promise.all([getSiteSettings(), loginRequest])
      .then(([siteResult, loginResult]) => {
        if (!active) {
          return
        }
        setSite(siteResult.site)
        setSiteForm(toSiteForm(siteResult.site))
        if (loginResult) {
          setLoginForm(toLoginForm(loginResult.loginMethods))
        }
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
  }, [attempt, canLoginRead, canRead, session.loading])

  if (!session.loading && !canRead) {
    return <Navigate to="/forbidden" replace />
  }

  function reload() {
    setLoading(true)
    setFailed(false)
    setAttempt((value) => value + 1)
  }

  function updateSiteField(field: SiteField, value: string) {
    setSiteForm((form) => ({ ...form, [field]: value }))
  }

  async function onSubmitSite(event: React.FormEvent) {
    event.preventDefault()
    if (!canWrite || savingSite) {
      return
    }
    const values = trimSiteForm(siteForm)
    const errors: SiteFieldErrors = {}
    if (values.contactEmail && !isEmail(values.contactEmail)) {
      errors.contactEmail = t("form.invalidEmail")
    }
    const urls: Array<[SiteField, string]> = [
      ["logoUrl", values.logoUrl],
      ["faviconUrl", values.faviconUrl],
      ["repositoryUrl", values.repositoryUrl],
      ["filingIcpUrl", values.filingIcpUrl],
      ["filingPoliceUrl", values.filingPoliceUrl],
    ]
    for (const [field, value] of urls) {
      if (value && !isHTTPURL(value)) {
        errors[field] = t("form.invalidUrl")
      }
    }
    if (Object.keys(errors).length > 0) {
      setSiteFieldErrors(errors)
      setSiteFormError("")
      return
    }
    setSavingSite(true)
    setSiteFieldErrors({})
    setSiteFormError("")
    try {
      const result = await updateSiteSettings(values)
      setSite(result.site)
      setSiteForm(toSiteForm(result.site))
      toast.success(t("systemSite.saved"))
    } catch (err) {
      const message = readErrorMessage(err, t("state.failed"))
      const raw = err instanceof ApiError ? err.raw : ""
      const field = messageFields[raw]
      if (field) {
        setSiteFieldErrors({ [field]: message })
      } else {
        setSiteFormError(message)
      }
    } finally {
      setSavingSite(false)
    }
  }

  async function onSubmitLogin(event: React.FormEvent) {
    event.preventDefault()
    if (!canLoginWrite || savingLogin) {
      return
    }
    if (!loginForm.emailEnabled && !loginForm.oauth2Enabled) {
      setLoginError(t("systemSite.login.atLeastOne"))
      return
    }
    setSavingLogin(true)
    setLoginError("")
    try {
      const result = await updateLoginMethods(loginForm)
      setLoginForm(toLoginForm(result.loginMethods))
      toast.success(t("systemSite.loginSaved"))
    } catch (err) {
      setLoginError(readErrorMessage(err, t("state.failed")))
    } finally {
      setSavingLogin(false)
    }
  }

  const initialLoading = session.loading || (loading && !site)

  return (
    // 标题区固定在滚动区之外：点左侧快捷导航会滚动内容，标题跟着滚出视口就「消失」了。
    <div className="flex min-h-0 flex-1 flex-col gap-4">
      <PageHeader
        title={t("systemSite.title")}
        description={t("systemSite.description")}
      >
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={reload}
          disabled={initialLoading}
        >
          <RefreshCw
            className={initialLoading ? "motion-safe:animate-spin" : undefined}
            aria-hidden="true"
          />
          {t("action.refresh")}
        </Button>
      </PageHeader>

      {/* 唯一滚动所有者：内容区自己滚动，标题与刷新按钮常驻。 */}
      <div
        ref={scrollRef}
        className="flex min-h-0 flex-1 flex-col gap-5 overflow-y-auto [&>*]:shrink-0"
      >
        {/* 窄屏：横向滚动的快捷导航；宽屏：左侧粘性目录。 */}
        <nav
          aria-label={t("systemSite.navLabel")}
          className="-mt-1 flex gap-2 overflow-x-auto pb-1 lg:hidden"
        >
          {SECTIONS.map((section) => (
            <Button
              key={section.key}
              type="button"
              size="xs"
              variant={activeSection === section.key ? "secondary" : "outline"}
              className="shrink-0"
              onClick={() => jumpTo(section.key)}
            >
              {t(section.labelKey)}
            </Button>
          ))}
        </nav>

        <div className="flex min-h-0 items-start gap-6">
          <nav
            aria-label={t("systemSite.navLabel")}
            className="sticky top-0 hidden w-40 shrink-0 lg:block"
          >
            <ul className="grid gap-0.5">
              {SECTIONS.map((section) => (
                <li key={section.key}>
                  <Button
                    type="button"
                    variant={
                      activeSection === section.key ? "secondary" : "ghost"
                    }
                    className="w-full justify-start text-sm font-normal"
                    aria-current={
                      activeSection === section.key ? "location" : undefined
                    }
                    onClick={() => jumpTo(section.key)}
                  >
                    <span
                      aria-hidden="true"
                      className="size-1.5 rounded-full bg-primary-ink"
                    />
                    {t(section.labelKey)}
                  </Button>
                </li>
              ))}
            </ul>
          </nav>
          <div className="grid min-w-0 flex-1 gap-5">
            {initialLoading ? (
              <div
                className="grid gap-4"
                aria-busy="true"
                aria-label={t("state.loading")}
              >
                <Skeleton className="h-72 rounded-xl" />
                <Skeleton className="h-48 rounded-xl" />
              </div>
            ) : failed && !site ? (
              <Card>
                <CardContent className="flex flex-col items-center gap-3 py-10 text-center">
                  <CircleAlert
                    className="size-6 text-destructive"
                    aria-hidden="true"
                  />
                  <p className="font-medium">{t("systemSite.loadFailed")}</p>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={reload}
                  >
                    {t("action.retry")}
                  </Button>
                </CardContent>
              </Card>
            ) : site ? (
              <>
                <div id="site-settings-site">
                  <Card>
                    <CardHeader>
                      <CardTitle>{t("systemSite.site.title")}</CardTitle>
                      <CardDescription>
                        {t("systemSite.site.description")}
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <form
                        noValidate
                        className="grid gap-4"
                        onSubmit={onSubmitSite}
                      >
                        <div className="grid gap-4 sm:grid-cols-2">
                          <FormField
                            label={t("systemSite.fields.siteName")}
                            htmlFor="site-name"
                            hint={t("systemSite.fields.siteNameHint")}
                            error={siteFieldErrors.siteName}
                          >
                            <Input
                              id="site-name"
                              value={siteForm.siteName}
                              placeholder={t(
                                "systemSite.fields.siteNamePlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.siteName ? true : undefined
                              }
                              onChange={(event) =>
                                updateSiteField("siteName", event.target.value)
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.contactEmail")}
                            htmlFor="site-contact-email"
                            error={siteFieldErrors.contactEmail}
                          >
                            <Input
                              id="site-contact-email"
                              type="email"
                              value={siteForm.contactEmail}
                              placeholder={t(
                                "systemSite.fields.contactEmailPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.contactEmail ? true : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "contactEmail",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.logoUrl")}
                            htmlFor="site-logo-url"
                            error={siteFieldErrors.logoUrl}
                          >
                            <Input
                              id="site-logo-url"
                              value={siteForm.logoUrl}
                              placeholder={t(
                                "systemSite.fields.logoUrlPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.logoUrl ? true : undefined
                              }
                              onChange={(event) =>
                                updateSiteField("logoUrl", event.target.value)
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.faviconUrl")}
                            htmlFor="site-favicon-url"
                            error={siteFieldErrors.faviconUrl}
                          >
                            <Input
                              id="site-favicon-url"
                              value={siteForm.faviconUrl}
                              placeholder={t(
                                "systemSite.fields.faviconUrlPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.faviconUrl ? true : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "faviconUrl",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.repositoryUrl")}
                            htmlFor="site-repository-url"
                            className="sm:col-span-full"
                            error={siteFieldErrors.repositoryUrl}
                          >
                            <Input
                              id="site-repository-url"
                              value={siteForm.repositoryUrl}
                              placeholder={t(
                                "systemSite.fields.repositoryUrlPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.repositoryUrl ? true : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "repositoryUrl",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.copyrightText")}
                            htmlFor="site-copyright"
                            hint={t("systemSite.fields.copyrightTextHint")}
                            className="sm:col-span-full"
                            error={siteFieldErrors.copyrightText}
                          >
                            <Input
                              id="site-copyright"
                              value={siteForm.copyrightText}
                              placeholder={t(
                                "systemSite.fields.copyrightTextPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.copyrightText ? true : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "copyrightText",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.filingIcpNumber")}
                            htmlFor="site-filing-icp-number"
                            error={siteFieldErrors.filingIcpNumber}
                          >
                            <Input
                              id="site-filing-icp-number"
                              value={siteForm.filingIcpNumber}
                              placeholder={t(
                                "systemSite.fields.filingIcpNumberPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.filingIcpNumber
                                  ? true
                                  : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "filingIcpNumber",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.filingIcpUrl")}
                            htmlFor="site-filing-icp-url"
                            hint={t("systemSite.fields.filingIcpUrlHint")}
                            error={siteFieldErrors.filingIcpUrl}
                          >
                            <Input
                              id="site-filing-icp-url"
                              value={siteForm.filingIcpUrl}
                              placeholder={t(
                                "systemSite.fields.filingIcpUrlPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.filingIcpUrl ? true : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "filingIcpUrl",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.filingPoliceNumber")}
                            htmlFor="site-filing-police-number"
                            error={siteFieldErrors.filingPoliceNumber}
                          >
                            <Input
                              id="site-filing-police-number"
                              value={siteForm.filingPoliceNumber}
                              placeholder={t(
                                "systemSite.fields.filingPoliceNumberPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.filingPoliceNumber
                                  ? true
                                  : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "filingPoliceNumber",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.filingPoliceUrl")}
                            htmlFor="site-filing-police-url"
                            error={siteFieldErrors.filingPoliceUrl}
                          >
                            <Input
                              id="site-filing-police-url"
                              value={siteForm.filingPoliceUrl}
                              placeholder={t(
                                "systemSite.fields.filingPoliceUrlPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.filingPoliceUrl
                                  ? true
                                  : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "filingPoliceUrl",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("systemSite.fields.filingExtraText")}
                            htmlFor="site-filing-extra"
                            hint={t("systemSite.fields.filingExtraTextHint")}
                            className="sm:col-span-full"
                            error={siteFieldErrors.filingExtraText}
                          >
                            <Input
                              id="site-filing-extra"
                              value={siteForm.filingExtraText}
                              placeholder={t(
                                "systemSite.fields.filingExtraTextPlaceholder"
                              )}
                              readOnly={!canWrite}
                              aria-invalid={
                                siteFieldErrors.filingExtraText
                                  ? true
                                  : undefined
                              }
                              onChange={(event) =>
                                updateSiteField(
                                  "filingExtraText",
                                  event.target.value
                                )
                              }
                            />
                          </FormField>
                        </div>
                        <FooterPreview form={siteForm} />
                        {siteFormError ? (
                          <p role="alert" className="text-sm text-destructive">
                            {siteFormError}
                          </p>
                        ) : null}
                        {canWrite ? (
                          <div className="flex justify-end">
                            <Button type="submit" disabled={savingSite}>
                              {savingSite ? (
                                <Loader2
                                  className="size-4 animate-spin"
                                  aria-hidden="true"
                                />
                              ) : (
                                <Save aria-hidden="true" />
                              )}
                              {savingSite
                                ? t("systemSite.saving")
                                : t("systemSite.save")}
                            </Button>
                          </div>
                        ) : (
                          <p className="text-sm text-muted-foreground">
                            {t("systemSite.readOnly")}
                          </p>
                        )}
                      </form>
                    </CardContent>
                  </Card>
                </div>

                <div id="site-settings-login">
                  <Card>
                    <CardHeader>
                      <CardTitle>{t("systemSite.login.title")}</CardTitle>
                      <CardDescription>
                        {t("systemSite.login.description")}
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      {canLoginRead ? (
                        <form
                          noValidate
                          className="grid gap-4"
                          onSubmit={onSubmitLogin}
                        >
                          <div className="grid gap-3">
                            <label className="flex items-start gap-3 rounded-lg border border-border/60 p-4">
                              <Checkbox
                                id="login-email-enabled"
                                className="mt-0.5"
                                checked={loginForm.emailEnabled}
                                disabled={!canLoginWrite}
                                onCheckedChange={(value) =>
                                  setLoginForm((form) => ({
                                    ...form,
                                    emailEnabled: value === true,
                                  }))
                                }
                              />
                              <span className="grid gap-1">
                                <span className="text-sm font-medium">
                                  {t("systemSite.login.emailEnabled")}
                                </span>
                                <span className="text-xs text-muted-foreground">
                                  {t("systemSite.login.emailEnabledHint")}
                                </span>
                              </span>
                            </label>
                            <label className="flex items-start gap-3 rounded-lg border border-border/60 p-4">
                              <Checkbox
                                id="login-oauth2-enabled"
                                className="mt-0.5"
                                checked={loginForm.oauth2Enabled}
                                disabled={!canLoginWrite}
                                onCheckedChange={(value) =>
                                  setLoginForm((form) => ({
                                    ...form,
                                    oauth2Enabled: value === true,
                                  }))
                                }
                              />
                              <span className="grid gap-1">
                                <span className="text-sm font-medium">
                                  {t("systemSite.login.oauth2Enabled")}
                                </span>
                                <span className="text-xs text-muted-foreground">
                                  {t("systemSite.login.oauth2EnabledHint")}
                                </span>
                              </span>
                            </label>
                          </div>
                          {loginError ? (
                            <p
                              role="alert"
                              className="text-sm text-destructive"
                            >
                              {loginError}
                            </p>
                          ) : null}
                          {canLoginWrite ? (
                            <div className="flex justify-end">
                              <Button type="submit" disabled={savingLogin}>
                                {savingLogin ? (
                                  <Loader2
                                    className="size-4 animate-spin"
                                    aria-hidden="true"
                                  />
                                ) : (
                                  <Save aria-hidden="true" />
                                )}
                                {savingLogin
                                  ? t("systemSite.saving")
                                  : t("systemSite.saveLogin")}
                              </Button>
                            </div>
                          ) : (
                            <p className="text-sm text-muted-foreground">
                              {t("systemSite.readOnly")}
                            </p>
                          )}
                        </form>
                      ) : (
                        <p className="text-sm text-muted-foreground">
                          {t("systemSite.login.noPermission")}
                        </p>
                      )}
                    </CardContent>
                  </Card>
                </div>

                <div id="site-settings-email">
                  <EmailSettingsCard
                    canRead={canLoginRead}
                    canWrite={canLoginWrite}
                    refreshKey={attempt}
                  />
                </div>

                <div id="site-settings-oauth">
                  <OAuthProvidersCard
                    canRead={canLoginRead}
                    canWrite={canLoginWrite}
                    refreshKey={attempt}
                  />
                </div>

                <div id="site-settings-home">
                  <HomeContentCard
                    canRead={canRead}
                    canWrite={canWrite}
                    refreshKey={attempt}
                  />
                </div>
              </>
            ) : null}
          </div>
        </div>
      </div>
    </div>
  )
}
