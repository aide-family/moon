import type { OnMount } from "@monaco-editor/react"
import { useId, useState } from "react"
import { useTranslation } from "react-i18next"

import { CodeEditor } from "@/components/code-editor"
import {
  isKnownKubernetesKind,
  readManifestKind,
} from "@/components/kubernetes-yaml-schema"
import { cn } from "@/lib/class-names"

export interface YamlEditorProps {
  value: string
  onChange: (value: string) => void
  id?: string
  label?: string
  height?: number
  placeholder?: string
  ariaLabel?: string
  className?: string
}

interface YamlIssue {
  key: string
  severity: "error" | "warning"
  line: number
  message: string
}

/**
 * Kubernetes YAML 编辑器：monaco-yaml 提供 schema 补全、悬浮说明和红波浪线校验，
 * 校验结果显示在编辑器下方。kind 在内置 schema 中存在时按 kind 校验；无法识别
 * kind 时退化为 apiVersion / kind / metadata 的基础提示。
 */
export function YamlEditor({
  value,
  onChange,
  id,
  label,
  height = 320,
  placeholder,
  ariaLabel,
  className,
}: YamlEditorProps) {
  const { t } = useTranslation()
  // 每次编辑器实例使用唯一 model 路径：路径决定 schema，也避免不同对话框
  // （例如新增与编辑）共用同一个 Monaco 文本模型。
  const instance = useId().replace(/[^a-zA-Z0-9]/g, "")
  const kind = readManifestKind(value)
  const schemaKind = isKnownKubernetesKind(kind) ? kind : "Unknown"
  const path = `file:///alioth/appstore/${instance}/${schemaKind}.yaml`
  const [issues, setIssues] = useState<YamlIssue[]>([])

  const handleMount: OnMount = (editor, monaco) => {
    let serial = 0
    const refresh = () => {
      const model = editor.getModel()
      if (!model) {
        return
      }
      const error = monaco.MarkerSeverity.Error
      const warning = monaco.MarkerSeverity.Warning
      const markers: import("monaco-editor").editor.IMarker[] = monaco.editor.getModelMarkers({ resource: model.uri })
      const next = markers
        .filter((marker) => marker.severity === error || marker.severity === warning)
        .slice(0, 8)
        .map((marker) => {
          serial += 1
          return {
            key: `${marker.startLineNumber}-${marker.startColumn}-${serial}`,
            severity: marker.severity === error ? ("error" as const) : ("warning" as const),
            line: marker.startLineNumber,
            message: marker.message,
          }
        })
      setIssues(next)
    }
    const markersListener = monaco.editor.onDidChangeMarkers(() => refresh())
    const modelListener = editor.onDidChangeModel(() => {
      setIssues([])
      refresh()
    })
    editor.onDidDispose(() => {
      markersListener.dispose()
      modelListener.dispose()
    })
    refresh()
  }

  const errors = issues.filter((issue) => issue.severity === "error")
  const warnings = issues.filter((issue) => issue.severity === "warning")

  return (
    <div className="grid min-w-0 gap-2">
      <CodeEditor
        id={id}
        value={value}
        onChange={onChange}
        language="yaml"
        label={label}
        height={height}
        placeholder={placeholder}
        ariaLabel={ariaLabel}
        className={className}
        path={path}
        enableSuggestions
        onMount={handleMount}
      />
      {!isKnownKubernetesKind(kind) ? (
        <p className="text-xs text-muted-foreground">{t("appStore.yamlEditor.unknownKindHint")}</p>
      ) : null}
      {errors.length > 0 ? (
        <IssueList title={t("appStore.yamlEditor.errorTitle")} issues={errors} tone="error" />
      ) : null}
      {warnings.length > 0 ? (
        <IssueList title={t("appStore.yamlEditor.warningTitle")} issues={warnings} tone="warning" />
      ) : null}
    </div>
  )
}

function IssueList({
  title,
  issues,
  tone,
}: {
  title: string
  issues: YamlIssue[]
  tone: "error" | "warning"
}) {
  const { t } = useTranslation()
  return (
    <div
      role={tone === "error" ? "alert" : "status"}
      className={cn(
        "min-w-0 rounded-md border p-2 text-xs",
        tone === "error"
          ? "border-destructive/40 bg-destructive/10 text-destructive"
          : "border-warning-border bg-warning-soft text-warning"
      )}
    >
      <p className="mb-1 font-medium">{title}</p>
      <ul className="grid gap-0.5">
        {issues.map((issue) => (
          <li key={issue.key} className="break-words">
            {t("appStore.yamlEditor.line", { line: issue.line })} · {issue.message}
          </li>
        ))}
      </ul>
    </div>
  )
}
