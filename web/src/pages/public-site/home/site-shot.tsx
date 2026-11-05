import type { ReactNode } from "react"

import { cn } from "@/lib/class-names"

/**
 * 官网用的产品截图。每张都准备了浅色与深色两份：
 * 站点的深色是手动切换的 class（html.dark），所以用 `dark:` 变体切换图片，
 * 而不是 `<picture media="prefers-color-scheme">`——访客的主题选择与系统偏好是两回事。
 */
const SHOTS = {
  overview: "/site/overview",
  "alert-rules": "/site/alert-rules",
  permissions: "/site/permissions",
  workers: "/site/workers",
} as const

export type ShotName = keyof typeof SHOTS

/** 截图外框：固定 16:10 槽位 + 圆角描边，深浅色都由语义 token 决定。 */
export function ShotFrame({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        "aspect-[16/10] w-full overflow-hidden rounded-[22px] border border-border bg-card p-2",
        className
      )}
    >
      <div className="size-full overflow-hidden rounded-[16px]">{children}</div>
    </div>
  )
}

export function SiteShot({
  name,
  priority = false,
  className,
}: {
  name: ShotName
  priority?: boolean
  className?: string
}) {
  const base = SHOTS[name]
  const loading = priority ? "eager" : "lazy"
  return (
    <ShotFrame className={className}>
      <img
        src={`${base}-light.webp`}
        alt=""
        width={1600}
        height={1000}
        loading={loading}
        fetchPriority={priority ? "high" : "auto"}
        decoding="async"
        className="size-full object-cover dark:hidden"
      />
      <img
        src={`${base}-dark.webp`}
        alt=""
        width={1600}
        height={1000}
        loading={loading}
        fetchPriority={priority ? "high" : "auto"}
        decoding="async"
        className="hidden size-full object-cover dark:block"
      />
    </ShotFrame>
  )
}
