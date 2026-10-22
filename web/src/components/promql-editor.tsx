import Editor, { loader, type Monaco, type OnMount } from "@monaco-editor/react"
import { promLanguageDefinition } from "monaco-promql"
import { useEffect, useMemo, useRef } from "react"
import * as monacoEditor from "monaco-editor"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker"

import { listDatasourceMetrics } from "@/api/datasource"
import { proxyQuery } from "@/api/metric-query"
import { useMonacoOverflowNode } from "@/components/monaco-overflow"
import { useTheme } from "@/components/theme-provider"

let monacoReady = false
let promqlLanguageReady = false
let logqlLanguageReady = false
let logsqlLanguageReady = false

const LINE_HEIGHT = 20
const EDITOR_PADDING_Y = 8
const MIN_HEIGHT = LINE_HEIGHT + EDITOR_PADDING_Y * 2
const MAX_HEIGHT = 180

function ensureMonaco() {
  if (monacoReady) {
    return
  }
  monacoReady = true
  window.MonacoEnvironment = {
    getWorker(_workerId, label) {
      return label === "json" ? new jsonWorker() : new editorWorker()
    },
  }
  loader.config({ monaco: monacoEditor })
}

ensureMonaco()

function setupPromqlLanguage(monaco: Monaco) {
  if (promqlLanguageReady) {
    return
  }
  const languageId = promLanguageDefinition.id
  monaco.languages.register(promLanguageDefinition)
  monaco.languages.onLanguage(languageId, () => {
    void promLanguageDefinition.loader().then((mod) => {
      monaco.languages.setMonarchTokensProvider(languageId, mod.language)
      monaco.languages.setLanguageConfiguration(
        languageId,
        mod.languageConfiguration
      )
      monaco.languages.registerCompletionItemProvider(
        languageId,
        mod.completionItemProvider
      )
    })
  })
  promqlLanguageReady = true
}

function setupLogqlLanguage(monaco: Monaco) {
  if (logqlLanguageReady) {
    return
  }
  monaco.languages.register({ id: "logql" })
  monaco.languages.setLanguageConfiguration("logql", {
    comments: { lineComment: "#" },
    brackets: [
      ["{", "}"],
      ["[", "]"],
      ["(", ")"],
    ],
    autoClosingPairs: [
      { open: "{", close: "}" },
      { open: "[", close: "]" },
      { open: "(", close: ")" },
      { open: '"', close: '"', notIn: ["string"] },
    ],
    surroundingPairs: [
      { open: "{", close: "}" },
      { open: "[", close: "]" },
      { open: "(", close: ")" },
      { open: '"', close: '"' },
    ],
  })
  monaco.languages.setMonarchTokensProvider("logql", {
    keywords: [
      "and",
      "or",
      "unless",
      "by",
      "without",
      "on",
      "ignoring",
      "group_left",
      "group_right",
      "sum",
      "avg",
      "count",
      "min",
      "max",
      "topk",
      "bottomk",
      "sort",
      "sort_desc",
      "stddev",
      "stdvar",
      "quantile",
      "absent",
      "rate",
      "rate_counter",
      "count_over_time",
      "bytes_over_time",
      "bytes_rate",
      "sum_over_time",
      "avg_over_time",
      "min_over_time",
      "max_over_time",
      "first_over_time",
      "last_over_time",
      "quantile_over_time",
      "stdvar_over_time",
      "stddev_over_time",
      "absent_over_time",
      "unwrap",
      "json",
      "logfmt",
      "pattern",
      "regexp",
      "line_format",
      "label_format",
      "decolorize",
      "drop",
      "keep",
      "duration",
      "bytes",
    ],
    tokenizer: {
      root: [
        [/#.*$/, "comment"],
        [/"(?:[^"\\]|\\.)*"/, "string"],
        [/\b\d+(?:\.\d+)?(?:ms|[smhdwy])\b/, "number"],
        [/\b\d+(?:\.\d+)?\b/, "number"],
        [/\|=|\|~|\|>|\|!|!=|!~|=~|==|>=|<=|>|<|=|!|\|\||&&|\|/, "operator"],
        [/[{}()[\]]/, "@brackets"],
        [/[,+\-*/%^]/, "operator"],
        [
          /[a-zA-Z_][\w:]*/,
          { cases: { "@keywords": "keyword", "@default": "identifier" } },
        ],
        [/\s+/, "white"],
      ],
    },
  })
  logqlLanguageReady = true
}

// LogsQL 的管道、统计函数和关键字。语法定义只负责着色，补全的字段名与字段值
// 由数据源的 /select/logsql/field_names 与 field_values 提供。
const logsqlKeywords = [
  "AND",
  "and",
  "OR",
  "or",
  "NOT",
  "not",
  "in",
  "not_in",
  "by",
  "as",
  "if",
  "stats",
  "sort",
  "limit",
  "first",
  "last",
  "offset",
  "fields",
  "filter",
  "math",
  "format",
  "rename",
  "copy",
  "delete",
  "uniq",
  "top",
  "unpack_json",
  "unpack_logfmt",
  "extract",
  "extract_regexp",
  "stream_context",
  "field_names",
  "field_values",
  "join",
  "union",
  "count",
  "count_uniq",
  "count_empty",
  "sum",
  "sum_len",
  "avg",
  "min",
  "max",
  "median",
  "quantile",
  "stddev",
  "rate",
  "rate_sum",
  "histogram",
  "values",
  "_time",
  "_stream",
  "_stream_id",
  "_msg",
]

function setupLogsqlLanguage(monaco: Monaco) {
  if (logsqlLanguageReady) {
    return
  }
  monaco.languages.register({ id: "logsql" })
  monaco.languages.setLanguageConfiguration("logsql", {
    comments: { lineComment: "#" },
    brackets: [
      ["{", "}"],
      ["[", "]"],
      ["(", ")"],
    ],
    autoClosingPairs: [
      { open: "{", close: "}" },
      { open: "[", close: "]" },
      { open: "(", close: ")" },
      { open: '"', close: '"', notIn: ["string"] },
    ],
    surroundingPairs: [
      { open: "{", close: "}" },
      { open: "[", close: "]" },
      { open: "(", close: ")" },
      { open: '"', close: '"' },
    ],
  })
  monaco.languages.setMonarchTokensProvider("logsql", {
    keywords: logsqlKeywords,
    tokenizer: {
      root: [
        [/#.*$/, "comment"],
        [/"(?:[^"\\]|\\.)*"/, "string"],
        [/'(?:[^'\\]|\\.)*'/, "string"],
        [/`[^`]*`/, "string"],
        [/\b\d+(?:\.\d+)?(?:ns|us|µs|ms|[smhdwy])\b/, "number"],
        [/\b\d+(?:\.\d+)?\b/, "number"],
        [/:>=|:<=|:>|:<|:=|!=|!~|=~|>=|<=|==|[=~<>:]/, "operator"],
        [/[|,+\-*/%]/, "operator"],
        [/[{}()[\]]/, "@brackets"],
        [
          /[a-zA-Z_][\w.]*/,
          { cases: { "@keywords": "keyword", "@default": "identifier" } },
        ],
        [/\s+/, "white"],
      ],
    },
  })
  logsqlLanguageReady = true
}

function setupExpressionLanguages(monaco: Monaco) {
  setupPromqlLanguage(monaco)
  setupLogqlLanguage(monaco)
  setupLogsqlLanguage(monaco)
}

function parseStringList(body: string): string[] {
  try {
    const payload = JSON.parse(body) as { status?: string; data?: unknown }
    if (payload.status !== "success" || !Array.isArray(payload.data)) {
      return []
    }
    return payload.data.filter(
      (item): item is string => typeof item === "string"
    )
  } catch {
    return []
  }
}

// VictoriaLogs 的 field_names / field_values 返回 {values: [{value, hits}]}，
// 与 Prometheus / Loki 的字符串数组不同，需要单独解析。
function parseVlogsValueList(body: string): string[] {
  try {
    const payload = JSON.parse(body) as { values?: unknown }
    if (!Array.isArray(payload.values)) {
      return []
    }
    const values: string[] = []
    for (const item of payload.values) {
      if (typeof item === "string") {
        values.push(item)
        continue
      }
      if (item && typeof item === "object" && "value" in item) {
        const value = (item as { value?: unknown }).value
        if (typeof value === "string") {
          values.push(value)
        }
      }
    }
    return values
  } catch {
    return []
  }
}

type CompletionCache = {
  metrics: string[] | null
  labels: string[] | null
  fields: string[] | null
  labelValues: Map<string, string[]>
  fieldValues: Map<string, string[]>
  uid: string
}

// VictoriaLogs 的 field_names / field_values 要求 query 参数；补全只看最近一天，
// 避免因为缺少时间范围被 -search.maxQueryTimeRange 拒绝或触发全量扫描。
const LOGSQL_COMPLETION_QUERY = encodeURIComponent("_time:1d *")

const logsqlStatics: {
  label: string
  detail: string
  kind: "keyword" | "function"
}[] = [
  { label: "AND", detail: "operator", kind: "keyword" },
  { label: "OR", detail: "operator", kind: "keyword" },
  { label: "NOT", detail: "operator", kind: "keyword" },
  { label: "_time", detail: "field", kind: "keyword" },
  { label: "_stream", detail: "field", kind: "keyword" },
  { label: "_msg", detail: "field", kind: "keyword" },
  { label: "stats", detail: "pipe", kind: "keyword" },
  { label: "by", detail: "pipe", kind: "keyword" },
  { label: "as", detail: "pipe", kind: "keyword" },
  { label: "sort", detail: "pipe", kind: "keyword" },
  { label: "limit", detail: "pipe", kind: "keyword" },
  { label: "fields", detail: "pipe", kind: "keyword" },
  { label: "filter", detail: "pipe", kind: "keyword" },
  { label: "uniq", detail: "pipe", kind: "keyword" },
  { label: "top", detail: "pipe", kind: "keyword" },
  { label: "first", detail: "pipe", kind: "keyword" },
  { label: "last", detail: "pipe", kind: "keyword" },
  { label: "math", detail: "pipe", kind: "keyword" },
  { label: "format", detail: "pipe", kind: "keyword" },
  { label: "unpack_json", detail: "pipe", kind: "keyword" },
  { label: "extract", detail: "pipe", kind: "keyword" },
  { label: "count", detail: "stats function", kind: "function" },
  { label: "count_uniq", detail: "stats function", kind: "function" },
  { label: "sum", detail: "stats function", kind: "function" },
  { label: "avg", detail: "stats function", kind: "function" },
  { label: "min", detail: "stats function", kind: "function" },
  { label: "max", detail: "stats function", kind: "function" },
  { label: "median", detail: "stats function", kind: "function" },
  { label: "quantile", detail: "stats function", kind: "function" },
  { label: "stddev", detail: "stats function", kind: "function" },
  { label: "rate", detail: "stats function", kind: "function" },
]

function isLogsqlWord(value: string): boolean {
  return /^[A-Za-z0-9_.-]+$/.test(value)
}

// 判断光标前是否还停在一个未闭合的引号里。只看奇偶性，避免把已经闭合的
// `app="api"` 误判成值上下文。
function hasOpenLogsqlQuote(text: string): boolean {
  let single = 0
  let double = 0
  let escaped = false
  for (const char of text) {
    if (escaped) {
      escaped = false
      continue
    }
    if (char === "\\") {
      escaped = true
      continue
    }
    if (char === '"') {
      double += 1
    } else if (char === "'") {
      single += 1
    }
  }
  return double % 2 === 1 || single % 2 === 1
}

function insertLogsqlValue(value: string, quote: string | null): string {
  if (quote === '"') {
    return value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')
  }
  if (quote === "'") {
    return value.replace(/\\/g, "\\\\").replace(/'/g, "\\'")
  }
  return isLogsqlWord(value) ? value : JSON.stringify(value)
}

async function provideLogsqlCompletion(
  monaco: Monaco,
  cache: CompletionCache,
  uid: string,
  position: monacoEditor.Position,
  before: string
): Promise<monacoEditor.languages.CompletionList> {
  const rangeOf = (length: number) => ({
    startLineNumber: position.lineNumber,
    startColumn: position.column - length,
    endLineNumber: position.lineNumber,
    endColumn: position.column,
  })

  // 字段值：`field:prefix`、`field:="prefix"`、`field:~"prefix"`、
  // `field:!~"prefix"`、`field:in("prefix`，以及流选择器 `{app="prefix`。
  // _time 的值是时长，_msg/_stream 的值没有可枚举语义，不做字段值补全。
  const filterValue = before.match(
    /([a-zA-Z_][\w.]*)\s*:\s*[=!~]{0,2}\s*(?:i\(|seq\(|in\(|contains_all\(|contains_any\(|equals_common_case\(|contains_common_case\()?["']?([^"'\s)]*)$/
  )
  const streamValue = before.match(
    /([a-zA-Z_][\w.]*)\s*(?:=~|!~|!=|=)\s*["']?([^"'\s}]*)$/
  )
  // `_stream:{app="..."}` 的 `_stream:` 会先匹配 filterValue，但真正的
  // 字段是花括号里的 app；这种情况下优先用流选择器的字段。
  const valueContext =
    filterValue && filterValue[1] !== "_stream"
      ? filterValue
      : (streamValue ?? filterValue)
  if (valueContext) {
    const field = valueContext[1]
    const prefix = valueContext[2]
    if (
      field === "_time" ||
      field === "_msg" ||
      field === "_stream" ||
      field === "_stream_id"
    ) {
      return { suggestions: [] }
    }
    if (!cache.fieldValues.has(field)) {
      try {
        const response = await proxyQuery({
          datasourceUid: uid,
          method: "GET",
          path: `/select/logsql/field_values?field=${encodeURIComponent(field)}&query=${LOGSQL_COMPLETION_QUERY}&limit=500`,
        })
        cache.fieldValues.set(field, parseVlogsValueList(response.body))
      } catch {
        cache.fieldValues.set(field, [])
      }
    }
    const values = cache.fieldValues.get(field) ?? []
    const quoted = before.match(/(["'])[^"']*$/)?.[1] ?? null
    return {
      suggestions: values
        .filter((value) => value.startsWith(prefix))
        .slice(0, 100)
        .map((value) => ({
          label: value,
          kind: monaco.languages.CompletionItemKind.Value,
          detail: field,
          insertText: insertLogsqlValue(value, quoted),
          range: rangeOf(prefix.length),
        })),
    }
  }

  // 无法确定当前在字段名位置时（例如引号内的第二个参数），不要插入字段名，
  // 否则会把字段名写进值里破坏查询。
  if (hasOpenLogsqlQuote(before)) {
    return { suggestions: [] }
  }

  const prefix =
    before.match(/(?:^|[\s{|(),!~=:+-])([a-zA-Z_][\w.]*)$/)?.[1] ?? ""
  if (!cache.fields) {
    try {
      const response = await proxyQuery({
        datasourceUid: uid,
        method: "GET",
        path: `/select/logsql/field_names?query=${LOGSQL_COMPLETION_QUERY}&limit=500`,
      })
      cache.fields = parseVlogsValueList(response.body)
    } catch {
      cache.fields = []
    }
  }
  const fields = (cache.fields ?? []).filter((field) =>
    field.startsWith(prefix)
  )
  const statics = logsqlStatics.filter((item) =>
    item.label.toLowerCase().startsWith(prefix.toLowerCase())
  )
  return {
    suggestions: [
      ...fields.slice(0, 100).map((field) => ({
        label: field,
        kind: monaco.languages.CompletionItemKind.Field,
        detail: "field",
        insertText: /^[A-Za-z_][\w.]*$/.test(field)
          ? field
          : JSON.stringify(field),
        range: rangeOf(prefix.length),
      })),
      ...statics.map((item) => ({
        label: item.label,
        kind:
          item.kind === "function"
            ? monaco.languages.CompletionItemKind.Function
            : monaco.languages.CompletionItemKind.Keyword,
        detail: item.detail,
        insertText: item.label,
        range: rangeOf(prefix.length),
      })),
    ],
  }
}

// 三种查询语言共享 Monaco 运行时；PromQL/LogQL 使用各自的标签补全，
// LogsQL 使用 VictoriaLogs 的字段名和字段值接口。
function registerDataCompletion(
  monaco: Monaco,
  languageId: "promql" | "logql" | "logsql",
  uidRef: { current: string },
  variantRef: { current: string }
) {
  const cache: CompletionCache = {
    metrics: null,
    labels: null,
    fields: null,
    labelValues: new Map<string, string[]>(),
    fieldValues: new Map<string, string[]>(),
    uid: "",
  }

  return monaco.languages.registerCompletionItemProvider(languageId, {
    triggerCharacters: ["{", '"', "'", ",", "=", ":", "|"],
    provideCompletionItems: async (
      model: monacoEditor.editor.ITextModel,
      position: monacoEditor.Position
    ) => {
      const loki = languageId === "logql"
      if (variantRef.current !== languageId) {
        return { suggestions: [] }
      }
      const uid = uidRef.current
      if (!uid) {
        return { suggestions: [] }
      }
      if (cache.uid !== uid) {
        cache.uid = uid
        cache.metrics = null
        cache.labels = null
        cache.fields = null
        cache.labelValues.clear()
        cache.fieldValues.clear()
      }
      const line = model.getLineContent(position.lineNumber)
      const before = line.slice(0, position.column - 1)
      if (languageId === "logsql") {
        return provideLogsqlCompletion(monaco, cache, uid, position, before)
      }
      const labelValue = before.match(/([a-zA-Z_]\w*)\s*(=~?|!~?)\s*"([^"]*)$/)
      if (labelValue) {
        const labelName = labelValue[1]
        const prefix = labelValue[3]
        if (!cache.labelValues.has(labelName)) {
          try {
            const response = await proxyQuery({
              datasourceUid: uid,
              method: "GET",
              path: loki
                ? `/loki/api/v1/label/${encodeURIComponent(labelName)}/values`
                : `/api/v1/label/${encodeURIComponent(labelName)}/values`,
            })
            cache.labelValues.set(labelName, parseStringList(response.body))
          } catch {
            cache.labelValues.set(labelName, [])
          }
        }
        const values = cache.labelValues.get(labelName) ?? []
        return {
          suggestions: values
            .filter((value) => value.startsWith(prefix))
            .slice(0, 100)
            .map((value) => ({
              label: value,
              kind: monaco.languages.CompletionItemKind.Value,
              insertText: value,
              range: {
                startLineNumber: position.lineNumber,
                startColumn: position.column - prefix.length,
                endLineNumber: position.lineNumber,
                endColumn: position.column,
              },
            })),
        }
      }

      if (/\{[^}]*$/.test(before)) {
        if (!cache.labels) {
          try {
            const response = await proxyQuery({
              datasourceUid: uid,
              method: "GET",
              path: loki ? "/loki/api/v1/labels" : "/api/v1/labels",
            })
            cache.labels = parseStringList(response.body).filter(
              (label) => label !== "__name__"
            )
          } catch {
            cache.labels = []
          }
        }
        const prefix = before.match(/[,{]\s*([a-zA-Z_]\w*)$/)?.[1] ?? ""
        return {
          suggestions: (cache.labels ?? [])
            .filter((label) => label.startsWith(prefix))
            .slice(0, 100)
            .map((label) => ({
              label,
              kind: monaco.languages.CompletionItemKind.Field,
              insertText: label,
              range: {
                startLineNumber: position.lineNumber,
                startColumn: position.column - prefix.length,
                endLineNumber: position.lineNumber,
                endColumn: position.column,
              },
            })),
        }
      }

      // LogQL 没有指标名，补充标签和标签值即可。
      if (loki) {
        return { suggestions: [] }
      }

      if (!cache.metrics) {
        try {
          const response = await listDatasourceMetrics(uid)
          cache.metrics = response.names ?? []
        } catch {
          cache.metrics = []
        }
      }
      const prefix =
        before.match(/(?:^|[\s+\-*/%<>!&|,({])([a-zA-Z_:]\w*)$/)?.[1] ?? ""
      return {
        suggestions: (cache.metrics ?? [])
          .filter((metric) => metric.startsWith(prefix))
          .slice(0, 100)
          .map((metric) => ({
            label: metric,
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: metric,
            detail: "metric",
            range: {
              startLineNumber: position.lineNumber,
              startColumn: position.column - prefix.length,
              endLineNumber: position.lineNumber,
              endColumn: position.column,
            },
          })),
      }
    },
  })
}

function lineCountOf(value: string): number {
  if (!value) {
    return 1
  }
  return value.split("\n").length
}

interface QueryExpressionEditorProps {
  value: string
  onChange: (value: string) => void
  datasourceUid?: string
  onSubmit?: () => void
  /** LogQL 使用 Loki 标签补全，LogsQL 使用 VictoriaLogs 字段补全，json 用于 Elasticsearch Query DSL。 */
  variant?: "promql" | "logql" | "logsql" | "json"
}

export function QueryExpressionEditor({
  value,
  onChange,
  datasourceUid = "",
  onSubmit,
  variant = "promql",
}: QueryExpressionEditorProps) {
  const { theme } = useTheme()
  const overflowWidgetsDomNode = useMonacoOverflowNode()
  const uidRef = useRef(datasourceUid)
  const variantRef = useRef(variant)
  const onSubmitRef = useRef(onSubmit)
  const disposeRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    uidRef.current = datasourceUid
  }, [datasourceUid])

  useEffect(() => {
    variantRef.current = variant
  }, [variant])

  useEffect(() => {
    onSubmitRef.current = onSubmit
  }, [onSubmit])

  useEffect(() => {
    return () => {
      disposeRef.current?.()
    }
  }, [])

  const height = useMemo(() => {
    const lines = lineCountOf(value)
    return Math.min(
      MAX_HEIGHT,
      Math.max(MIN_HEIGHT, lines * LINE_HEIGHT + EDITOR_PADDING_Y * 2)
    )
  }, [value])

  const editorTheme =
    theme === "dark" ||
    (theme === "system" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)
      ? "vs-dark"
      : "vs"

  const onMount: OnMount = (editor, monaco) => {
    disposeRef.current?.()
    const providers = (["promql", "logql", "logsql"] as const).map(
      (languageId) =>
        registerDataCompletion(monaco, languageId, uidRef, variantRef)
    )
    disposeRef.current = () =>
      providers.forEach((provider) => provider.dispose())

    editor.addCommand(
      monaco.KeyCode.Enter,
      () => {
        if (onSubmitRef.current) {
          onSubmitRef.current()
          return
        }
        editor.trigger("keyboard", "type", { text: "\n" })
      },
      "!suggestWidgetVisible"
    )
    editor.addCommand(monaco.KeyMod.Shift | monaco.KeyCode.Enter, () => {
      editor.trigger("keyboard", "type", { text: "\n" })
    })
    editor.layout()
  }

  return (
    <div
      className="overflow-hidden rounded-md border border-input bg-background px-2.5 [&_.margin]:!bg-background [&_.monaco-editor]:!bg-background [&_.monaco-editor-background]:!bg-background"
      style={{ height }}
    >
      <Editor
        height={height}
        language={
          variant === "json"
            ? "json"
            : variant === "logsql"
              ? "logsql"
              : variant === "logql"
                ? "logql"
                : "promql"
        }
        theme={editorTheme}
        value={value}
        beforeMount={setupExpressionLanguages}
        onMount={onMount}
        onChange={(next) => onChange(next ?? "")}
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
          overviewRulerBorder: false,
          hideCursorInOverviewRuler: true,
          scrollBeyondLastLine: false,
          wordWrap: "on",
          automaticLayout: true,
          padding: { top: EDITOR_PADDING_Y, bottom: EDITOR_PADDING_Y },
          scrollbar: {
            vertical: height >= MAX_HEIGHT ? "auto" : "hidden",
            horizontal: "hidden",
            verticalScrollbarSize: 8,
            handleMouseWheel: height >= MAX_HEIGHT,
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

export function PromqlEditor(props: QueryExpressionEditorProps) {
  return <QueryExpressionEditor {...props} />
}

export function LogqlEditor(
  props: Omit<QueryExpressionEditorProps, "variant">
) {
  return <QueryExpressionEditor {...props} variant="logql" />
}

export function LogsqlEditor(
  props: Omit<QueryExpressionEditorProps, "variant">
) {
  return <QueryExpressionEditor {...props} variant="logsql" />
}
