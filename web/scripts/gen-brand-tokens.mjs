#!/usr/bin/env node
/**
 * 品牌主题 token 生成器。
 *
 * 单一事实来源是下面 BRANDS 里的 7 个品牌基色。脚本用 OKLCH 生成 50–950 色阶
 * （色相固定、明度按固定刻度、彩度按品牌锚定后受 sRGB 色域钳制），再按 WCAG
 * 对比度为浅色/深色两种模式各挑出语义角色色（填充、填充前景、可读文字色、环色）。
 *
 * 用法：
 *   node scripts/gen-brand-tokens.mjs            # 只打印校验报告
 *   node scripts/gen-brand-tokens.mjs --write    # 写入 tokens.brand.css 与 brand-palette.generated.ts
 *   node scripts/gen-brand-tokens.mjs --check    # CI 用：产物过期即退出码 1
 */
import { mkdirSync, readFileSync, writeFileSync } from "node:fs"
import { dirname, resolve } from "node:path"
import { fileURLToPath } from "node:url"

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..")
const CSS_OUT = resolve(ROOT, "src/styles/theme/tokens.brand.css")
const TS_OUT = resolve(ROOT, "src/lib/brand-palette.generated.ts")

/* ------------------------------------------------------------------ 色彩数学 */

const srgbToLinear = (c) =>
  c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
const linearToSrgb = (c) =>
  c <= 0.0031308 ? 12.92 * c : 1.055 * Math.pow(c, 1 / 2.4) - 0.055

const hexToLinear = (hex) => {
  const n = parseInt(hex.replace("#", ""), 16)
  return [
    srgbToLinear(((n >> 16) & 255) / 255),
    srgbToLinear(((n >> 8) & 255) / 255),
    srgbToLinear((n & 255) / 255),
  ]
}

const linearToOklab = ([r, g, b]) => {
  const l = Math.cbrt(0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b)
  const m = Math.cbrt(0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b)
  const s = Math.cbrt(0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b)
  return [
    0.2104542553 * l + 0.7936177850 * m - 0.0040720468 * s,
    1.9779984951 * l - 2.4285922050 * m + 0.4505937099 * s,
    0.0259040371 * l + 0.7827717662 * m - 0.8086757660 * s,
  ]
}

const oklabToLinear = ([L, a, b]) => {
  const l_ = L + 0.3963377774 * a + 0.2158037573 * b
  const m_ = L - 0.1055613458 * a - 0.0638541728 * b
  const s_ = L - 0.0894841775 * a - 1.2914855480 * b
  const l = l_ ** 3
  const m = m_ ** 3
  const s = s_ ** 3
  return [
    4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
    -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
    -0.0041960863 * l - 0.7034186147 * m + 1.7076147010 * s,
  ]
}

const inGamut = ([r, g, b], eps = 1e-4) =>
  r >= -eps && r <= 1 + eps && g >= -eps && g <= 1 + eps && b >= -eps && b <= 1 + eps

const linearToHex = (lin) =>
  "#" +
  lin
    .map((c) => Math.round(Math.min(1, Math.max(0, linearToSrgb(c))) * 255))
    .map((c) => c.toString(16).padStart(2, "0"))
    .join("")

const toOklch = (hex) => {
  const [L, a, b] = linearToOklab(hexToLinear(hex))
  return {
    L,
    C: Math.hypot(a, b),
    H: ((Math.atan2(b, a) * 180) / Math.PI + 360) % 360,
  }
}

const fromOklch = (L, C, H) => {
  const rad = (H * Math.PI) / 180
  const direct = oklabToLinear([L, C * Math.cos(rad), C * Math.sin(rad)])
  if (inGamut(direct)) return linearToHex(direct)
  let lo = 0
  let hi = C
  let best = oklabToLinear([L, 0, 0])
  for (let i = 0; i < 40; i++) {
    const mid = (lo + hi) / 2
    const candidate = oklabToLinear([L, mid * Math.cos(rad), mid * Math.sin(rad)])
    if (inGamut(candidate)) {
      lo = mid
      best = candidate
    } else {
      hi = mid
    }
  }
  return linearToHex(best)
}

const maxChroma = (L, H) => {
  const rad = (H * Math.PI) / 180
  let lo = 0
  let hi = 0.45
  for (let i = 0; i < 36; i++) {
    const mid = (lo + hi) / 2
    if (inGamut(oklabToLinear([L, mid * Math.cos(rad), mid * Math.sin(rad)]))) lo = mid
    else hi = mid
  }
  return lo
}

const luminance = (hex) => {
  const [r, g, b] = hexToLinear(hex)
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

const contrast = (a, b) => {
  const x = luminance(a)
  const y = luminance(b)
  return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05)
}

/* -------------------------------------------------------------- 品牌与色阶定义 */

const INK = "#171717"
const WHITE = "#ffffff"
const LIGHT_BG = "#f8fafc"
const LIGHT_CARD = "#ffffff"
const DARK_BG = "#0b1120"

/** 明度刻度（OKLCH L）与对应的色域占用上限，浅端压制彩度避免“刺眼的 50 号色”。 */
const STOPS = [
  [50, 0.965, 0.6],
  [100, 0.935, 0.72],
  [200, 0.878, 0.76],
  [300, 0.808, 0.88],
  [400, 0.726, 0.95],
  [500, 0.648, 0.95],
  [600, 0.566, 0.95],
  [700, 0.486, 0.95],
  [800, 0.408, 0.95],
  [900, 0.33, 0.92],
  [950, 0.248, 0.88],
]

const BRANDS = [
  { id: "sunset", zh: "暖阳橙", en: "Sunset Orange", hex: "#DB7A0E" },
  { id: "cornflower", zh: "矢车菊蓝", en: "Cornflower Blue", hex: "#5A92E5" },
  { id: "jade", zh: "翠微", en: "Jade Green", hex: "#5E8034" },
  { id: "lagoon", zh: "碧青", en: "Lagoon Teal", hex: "#41B5C2" },
  { id: "lemon", zh: "柠檬黄", en: "Lemon Yellow", hex: "#FACA2E" },
  { id: "merlot", zh: "梅洛粉", en: "Merlot Pink", hex: "#EB6D98" },
  { id: "royal", zh: "皇家紫", en: "Royal Purple", hex: "#722169" },
  { id: "dawn", zh: "拂晓蓝", en: "Dawn Blue", hex: "#1677ff" },
  { id: "cyan", zh: "明青", en: "Cyan", hex: "#13c2c2" },
  { id: "purple", zh: "酱紫", en: "Golden Purple", hex: "#722ed1" },
  { id: "geekblue", zh: "极客蓝", en: "Geek Blue", hex: "#2f54eb" },
  { id: "magenta", zh: "法式洋红", en: "French Magenta", hex: "#eb2f96" },
  { id: "marigold", zh: "金盏花", en: "Marigold", hex: "#faad14" },
  { id: "lime", zh: "青柠", en: "Lime", hex: "#a0d911" },
  { id: "dusk", zh: "日暮", en: "Dusk", hex: "#fa8c16" },
]

function buildScale(baseHex) {
  const { L: baseL, C: baseC, H } = toOklch(baseHex)
  const scale = {}
  for (const [name, L, cap] of STOPS) {
    const amp =
      L <= baseL
        ? Math.max(0.28, 1 - 1.35 * (baseL - L))
        : (1 + 0.55 * Math.min(1, (L - baseL) / 0.24)) *
          (L > 0.84 ? Math.max(0.25, 1 - 2.4 * (L - 0.84)) : 1)
    const C = Math.min(baseC * amp, cap * maxChroma(L, H))
    scale[name] = fromOklch(L, C, H)
  }
  // 品牌色本身必须原样落在最接近的色阶上，否则主题页和设计稿对不上。
  const nearest = STOPS.reduce((best, stop) =>
    Math.abs(stop[1] - baseL) < Math.abs(best[1] - baseL) ? stop : best
  )
  scale[nearest[0]] = baseHex.toLowerCase()
  return { scale, baseStop: nearest[0], baseL, baseC, H }
}

const bestText = (bg) => {
  const cw = contrast(WHITE, bg)
  const ci = contrast(INK, bg)
  return cw >= ci ? { fg: WHITE, ratio: cw } : { fg: INK, ratio: ci }
}

/** 填充色：优先用最接近品牌的位置；该位置放不下 4.5:1 时向深/浅各让一拍。 */
function pickFill(scale, { order, minContrast = 4.5 }) {
  for (const stop of order) {
    const t = bestText(scale[stop])
    if (t.ratio >= minContrast) return { stop, color: scale[stop], ...t }
  }
  const fallback = order
    .map((stop) => ({ stop, ...bestText(scale[stop]) }))
    .sort((a, b) => b.ratio - a.ratio)[0]
  return { ...fallback, color: scale[fallback.stop] }
}

/**
 * 浅色模式可读文字色：从品牌所在色阶开始往深走，取第一个与背景 ≥4.5:1 的色阶。
 *
 * 不能“从 50 号往深扫第一个达标的”：那会取到最亮但仍够深的色阶，
 * 对深色品牌（皇家紫）会跳到色相更艳的 600 号，正文链接变成玫红。
 * 从品牌档位出发，皇家紫直接落到品牌原色 #722169（9.4:1），识别度最高。
 */
function pickInk(scale, bg, baseStop) {
  const start = STOPS.findIndex(([stop]) => stop === Number(baseStop))
  for (const [stop] of STOPS.slice(start >= 0 ? start : 0)) {
    const ratio = contrast(scale[stop], bg)
    // 取 5.2 而不是 4.5：品牌文字还会落在 --primary-soft（12% 品牌色底）上，
    // 浅底会把对比度再吃掉 0.3~0.5。留出余量后，chip/标签里的品牌字也不会掉到 AA 以下。
    if (ratio >= 5.2) return { stop, color: scale[stop], ratio }
  }
  return { stop: 950, color: scale[950], ratio: contrast(scale[950], bg) }
}

/** 深色模式下的品牌文字色：要亮，但不能白到失去品牌感。 */
function pickInkOnDark(scale, bg) {
  for (const stop of [300, 400, 200, 500, 100, 600, 50]) {
    const ratio = contrast(scale[stop], bg)
    if (ratio >= 4.5) return { stop, color: scale[stop], ratio }
  }
  return { stop: 300, color: scale[300], ratio: contrast(scale[300], bg) }
}

/**
 * 深色角色色的彩度上限。
 *
 * 品牌基色自身的彩度就是"识别度锚点"。深色模式为了对比度必须把明度拉到
 * 0.70~0.79，如果同时沿用色阶在高明度端的彩度放大，皇家紫会从深紫变成
 * 荧光玫红。这里把深色填充/文字的彩度钳回基色附近：明度可以变，色相和
 * "浓淡"不许跑。
 */
function capChroma(hex, limit) {
  const { L, C, H } = toOklch(hex)
  return C <= limit ? hex : fromOklch(L, limit, H)
}

function buildBrand(brand) {
  const { scale, baseStop, baseL, baseC } = buildScale(brand.hex)
  const stopIndex = (n) => STOPS.findIndex(([s]) => s === Number(n))

  // 浅色模式：从品牌所在色阶向两侧扩散，先深后浅（浅色底上深色填充更常见）。
  // 浅色模式：从品牌色阶出发，先往深走（深色填充 + 白字更常见），再往浅走。
  const idx = STOPS.findIndex(([stop]) => stop === Number(baseStop))
  const lightOrder = STOPS.slice(idx)
    .map(([stop]) => stop)
    .concat(STOPS.slice(0, idx).reverse().map(([stop]) => stop))
  const lightFill = pickFill(scale, { order: lightOrder })
  // 深色模式：填充不跟着品牌走，而是找“在深底上能分离出来、且前景放得下”的明度区间。
  const darkTarget = Math.min(0.79, Math.max(0.7, baseL))
  const darkOrder = STOPS.map(([stop, L]) => ({ stop, d: Math.abs(L - darkTarget) }))
    .sort((a, b) => a.d - b.d)
    .map((x) => x.stop)
  const darkFill = pickFill(scale, { order: darkOrder })
  // 深色填充允许比基色略艳（1.15 倍），再高就开始发光刺眼。
  const darkFillLimit = baseC * 1.15
  const darkFillColor = capChroma(darkFill.color, darkFillLimit)
  const darkFillText = bestText(darkFillColor)

  // 文字色要同时对页面底（更暗）和卡片（更亮）达标，取更严格的那个：页面底。
  const lightInk = pickInk(scale, LIGHT_BG, baseStop)
  const darkInk = pickInkOnDark(scale, DARK_BG)
  const darkInkColor = capChroma(darkInk.color, baseC)

  const hoverOf = (fill, direction, limit) => {
    const i = stopIndex(fill.stop)
    const next = STOPS[i + direction] ?? STOPS[i - direction]
    return next ? capChroma(scale[next[0]], limit ?? Number.POSITIVE_INFINITY) : fill.color
  }

  return {
    brand,
    scale,
    baseStop,
    light: {
      fill: lightFill,
      fillHover: hoverOf(lightFill, 1),
      ink: lightInk,
      // 焦点环需要和页面底色有 3:1
      ring: lightInk.color,
      ringRatio: contrast(lightInk.color, LIGHT_BG),
    },
    dark: {
      fill: { ...darkFill, color: darkFillColor, fg: darkFillText.fg, ratio: darkFillText.ratio },
      fillHover: hoverOf(darkFill, -1, darkFillLimit),
      ink: { ...darkInk, color: darkInkColor, ratio: contrast(darkInkColor, DARK_BG) },
      ring: darkInkColor,
      ringRatio: contrast(darkInkColor, DARK_BG),
    },
  }
}

const results = BRANDS.map(buildBrand)

/* ------------------------------------------------------------------ 校验断言 */

const failures = []
const rows = []

for (const r of results) {
  const checks = [
    ["浅色 填充/前景", contrast(r.light.fill.color, r.light.fill.fg), 4.5],
    ["浅色 文字/卡片", contrast(r.light.ink.color, LIGHT_CARD), 4.5],
    // 品牌文字色也会直接落在页面底色上（正文链接、侧栏选中项），只对卡片达标不够。
    ["浅色 文字/页面底", contrast(r.light.ink.color, LIGHT_BG), 4.5],
    ["浅色 焦点环/页面底", contrast(r.light.ring, LIGHT_BG), 3],
    ["深色 填充/前景", contrast(r.dark.fill.color, r.dark.fill.fg), 4.5],
    ["深色 填充/页面底", contrast(r.dark.fill.color, DARK_BG), 3],
    // 深色卡片不是页面底色，按钮和选中项多半落在卡片上，需要单独校验。
    ["深色 填充/卡片", contrast(r.dark.fill.color, "#111827"), 3],
    ["深色 文字/页面底", contrast(r.dark.ink.color, DARK_BG), 4.5],
    ["深色 焦点环/页面底", contrast(r.dark.ring, DARK_BG), 3],
  ]
  for (const [name, value, min] of checks) {
    if (value < min) failures.push(`${r.brand.id} ${name} = ${value.toFixed(2)} < ${min}`)
  }
  rows.push({
    brand: r.brand,
    baseStop: r.baseStop,
    light: {
      fill: r.light.fill.stop,
      fillFg: r.light.fill.fg === WHITE ? "white" : "ink",
      fillContrast: contrast(r.light.fill.color, r.light.fill.fg),
      ink: r.light.ink.stop,
      inkContrast: contrast(r.light.ink.color, LIGHT_CARD),
      ringContrast: contrast(r.light.ring, LIGHT_BG),
    },
    dark: {
      fill: r.dark.fill.stop,
      fillFg: r.dark.fill.fg === WHITE ? "white" : "ink",
      fillContrast: contrast(r.dark.fill.color, r.dark.fill.fg),
      ink: r.dark.ink.stop,
      inkContrast: contrast(r.dark.ink.color, DARK_BG),
      ringContrast: contrast(r.dark.ring, DARK_BG),
    },
  })
}

/* --------------------------------------------------------------------- 产物 */

const cssBlock = (r) => {
  const stops = STOPS.map(([name]) => `  --brand-${name}: ${r.scale[name]};`).join("\n")
  return `/* ${r.brand.zh} ${r.brand.en} — ${r.brand.hex} */
:root[data-brand="${r.brand.id}"] {
${stops}

  /* 浅色模式角色色 */
  --brand-fill: ${r.light.fill.color};
  --brand-fill-fg: ${r.light.fill.fg};
  --brand-fill-hover: ${r.light.fillHover};
  --brand-ink: ${r.light.ink.color};
  --brand-ring: ${r.light.ring};

  /* 深色模式角色色 */
  --brand-fill-dark: ${r.dark.fill.color};
  --brand-fill-fg-dark: ${r.dark.fill.fg};
  --brand-fill-hover-dark: ${r.dark.fillHover};
  --brand-ink-dark: ${r.dark.ink.color};
  --brand-ring-dark: ${r.dark.ring};
}`
}

const cssFile = `/*
 * 本文件由 scripts/gen-brand-tokens.mjs 生成，不要手改。
 * 新增品牌：改脚本里的 BRANDS，然后 npm run tokens:gen。
 */
${results.map(cssBlock).join("\n\n")}
`

const tsFile = `// 本文件由 scripts/gen-brand-tokens.mjs 生成，不要手改。
export type BrandId = ${BRANDS.map((b) => `"${b.id}"`).join(" | ")}

export interface BrandPalette {
  id: BrandId
  labelZh: string
  labelEn: string
  /** 品牌基色，仅用于展示与设计对齐，界面里请用语义 token。 */
  base: string
  /** 预览色卡用的浅色模式取值，和当前主题无关，保证 7 张卡片永远各自正确。 */
  previewLight: { bg: string; card: string; fill: string; fillFg: string; ink: string; tint: string }
  /** 预览色卡用的深色模式取值。 */
  previewDark: { bg: string; card: string; fill: string; fillFg: string; ink: string; tint: string }
}

export const BRAND_IDS = [${BRANDS.map((b) => `"${b.id}"`).join(", ")}] as const

export const BRAND_PALETTES: Record<BrandId, BrandPalette> = {
${results
  .map(
    (r) => `  ${r.brand.id}: {
    id: "${r.brand.id}",
    labelZh: "${r.brand.zh}",
    labelEn: "${r.brand.en}",
    base: "${r.brand.hex}",
    previewLight: {
      bg: "#f8fafc",
      card: "#ffffff",
      fill: "${r.light.fill.color}",
      fillFg: "${r.light.fill.fg}",
      ink: "${r.light.ink.color}",
      tint: "${r.scale[100]}",
    },
    previewDark: {
      bg: "#0b1120",
      card: "#111827",
      fill: "${r.dark.fill.color}",
      fillFg: "${r.dark.fill.fg}",
      ink: "${r.dark.ink.color}",
      tint: "${r.scale[900]}",
    },
  },`
  )
  .join("\n")}
}
`

const mode = process.argv.includes("--write")
  ? "write"
  : process.argv.includes("--check")
    ? "check"
    : "report"

if (mode === "write") {
  mkdirSync(dirname(CSS_OUT), { recursive: true })
  writeFileSync(CSS_OUT, cssFile)
  writeFileSync(TS_OUT, tsFile)
  console.log(`wrote ${CSS_OUT}`)
  console.log(`wrote ${TS_OUT}`)
}

if (mode === "check") {
  const stale = [
    [CSS_OUT, cssFile],
    [TS_OUT, tsFile],
  ].filter(([path, next]) => {
    try {
      return readFileSync(path, "utf8") !== next
    } catch {
      return true
    }
  })
  if (stale.length > 0) {
    console.error("品牌 token 产物已过期，请运行 npm run tokens:gen：")
    for (const [path] of stale) console.error(`  ${path}`)
    process.exit(1)
  }
}

if (mode === "report") {
  for (const r of rows) {
    console.log(
      `${r.brand.id.padEnd(11)} ${r.brand.zh.padEnd(5)} 浅色 fill#${String(r.light.fill).padEnd(4)} ${r.light.fillFg.padEnd(5)} ${r.light.fillContrast
        .toFixed(2)
        .padStart(5)}:1  ink#${String(r.light.ink).padEnd(4)} ${r.light.inkContrast
        .toFixed(2)
        .padStart(5)}:1  ring ${r.light.ringContrast.toFixed(2)}:1 | 深色 fill#${String(r.dark.fill).padEnd(4)} ${
        r.dark.fillFg
      } ${r.dark.fillContrast.toFixed(2)}:1  ink#${String(r.dark.ink).padEnd(4)} ${r.dark.inkContrast
        .toFixed(2)
        .padStart(5)}:1  ring ${r.dark.ringContrast.toFixed(2)}:1`
    )
  }
}

if (failures.length > 0) {
  console.error("\n对比度断言失败：")
  for (const f of failures) console.error(`  ${f}`)
  process.exit(1)
}
