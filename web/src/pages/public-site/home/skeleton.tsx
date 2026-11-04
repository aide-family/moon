import { useTranslation } from "react-i18next"

import { Skeleton } from "@/components/ui/skeleton"

/**
 * 首页加载骨架：复刻固定版式的容器、区块节奏与 1/2/3 列网格，避免内容到位后跳动。
 *
 * Hero 是「居中文字 + 大截图」的堆叠版式，三个解释区块是「左文字结构 + 右截图」，
 * 骨架按同样的占位块高度画，加载完成时不会整页位移。
 */
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
        <Skeleton className="mt-12 aspect-[16/10] w-full rounded-[22px] sm:mt-16" />
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

      {[0, 1, 2].map((index) => (
        <div
          key={index}
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
          <Skeleton className="aspect-[16/10] rounded-[22px] lg:col-span-7 lg:self-start" />
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
