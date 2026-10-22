import Editor, { loader, type Monaco, type OnMount } from "@monaco-editor/react"
import { promLanguageDefinition } from "monaco-promql"
import { useEffect, useMemo, useRef } from "react"
import * as monacoEditor from "monaco-editor"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker"

import { listDatasourceMetrics } from "@/api/datasource"
import { proxyQuery } from "@/api/metric-query"
import { useTheme } from "@/components/theme-provider"

let monacoReady = false
let promqlLanguageReady = false
let logqlLanguageReady = false

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
      monaco.languages.setLanguageConfiguration(languageId, mod.languageConfiguration)
      monaco.languages.registerCompletionItemProvider(languageId, mod.completionItemProvider)
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
        [/[a-zA-Z_][\w:]*/, { cases: { "@keywords": "keyword", "@default": "identifier" } }],
        [/\s+/, "white"],
      ],
    },
  })
  logqlLanguageReady = true
}

function setupExpressionLanguages(monaco: Monaco) {
  setupPromqlLanguage(monaco)
  setupLogqlLanguage(monaco)
}

function parseStringList(body: string): string[] {
  try {
    const payload = JSON.parse(body) as { status?: string; data?: unknown }
    if (payload.status !== "success" || !Array.isArray(payload.data)) {
      return []
    }
    return payload.data.filter((item): item is string => typeof item === "string")
  } catch {
    return []
  }
}

// 两种查询语言共享 Monaco 运行时和标签补全逻辑；LogQL 使用独立语言 ID 与语法定义。
function registerDataCompletion(
  monaco: Monaco,
  languageId: "promql" | "logql",
  uidRef: { current: string },
  lokiRef: { current: boolean },
  plainRef: { current: boolean },
) {
  const cache = {
    metrics: null as string[] | null,
    labels: null as string[] | null,
    labelValues: new Map<string, string[]>(),
    uid: "",
  }

  return monaco.languages.registerCompletionItemProvider(languageId, {
    triggerCharacters: ["{", '"', ",", "="],
    provideCompletionItems: async (model: monacoEditor.editor.ITextModel, position: monacoEditor.Position) => {
      const loki = languageId === "logql"
      if (plainRef.current || lokiRef.current !== loki) {
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
        cache.labelValues.clear()
      }
      const line = model.getLineContent(position.lineNumber)
      const before = line.slice(0, position.column - 1)
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
            cache.labels = parseStringList(response.body).filter((label) => label !== "__name__")
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
      const prefix = before.match(/(?:^|[\s+\-*/%<>!&|,({])([a-zA-Z_:]\w*)$/)?.[1] ?? ""
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
  /** LogQL 使用独立语法与 Loki 标签补全；LogsQL 为纯文本，json 用于 Elasticsearch Query DSL。 */
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
  const uidRef = useRef(datasourceUid)
  const lokiRef = useRef(variant === "logql")
  const plainRef = useRef(variant === "logsql")
  const onSubmitRef = useRef(onSubmit)
  const disposeRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    uidRef.current = datasourceUid
  }, [datasourceUid])

  useEffect(() => {
    lokiRef.current = variant === "logql"
  }, [variant])

  useEffect(() => {
    plainRef.current = variant === "logsql"
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
    return Math.min(MAX_HEIGHT, Math.max(MIN_HEIGHT, lines * LINE_HEIGHT + EDITOR_PADDING_Y * 2))
  }, [value])

  const editorTheme =
    theme === "dark" || (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)
      ? "vs-dark"
      : "vs"

  const onMount: OnMount = (editor, monaco) => {
    disposeRef.current?.()
    const providers = (["promql", "logql"] as const).map((languageId) =>
      registerDataCompletion(monaco, languageId, uidRef, lokiRef, plainRef),
    )
    disposeRef.current = () => providers.forEach((provider) => provider.dispose())

    editor.addCommand(
      monaco.KeyCode.Enter,
      () => {
        if (onSubmitRef.current) {
          onSubmitRef.current()
          return
        }
        editor.trigger("keyboard", "type", { text: "\n" })
      },
      "!suggestWidgetVisible",
    )
    editor.addCommand(monaco.KeyMod.Shift | monaco.KeyCode.Enter, () => {
      editor.trigger("keyboard", "type", { text: "\n" })
    })
    editor.layout()
  }

  return (
    <div
      className="overflow-hidden rounded-md border border-input bg-background px-2.5 [&_.monaco-editor]:!bg-background [&_.monaco-editor-background]:!bg-background [&_.margin]:!bg-background"
      style={{ height }}
    >
      <Editor
        height={height}
        language={
          variant === "json"
            ? "json"
            : variant === "logsql"
              ? "plaintext"
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

export function LogqlEditor(props: Omit<QueryExpressionEditorProps, "variant">) {
  return <QueryExpressionEditor {...props} variant="logql" />
}
