import React, { useContext } from 'react'
import { useRef, useState, useEffect } from 'react'

import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import { GlobalToken, theme } from 'antd'
import './userWorker'

import './style.css'
import { GlobalContext, ThemeType } from '@/context'
import { defaultTheme } from './color'

export interface JsonTemplateEditorProps {
    value?: string
    defaultValue?: string
    onChange?: (value: string) => void
    width?: number | string
    height?: number | string
}

const { useToken } = theme

const JsonTemplate = 'json'
const JsonTemplateTheme = 'JsonTemplateTheme'

const tpl = `Moon监控系统告警通知
告警状态: {{ .Status }}
机器实例: {{ .Labels.instance }}
规则名称: {{ .Labels.alertname }}
告警描述: {{ .Annotations.summary }}
告警详情: {{ .Annotations.description }}
告警时间: {{ .StartsAt }}
恢复时间: {{ .EndsAt }}
链接地址: {{ .GeneratorURL }}
告警指纹: {{ .Fingerprint }}
当前值: {{ .Value }}`

function createDependencyProposals(range: monaco.IRange) {
    return [
        {
            label: '"Labels"',
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: '{{ .Labels.${1:labelName} }}',
            range: range,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
        },
        {
            label: '"Annotations"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: '{{ .Annotations.${1:annotationName} }}',
            range: range,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
        },
        {
            label: '"summary"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: 'summary',
            range: range
        },
        {
            label: '"description"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: 'description',
            range: range
        },
        {
            label: '"Status"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: '{{ .Status }}',
            range: range
        },
        {
            label: '"StartsAt"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: '{{ .StartsAt }}',
            range: range
        },
        {
            label: '"EndsAt"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: '{{ .EndsAt }}',
            range: range
        },
        {
            label: '"GeneratorURL"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: '{{ .GeneratorURL }}',
            range: range
        },
        {
            label: '"Fingerprint"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: '{{ .Fingerprint }}',
            range: range
        },
        {
            label: '"Value"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: '{{ .Value }}',
            range: range
        },

        {
            label: 'tpl',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tpl,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        }
    ]
}

const provideCompletionItems = (
    model: monaco.editor.ITextModel,
    position: monaco.Position
) => {
    const word = model.getWordUntilPosition(position)

    const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn
    }
    return {
        suggestions: createDependencyProposals(range)
    }
}

const modelUri = monaco.Uri.parse('./json/other.json')

const model = monaco.editor.createModel('', JsonTemplate, modelUri)

const init = (token: GlobalToken, theme?: ThemeType) => {
    monaco.languages.setMonarchTokensProvider(JsonTemplate, {
        tokenizer: {
            root: [[/\{\{[ ]*\.[ ]*[^}]*[ ]*\}\}/, 'keyword']]
        }
    })

    monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
        validate: false,
        schemas: [
            {
                uri: './json/other.json', // id of the first schema
                fileMatch: [modelUri.toString()], // associate with our model
                schema: {
                    type: 'object',
                    properties: {}
                }
            }
        ]
    })

    // Define a new theme that contains only rules that match this language
    monaco.editor.defineTheme(JsonTemplateTheme, defaultTheme(token, theme))

    monaco.languages.registerCompletionItemProvider(JsonTemplate, {
        provideCompletionItems: provideCompletionItems
    })
}

export const JsonTemplateEditor: React.FC<JsonTemplateEditorProps> = (
    props
) => {
    const {
        value,
        defaultValue,
        onChange,
        width = '100%',
        height = '100%'
    } = props

    const { token } = useToken()
    const { sysTheme } = useContext(GlobalContext)

    const [editor, setEditor] =
        useState<monaco.editor.IStandaloneCodeEditor | null>(null)
    const monacoEl = useRef(null)

    useEffect(() => {
        setEditor((editor) => {
            if (editor) {
                return editor
            }

            const curr = monacoEl.current!
            const e = monaco.editor.create(curr, {
                model: model,
                theme: JsonTemplateTheme,
                language: JsonTemplate,
                value: value || defaultValue,
                // 展示行号和内容的边框
                lineNumbersMinChars: 4,
                minimap: {
                    // enabled: false
                    size: 'fit'
                }
            })
            e.onDidChangeModelContent(() => {
                onChange?.(e.getValue())
            })
            return e
        })
    }, [defaultValue, editor, monacoEl, onChange, value])

    useEffect(() => {
        init(token, sysTheme)
    }, [])

    return (
        <div
            style={{
                width: width,
                height: height,
                borderColor: token.colorBorder
            }}
            className="editorInput"
            ref={monacoEl}
        />
    )
}
