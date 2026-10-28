#!/usr/bin/env node
/**
 * 用生成产物重建 docs/theme-preview.html 的动态部分：
 *   - 顶部品牌色卡按钮
 *   - 每个品牌一块的「浅色 / 深色」面板（含内联的 --brand-* 变量）
 *   - 标题里的品牌数量
 *
 * 静态预览页要能直接双击打开（file://），所以变量和面板只能内联；
 * 生成器一跑它就会过期。面板样板取自文件里已有的第一块 <section class="brand">，
 * 因此新增品牌不需要手写 400 行 HTML。
 *
 *   node scripts/sync-theme-preview.mjs          # 写入
 *   node scripts/sync-theme-preview.mjs --check  # 过期则退出码 1
 */
import { readFileSync, writeFileSync } from "node:fs"
import { dirname, resolve } from "node:path"
import { fileURLToPath } from "node:url"

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..")
const CSS = resolve(ROOT, "src/styles/theme/tokens.brand.css")
const TS = resolve(ROOT, "src/lib/brand-palette.generated.ts")
const PREVIEW = resolve(ROOT, "../docs/theme-preview.html")

const css = readFileSync(CSS, "utf8")
const palette = readFileSync(TS, "utf8")
const html = readFileSync(PREVIEW, "utf8")

/** 品牌元信息与顺序以 generated 注册表为准。 */
const brands = [
  ...palette.matchAll(
    /^ {2}(\w+): \{\n {4}id: "([^"]+)",\n {4}labelZh: "([^"]+)",\n {4}labelEn: "([^"]+)",\n {4}base: "([^"]+)",/gm
  ),
].map(([, , id, zh, en, base]) => ({ id, zh, en, base }))

if (brands.length === 0) {
  console.error("没有从 brand-palette.generated.ts 解析出品牌")
  process.exit(1)
}

/** id → 该品牌全部 --brand-* 声明（含浅/深两套角色色）。 */
const declarations = new Map()
for (const match of css.matchAll(/:root\[data-brand="([^"]+)"\]\s*\{([\s\S]*?)\n\}/g)) {
  declarations.set(
    match[1],
    match[2]
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line.startsWith("--brand-"))
      .join("")
  )
}

const failures = []
const sections = html.match(/<section class="brand">[\s\S]*?<\/section>/g)
if (!sections || sections.length === 0) {
  failures.push('预览页里找不到 <section class="brand"> 样板')
}
const template = sections?.[0] ?? ""

if (template && !/data-brand="[^"]+"/.test(template)) {
  failures.push("样板 section 里缺少 data-brand 标记")
}
if (template && (template.match(/style="--brand-50:/g) ?? []).length !== 2) {
  failures.push("样板 section 里应当有两个内联变量块（浅色 / 深色）")
}

if (failures.length > 0) {
  for (const failure of failures) console.error(`  - ${failure}`)
  process.exit(1)
}

const rendered = brands.map((brand) => {
  const decl = declarations.get(brand.id)
  if (!decl) {
    failures.push(`tokens.brand.css 缺少品牌：${brand.id}`)
    return ""
  }
  return template
    .replace(/<span class="brand-name">[^<]*<\/span>/, `<span class="brand-name">${brand.zh} ${brand.en}</span>`)
    .replace(/<code>#[0-9A-Fa-f]{6}<\/code>/, `<code>${brand.base.toUpperCase()}</code>`)
    .replace(/data-brand="[^"]+"/, `data-brand="${brand.id}"`)
    .replace(/style="--brand-50:[^"]*"/g, `style="${decl}"`)
})

if (failures.length > 0) {
  for (const failure of failures) console.error(`  - ${failure}`)
  process.exit(1)
}

const buttons = brands
  .map((brand, index) => {
    const fill = declarations.get(brand.id)?.match(/--brand-fill:\s*(#[0-9a-f]{6});/)?.[1] ?? brand.base
    return `<button role="radio" aria-checked="${String(index === 0)}" data-brand="${brand.id}"><i style="background:${fill}"></i>${brand.zh} ${brand.en}</button>`
  })
  .join("")

let next = html
  .replace(
    /(<div class="swatches" role="radiogroup" aria-label="品牌">)[\s\S]*?(<\/div>)/,
    `$1${buttons}$2`
  )
  .replace(/<section class="brand">[\s\S]*<\/section>/, rendered.join("\n  "))
  .replace(/\d+ 品牌 × 浅\/深/g, `${brands.length} 品牌 × 浅/深`)

const check = process.argv.includes("--check")
if (check) {
  if (next !== html) {
    console.error("docs/theme-preview.html 已过期，请运行 npm run tokens:gen")
    process.exit(1)
  }
  console.log(`主题预览页已同步（${brands.length} 个品牌）。`)
} else {
  writeFileSync(PREVIEW, next)
  console.log(`已同步 docs/theme-preview.html（${brands.length} 个品牌）`)
}
