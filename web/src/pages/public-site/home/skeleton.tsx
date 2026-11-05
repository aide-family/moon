import { useTranslation } from "react-i18next"

import { Skeleton } from "@/components/ui/skeleton"

import { ShotFrame } from "./site-shot"

type PanelVariant = "hero" | "flow" | "permission" | "deploy"

/** 首页加载骨架：复刻固定版式的容器、区块节奏与 1/2/3 列网格，避免内容到位后跳动。 */
function PanelPlaceholder({ variant }: { variant: PanelVariant }) {
  if (variant === "hero") {
    return (
      <div className="flex h-full flex-col">
        <div className="flex h-6 shrink-0 items-center gap-2 border-b border-border/60 bg-card/60 px-2 sm:h-8 sm:px-3">
          <Skeleton className="size-1.5 rounded-full sm:size-2" />
          <Skeleton className="size-1.5 rounded-full sm:size-2" />
          <Skeleton className="size-1.5 rounded-full sm:size-2" />
          <Skeleton className="h-2.5 w-24 max-w-[40%] rounded-[4px] sm:h-3.5" />
          <Skeleton className="ml-auto h-2 w-10 rounded-full" />
        </div>
        <div className="flex min-h-0 flex-1">
          <div className="hidden w-[18%] shrink-0 border-r border-border/60 bg-muted/20 p-2 sm:block sm:p-2.5">
            <div className="grid gap-1.5">
              <Skeleton className="size-4 rounded-[5px]" />
              <Skeleton className="h-2 w-full rounded-full" />
              <Skeleton className="h-2 w-4/5 rounded-full" />
              <Skeleton className="h-2 w-3/5 rounded-full" />
              <Skeleton className="h-2 w-4/5 rounded-full" />
            </div>
          </div>
          <div className="grid min-h-0 flex-1 grid-rows-[auto_minmax(0,1fr)] gap-1.5 p-2 sm:gap-3 sm:p-3">
            <div className="grid grid-cols-3 gap-1.5 sm:gap-2">
              {[0, 1, 2].map((index) => (
                <Skeleton
                  key={index}
                  className="h-8 rounded-[6px] sm:h-11 sm:rounded-[8px]"
                />
              ))}
            </div>
            <div className="grid min-h-0 grid-cols-5 gap-1.5 sm:gap-3">
              <Skeleton className="col-span-3 rounded-[6px] sm:rounded-[8px]" />
              <Skeleton className="col-span-2 rounded-[6px] sm:rounded-[8px]" />
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (variant === "flow") {
    return (
      <div className="flex h-full items-center justify-center gap-1.5 p-2.5 sm:gap-4 sm:p-5">
        {[0, 1, 2].map((index) => (
          <Skeleton
            key={index}
            className="h-[68%] flex-1 rounded-[10px] bg-muted/70"
          />
        ))}
      </div>
    )
  }

  if (variant === "permission") {
    return (
      <div className="grid h-full grid-cols-2 gap-2 p-2.5 sm:gap-4 sm:p-5">
        {[0, 1].map((column) => (
          <div
            key={column}
            className="flex min-h-0 flex-col gap-2 rounded-[10px] border border-border/60 bg-card/50 p-2 sm:gap-3 sm:p-3"
          >
            <Skeleton className="h-2 w-14 rounded-full" />
            {[0, 1, 2].map((row) => (
              <Skeleton key={row} className="h-6 rounded-[6px] sm:h-8" />
            ))}
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="grid h-full grid-cols-3 content-center gap-2 p-3 sm:gap-4 sm:p-5">
      {[0, 1, 2, 3, 4].map((index) => (
        <Skeleton key={index} className="h-10 rounded-[8px] sm:h-14" />
      ))}
    </div>
  )
}

/** 首页加载骨架：Hero 与三个解释区块的面板占位与正式面板同尺寸同外框。 */
export function HomeSkeleton() {
  const { t } = useTranslation()
  return (
    <div
      className="site-container grid gap-[var(--site-section-gap)] pt-10 pb-16 md:pt-14 md:pb-20 lg:pt-16 lg:pb-24"
      aria-busy="true"
      aria-label={t("site.loading")}
    >
      <div className="flex flex-col items-center text-center">
        <Skeleton className="h-3 w-24" />
        <div className="mt-4 grid w-full justify-items-center gap-3">
          <Skeleton className="h-[44px] w-full max-w-xl md:h-[56px] lg:h-[72px]" />
          <Skeleton className="h-[44px] w-3/4 max-w-md md:h-[56px] lg:h-[72px]" />
        </div>
        <Skeleton className="mt-5 h-5 w-full max-w-lg" />
        <Skeleton className="mt-2 h-5 w-2/3 max-w-md" />
        <div className="mt-8 flex gap-3">
          <Skeleton className="h-11 w-32 rounded-[10px]" />
          <Skeleton className="h-11 w-28 rounded-[10px]" />
        </div>
        <ShotFrame className="mt-12 w-full sm:mt-16">
          <PanelPlaceholder variant="hero" />
        </ShotFrame>
      </div>

      <div className="grid gap-6">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-5 w-full max-w-md" />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[0, 1, 2].map((index) => (
            <Skeleton key={index} className="h-56 rounded-[18px]" />
          ))}
        </div>
      </div>

      {(["flow", "permission", "deploy"] as const).map((variant) => (
        <div
          key={variant}
          className="grid gap-8 border-t border-border/60 pt-10 md:pt-12 lg:grid-cols-12 lg:gap-12"
        >
          <div className="grid content-start gap-4 lg:col-span-5">
            <Skeleton className="h-8 w-44" />
            <Skeleton className="h-5 w-full max-w-sm" />
            <Skeleton className="h-5 w-2/3 max-w-xs" />
            <div className="mt-2 grid gap-3">
              <Skeleton className="h-20 rounded-[18px]" />
              <Skeleton className="h-20 rounded-[18px]" />
              <Skeleton className="h-20 rounded-[18px]" />
            </div>
          </div>
          <ShotFrame className="lg:col-span-7 lg:self-start">
            <PanelPlaceholder variant={variant} />
          </ShotFrame>
        </div>
      ))}

      <div className="grid gap-8 lg:grid-cols-[minmax(0,7fr)_minmax(0,5fr)] lg:gap-16">
        <div className="grid content-start gap-4">
          <Skeleton className="h-8 w-40" />
          <Skeleton className="h-5 w-full max-w-xl" />
          <Skeleton className="h-5 w-full max-w-xl" />
          <Skeleton className="h-5 w-2/3 max-w-md" />
        </div>
        <div className="grid content-start gap-3">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      </div>

      <Skeleton className="h-44 rounded-[18px]" />
    </div>
  )
}
