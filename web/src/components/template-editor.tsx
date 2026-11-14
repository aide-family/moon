import Editor, { loader, type Monaco, type OnMount } from "@monaco-editor/react"
import { useEffect, useMemo, useRef, useState } from "react"
import * as monacoEditor from "monaco-editor"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"

import { fetchDatasourceLabelKeys } from "@/components/datasource-label-keys"
import { useTheme } from "@/components/theme-provider"
import { useMonacoOverflowNode } from "@/hooks/use-monaco-overflow-node"

let monacoReady = false
let templateLanguageReady = false

// 编辑器文字与 placeholder 覆盖层共用这套字体/字号/行高/内边距，保证两者像素级对齐。
const FONT_FAMILY = 'JetBrains Mono, Menlo, Monaco, "Courier New", monospace'
const FONT_SIZE = 13
const LINE_HEIGHT = 20
const EDITOR_PADDING_Y = 8
// 与容器上的 px-2.5 一致（0.625rem）；Monaco 内容区从容器内边距内侧开始。
const EDITOR_PADDING_X = "0.625rem"

function ensureMonaco() {
  if (monacoReady) {
    return
  }
  monacoReady = true
  window.MonacoEnvironment = {
    getWorker() {
      return new editorWorker()
    },
  }
  loader.config({ monaco: monacoEditor })
}

ensureMonaco()

interface TemplateField {
  label: string
  detail: string
  insertText?: string
}

// notificationTemplateFields 是通知投递参数里的字段。
const notificationTemplateFields: TemplateField[] = [
  { label: ".summary", detail: "告警摘要" },
  { label: ".description", detail: "告警描述" },
  { label: ".phase", detail: "事件阶段：FIRING / RECOVERED" },
  { label: ".eventUid", detail: "事件标识" },
  { label: ".ruleUid", detail: "规则标识" },
  { label: ".subscriptionUid", detail: "命中订阅标识" },
  { label: ".groupUid", detail: "通知组标识" },
  { label: ".fingerprint", detail: "事件指纹" },
  { label: ".labels", detail: "事件标签 JSON 字符串" },
  { label: ".Labels", detail: "事件标签 map" },
  { label: ".value", detail: "采样值" },
  { label: ".Value", detail: "采样值（Prometheus 写法）" },
  { label: ".externalURL", detail: "数据源外部地址" },
  { label: ".ExternalURL", detail: "数据源外部地址（Prometheus 写法）" },
  { label: ".externalLabels", detail: "外部标签 map" },
  { label: ".ExternalLabels", detail: "外部标签 map（Prometheus 写法）" },
  { label: "$labels", detail: "事件标签 map" },
  { label: "$value", detail: "采样值" },
  { label: "$externalURL", detail: "数据源外部地址" },
  { label: "$externalLabels", detail: "外部标签 map" },
  { label: "{{ . }}", detail: "采样值快捷写法", insertText: "{{ . }}" },
]

// alertTemplateFields 是告警规则摘要/描述在评测期能拿到的全部数据（megrez 展开），
// 没有事件字段。两套模板的补全清单必须分开，否则用户会写出评测时永远为空的模板。
// 自定义变量的结果（.Vars / .VarLabels 和 $变量名）由 variableNames 动态补全：
// 变量查询语句在变量求值前展开，拿不到 .Vars，所以只在声明了变量时才提示。
const alertTemplateFields: TemplateField[] = [
  { label: ".labels", detail: "采样标签 JSON 字符串" },
  { label: ".Labels", detail: "采样标签 map（Prometheus 写法）" },
  { label: ".value", detail: "采样值" },
  { label: ".Value", detail: "采样值（Prometheus 写法）" },
  { label: ".externalURL", detail: "数据源外部地址" },
  { label: ".ExternalURL", detail: "数据源外部地址（Prometheus 写法）" },
  { label: ".externalLabels", detail: "外部标签 map" },
  { label: ".ExternalLabels", detail: "外部标签 map（Prometheus 写法）" },
  { label: "$labels", detail: "采样标签 map" },
  { label: "$value", detail: "采样值" },
  { label: "$externalURL", detail: "数据源外部地址" },
  { label: "$externalLabels", detail: "外部标签 map" },
  { label: "{{ . }}", detail: "采样值快捷写法", insertText: "{{ . }}" },
]

const alertVariableResultFields: TemplateField[] = [
  { label: ".Vars", detail: "自定义变量结果 map" },
  { label: ".VarLabels", detail: "自定义变量结果的标签 map" },
]

// inspectionTemplateFields 是巡检报告模板的数据契约（bo.InspectionReportData）。
// 字段在 range .Sections 内时指向单个检查项；补全只负责给名字，具体结构由报告说明描述。
const inspectionTemplateFields: TemplateField[] = [
  { label: ".Title", detail: "报告标题" },
  { label: ".GeneratedAt", detail: "报告生成时间" },
  { label: ".Tenant.Name", detail: "租户名称" },
  { label: ".Plan.Name", detail: "巡检计划名称" },
  { label: ".Plan.Description", detail: "巡检计划说明" },
  { label: ".Plan.Labels", detail: "计划标签 JSON" },
  { label: ".Run.Trigger", detail: "触发方式：SCHEDULED / MANUAL" },
  { label: ".Run.StartedAt", detail: "执行开始时间" },
  { label: ".Run.FinishedAt", detail: "执行结束时间" },
  { label: ".Run.Duration", detail: "执行耗时文本" },
  { label: ".Run.WindowStart", detail: "数据窗口开始时间" },
  { label: ".Run.WindowEnd", detail: "数据窗口结束时间" },
  { label: ".Summary.Total", detail: "检查项总数" },
  { label: ".Summary.OK", detail: "正常项数" },
  { label: ".Summary.Warn", detail: "警告项数" },
  { label: ".Summary.Failed", detail: "失败项数" },
  { label: ".Summary.Abnormal", detail: "异常项数（警告 + 失败）" },
  { label: ".Sections", detail: "全部检查项，配合 range 遍历" },
  { label: ".Abnormal", detail: "仅异常检查项" },
  { label: ".HasAbnormal", detail: "是否存在异常项" },
  { label: ".Status", detail: "章节状态：OK / WARN / FAILED（range .Sections 内）" },
  { label: ".Kind", detail: "章节类型（range .Sections 内）" },
  { label: ".Error", detail: "取数失败原因（range .Sections 内）" },
  { label: ".Note", detail: "章节说明，可用 markdown 渲染（range .Sections 内）" },
  { label: ".Metrics", detail: "指标卡（range .Sections 内）" },
  { label: ".Table", detail: "表格（range .Sections 内）" },
  { label: ".Series", detail: "趋势序列（range .Sections 内）" },
  { label: ".Logs", detail: "日志样本（range .Sections 内）" },
  { label: ".Alerts", detail: "告警行（range .Sections 内）" },
  { label: ".Raw", detail: "原始响应（range .Sections 内）" },
]

const inspectionTemplateFunctions: Array<{ label: string; detail: string }> = [
  { label: "date", detail: '格式化时间：date "2006-01-02 15:04" .Run.StartedAt' },
  { label: "duration", detail: "格式化时长：duration .Run.Duration" },
  { label: "percent", detail: "格式化百分比：percent 85.5" },
  { label: "default", detail: '空值兜底：default "—" .Plan.Description' },
  { label: "join", detail: '拼接切片：join .Plan.Labels ","' },
  { label: "markdown", detail: "渲染说明字段的 Markdown：markdown .Note" },
  { label: "metricCards", detail: "渲染指标卡：metricCards .Metrics" },
  { label: "tableHTML", detail: "渲染表格：tableHTML .Table" },
  { label: "seriesChart", detail: "渲染趋势图：seriesChart .Series" },
  { label: "logsHTML", detail: "渲染日志样本：logsHTML .Logs" },
  { label: "alertsHTML", detail: "渲染告警表：alertsHTML .Alerts" },
  { label: "sectionHTML", detail: "渲染整个章节：sectionHTML ." },
  { label: "json", detail: "输出 JSON：json ." },
  { label: "add", detail: "整数相加" },
  { label: "sub", detail: "整数相减" },
  { label: "mul", detail: "整数相乘" },
  { label: "div", detail: "整数相除" },
]

const templateLabelKeys = [
  "app",
  "namespace",
  "pod",
  "job",
  "instance",
  "node",
  "alert_level",
  "datasource",
  "datasource_level",
]

// templateFunctions 对应 Prometheus 告警模板函数。query 依赖数据源，只在评测阶段可用，
// 通知投递阶段调用会返回空结果，这里仍然列出便于用户保留同一套模板写法。
const templateFunctions: Array<{ label: string; detail: string }> = [
  { label: "humanize", detail: "格式化数字：humanize $value" },
  { label: "humanizePercentage", detail: "格式化为百分比" },
  { label: "humanizeDuration", detail: "格式化时长（秒）" },
  { label: "humanizeTimestamp", detail: "格式化时间戳" },
  { label: "toUpper", detail: "转大写" },
  { label: "toLower", detail: "转小写" },
  { label: "title", detail: "首字母大写" },
  { label: "trimSpace", detail: "去掉首尾空白" },
  { label: "match", detail: "正则匹配" },
  { label: "reReplaceAll", detail: "正则替换" },
  { label: "safeHtml", detail: "标记为安全 HTML" },
  { label: "printf", detail: "格式化字符串" },
  { label: "print", detail: "拼接字符串" },
  { label: "println", detail: "拼接字符串并换行" },
  { label: "len", detail: "长度" },
  { label: "index", detail: "取下标" },
  { label: "first", detail: "取查询结果第一条" },
  { label: "label", detail: "读取标签值" },
  { label: "value", detail: "读取采样值" },
  { label: "strvalue", detail: "读取采样值字符串" },
  { label: "graphLink", detail: "生成图表链接" },
  { label: "tableLink", detail: "生成表格链接" },
  { label: "sortByLabel", detail: "按标签排序" },
  { label: "query", detail: "即时查询（仅评测阶段）" },
  { label: "range", detail: "自定义数字范围" },
]

export type TemplateScope = "notification" | "alert" | "inspection"

interface TemplateContext {
  scope: TemplateScope
  labelKeys: string[]
  /** 告警规则已声明的自定义变量名，摘要/描述里用 $name 引用。 */
  variableNames: string[]
}

// 补全 provider 全局只注册一次，按 model URI 查当前编辑器的上下文。弹窗里可能同时
// 存在摘要和描述两个编辑器，按实例各注册 provider 会出现重复补全项。
const editorContexts = new Map<string, TemplateContext>()

function setupTemplateLanguage(monaco: Monaco) {
  if (templateLanguageReady) {
    return
  }
  templateLanguageReady = true
  monaco.languages.register({ id: "gotmpl" })
  monaco.languages.setLanguageConfiguration("gotmpl", {
    comments: { lineComment: "" },
    brackets: [
      ["{", "}"],
      ["(", ")"],
    ],
    autoClosingPairs: [
      { open: "{", close: "}" },
      { open: "(", close: ")" },
      { open: '"', close: '"', notIn: ["string"] },
    ],
  })
  monaco.languages.setMonarchTokensProvider("gotmpl", {
    defaultToken: "",
    tokenPostfix: ".gotmpl",
    keywords: templateFunctions.map((item) => item.label),
    tokenizer: {
      root: [
        [/\{\{-?/, "delimiter", "@template"],
        [/[^{]+/, ""],
        [/\{/, ""],
      ],
      template: [
        [/-?\}\}/, "delimiter", "@pop"],
        [/\$[a-zA-Z_]\w*/, "variable"],
        [/\.[a-zA-Z_]\w*/, "variable"],
        [/"(?:[^"\\]|\\.)*"/, "string"],
        [/\|/, "operator"],
        [/\b\d+(\.\d+)?\b/, "number"],
        [/[a-zA-Z_]\w*/, {
          cases: {
            "@keywords": "keyword",
            "@default": "identifier",
          },
        }],
      ],
    },
  })
  monaco.languages.registerCompletionItemProvider("gotmpl", {
    triggerCharacters: ["{", "$", "."],
    provideCompletionItems(model: monacoEditor.editor.ITextModel, position: monacoEditor.Position) {
      const before = model.getValueInRange({
        startLineNumber: 1,
        startColumn: 1,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      })
      const lastOpen = before.lastIndexOf("{{")
      const lastClose = before.lastIndexOf("}}")
      const word = before.match(/[A-Za-z0-9_$.]*$/)?.[0] ?? ""
      const range = {
        startLineNumber: position.lineNumber,
        startColumn: position.column - word.length,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      }
      const context = editorContexts.get(model.uri.toString())
      const scope = context?.scope ?? "notification"
      const labelKeys = context?.labelKeys ?? templateLabelKeys
      const variableNames =
        scope === "alert" ? (context?.variableNames ?? []) : []
      const fields =
        scope === "alert"
          ? variableNames.length > 0
            ? [...alertTemplateFields, ...alertVariableResultFields]
            : alertTemplateFields
          : scope === "inspection"
            ? inspectionTemplateFields
            : notificationTemplateFields
      const functions =
        scope === "inspection" ? inspectionTemplateFunctions : templateFunctions
      const suggestions: monacoEditor.languages.CompletionItem[] = []
      const insideTemplate = lastOpen > lastClose
      if (!insideTemplate) {
        const snippet =
          scope === "alert"
            ? "{{ $labels.instance }}"
            : scope === "inspection"
              ? "{{ .Summary.Abnormal }}"
              : "{{ .summary }}"
        suggestions.push({
          label: snippet,
          kind: monaco.languages.CompletionItemKind.Snippet,
          detail: "插入模板表达式",
          insertText: snippet,
          range,
        })
        return { suggestions }
      }
      const labelPrefix = word.match(/^(?:\$labels|\.Labels)\.(\w*)$/)
      if (labelPrefix) {
        for (const key of labelKeys) {
          if (!key.startsWith(labelPrefix[1])) {
            continue
          }
          const prefix = word.slice(0, -labelPrefix[1].length)
          suggestions.push({
            label: key,
            kind: monaco.languages.CompletionItemKind.Field,
            detail: scope === "alert" ? "采样标签" : "事件标签",
            insertText: `${prefix}${key}`,
            range,
          })
        }
        return { suggestions }
      }
      for (const name of variableNames) {
        if (word && !`$${name}`.startsWith(word)) {
          continue
        }
        suggestions.push({
          label: `$${name}`,
          kind: monaco.languages.CompletionItemKind.Variable,
          detail: "自定义变量（评测时查询）",
          insertText: `$${name}`,
          range,
        })
      }
      for (const field of fields) {
        if (word && !field.label.startsWith(word)) {
          continue
        }
        suggestions.push({
          label: field.label,
          kind: monaco.languages.CompletionItemKind.Variable,
          detail: field.detail,
          insertText: field.insertText ?? field.label,
          range,
        })
      }
      for (const fn of functions) {
        if (word && !fn.label.startsWith(word.replace(/^\$/, ""))) {
          continue
        }
        suggestions.push({
          label: fn.label,
          kind: monaco.languages.CompletionItemKind.Function,
          detail: fn.detail,
          insertText: `${fn.label} `,
          range,
        })
      }
      return { suggestions }
    },
  })
}

interface TemplateEditorProps {
  value: string
  onChange: (value: string) => void
  height?: number
  singleLine?: boolean
  placeholder?: string
  ariaLabel?: string
  /** 模板作用域：通知投递包含事件字段，告警规则的摘要/描述只有评测期数据。 */
  scope?: TemplateScope
  /** 探测标签用的数据源；留空只用内置标签键。 */
  datasourceUid?: string
  datasourceDriver?: string
  expression?: string
  /** 规则自身标签的键，评测时会并入采样标签。 */
  labelKeys?: string[]
  /**
   * 告警规则已声明的自定义变量名（只在 scope="alert" 生效）。摘要/描述编辑器
   * 会补全 `$name` 和 `.Vars`；变量查询语句编辑器不传，避免写出求值前不存在的引用。
   */
  variableNames?: string[]
  /** 编辑器挂载后返回插入文本的入口，用于「可用变量」快捷插入。 */
  onEditorReady?: (insert: (text: string) => void) => void
}

/**
 * 模板编辑器：Go template + Prometheus 模板函数的补全和语法高亮。通知模板、
 * 告警规则的摘要和描述共用；告警作用域只提示评测期真实存在的字段，并按数据源
 * 查询到的标签键补全 $labels。
 */
export function TemplateEditor({
  value,
  onChange,
  height = 220,
  singleLine = false,
  placeholder,
  ariaLabel,
  scope = "notification",
  datasourceUid = "",
  datasourceDriver = "",
  expression = "",
  labelKeys,
  variableNames,
  onEditorReady,
}: TemplateEditorProps) {
  const { theme } = useTheme()
  const onChangeRef = useRef(onChange)
  const onEditorReadyRef = useRef(onEditorReady)
  const overflowWidgetsDomNode = useMonacoOverflowNode()
  const editorHeight = singleLine ? 40 : height
  const modelUriRef = useRef<string | null>(null)
  const [probedLabels, setProbedLabels] = useState<{
    query: string
    keys: string[]
  }>({ query: "", keys: [] })
  const labelQuery = datasourceUid
    ? `${datasourceUid}|${datasourceDriver}|${expression}`
    : ""
  // 查询串变了就不再使用旧结果，避免上一个表达式的标签短暂出现在补全里。
  const fetchedLabelKeys = useMemo(
    () => (probedLabels.query === labelQuery ? probedLabels.keys : []),
    [probedLabels, labelQuery]
  )

  const mergedLabelKeys = useMemo(() => {
    const keys = new Set<string>([
      ...templateLabelKeys,
      ...(labelKeys ?? []),
      ...fetchedLabelKeys,
    ])
    return [...keys]
  }, [labelKeys, fetchedLabelKeys])

  // 只保留非空且去重后的变量名，避免空行把 `$` 补全塞满。
  const mergedVariableNames = useMemo(() => {
    const names = new Set<string>()
    for (const name of variableNames ?? []) {
      const trimmed = name.trim()
      if (trimmed) {
        names.add(trimmed)
      }
    }
    return [...names]
  }, [variableNames])

  const contextRef = useRef<TemplateContext>({
    scope,
    labelKeys: mergedLabelKeys,
    variableNames: mergedVariableNames,
  })

  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  useEffect(() => {
    onEditorReadyRef.current = onEditorReady
  }, [onEditorReady])

  // 数据源标签是异步查询的：防抖后探测一次，失败或没有数据源时只用内置键。
  useEffect(() => {
    if (!labelQuery) {
      return
    }
    let active = true
    const timer = window.setTimeout(() => {
      void fetchDatasourceLabelKeys({
        datasourceUid,
        driver: datasourceDriver,
        expression,
      }).then((keys) => {
        if (active) {
          setProbedLabels({ query: labelQuery, keys })
        }
      })
    }, 400)
    return () => {
      active = false
      window.clearTimeout(timer)
    }
  }, [labelQuery, datasourceUid, datasourceDriver, expression])

  useEffect(() => {
    contextRef.current = {
      scope,
      labelKeys: mergedLabelKeys,
      variableNames: mergedVariableNames,
    }
    const uri = modelUriRef.current
    if (uri) {
      editorContexts.set(uri, contextRef.current)
    }
  }, [scope, mergedLabelKeys, mergedVariableNames])

  useEffect(
    () => () => {
      const uri = modelUriRef.current
      if (uri) {
        editorContexts.delete(uri)
        modelUriRef.current = null
      }
    },
    []
  )

  const dark =
    theme === "dark" ||
    (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)
  const editorTheme = dark ? "vs-dark" : "vs"

  const onMount: OnMount = (editor, monaco) => {
    const uri = editor.getModel()?.uri.toString()
    if (uri) {
      modelUriRef.current = uri
      editorContexts.set(uri, contextRef.current)
    }
    editor.addCommand(monaco.KeyCode.Enter, () => {
      if (singleLine) {
        return
      }
      editor.trigger("keyboard", "type", { text: "\n" })
    }, "!suggestWidgetVisible")
    if (singleLine) {
      editor.addCommand(monaco.KeyMod.Shift | monaco.KeyCode.Enter, () => {
        return
      })
    }
    onEditorReadyRef.current?.((text) => {
      const selection = editor.getSelection()
      if (!selection) {
        return
      }
      editor.pushUndoStop()
      editor.executeEdits("template-insert", [
        { range: selection, text, forceMoveMarkers: true },
      ])
      editor.pushUndoStop()
      editor.focus()
    })
    editor.layout()
  }

  return (
    <div
      data-slot="template-editor"
      className="relative overflow-hidden rounded-md border border-input bg-background px-2.5 [&_.margin]:!bg-background [&_.monaco-editor]:!bg-background [&_.monaco-editor-background]:!bg-background"
      style={{ height: editorHeight }}
    >
      {value === "" && placeholder ? (
        // 覆盖层与 Monaco 内容使用同一套字体/行高/内边距，left/top 对齐到内容区；
        // pointer-events-none 保证点击和拖选仍落在编辑器上。单行模式不换行并省略溢出。
        <span
          className={`pointer-events-none absolute z-10 select-none text-muted-foreground ${
            singleLine ? "truncate" : "whitespace-pre-wrap"
          }`}
          style={{
            top: EDITOR_PADDING_Y,
            left: EDITOR_PADDING_X,
            right: EDITOR_PADDING_X,
            fontFamily: FONT_FAMILY,
            fontSize: FONT_SIZE,
            lineHeight: `${LINE_HEIGHT}px`,
          }}
        >
          {placeholder}
        </span>
      ) : null}
      <Editor
        height={editorHeight}
        language="gotmpl"
        theme={editorTheme}
        value={value}
        beforeMount={setupTemplateLanguage}
        onMount={onMount}
        onChange={(next) => onChangeRef.current(next ?? "")}
        options={{
          fixedOverflowWidgets: true,
          overflowWidgetsDomNode: overflowWidgetsDomNode ?? undefined,
          minimap: { enabled: false },
          fontSize: FONT_SIZE,
          fontFamily: FONT_FAMILY,
          lineNumbers: "off",
          glyphMargin: false,
          folding: false,
          lineDecorationsWidth: 0,
          lineNumbersMinChars: 0,
          renderLineHighlight: "none",
          overviewRulerLanes: 0,
          hideCursorInOverviewRuler: true,
          scrollBeyondLastLine: false,
          wordWrap: singleLine ? "off" : "on",
          automaticLayout: true,
          ariaLabel,
          padding: { top: EDITOR_PADDING_Y, bottom: EDITOR_PADDING_Y },
          scrollbar: {
            vertical: singleLine ? "hidden" : height >= 220 ? "auto" : "hidden",
            horizontal: "hidden",
            verticalScrollbarSize: 8,
          },
          guides: { indentation: false },
          renderWhitespace: "none",
          contextmenu: false,
          quickSuggestions: true,
          suggestOnTriggerCharacters: true,
          tabCompletion: "on",
        }}
      />
    </div>
  )
}
