import Editor, { type Monaco, type OnMount } from "@monaco-editor/react"
import { useEffect, useRef } from "react"

import { ensureMonacoRuntime, setupMonaco } from "@/components/monaco-runtime"
import { useTheme } from "@/components/theme-provider"
import { useMonacoOverflowNode } from "@/hooks/use-monaco-overflow-node"
import { cn } from "@/lib/class-names"

ensureMonacoRuntime()

// 编辑器文字与 placeholder 覆盖层共用这套字体/字号/行高/内边距，保证两者像素级对齐。
const FONT_FAMILY = 'JetBrains Mono, Menlo, Monaco, "Courier New", monospace'
const FONT_SIZE = 13
const LINE_HEIGHT = 20
const EDITOR_PADDING_Y = 8
// 与容器上的 px-2.5 一致（0.625rem）；Monaco 内容区从容器内边距内侧开始。
const EDITOR_PADDING_X = "0.625rem"

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
  /**
   * Monaco 模型的 URI。同一路径的多个编辑器会共享同一个文本模型，
   * 因此需要独立内容的编辑器必须传唯一路径（例如带实例前缀的文件名）。
   */
  path?: string
  /** 打开 schema 补全/悬浮提示；默认关闭，YAML 编辑器按需打开。 */
  enableSuggestions?: boolean
  /** 编辑器挂载完成后的回调，用于订阅 marker（校验错误）等能力。 */
  onMount?: (editor: Parameters<OnMount>[0], monaco: Monaco) => void
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
  path,
  enableSuggestions = false,
  onMount,
}: CodeEditorProps) {
  const { theme } = useTheme()
  const onChangeRef = useRef(onChange)
  const onMountRef = useRef(onMount)
  const overflowWidgetsDomNode = useMonacoOverflowNode()

  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  useEffect(() => {
    onMountRef.current = onMount
  }, [onMount])

  const dark =
    theme === "dark" ||
    (theme === "system" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)
  const editorTheme = dark ? "github-dark" : "github-light"

  const handleMount: OnMount = (editor, monaco) => {
    editor.layout()
    onMountRef.current?.(editor, monaco)
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
          path={path}
          theme={editorTheme}
          value={value}
          beforeMount={setupMonaco}
          onMount={handleMount}
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
            quickSuggestions: enableSuggestions,
            suggestOnTriggerCharacters: enableSuggestions,
          }}
        />
      </div>
    </div>
  )
}
