import { useEffect, useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"

import alertRulesIllustration from "@/assets/undraw/alert-rules.svg"
import auditSearchIllustration from "@/assets/undraw/audit-search.svg"
import dashboardIllustration from "@/assets/undraw/dashboard.svg"
import dataSourcesIllustration from "@/assets/undraw/data-sources.svg"
import deployGloballyIllustration from "@/assets/undraw/deploy-globally.svg"
import identityPermissionsIllustration from "@/assets/undraw/identity-permissions.svg"
import notificationsIllustration from "@/assets/undraw/notifications.svg"
import serviceHealthIllustration from "@/assets/undraw/service-health.svg"
import sshWebShellIllustration from "@/assets/undraw/ssh-web-shell.svg"
import teamPermissionsIllustration from "@/assets/undraw/team-permissions.svg"
import { AppearanceMenu } from "@/components/appearance-menu"
import { GitHubIcon } from "@/components/github-icon"
import { LocaleMenu } from "@/components/locale-menu"
import { RepositoryStarLink } from "@/components/repository-star-link"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  type CarouselApi,
} from "@/components/ui/carousel"

const slides = [
  {
    image: dashboardIllustration,
    titleKey: "login.intro.title1",
    descKey: "login.intro.desc1",
  },
  {
    image: teamPermissionsIllustration,
    titleKey: "login.intro.title2",
    descKey: "login.intro.desc2",
  },
  {
    image: deployGloballyIllustration,
    titleKey: "login.intro.title3",
    descKey: "login.intro.desc3",
  },
  {
    image: dataSourcesIllustration,
    titleKey: "login.intro.title4",
    descKey: "login.intro.desc4",
  },
  {
    image: alertRulesIllustration,
    titleKey: "login.intro.title5",
    descKey: "login.intro.desc5",
  },
  {
    image: notificationsIllustration,
    titleKey: "login.intro.title6",
    descKey: "login.intro.desc6",
  },
  {
    image: identityPermissionsIllustration,
    titleKey: "login.intro.title7",
    descKey: "login.intro.desc7",
  },
  {
    image: sshWebShellIllustration,
    titleKey: "login.intro.title8",
    descKey: "login.intro.desc8",
  },
  {
    image: auditSearchIllustration,
    titleKey: "login.intro.title9",
    descKey: "login.intro.desc9",
  },
  {
    image: serviceHealthIllustration,
    titleKey: "login.intro.title10",
    descKey: "login.intro.desc10",
  },
]

function AuthIntro() {
  const { t } = useTranslation()
  const [api, setApi] = useState<CarouselApi>()

  useEffect(() => {
    if (!api) {
      return
    }
    const timer = window.setInterval(() => {
      api.scrollNext()
    }, 5000)
    return () => window.clearInterval(timer)
  }, [api])

  return (
    <aside className="hidden min-h-svh w-[55%] shrink-0 lg:block">
      <Carousel
        setApi={setApi}
        opts={{ loop: true }}
        className="h-svh"
        aria-label={t("login.intro.label")}
      >
        <CarouselContent className="ml-0 h-svh">
          {slides.map((slide) => (
            <CarouselItem key={slide.titleKey} className="pl-0">
              <div className="flex h-svh flex-col items-center justify-center gap-4 px-10">
                <img
                  src={slide.image}
                  alt=""
                  className="h-80 w-full max-w-lg object-contain"
                />
                <h2 className="font-heading text-3xl font-semibold tracking-tight">
                  {t(slide.titleKey)}
                </h2>
                <p className="max-w-md text-center text-lg text-muted-foreground">
                  {t(slide.descKey)}
                </p>
              </div>
            </CarouselItem>
          ))}
        </CarouselContent>
      </Carousel>
    </aside>
  )
}

export function AuthShell({
  title,
  subtitle,
  children,
  repositoryURL,
}: {
  title: string
  subtitle: string
  children: ReactNode
  /** 传入后右上角显示 GitHub 仓库入口与 Star 数量；登录页使用。 */
  repositoryURL?: string
}) {
  const { t } = useTranslation()
  return (
    <main className="relative flex min-h-svh">
      {/* 品牌标识固定在左上角，与右上角的主题/语言/GitHub/Star 入口同一行；
          窄屏只保留图标，内容区多留顶部内边距，避免遮挡表单。 */}
      <div className="absolute inset-x-0 top-0 z-20 flex items-start justify-between gap-3 p-4 sm:p-6">
        <div className="inline-flex items-center gap-2">
          <img
            src="/alioth.svg"
            alt="Alioth"
            className="size-9 rounded-lg sm:size-10"
          />
          <span
            aria-hidden="true"
            className="hidden font-heading text-lg font-semibold tracking-tight sm:inline"
          >
            Alioth
          </span>
        </div>
        <div className="flex items-center gap-2">
          <AppearanceMenu />
          <LocaleMenu />
          {repositoryURL ? (
            <>
              <Button variant="ghost" size="icon" className="shrink-0" asChild>
                <a
                  href={repositoryURL}
                  target="_blank"
                  rel="noreferrer"
                  aria-label={t("shell.github")}
                  title={t("shell.github")}
                >
                  <GitHubIcon />
                </a>
              </Button>
              {/* Star 入口与仓库 icon 同一行；数量拿不到时只显示 Star 文案。 */}
              <RepositoryStarLink repositoryURL={repositoryURL} />
            </>
          ) : null}
        </div>
      </div>
      <AuthIntro />
      <div className="flex min-w-0 flex-1 flex-col items-center justify-center px-4 pt-20 pb-12 sm:px-10">
        <Card className="glass-card w-full max-w-md rounded-2xl border-0 shadow-none">
          <CardHeader className="items-center text-center">
            <CardTitle className="font-heading text-2xl">{title}</CardTitle>
            <p className="text-sm text-muted-foreground">{subtitle}</p>
          </CardHeader>
          <CardContent className="grid gap-4">{children}</CardContent>
        </Card>
      </div>
      {/* 备案号暂时是占位值（见 i18n 的 login.icp），接入真实备案后只需改文案。 */}
      <footer className="absolute inset-x-0 bottom-4 z-20 flex justify-center px-4">
        <a
          href="https://beian.miit.gov.cn/"
          target="_blank"
          rel="noreferrer"
          className="text-xs text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
        >
          {t("login.icp")}
        </a>
      </footer>
    </main>
  )
}
