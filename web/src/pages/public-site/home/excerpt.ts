/** 产品说明摘要：最多 3 段、总共 480 字符。 */
const MAX_PARAGRAPHS = 3
const MAX_CHARS = 480

/**
 * 去掉一段文本里常见的 Markdown 标记，保留可读的纯文本。
 *
 * 标题行整行丢弃：首页摘要区自带 docTitle 标题，再把正文里的 `# 产品说明` 当段落
 * 渲染会重复一次标题。列表/引用只保留文字，不解析行内 Markdown。
 */
function stripMarkdown(block: string): string {
  const body = block
    .split("\n")
    .filter((line) => !/^\s{0,3}#{1,6}\s+/.test(line))
    .join("\n")
  return body
    .replace(/!\[[^\]]*\]\([^)]*\)/g, " ")
    .replace(/\[([^\]]+)\]\([^)]*\)/g, "$1")
    .replace(/^\s{0,3}>\s?/gm, "")
    .replace(/^\s{0,3}(?:[-*+]|\d+\.)\s+/gm, "")
    .replace(/[*_~`]/g, "")
    .replace(/\s+/g, " ")
    .trim()
}

/** 截断到 max 个字符；英文在空格处收口，中文按字截断。 */
function clip(chars: string[], max: number): string {
  const head = chars.slice(0, max).join("")
  const lastSpace = head.lastIndexOf(" ")
  const base = lastSpace > max * 0.6 ? head.slice(0, lastSpace) : head
  return `${base.trimEnd()}…`
}

/**
 * 官网首页的「产品说明摘要」：把 home 页正文压成纯文本段落。
 *
 * 版式只留 3 段、480 字符，所以这里在进入版式前就截断，避免把整篇文档塞进首页。
 * 代码块整体丢弃，标题/列表/引用只保留文字，不解析行内 Markdown。
 */
export function plainParagraphs(bodyMd: string): string[] {
  const text = bodyMd
    .replace(/\r\n?/g, "\n")
    .replace(/```[\s\S]*?```/g, "\n\n")
    .replace(/~~~[\s\S]*?~~~/g, "\n\n")

  const paragraphs: string[] = []
  let used = 0
  for (const block of text.split(/\n{2,}/)) {
    if (paragraphs.length >= MAX_PARAGRAPHS || used >= MAX_CHARS) {
      break
    }
    const plain = stripMarkdown(block)
    if (!plain) {
      continue
    }
    const remaining = MAX_CHARS - used
    const chars = Array.from(plain)
    const piece = (
      chars.length > remaining ? clip(chars, remaining) : plain
    ).trim()
    if (!piece) {
      continue
    }
    paragraphs.push(piece)
    used += Array.from(piece).length
  }
  return paragraphs
}
