import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import type { ReactNode } from "react"

import type {
  PublicSitePage,
  SiteConfig,
  SiteHomeContent,
  SiteNavItem,
} from "@/api/site/types"
import { useSiteAccount } from "@/components/site-shell"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/class-names"

import { plainParagraphs } from "./excerpt"
import { FeatureIcon } from "./feature-icons"

/** 版式固定 6 张卡；后端也按 6 张校验。 */
const MAX_FEATURES = 6

/** 后端字段可能整段为空：空字符串表示「用前端默认文案」。 */
const EMPTY_HOME: SiteHomeContent = {
  locale: "",
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
  updatedBy: "",
  updatedAt: "",
}

/** 次 CTA 只允许站内路径或页内锚点，避免首页出现外链跳转。 */
function isInSiteHref(href: string): boolean {
  return (
    (href.startsWith("/") && !href.startsWith("//")) || href.startsWith("#")
  )
}

/** 主 CTA：未登录去注册，已登录调 SiteShell 的进入后台逻辑；Hero 与尾部共用。 */
function PrimaryCta({ label }: { label: string }) {
  const { t } = useTranslation()
  const { loggedIn, entering, enterAdmin, registerHref } = useSiteAccount()
  const className = "h-11 rounded-[10px] px-5"

  if (loggedIn) {
    return (
      <Button className={className} disabled={entering} onClick={enterAdmin}>
        {entering ? t("site.tenant.entering") : label}
      </Button>
    )
  }
  return (
    <Button className={className} asChild>
      <Link to={registerHref}>{label}</Link>
    </Button>
  )
}

function SecondaryCta({ label, href }: { label: string; href: string }) {
  const className = "h-11 rounded-[10px] px-5"
  if (href.startsWith("#")) {
    return (
      <Button variant="outline" className={className} asChild>
        <a href={href}>{label}</a>
      </Button>
    )
  }
  return (
    <Button variant="outline" className={className} asChild>
      <Link to={href}>{label}</Link>
    </Button>
  )
}

/** Hero 主视觉：有图用 4:3 截图框，没图用品牌淡底 + logo 占位，两种都不塌尺寸。 */
function HeroVisual({
  imageUrl,
  logoUrl,
  siteName,
}: {
  imageUrl: string
  logoUrl: string
  siteName: string
}) {
  if (imageUrl) {
    return (
      <div className="overflow-hidden rounded-[22px] border border-border/60 bg-card">
        <img
          src={imageUrl}
          alt=""
          className="aspect-[4/3] w-full object-cover"
        />
      </div>
    )
  }
  return (
    <div
      aria-hidden="true"
      className="grid aspect-[4/3] w-full place-items-center overflow-hidden rounded-[22px] border border-border/60 bg-primary-soft"
    >
      <div className="flex flex-col items-center gap-3 text-primary-ink">
        <img src={logoUrl} alt="" className="size-16 rounded-2xl" />
        <span className="font-mono text-meta tracking-[0.16em] uppercase">
          {siteName}
        </span>
      </div>
    </div>
  )
}

function HeroSection({
  home,
  siteName,
  logoUrl,
}: {
  home: SiteHomeContent
  siteName: string
  logoUrl: string
}) {
  const { t } = useTranslation()
  const heroTitle = home.heroTitle.trim() || t("site.landing.heroTitleDefault")
  const heroSubtitle =
    home.heroSubtitle.trim() || t("site.landing.heroSubtitleDefault")
  const primaryLabel =
    home.ctaPrimaryLabel.trim() || t("site.landing.ctaPrimaryDefault")
  const secondaryLabel = home.ctaSecondaryLabel.trim()
  const secondaryHref = home.ctaSecondaryHref.trim()
  const showSecondary =
    secondaryLabel !== "" && secondaryHref !== "" && isInSiteHref(secondaryHref)

  return (
    <section
      id="site-hero"
      aria-labelledby="site-hero-title"
      className="grid scroll-mt-[88px] items-center gap-10 lg:grid-cols-[minmax(0,7fr)_minmax(0,5fr)] lg:gap-16"
    >
      <div className="min-w-0">
        <p className="font-mono text-meta tracking-[0.16em] text-muted-foreground uppercase">
          {siteName}
        </p>
        <h1 id="site-hero-title" className="mt-4 text-display-1 break-words">
          {heroTitle}
        </h1>
        <p className="mt-5 max-w-[34em] text-lead break-words text-muted-foreground">
          {heroSubtitle}
        </p>
        <div className="mt-8 flex flex-wrap items-center gap-3">
          <PrimaryCta label={primaryLabel} />
          {showSecondary ? (
            <SecondaryCta label={secondaryLabel} href={secondaryHref} />
          ) : null}
        </div>
      </div>
      <HeroVisual
        imageUrl={home.heroImageUrl.trim()}
        logoUrl={logoUrl}
        siteName={siteName}
      />
    </section>
  )
}

function FeaturesSection({
  home,
  features,
}: {
  home: SiteHomeContent
  features: SiteConfig["features"]
}) {
  const { t } = useTranslation()
  if (features.length === 0) {
    return null
  }
  const title =
    home.featuresTitle.trim() || t("site.landing.featuresTitleDefault")
  const subtitle = home.featuresSubtitle.trim()

  return (
    <section
      id="site-features"
      aria-labelledby="site-features-title"
      className="scroll-mt-[88px]"
    >
      <h2 id="site-features-title" className="text-section">
        {title}
      </h2>
      {subtitle ? (
        <p className="mt-3 max-w-[52ch] text-lead text-muted-foreground">
          {subtitle}
        </p>
      ) : null}
      <ul className="mt-8 grid auto-rows-fr gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {features.slice(0, MAX_FEATURES).map((feature) => (
          <li
            key={feature.uid}
            className="min-w-0 rounded-[18px] border border-border/60 bg-card p-6"
          >
            <span
              aria-hidden="true"
              className="grid size-10 place-items-center rounded-[10px] bg-primary-soft text-primary-ink"
            >
              <FeatureIcon iconKey={feature.iconKey} />
            </span>
            <p className="mt-4 font-heading text-base font-semibold break-words">
              {feature.title}
            </p>
            <p className="mt-2 text-body break-words text-muted-foreground">
              {feature.description}
            </p>
          </li>
        ))}
      </ul>
    </section>
  )
}

/** ③④⑤ 共用骨架：左侧 5 列标题与说明，右侧 7 列固定结构；移动单列。 */
function SplitSection({
  id,
  title,
  body,
  children,
}: {
  id: string
  title: string
  body: string
  children: ReactNode
}) {
  return (
    <section
      id={id}
      aria-labelledby={`${id}-title`}
      className="grid scroll-mt-[88px] gap-6 border-t border-border/60 pt-10 md:pt-12 lg:grid-cols-12 lg:gap-12"
    >
      <div className="min-w-0 lg:col-span-5">
        <h2 id={`${id}-title`} className="text-section">
          {title}
        </h2>
        <p className="mt-4 max-w-[38ch] text-body break-words text-muted-foreground">
          {body}
        </p>
      </div>
      <div className="min-w-0 lg:col-span-7">{children}</div>
    </section>
  )
}

function HowSection({ home }: { home: SiteHomeContent }) {
  const { t } = useTranslation()
  const steps = [
    {
      title: t("site.landing.steps.collect.title"),
      description: t("site.landing.steps.collect.description"),
    },
    {
      title: t("site.landing.steps.evaluate.title"),
      description: t("site.landing.steps.evaluate.description"),
    },
    {
      title: t("site.landing.steps.notify.title"),
      description: t("site.landing.steps.notify.description"),
    },
  ]
  return (
    <SplitSection
      id="site-how"
      title={home.howTitle.trim() || t("site.landing.howTitleDefault")}
      body={home.howBody.trim() || t("site.landing.howBodyDefault")}
    >
      <ol className="grid gap-3">
        {steps.map((step, index) => (
          <li
            key={step.title}
            className="flex min-w-0 gap-4 rounded-[18px] border border-border/60 bg-card p-4 sm:p-5"
          >
            <span
              aria-hidden="true"
              className="font-mono text-sm text-primary-ink tabular-nums"
            >
              {String(index + 1).padStart(2, "0")}
            </span>
            <div className="min-w-0">
              <p className="font-heading text-base font-semibold">
                {step.title}
              </p>
              <p className="mt-1 text-body break-words text-muted-foreground">
                {step.description}
              </p>
            </div>
          </li>
        ))}
      </ol>
    </SplitSection>
  )
}

function PermSection({ home }: { home: SiteHomeContent }) {
  const { t } = useTranslation()
  const columns = [
    {
      label: t("site.landing.permTenantLabel"),
      items: [
        t("site.landing.permTenant.roles"),
        t("site.landing.permTenant.members"),
        t("site.landing.permTenant.isolation"),
      ],
    },
    {
      label: t("site.landing.permPlatformLabel"),
      items: [
        t("site.landing.permPlatform.systemPerms"),
        t("site.landing.permPlatform.siteLogin"),
        t("site.landing.permPlatform.superAdmin"),
      ],
    },
  ]
  return (
    <SplitSection
      id="site-permissions"
      title={home.permTitle.trim() || t("site.landing.permTitleDefault")}
      body={home.permBody.trim() || t("site.landing.permBodyDefault")}
    >
      <div className="grid gap-4 sm:grid-cols-2">
        {columns.map((column) => (
          <div
            key={column.label}
            className="min-w-0 rounded-[18px] border border-border/60 bg-card p-5"
          >
            <p className="font-mono text-meta tracking-[0.12em] text-muted-foreground uppercase">
              {column.label}
            </p>
            <ul className="mt-3 grid gap-2.5">
              {column.items.map((item) => (
                <li
                  key={item}
                  className="flex min-w-0 items-center gap-2.5 text-body break-words"
                >
                  <span
                    aria-hidden="true"
                    className="size-1.5 shrink-0 rounded-full bg-primary-ink"
                  />
                  {item}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </SplitSection>
  )
}

function DeploySection({ home }: { home: SiteHomeContent }) {
  const { t } = useTranslation()
  // 措辞按仓库 README / Makefile 事实：四个 Go 服务各自构建、compose 起全栈、
  // 数据保存在自建 PostgreSQL；没有「单二进制」这种交付方式。
  const chips = [
    t("site.landing.deployChips.services"),
    t("site.landing.deployChips.compose"),
    t("site.landing.deployChips.database"),
  ]
  return (
    <SplitSection
      id="site-deploy"
      title={home.deployTitle.trim() || t("site.landing.deployTitleDefault")}
      body={home.deployBody.trim() || t("site.landing.deployBodyDefault")}
    >
      <ul className="flex flex-wrap gap-2.5">
        {chips.map((chip) => (
          <li
            key={chip}
            className="flex items-center gap-2 rounded-[10px] border border-border bg-card px-3.5 py-2 text-sm"
          >
            <span
              aria-hidden="true"
              className="size-1.5 shrink-0 rounded-full bg-primary-ink"
            />
            {chip}
          </li>
        ))}
      </ul>
    </SplitSection>
  )
}

function DocNav({ items, label }: { items: SiteNavItem[]; label: string }) {
  return (
    <nav aria-label={label} className="min-w-0">
      <p className="font-mono text-meta tracking-[0.12em] text-muted-foreground uppercase">
        {label}
      </p>
      <ul className="mt-3">
        {items.map((item) => (
          <li
            key={item.slug}
            className="border-b border-border/60 last:border-b-0"
          >
            <Link
              to={`/p/${encodeURIComponent(item.slug)}`}
              className="flex min-w-0 items-center justify-between gap-3 py-3 text-body break-words text-foreground underline-offset-4 hover:text-primary-ink hover:underline focus-visible:ring-3 focus-visible:ring-ring focus-visible:outline-none"
            >
              <span className="min-w-0 break-words">
                {item.title || item.slug}
              </span>
              <span
                aria-hidden="true"
                className="font-mono text-xs text-muted-foreground"
              >
                →
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  )
}

function DocSection({
  home,
  page,
  navItems,
}: {
  home: SiteHomeContent
  page: PublicSitePage | null
  navItems: SiteNavItem[]
}) {
  const { t } = useTranslation()
  const paragraphs = plainParagraphs(page?.bodyMd ?? "")
  const hasBody = paragraphs.length > 0
  const hasNav = navItems.length > 0
  if (!hasBody && !hasNav) {
    return null
  }
  const title = home.docTitle.trim() || t("site.landing.docTitleDefault")
  const navLabel = t("site.landing.navLabel")

  return (
    <section
      id="site-docs"
      aria-labelledby="site-docs-title"
      className={cn(
        "grid scroll-mt-[88px] gap-8",
        hasBody && hasNav
          ? "lg:grid-cols-[minmax(0,7fr)_minmax(0,5fr)] lg:gap-16"
          : ""
      )}
    >
      <div className="min-w-0">
        <h2 id="site-docs-title" className="text-section">
          {title}
        </h2>
        {hasBody ? (
          <div className="mt-4 grid max-w-[68ch] gap-4">
            {paragraphs.map((paragraph, index) => (
              <p
                key={`${index}-${paragraph.slice(0, 24)}`}
                className="text-body break-words text-muted-foreground"
              >
                {paragraph}
              </p>
            ))}
          </div>
        ) : null}
        {!hasBody && hasNav ? (
          <div className="mt-4 max-w-[68ch]">
            <DocNav items={navItems} label={navLabel} />
          </div>
        ) : null}
      </div>
      {hasBody && hasNav ? <DocNav items={navItems} label={navLabel} /> : null}
    </section>
  )
}

function ClosingSection({ home }: { home: SiteHomeContent }) {
  const { t } = useTranslation()
  const title =
    home.closingTitle.trim() || t("site.landing.closingTitleDefault")
  const description =
    home.closingDescription.trim() ||
    t("site.landing.closingDescriptionDefault")
  const label =
    home.closingButtonLabel.trim() ||
    home.ctaPrimaryLabel.trim() ||
    t("site.landing.ctaPrimaryDefault")

  return (
    <section
      id="site-closing"
      aria-labelledby="site-closing-title"
      className="scroll-mt-[88px]"
    >
      <div className="rounded-[18px] border border-border bg-card p-6 md:p-8 lg:p-12">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between lg:gap-12">
          <div className="min-w-0">
            <h2 id="site-closing-title" className="text-section">
              {title}
            </h2>
            {description ? (
              <p className="mt-3 max-w-[52ch] text-body break-words text-muted-foreground">
                {description}
              </p>
            ) : null}
          </div>
          <div className="shrink-0">
            <PrimaryCta label={label} />
          </div>
        </div>
      </div>
    </section>
  )
}

/**
 * 官网首页的固定版式。
 *
 * 区块顺序写死在组件里，后台只能填短文本；空字段一律回落到默认文案，能力卡片
 * 为空时整块（含 h2）不渲染。版式间距只由最外层容器的 row-gap 控制。
 */
export function SiteHomeSections({
  config,
  page,
  navItems,
}: {
  config: SiteConfig
  page: PublicSitePage | null
  navItems: SiteNavItem[]
}) {
  const { t } = useTranslation()
  const home = { ...EMPTY_HOME, ...config.home }
  const siteName = config.site.siteName.trim() || t("site.brandFallback")
  const logoUrl = config.site.logoUrl.trim() || "/alioth.svg"
  const features = config.features ?? []

  return (
    <div className="relative">
      <a
        href="#site-home-main"
        className="absolute top-3 left-3 z-50 -translate-y-24 rounded-[10px] border border-border bg-card px-3 py-2 text-sm text-foreground focus:translate-y-0 focus:ring-3 focus:ring-ring focus:outline-none"
      >
        {t("site.landing.skipToContent")}
      </a>
      <div
        id="site-home-main"
        tabIndex={-1}
        className="site-container site-enter grid gap-[var(--site-section-gap)] pt-10 pb-16 focus:outline-none md:pt-14 md:pb-20 lg:pt-16 lg:pb-24"
      >
        <HeroSection home={home} siteName={siteName} logoUrl={logoUrl} />
        <FeaturesSection home={home} features={features} />
        <HowSection home={home} />
        <PermSection home={home} />
        <DeploySection home={home} />
        <DocSection home={home} page={page} navItems={navItems} />
        <ClosingSection home={home} />
      </div>
    </div>
  )
}
