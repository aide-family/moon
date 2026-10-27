import type { HighlighterCore, LanguageRegistration } from "shiki/core"

/**
 * 代码块语法高亮。
 *
 * 只有同时满足「语种受支持」「内容长度可控」时才走 Shiki：highlighter、两个主题和
 * 每个 grammar 都按需动态 import，首屏和未打开代码块的页面不会付这份体积。
 * light 固定 github-light、dark 固定 github-dark，靠 CSS 变量在 .dark 下切换，
 * 所以切换主题不需要重新高亮，也不会闪一下未着色的文本。
 */
export type CodeLanguage = "yaml" | "json" | "bash" | "log" | "text"

type LoadedLanguage = Exclude<CodeLanguage, "text">

// 超过这个长度就退回纯文本：语法高亮是渐进增强，不能为一段超长日志阻塞主线程。
const MAX_HIGHLIGHT_CHARS = 80_000
const CACHE_LIMIT = 32

const languageAliases: Record<string, CodeLanguage> = {
  yaml: "yaml",
  yml: "yaml",
  json: "json",
  jsonc: "json",
  bash: "bash",
  sh: "bash",
  shell: "bash",
  zsh: "bash",
  log: "log",
  logs: "log",
  text: "text",
  plain: "text",
  plaintext: "text",
  txt: "text",
}

const languageLoaders: Record<
  LoadedLanguage,
  () => Promise<{ default: LanguageRegistration[] }>
> = {
  yaml: () => import("shiki/langs/yaml.mjs"),
  json: () => import("shiki/langs/json.mjs"),
  bash: () => import("shiki/langs/bash.mjs"),
  log: () => import("shiki/langs/log.mjs"),
}

// key → 已高亮的 HTML；只保留最近若干块，避免长时间浏览时无限增长。
const htmlCache = new Map<string, string>()
const languageTasks = new Map<string, Promise<void>>()

let highlighterTask: Promise<HighlighterCore> | null = null

/** 把调用方传入的语言标识归一成受支持的语种，未知语言按纯文本处理。 */
export function normalizeCodeLanguage(language?: string): CodeLanguage {
  if (!language) {
    return "text"
  }
  return languageAliases[language.trim().toLowerCase()] ?? "text"
}

function cacheKey(language: CodeLanguage, code: string): string {
  return `${language}\u0000${code}`
}

function readCache(language: CodeLanguage, code: string): string | null {
  return htmlCache.get(cacheKey(language, code)) ?? null
}

function writeCache(language: CodeLanguage, code: string, html: string) {
  if (htmlCache.size >= CACHE_LIMIT) {
    const oldest = htmlCache.keys().next().value
    if (oldest !== undefined) {
      htmlCache.delete(oldest)
    }
  }
  htmlCache.set(cacheKey(language, code), html)
}

async function createHighlighter(): Promise<HighlighterCore> {
  const [core, engine, githubLight, githubDark] = await Promise.all([
    import("shiki/core"),
    import("shiki/engine/javascript"),
    import("shiki/themes/github-light.mjs"),
    import("shiki/themes/github-dark.mjs"),
  ])
  return core.createHighlighterCore({
    themes: [githubLight.default, githubDark.default],
    langs: [],
    engine: engine.createJavaScriptRegexEngine(),
  })
}

function getHighlighter(): Promise<HighlighterCore> {
  highlighterTask ??= createHighlighter()
  return highlighterTask
}

async function ensureLanguage(
  highlighter: HighlighterCore,
  language: LoadedLanguage
): Promise<void> {
  let task = languageTasks.get(language)
  if (!task) {
    task = (async () => {
      const grammar = await languageLoaders[language]()
      await highlighter.loadLanguage(grammar.default)
    })()
    languageTasks.set(language, task)
    // 失败（例如 grammar 用了引擎不支持的语法）时不要缓存 rejected promise，
    // 下一次调用可以重试；调用方会退回纯文本。
    task.catch(() => languageTasks.delete(language))
  }
  await task
}

/** 读取同步可用的高亮结果，用于已缓存内容的首帧，避免先闪纯文本。 */
export function peekHighlightedHtml(
  code: string,
  language?: string
): string | null {
  if (!code) {
    return ""
  }
  return readCache(normalizeCodeLanguage(language), code)
}

/**
 * 返回可直接放进 <code> 的高亮 HTML（行间用 <br>），纯文本语言或超长内容返回 null，
 * 由调用方按原文渲染。永不抛出：高亮失败只是降级，不影响代码可读性。
 */
export async function highlightCode(
  code: string,
  language?: string
): Promise<string | null> {
  if (!code || code.length > MAX_HIGHLIGHT_CHARS) {
    return null
  }
  const lang = normalizeCodeLanguage(language)
  if (lang === "text") {
    return null
  }
  const cached = readCache(lang, code)
  if (cached !== null) {
    return cached
  }
  try {
    const highlighter = await getHighlighter()
    await ensureLanguage(highlighter, lang)
    const html = highlighter.codeToHtml(code, {
      lang,
      themes: { light: "github-light", dark: "github-dark" },
      defaultColor: false,
      structure: "inline",
    })
    writeCache(lang, code, html)
    return html
  } catch {
    return null
  }
}
