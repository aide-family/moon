import { useEffect, useState, type MouseEvent } from "react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/class-names"
import type { MarkdownHeading } from "@/lib/markdown-headings"

/** 吸顶页头约 57px；标题用 scroll-mt-24（96px）留白，高亮阈值与它对齐。 */
const ACTIVE_OFFSET = 96

function isEnglish(language: string): boolean {
  return language.startsWith("en")
}

/** 平滑滚动到标题；scroll-mt-24 负责避开吸顶页头。 */
function scrollToHeading(event: MouseEvent<HTMLAnchorElement>, id: string) {
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
  const element = document.getElementById(id)
  if (!element) {
    return
  }
  event.preventDefault()
  const reduceMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)"
  ).matches
  element.scrollIntoView({
    behavior: reduceMotion ? "auto" : "smooth",
    block: "start",
  })
  window.history.replaceState(window.history.state, "", `#${id}`)
}

/**
 * 滚动高亮：取最后一个已经越过吸顶页头阈值的标题；滚到文档底部时高亮最后一节。
 *
 * 用 window 滚动而不是 IntersectionObserver，因为正文页的滚动主人就是文档，
 * 且高亮语义是「当前读到哪一节」，需要按标题顺序而不是可见性判断。
 */
function useActiveHeading(headings: readonly MarkdownHeading[]): string {
  const [activeId, setActiveId] = useState("")

  useEffect(() => {
    if (headings.length === 0) {
      return
    }
    let frame = 0
    const update = () => {
      frame = 0
      let next = ""
      for (const heading of headings) {
        const element = document.getElementById(heading.id)
        if (!element) {
          continue
        }
        if (element.getBoundingClientRect().top <= ACTIVE_OFFSET + 8) {
          next = heading.id
        } else {
          break
        }
      }
      const atBottom =
        document.documentElement.scrollHeight > window.innerHeight + 2 &&
        window.innerHeight + window.scrollY >=
          document.documentElement.scrollHeight - 2
      if (atBottom) {
        next = headings[headings.length - 1].id
      }
      setActiveId(next)
    }
    const onScroll = () => {
      if (frame === 0) {
        frame = requestAnimationFrame(update)
      }
    }

    onScroll()
    window.addEventListener("scroll", onScroll, { passive: true })
    window.addEventListener("resize", onScroll)
    return () => {
      window.removeEventListener("scroll", onScroll)
      window.removeEventListener("resize", onScroll)
      if (frame !== 0) {
        cancelAnimationFrame(frame)
      }
    }
  }, [headings])

  return activeId
}

interface SiteTocProps {
  headings: readonly MarkdownHeading[]
}

/**
 * 文档页右侧目录（lg 及以上显示，1024px 起）。
 *
 * 没有 h2/h3 时由调用方整列不渲染；这里只负责列表、当前项高亮与键盘焦点环。
 * lg 以下不显示，避免正文被横向挤压（窄屏仍可按标题层级顺序阅读正文）。
 */
export function SiteToc({ headings }: SiteTocProps) {
  const { t, i18n } = useTranslation()
  const activeId = useActiveHeading(headings)
  const label = t("site.tocTitle", {
    defaultValue: isEnglish(i18n.resolvedLanguage ?? i18n.language ?? "")
      ? "On this page"
      : "本页目录",
  })

  return (
    <nav
      aria-label={label}
      className="hidden px-1 lg:sticky lg:top-24 lg:block lg:max-h-[calc(100svh-8rem)] lg:self-start lg:overflow-y-auto"
    >
      <p className="font-mono text-meta tracking-[0.12em] text-muted-foreground uppercase">
        {label}
      </p>
      <ul className="mt-3 border-l border-border/70">
        {headings.map((heading) => {
          const current = heading.id === activeId
          return (
            <li key={heading.id}>
              <a
                href={`#${heading.id}`}
                aria-current={current ? "location" : undefined}
                onClick={(event) => scrollToHeading(event, heading.id)}
                className={cn(
                  "-ml-px block border-l-2 py-1.5 pr-2 pl-3 text-sm leading-6 break-words transition-colors outline-none focus-visible:ring-3 focus-visible:ring-ring",
                  heading.level === 3 && "pl-6",
                  current
                    ? "border-primary-ink font-medium text-foreground"
                    : "border-transparent text-muted-foreground hover:text-foreground"
                )}
              >
                {heading.text}
              </a>
            </li>
          )
        })}
      </ul>
    </nav>
  )
}
