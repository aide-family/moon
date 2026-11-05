// UI 约定检查：页面里不允许出现原生表单控件，也不允许字面色/调色板类。
// 历史遗留用基线文件兜底：只拦「新增」违规，基线里的旧账单独排期清理。
import fs from "node:fs"
import path from "node:path"

const root = path.resolve(import.meta.dirname, "../src")
const baselinePath = path.resolve(
  import.meta.dirname,
  "ui-conventions-baseline.json"
)
const update = process.argv.includes("--update")

// 历史遗留：约定生效前就存在的原生控件，改造成本单独排期；新代码不允许再加。
const RAW_CONTROL_ALLOWLIST = new Map([
  ["pages/datasource/query-panel.tsx", "查询正文需要原生 textarea 的行高与快捷键行为"],
])

const PALETTE =
  /(?:bg|text|border|ring|from|to|via)-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\d{2,3}/
const LITERAL = /#[0-9a-fA-F]{3,8}\b|\brgba?\(|\bhsla?\(/

const problems = []

function walk(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      walk(full)
      continue
    }
    if (!/\.(tsx|ts)$/.test(entry.name)) continue
    const relative = path.relative(root, full)
    if (relative.startsWith("components/ui/")) continue
    const lines = fs.readFileSync(full, "utf8").split("\n")
    lines.forEach((line, index) => {
      if (line.trimStart().startsWith("//")) return
      if (/<textarea\b|<select\b/.test(line)) {
        if (!RAW_CONTROL_ALLOWLIST.has(relative)) {
          problems.push({
            file: relative,
            line: index + 1,
            rule: "raw-control",
            message: "使用了原生表单控件，改用 ui/textarea 或 ui/select",
          })
        }
      }
      if (LITERAL.test(line)) {
        problems.push({
          file: relative,
          line: index + 1,
          rule: "color-literal",
          message: "出现字面色值，改用语义 token",
        })
      }
      if (PALETTE.test(line)) {
        problems.push({
          file: relative,
          line: index + 1,
          rule: "palette-class",
          message: "出现 Tailwind 调色板类，改用语义 token",
        })
      }
    })
  }
}

walk(root)

const key = (problem) => `${problem.file}|${problem.rule}`
const baseline = fs.existsSync(baselinePath)
  ? new Set(JSON.parse(fs.readFileSync(baselinePath, "utf8")))
  : new Set()

if (update) {
  const entries = [...new Set(problems.map(key))].sort()
  fs.writeFileSync(baselinePath, `${JSON.stringify(entries, null, 2)}\n`)
  console.log(`baseline updated: ${entries.length} entries`)
  process.exit(0)
}

const fresh = problems.filter((problem) => !baseline.has(key(problem)))
if (fresh.length === 0) {
  console.log(
    `ui conventions ok（基线内 ${new Set(problems.map(key)).size} 项旧账不计入）`
  )
  process.exit(0)
}
for (const problem of fresh) {
  console.error(`  - ${problem.file}:${problem.line} ${problem.message}`)
}
console.error(`\n新增 ${fresh.length} 处不符合 UI 约定（历史遗留用 --update 落基线）`)
process.exit(1)
