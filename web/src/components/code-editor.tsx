import Editor, { loader, type Monaco, type OnMount } from "@monaco-editor/react"
import { useEffect, useRef } from "react"
import * as monacoEditor from "monaco-editor"
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker"
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker"

import { useTheme } from "@/components/theme-provider"
import { useMonacoOverflowNode } from "@/hooks/use-monaco-overflow-node"
import { cn } from "@/lib/class-names"

let monacoReady = false
let themesReady = false

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
    getWorker(_workerId, label) {
      return label === "json" ? new jsonWorker() : new editorWorker()
    },
  }
  loader.config({ monaco: monacoEditor })
}

ensureMonaco()

// 配色直接取自 Shiki 的 github-light / github-dark，与 CodeBlock 的
// --code-bg / --code-fg 是同一套 GitHub 主题，代码块和编辑器在明暗主题下保持一致。
const githubLightRules = [
  { token: "comment", foreground: "6a737d" },
  { token: "string", foreground: "032f62" },
  { token: "string.key", foreground: "005cc5" },
  { token: "string.value", foreground: "032f62" },
  { token: "constant", foreground: "005cc5" },
  { token: "keyword", foreground: "d73a49" },
  { token: "number", foreground: "005cc5" },
  { token: "type", foreground: "22863a" },
  { token: "tag", foreground: "22863a" },
  { token: "attribute", foreground: "6f42c1" },
  { token: "variable", foreground: "e36209" },
  { token: "operator", foreground: "24292e" },
  { token: "operators", foreground: "24292e" },
  { token: "delimiter", foreground: "24292e" },
  { token: "meta", foreground: "6a737d" },
  { token: "anchor", foreground: "005cc5" },
  { token: "invalid", foreground: "b31d28" },
]

const githubDarkRules = [
  { token: "comment", foreground: "6a737d" },
  { token: "string", foreground: "9ecbff" },
  { token: "string.key", foreground: "79b8ff" },
  { token: "string.value", foreground: "9ecbff" },
  { token: "constant", foreground: "79b8ff" },
  { token: "keyword", foreground: "f97583" },
  { token: "number", foreground: "79b8ff" },
  { token: "type", foreground: "85e89d" },
  { token: "tag", foreground: "85e89d" },
  { token: "attribute", foreground: "b392f0" },
  { token: "variable", foreground: "ffab70" },
  { token: "operator", foreground: "e1e4e8" },
  { token: "operators", foreground: "e1e4e8" },
  { token: "delimiter", foreground: "e1e4e8" },
  { token: "meta", foreground: "6a737d" },
  { token: "anchor", foreground: "79b8ff" },
  { token: "invalid", foreground: "fdaeb7" },
]

function setupGithubThemes(monaco: Monaco) {
  if (themesReady) {
    return
  }
  themesReady = true
  monaco.editor.defineTheme("github-light", {
    base: "vs",
    inherit: true,
    rules: githubLightRules,
    colors: {
      "editor.background": "#ffffff",
      "editor.foreground": "#24292e",
      "editorCursor.foreground": "#044289",
      "editor.lineHighlightBackground": "#f6f8fa",
      "editor.selectionBackground": "#0366d625",
      "editor.inactiveSelectionBackground": "#0366d611",
      "editorIndentGuide.background1": "#eff2f6",
      "editorIndentGuide.activeBackground1": "#d7dbe0",
      "editorWidget.background": "#f6f8fa",
      "editorWidget.border": "#d1d5da",
      "editorSuggestWidget.background": "#f6f8fa",
      "editorSuggestWidget.border": "#d1d5da",
      "editorSuggestWidget.foreground": "#24292e",
      "editorSuggestWidget.selectedBackground": "#e2e5e9",
      "editorHoverWidget.background": "#f6f8fa",
      "editorHoverWidget.border": "#d1d5da",
      "scrollbarSlider.background": "#959da533",
      "scrollbarSlider.hoverBackground": "#959da544",
      "scrollbarSlider.activeBackground": "#959da555",
      "editorOverviewRuler.border": "#ffffff",
    },
  })
  monaco.editor.defineTheme("github-dark", {
    base: "vs-dark",
    inherit: true,
    rules: githubDarkRules,
    colors: {
      "editor.background": "#24292e",
      "editor.foreground": "#e1e4e8",
      "editorCursor.foreground": "#c8e1ff",
      "editor.lineHighlightBackground": "#2b3036",
      "editor.selectionBackground": "#3392ff44",
      "editor.inactiveSelectionBackground": "#3392ff22",
      "editorIndentGuide.background1": "#2f363d",
      "editorIndentGuide.activeBackground1": "#444d56",
      "editorWidget.background": "#1f2428",
      "editorWidget.border": "#444d56",
      "editorSuggestWidget.background": "#1f2428",
      "editorSuggestWidget.border": "#444d56",
      "editorSuggestWidget.foreground": "#e1e4e8",
      "editorSuggestWidget.selectedBackground": "#39414a",
      "editorHoverWidget.background": "#1f2428",
      "editorHoverWidget.border": "#444d56",
      "scrollbarSlider.background": "#6a737d33",
      "scrollbarSlider.hoverBackground": "#6a737d44",
      "scrollbarSlider.activeBackground": "#6a737d55",
      "editorOverviewRuler.border": "#24292e",
    },
  })
}

export interface CodeEditorProps {
  value: string
  onChange: (value: string) => void
  /** 语法高亮语种；kubeconfig 用 yaml，JSON 粘贴内容也由 YAML 词法着色。 */
  language?: "yaml" | "json"
  /** 标题栏文案，默认显示语种大写。 */
  label?: string
  id?: string
  /** 编辑器内容区高度（px），内容超出时在编辑器内部滚动。 */
  height?: number
  placeholder?: string
  ariaLabel?: string
  className?: string
}

/**
 * GitHub 主题的代码块编辑器：沿用 CodeBlock 的卡片外观（语言标题栏 + 边框 + GitHub
 * light/dark 配色），但内容可编辑。用于 kubeconfig 这类需要粘贴/修改的配置片段，
 * 内部自行滚动，不撑开外层对话框内容区。
 */
export function CodeEditor({
  value,
  onChange,
  language = "yaml",
  label,
  id,
  height = 240,
  placeholder,
  ariaLabel,
  className,
}: CodeEditorProps) {
  const { theme } = useTheme()
  const onChangeRef = useRef(onChange)
  const overflowWidgetsDomNode = useMonacoOverflowNode()

  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  const dark =
    theme === "dark" ||
    (theme === "system" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)
  const editorTheme = dark ? "github-dark" : "github-light"

  const onMount: OnMount = (editor) => {
    editor.layout()
  }

  return (
    <div
      id={id}
      data-slot="code-editor"
      className={cn(
        "flex min-w-0 flex-col overflow-hidden rounded-lg border border-border/70",
        className
      )}
    >
      <div className="flex shrink-0 items-center justify-between gap-2 border-b border-border/60 bg-muted/40 px-3 py-1.5">
        <span className="min-w-0 truncate font-mono text-[11px] tracking-[0.12em] text-muted-foreground uppercase">
          {label ?? language}
        </span>
      </div>
      <div
        className="relative min-w-0 bg-[var(--code-bg)] px-2.5"
        style={{ height }}
      >
        {value === "" && placeholder ? (
          // 覆盖层不参与命中测试：点击、拖选仍落在 Monaco 上；输入任意字符后卸载。
          <span
            className="pointer-events-none absolute z-10 font-mono whitespace-pre-wrap text-muted-foreground select-none"
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
          height={height}
          language={language}
          theme={editorTheme}
          value={value}
          beforeMount={setupGithubThemes}
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
            overviewRulerBorder: false,
            hideCursorInOverviewRuler: true,
            scrollBeyondLastLine: false,
            wordWrap: "on",
            automaticLayout: true,
            ariaLabel,
            padding: { top: EDITOR_PADDING_Y, bottom: EDITOR_PADDING_Y },
            scrollbar: {
              vertical: "auto",
              horizontal: "hidden",
              verticalScrollbarSize: 8,
            },
            tabSize: 2,
            insertSpaces: true,
            guides: { indentation: false },
            renderWhitespace: "none",
            contextmenu: false,
            quickSuggestions: false,
            suggestOnTriggerCharacters: false,
          }}
        />
      </div>
    </div>
  )
}
