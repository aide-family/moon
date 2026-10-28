#!/usr/bin/env node
/**
 * 主题引导脚本守卫。
 *
 * index.html 里那段内联脚本为了"首帧生效"不能 import 模块，所以品牌 id 列表
 * 在每个文件里各写了一份。这份守卫保证两处不漂移：
 *   1. index.html 引导脚本里的品牌数组 == brand-palette.generated.ts 的 BRAND_IDS
 *   2. 引导脚本读取的 localStorage 键名 == lib/appearance.ts 的 STORAGE_KEYS
 * 用法：node scripts/check-theme-boot.mjs
 */
import { readFileSync } from "node:fs"
import { dirname, resolve } from "node:path"
import { fileURLToPath } from "node:url"

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..")
const html = readFileSync(resolve(ROOT, "index.html"), "utf8")
const palette = readFileSync(resolve(ROOT, "src/lib/brand-palette.generated.ts"), "utf8")
const appearance = readFileSync(resolve(ROOT, "src/lib/appearance.ts"), "utf8")

const failures = []

const bootMatch = html.match(/const brands = \[([^\]]+)\]/)
if (!bootMatch) {
  failures.push("index.html 中找不到引导脚本的品牌数组（const brands = [...]）")
} else {
  const bootBrands = [...bootMatch[1].matchAll(/"([^"]+)"/g)].map((m) => m[1])
  const idsMatch = palette.match(/export const BRAND_IDS = \[([^\]]+)\]/)
  const registry = idsMatch ? [...idsMatch[1].matchAll(/"([^"]+)"/g)].map((m) => m[1]) : []
  if (bootBrands.join(",") !== registry.join(",")) {
    failures.push(`品牌列表不一致：\n  引导脚本 ${bootBrands.join(",")}\n  BRAND_IDS ${registry.join(",")}`)
  }
}

const storageMatch = appearance.match(/STORAGE_KEYS = \{([^}]+)\}/)
const storageKeys = storageMatch ? [...storageMatch[1].matchAll(/"([^"]+)"/g)].map((m) => m[1]) : []
for (const key of ["theme", "brand"]) {
  if (!html.includes(`localStorage.getItem("${key}")`)) {
    failures.push(`index.html 引导脚本没有读取 localStorage 的 "${key}" 键`)
  }
  if (!storageKeys.includes(key)) {
    failures.push(`lib/appearance.ts 的 STORAGE_KEYS 缺少 "${key}"`)
  }
}

// 引导脚本必须早于样式表和 React：否则首帧会先按默认值绘制一遍。
const scriptIndex = html.indexOf("prefers-color-scheme: dark")
const moduleIndex = html.indexOf('type="module"')
if (scriptIndex === -1 || moduleIndex === -1 || scriptIndex > moduleIndex) {
  failures.push("引导脚本必须内联在 <head> 且早于入口 module 脚本")
}

if (failures.length > 0) {
  console.error("主题引导脚本校验失败：")
  for (const failure of failures) console.error(`  - ${failure}`)
  process.exit(1)
}

console.log("主题引导脚本与品牌注册表一致。")
