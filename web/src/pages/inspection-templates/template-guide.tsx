import ReactMarkdown from "react-markdown"
import { useTranslation } from "react-i18next"
import remarkGfm from "remark-gfm"

import { DetailSheet } from "@/components/detail-sheet"

const ZH_GUIDE = `## 报告模板怎么工作

一份模板只对应一种格式（HTML 或 Markdown），计划绑定模板后，报告格式跟随模板。模板由 Go 的 \`html/template\`（HTML）或 \`text/template\`（Markdown）渲染，数据来自一次巡检执行。

## 1. 可用数据

| 字段 | 说明 |
| --- | --- |
| \`{{.Title}}\` | 报告标题，来自计划的标题模板或默认标题 |
| \`{{.GeneratedAt}}\` | 报告生成时间 |
| \`{{.Tenant.Name}}\` | 租户名称 |
| \`{{.Plan.Name}}\` / \`.Plan.Description\` / \`.Plan.Labels\` | 巡检计划信息 |
| \`{{.Run.Trigger}}\` / \`.StartedAt\` / \`.FinishedAt\` / \`.Duration\` | 触发方式、起止时间与耗时 |
| \`{{.Run.WindowStart}}\` / \`.WindowEnd\` | 本次数据窗口 |
| \`{{.Summary.Total}}\` / \`.OK\` / \`.Warn\` / \`.Failed\` / \`.Abnormal\` | 检查项统计 |
| \`{{.Sections}}\` | 全部检查项结果，按计划顺序排列 |
| \`{{.Abnormal}}\` | 仅警告与失败的检查项 |
| \`{{.HasAbnormal}}\` | 是否存在异常项 |

每个 Section 包含：\`.Title\`、\`.Kind\`、\`.Status\`（OK / WARN / FAILED）、\`.Error\`、\`.Note\`、\`.Metrics\`、\`.Table\`、\`.Series\`、\`.Logs\`、\`.Alerts\`、\`.Raw\`。不是每个检查项都会填满全部字段，渲染前用 \`{{if}}\` 判断。

## 2. 可用函数

- \`date "2006-01-02 15:04" .Run.StartedAt\`：按布局格式化时间
- \`duration .Run.Duration\`：格式化时长
- \`percent 85.5\`：输出 \`85.50%\`
- \`default "—" .Plan.Description\`：空值兜底
- \`join .Plan.Labels ","\`：拼接字符串切片
- \`markdown .Note\`：把说明字段按安全 Markdown 渲染
- \`metricCards .Metrics\` / \`tableHTML .Table\` / \`seriesChart .Series\` / \`logsHTML .Logs\` / \`alertsHTML .Alerts\`：内置可视化组件
- \`sectionHTML .\`：把一个 Section 的标题、状态、指标卡、趋势、表格一次性渲染
- \`json .\`：输出 JSON 字符串
- \`add\` / \`sub\` / \`mul\` / \`div\`：整数运算

## 3. HTML 示例

~~~html
<h1>{{.Title}}</h1>
<p>{{.Tenant.Name}} · {{.Run.WindowStart | date "2006-01-02 15:04"}}</p>

{{if .HasAbnormal}}
  <p>发现 {{.Summary.Abnormal}} 项需要关注</p>
{{end}}

{{range .Sections}}
  <section>
    <h2>{{.Title}}（{{.Status}}）</h2>
    {{if .Error}}<p>取数失败：{{.Error}}</p>{{end}}
    {{metricCards .Metrics}}
    {{seriesChart .Series}}
    {{tableHTML .Table}}
  </section>
{{end}}
~~~

只想快速出一份完整报告时，正文可以只写：

~~~html
<h1>{{.Title}}</h1>
{{range .Sections}}{{sectionHTML .}}{{end}}
~~~

## 4. Markdown 示例

~~~markdown
# {{.Title}}

{{if .HasAbnormal}}发现 **{{.Summary.Abnormal}}** 项异常。{{end}}

{{range .Sections}}
## {{.Title}} · {{.Status}}
{{if .Table}}
| 序列 | 最新值 |
| --- | --- |
{{range .Table.Rows}}| {{index . 0}} | {{index . 1}} |
{{end}}{{end}}
{{end}}
~~~

## 5. 限制与调试

- 模板保存前会做 Parse 校验；正文本身出错会阻止保存。
- 渲染结果上限 2 MB、单次渲染 10 秒，超限或超时会让整次巡检标记为 FAILED。
- 不提供环境变量、文件、网络或任意函数调用；HTML 由 \`html/template\` 自动转义。
- 不要引用外部 CSS、JS、图片；请把样式内联，保证下载后的单文件报告可直接打开。
- 报告一旦生成就是快照，之后修改模板不会回写历史报告。`

const EN_GUIDE = `## How report templates work

One template has exactly one format (HTML or Markdown). Once a plan binds a template, the report format follows it. Templates are rendered with Go \`html/template\` (HTML) or \`text/template\` (Markdown) from one inspection run.

## 1. Data available

| Field | Meaning |
| --- | --- |
| \`{{.Title}}\` | Report title from the plan title template or the default |
| \`{{.GeneratedAt}}\` | Report generation time |
| \`{{.Tenant.Name}}\` | Tenant name |
| \`{{.Plan.Name}}\` / \`.Plan.Description\` / \`.Plan.Labels\` | Plan metadata |
| \`{{.Run.Trigger}}\` / \`.StartedAt\` / \`.FinishedAt\` / \`.Duration\` | Trigger, start/end time and duration |
| \`{{.Run.WindowStart}}\` / \`.WindowEnd\` | Data window of this run |
| \`{{.Summary.Total}}\` / \`.OK\` / \`.Warn\` / \`.Failed\` / \`.Abnormal\` | Item counts |
| \`{{.Sections}}\` | Every item result in plan order |
| \`{{.Abnormal}}\` | Only WARN and FAILED items |
| \`{{.HasAbnormal}}\` | Whether any abnormal item exists |

Each Section exposes \`.Title\`, \`.Kind\`, \`.Status\` (OK / WARN / FAILED), \`.Error\`, \`.Note\`, \`.Metrics\`, \`.Table\`, \`.Series\`, \`.Logs\`, \`.Alerts\` and \`.Raw\`. Not every item fills every field; guard with \`{{if}}\`.

## 2. Functions

- \`date "2006-01-02 15:04" .Run.StartedAt\`: format a time
- \`duration .Run.Duration\`: format a duration
- \`percent 85.5\`: render \`85.50%\`
- \`default "—" .Plan.Description\`: fallback for empty values
- \`join .Plan.Labels ","\`: join a string slice
- \`markdown .Note\`: safe Markdown rendering for notes
- \`metricCards .Metrics\` / \`tableHTML .Table\` / \`seriesChart .Series\` / \`logsHTML .Logs\` / \`alertsHTML .Alerts\`: built-in visual helpers
- \`sectionHTML .\`: render one Section (title, status, cards, trend, table)
- \`json .\`: render JSON
- \`add\` / \`sub\` / \`mul\` / \`div\`: integer math

## 3. HTML example

~~~html
<h1>{{.Title}}</h1>
<p>{{.Tenant.Name}} · {{.Run.WindowStart | date "2006-01-02 15:04"}}</p>

{{if .HasAbnormal}}
  <p>{{.Summary.Abnormal}} item(s) need attention</p>
{{end}}

{{range .Sections}}
  <section>
    <h2>{{.Title}} ({{.Status}})</h2>
    {{if .Error}}<p>Query failed: {{.Error}}</p>{{end}}
    {{metricCards .Metrics}}
    {{seriesChart .Series}}
    {{tableHTML .Table}}
  </section>
{{end}}
~~~

For a complete report with minimal markup:

~~~html
<h1>{{.Title}}</h1>
{{range .Sections}}{{sectionHTML .}}{{end}}
~~~

## 4. Markdown example

~~~markdown
# {{.Title}}

{{if .HasAbnormal}}**{{.Summary.Abnormal}}** item(s) abnormal.{{end}}

{{range .Sections}}
## {{.Title}} · {{.Status}}
{{if .Table}}
| Series | Latest |
| --- | --- |
{{range .Table.Rows}}| {{index . 0}} | {{index . 1}} |
{{end}}{{end}}
{{end}}
~~~

## 5. Limits and debugging

- The body is parsed before saving; invalid templates cannot be saved.
- Rendered output is capped at 2 MB and 10 seconds; exceeding either marks the whole run FAILED.
- No environment, file, network or arbitrary function access. HTML is auto-escaped.
- Do not reference external CSS, JS or images; inline styles so the downloaded single-file report opens standalone.
- A report is an immutable snapshot; later template edits never rewrite history.`

/** 报告模板编写说明抽屉。 */
export function InspectionTemplateGuide({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t, i18n } = useTranslation()
  const content = i18n.language.startsWith("zh") ? ZH_GUIDE : EN_GUIDE
  return (
    <DetailSheet
      open={open}
      onOpenChange={onOpenChange}
      size="80vw"
      title={t("inspection.templateGuideTitle")}
      description={t("inspection.templateGuideDesc")}
    >
      <div className="mx-auto w-full max-w-[960px]">
        <article className="prose prose-sm max-w-none rounded-xl border border-border bg-card px-6 py-6 text-sm shadow-sm sm:px-10 sm:py-8 dark:prose-invert">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
        </article>
      </div>
    </DetailSheet>
  )
}
