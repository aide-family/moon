import Editor, { loader, type Monaco, type OnMount } from "@monaco-editor/react"
import { promLanguageDefinition } from "monaco-promql"
import { useEffect, useRef } from "react"
import * as monacoEditor from "monaco-editor"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"

import { listDatasourceMetrics } from "@/api/datasource"
import { proxyQuery } from "@/api/metric-query"
import { useTheme } from "@/components/theme-provider"

let monacoReady = false
let promqlLanguageReady = false

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

function registerDataCompletion(monaco: Monaco, uidRef: { current: string }) {
  const cache = {
    metrics: null as string[] | null,
    labels: null as string[] | null,
    labelValues: new Map<string, string[]>(),
    uid: "",
  }

  return monaco.languages.registerCompletionItemProvider("promql", {
    triggerCharacters: ["{", '"', ",", "="],
    provideCompletionItems: async (model, position) => {
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
              path: `/api/v1/label/${encodeURIComponent(labelName)}/values`,
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
              path: "/api/v1/labels",
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

interface PromqlEditorProps {
  value: string
  onChange: (value: string) => void
  datasourceUid?: string
  height?: number
}

export function PromqlEditor({ value, onChange, datasourceUid = "", height = 160 }: PromqlEditorProps) {
  const { theme } = useTheme()
  const uidRef = useRef(datasourceUid)
  const disposeRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    uidRef.current = datasourceUid
  }, [datasourceUid])

  useEffect(() => {
    return () => {
      disposeRef.current?.()
    }
  }, [])

  const editorTheme =
    theme === "dark" || (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)
      ? "vs-dark"
      : "vs"

  const onMount: OnMount = (editor, monaco) => {
    disposeRef.current?.()
    const provider = registerDataCompletion(monaco, uidRef)
    disposeRef.current = () => provider.dispose()
    editor.layout()
  }

  return (
    <div className="overflow-hidden rounded-lg border" style={{ height }}>
      <Editor
        height={height}
        language="promql"
        theme={editorTheme}
        value={value}
        beforeMount={setupPromqlLanguage}
        onMount={onMount}
        onChange={(next) => onChange(next ?? "")}
        options={{
          minimap: { enabled: false },
          fontSize: 13,
          scrollBeyondLastLine: false,
          wordWrap: "on",
          automaticLayout: true,
        }}
      />
    </div>
  )
}
