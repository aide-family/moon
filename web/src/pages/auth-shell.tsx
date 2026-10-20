import { useEffect, useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"

import banner1 from "@/assets/banner/banner1.svg"
import banner2 from "@/assets/banner/banner2.svg"
import banner3 from "@/assets/banner/banner3.svg"
import { LocaleThemeMenu } from "@/components/locale-theme-menu"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  type CarouselApi,
} from "@/components/ui/carousel"

const slides = [
  {
    image: banner1,
    titleKey: "login.intro.title1",
    descKey: "login.intro.desc1",
  },
  {
    image: banner2,
    titleKey: "login.intro.title2",
    descKey: "login.intro.desc2",
  },
  {
    image: banner3,
    titleKey: "login.intro.title3",
    descKey: "login.intro.desc3",
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
}: {
  title: string
  subtitle: string
  children: ReactNode
}) {
  return (
    <main className="relative flex min-h-svh">
      <div className="fixed top-4 right-4 z-10">
        <LocaleThemeMenu />
      </div>
      <AuthIntro />
      <div className="flex min-w-0 flex-1 flex-col items-center justify-center px-4 py-16 sm:px-10">
        <Card className="glass-card w-full max-w-md rounded-2xl border-0 shadow-none">
          <CardHeader className="items-center text-center">
            <img
              src="/alioth.svg"
              alt="Alioth"
              className="mx-auto mb-2 size-12 rounded-xl"
            />
            <CardTitle className="font-heading text-2xl">{title}</CardTitle>
            <p className="text-sm text-muted-foreground">{subtitle}</p>
          </CardHeader>
          <CardContent className="grid gap-4">{children}</CardContent>
        </Card>
      </div>
    </main>
  )
}
