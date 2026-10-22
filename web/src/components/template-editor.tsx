import Editor, { loader, type Monaco, type OnMount } from "@monaco-editor/react"
import { useEffect, useRef } from "react"
import * as monacoEditor from "monaco-editor"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"

import { useMonacoOverflowNode } from "@/components/monaco-overflow"
import { useTheme } from "@/components/theme-provider"

let monacoReady = false
let templateLanguageReady = false

const EDITOR_PADDING_Y = 8

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

// templateFields 是投递参数里的字段。模板和告警摘要、描述共用 pkg/alerttmpl，
// 因此这里列出的写法在试发、预览和真实投递里含义一致。
const templateFields: Array<{ label: string; detail: string; insertText?: string }> = [
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
      const suggestions: monacoEditor.languages.CompletionItem[] = []
      const insideTemplate = lastOpen > lastClose
      if (!insideTemplate) {
        suggestions.push({
          label: "{{ .summary }}",
          kind: monaco.languages.CompletionItemKind.Snippet,
          detail: "插入模板表达式",
          insertText: "{{ .summary }}",
          range,
        })
        return { suggestions }
      }
      const labelPrefix = word.match(/^(?:\$labels|\.Labels)\.(\w*)$/)
      if (labelPrefix) {
        for (const key of templateLabelKeys) {
          if (!key.startsWith(labelPrefix[1])) {
            continue
          }
          const prefix = word.slice(0, -labelPrefix[1].length)
          suggestions.push({
            label: key,
            kind: monaco.languages.CompletionItemKind.Field,
            detail: "事件标签",
            insertText: `${prefix}${key}`,
            range,
          })
        }
        return { suggestions }
      }
      for (const field of templateFields) {
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
      for (const fn of templateFunctions) {
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
}

/**
 * 通知模板编辑器：Go template + Prometheus 模板函数的补全和语法高亮，
 * 用于邮件主题、正文和预设页。补全列表里的字段与 pkg/alerttmpl 保持一致。
 */
export function TemplateEditor({
  value,
  onChange,
  height = 220,
  singleLine = false,
  placeholder,
  ariaLabel,
}: TemplateEditorProps) {
  const { theme } = useTheme()
  const onChangeRef = useRef(onChange)
  const overflowWidgetsDomNode = useMonacoOverflowNode()
  const editorHeight = singleLine ? 40 : height

  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  const dark =
    theme === "dark" ||
    (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)
  const editorTheme = dark ? "vs-dark" : "vs"

  const onMount: OnMount = (editor, monaco) => {
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
    editor.layout()
  }

  return (
    <div
      className="relative overflow-hidden rounded-md border border-input bg-background px-2.5 [&_.margin]:!bg-background [&_.monaco-editor]:!bg-background [&_.monaco-editor-background]:!bg-background"
      style={{ height: editorHeight }}
    >
      {value === "" && placeholder ? (
        <span className="pointer-events-none absolute left-3 top-2 z-10 text-sm text-muted-foreground">
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
          fontSize: 13,
          fontFamily: 'JetBrains Mono, Menlo, Monaco, "Courier New", monospace',
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
