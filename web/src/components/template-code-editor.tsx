import Editor, { loader, type Monaco } from "@monaco-editor/react"
import { useEffect } from "react"
import * as monacoEditor from "monaco-editor"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker"
import htmlWorker from "monaco-editor/esm/vs/language/html/html.worker?worker"

import { useTheme } from "@/components/theme-provider"

let ready = false

function ensureMonaco() {
  if (ready) {
    return
  }
  ready = true
  window.MonacoEnvironment = {
    getWorker(_workerId, label) {
      if (label === "json") {
        return new jsonWorker()
      }
      if (label === "html" || label === "markdown") {
        return new htmlWorker()
      }
      return new editorWorker()
    },
  }
  loader.config({ monaco: monacoEditor })
}

ensureMonaco()

interface TemplateCodeEditorProps {
  value: string
  onChange: (value: string) => void
  language: "html" | "markdown"
  height?: number
  placeholder?: string
  readOnly?: boolean
}

/** HTML / Markdown 代码编辑器；模板正文用。 */
export function TemplateCodeEditor({
  value,
  onChange,
  language,
  height = 340,
  placeholder,
  readOnly = false,
}: TemplateCodeEditorProps) {
  const { theme } = useTheme()
  const editorTheme =
    theme === "dark" ||
    (theme === "system" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)
      ? "vs-dark"
      : "vs"
  useEffect(() => {
    void loader.init().then((monaco: Monaco) => {
      monaco.languages.typescript?.javascriptDefaults?.setDiagnosticsOptions?.({
        noSemanticValidation: true,
        noSyntaxValidation: true,
      })
    })
  }, [])
  return (
    <div className="overflow-hidden rounded-lg border border-input bg-background">
      <Editor
        height={`${height}px`}
        language={language}
        theme={editorTheme}
        value={value}
        onChange={(next) => onChange(next ?? "")}
        options={{
          readOnly,
          minimap: { enabled: false },
          fontSize: 12,
          lineNumbers: "on",
          wordWrap: "on",
          scrollBeyondLastLine: false,
          automaticLayout: true,
          padding: { top: 10, bottom: 10 },
          tabSize: 2,
          placeholder,
        }}
      />
    </div>
  )
}
