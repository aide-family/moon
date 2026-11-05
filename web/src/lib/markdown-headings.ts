/**
 * Markdown 正文的 h2 / h3 提取与锚点 id 生成。
 *
 * 只扫源码、不引第三方 slug 插件：目录与正文标题共用同一份结果（渲染标题时按源码
 * 行号反查 id），因此不会出现「目录一套算法、标题另一套算法」的漂移。
 *
 * 跳过围栏代码块、缩进代码块与 HTML 注释里的 `#`；标题 id 规则固定为：
 * 转小写 → 去掉标点 → 空白转连字符 → 合并/裁掉首尾连字符。中文按字母保留。
 * 重名标题按 GitHub 习惯追加 `-1`、`-2`。
 */

export interface MarkdownHeading {
  level: 2 | 3
  /** 目录里显示的纯文本（已去掉行内 Markdown 标记）。 */
  text: string
  /** 锚点 id，正文标题与目录共用。 */
  id: string
  /** 1-based 源码行号，渲染标题时用它反查 id。 */
  line: number
}

const ATX_HEADING = /^ {0,3}(#{1,6})(?:[ \t]+(.*))?$/
const FENCE = /^ {0,3}(`{3,}|~{3,})(.*)$/

/**
 * 生成标题锚点 id：小写、空白转连字符、去掉标点，保留中文。
 *
 * 允许字符是 Unicode 字母 / 数字 / 组合标记与连字符；其余标点一律删除。
 */
export function slugifyHeading(text: string): string {
  return text
    .normalize("NFKC")
    .trim()
    .toLocaleLowerCase()
    .replace(/[^\p{L}\p{N}\p{M}\s-]/gu, "")
    .replace(/\s+/g, "-")
    .replace(/-{2,}/g, "-")
    .replace(/^-+|-+$/g, "")
}

/** 去掉行内 Markdown 标记，得到目录里显示的纯文本。 */
export function plainHeadingText(raw: string): string {
  return (
    raw
      // 图片只保留 alt，链接只保留文字，行内代码只保留内容。
      .replace(/!\[([^\]]*)\]\([^)]*\)/g, "$1")
      .replace(/\[([^\]]*)\]\([^)]*\)/g, "$1")
      .replace(/`+([^`]*)`+/g, "$1")
      .replace(/\*\*(.+?)\*\*/g, "$1")
      .replace(/__(.+?)__/g, "$1")
      .replace(/~~(.+?)~~/g, "$1")
      .replace(/\*(.+?)\*/g, "$1")
      .replace(/(^|[\s(])_(.+?)_(?=$|[\s).,])/g, "$1$2")
      .replace(/<[^>]*>/g, "")
      .replace(/\\(.)/g, "$1")
      .replace(/\s+/g, " ")
      .trim()
  )
}

/** 提取正文里的 h2 / h3，按出现顺序返回。 */
export function extractHeadings(markdown: string): MarkdownHeading[] {
  const lines = markdown.replace(/\r\n?/g, "\n").split("\n")
  const headings: MarkdownHeading[] = []
  const used = new Set<string>()
  let fence: { char: string; size: number } | null = null
  let inComment = false

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index]

    // 围栏代码块：开合标记必须同字符、闭合至少与开启等长。
    if (fence) {
      const match = FENCE.exec(line)
      if (
        match &&
        match[1][0] === fence.char &&
        match[1].length >= fence.size &&
        match[2].trim() === ""
      ) {
        fence = null
      }
      continue
    }
    const fenceOpen = FENCE.exec(line)
    if (fenceOpen) {
      fence = { char: fenceOpen[1][0], size: fenceOpen[1].length }
      continue
    }

    // HTML 注释块里的标题也跳过（react-markdown 不渲染 raw HTML）。
    if (inComment) {
      if (line.includes("-->")) {
        inComment = false
      }
      continue
    }
    const commentAt = line.indexOf("<!--")
    if (commentAt !== -1 && !line.slice(commentAt).includes("-->")) {
      inComment = true
      continue
    }

    const match = ATX_HEADING.exec(line)
    if (!match) {
      continue
    }
    const level = match[1].length
    if (level !== 2 && level !== 3) {
      continue
    }
    const text = plainHeadingText(
      (match[2] ?? "").replace(/[ \t]+#+$/, "").trim()
    )
    if (text === "") {
      continue
    }

    const base = slugifyHeading(text) || "section"
    let id = base
    let suffix = 0
    while (used.has(id)) {
      suffix += 1
      id = `${base}-${suffix}`
    }
    used.add(id)
    headings.push({ level, text, id, line: index + 1 })
  }

  return headings
}
