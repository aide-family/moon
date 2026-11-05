// 校验 web/src/i18n/index.ts 里 zh 与 en 两份文案的 key 完全一致。
// 不引入依赖：直接截出两个对象字面量求值，再把 key 拍平成路径集合比较。
import fs from "node:fs"
import path from "node:path"

const file = path.resolve(import.meta.dirname, "../src/i18n/index.ts")
const source = fs.readFileSync(file, "utf8")

function extractObject(name) {
  const marker = `const ${name} = {`
  const start = source.indexOf(marker)
  if (start < 0) throw new Error(`找不到 ${name} 对象`)
  const bodyStart = start + marker.length - 1
  let depth = 0
  let inString = null
  let escaped = false
  for (let i = bodyStart; i < source.length; i += 1) {
    const char = source[i]
    if (inString) {
      if (escaped) escaped = false
      else if (char === "\\") escaped = true
      else if (char === inString) inString = null
      continue
    }
    if (char === '"' || char === "'" || char === "`") {
      inString = char
      continue
    }
    if (char === "{") depth += 1
    else if (char === "}") {
      depth -= 1
      if (depth === 0) return source.slice(bodyStart, i + 1)
    }
  }
  throw new Error(`${name} 对象没有闭合`)
}

function flatten(value, prefix = "", out = new Set()) {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    for (const [key, child] of Object.entries(value)) {
      flatten(child, prefix ? `${prefix}.${key}` : key, out)
    }
    return out
  }
  out.add(prefix)
  return out
}

// i18next 的复数形态只在需要的语言里出现（英文 _one/_other，中文通常没有），
// 比较时把复数后缀归一化，避免把合法的复数差异报成缺失。
const PLURAL = /_(zero|one|two|few|many|other)$/
const normalize = (keys) =>
  new Set([...keys].map((key) => key.replace(PLURAL, "")))

const zh = normalize(flatten(new Function(`return (${extractObject("zh")})`)()))
const en = normalize(flatten(new Function(`return (${extractObject("en")})`)()))

const missingInEn = [...zh].filter((key) => !en.has(key)).sort()
const missingInZh = [...en].filter((key) => !zh.has(key)).sort()

if (missingInEn.length === 0 && missingInZh.length === 0) {
  console.log(`i18n ok: ${zh.size} keys`)
  process.exit(0)
}
if (missingInEn.length > 0) {
  console.error("en 缺少 key:")
  for (const key of missingInEn) console.error(`  - ${key}`)
}
if (missingInZh.length > 0) {
  console.error("zh 缺少 key:")
  for (const key of missingInZh) console.error(`  - ${key}`)
}
process.exit(1)
