import { useEffect, useState } from "react"

import {
  highlightCode,
  normalizeCodeLanguage,
  peekHighlightedHtml,
} from "@/lib/code-highlight"

/**
 * 按需高亮一段代码，返回可注入 <code> 的 HTML；未完成或不可高亮时为 null。
 *
 * 同步命中的缓存直接参与渲染（不在 effect 里 setState），异步高亮结果按
 * 「语言 + 原文」键写入状态，输入变化时旧结果自然失效，不会出现错位闪烁。
 * 主题切换由 CSS 处理，不参与这里的依赖。
 */
export function useCodeHighlight(
  code: string,
  language?: string
): string | null {
  const normalized = normalizeCodeLanguage(language)
  const key = `${normalized}\u0000${code}`
  const cached = peekHighlightedHtml(code, normalized)
  const [resolved, setResolved] = useState<{
    key: string
    html: string | null
  }>(() => ({ key, html: cached }))

  useEffect(() => {
    if (cached !== null) {
      return
    }
    let active = true
    void highlightCode(code, normalized).then((next) => {
      if (active) {
        setResolved({ key, html: next })
      }
    })
    return () => {
      active = false
    }
  }, [code, normalized, key, cached])

  if (cached !== null) {
    return cached
  }
  return resolved.key === key ? resolved.html : null
}
