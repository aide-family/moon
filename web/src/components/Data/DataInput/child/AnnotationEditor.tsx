import React, { useContext } from 'react'
import { useRef, useState, useEffect } from 'react'

import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import { GlobalToken, theme } from 'antd'
import './userWorker'

import { GlobalContext, ThemeType } from '@/context'
import { defaultTheme } from './color'

import './style.css'

export interface AnnotationEditorProps {
    value?: string
    defaultValue?: string
    onChange?: (value: string) => void
    width?: number | string
    height?: number | string
    disabled?: boolean
}

const { useToken } = theme

const AnnotationTemplate = 'AnnotationTemplate'
const AnnotationTemplateTheme = 'AnnotationTemplateTheme'

const tpl = '告警实例: {{ \\$labels.instance }}告警，当前值: {{ \\$value }}'

// TODO 根据元数据补充
const keywords: string[] = [
    'app_kubernetes_io_managed_by',
    'chart',
    'component',
    'heritage',
    'instance',
    'job',
    'namespace',
    'node',
    'release',
    'service',
    'app',
    'instance'
]

function createDependencyProposals(range: monaco.IRange) {
    return [
        {
            label: '"labels"',
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: '{{ \\$labels.${1:labelName} }}',
            range: range,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
        },
        {
            label: '"eventsAt"',
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: '{{ \\$eventsAt }}',
            range: range,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
        },
        ...keywords.map((key) => {
            return {
                label: `"${key}"`,
                kind: monaco.languages.CompletionItemKind.Keyword,
                insertText: `{{ \\$labels.${key} }}`,
                range: range,
                insertTextRules:
                    monaco.languages.CompletionItemInsertTextRule
                        .InsertAsSnippet
            }
        }),
        {
            label: '"value"',
            kind: monaco.languages.CompletionItemKind.Function,
            insertText: '{{ \\$value }}',
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

const createKeyDependencyProposals = (range: monaco.IRange) => {
    return keywords.map((key) => {
        return {
            label: key,
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: key,
            range: range,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
        }
    })
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

    // 如果都匹配，返回空建议（这种情况理论上不应该发生，除非正则表达式有误）
    return {
        suggestions: [
            ...createDependencyProposals(range),
            ...createKeyDependencyProposals(range)
        ]
    }
}

const init = (token: GlobalToken, theme?: ThemeType) => {
    monaco.languages.register({ id: AnnotationTemplate })
    // Register a tokens provider for the language
    monaco.languages.setMonarchTokensProvider(AnnotationTemplate, {
        tokenizer: {
            root: [[/\{\{[ ]*\$[ ]*[^}]*[ ]*\}\}/, 'keyword']]
        }
    })

    // Define a new theme that contains only rules that match this language
    monaco.editor.defineTheme(
        AnnotationTemplateTheme,
        defaultTheme(token, theme)
    )

    monaco.languages.registerCompletionItemProvider(AnnotationTemplate, {
        provideCompletionItems: provideCompletionItems
    })
}

export const AnnotationEditor: React.FC<AnnotationEditorProps> = (props) => {
    const {
        value,
        defaultValue,
        onChange,
        disabled,
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
                if (!editor.getValue()) {
                    editor.setValue(value || defaultValue || '')
                }
                editor.updateOptions({
                    readOnly: disabled
                })
                editor.layout()
                return editor
            }

            const curr = monacoEl.current!
            const e = monaco.editor.create(curr, {
                theme: AnnotationTemplateTheme,
                language: AnnotationTemplate,
                value: value || defaultValue,
                lineNumbers: 'off',
                // 展示行号和内容的边框
                lineNumbersMinChars: 4,
                readOnly: disabled,
                minimap: {
                    enabled: false
                }
            })

            return e
        })
    }, [defaultValue, editor, monacoEl, onChange, value])

    useEffect(() => {
        init(token, sysTheme)
        if (editor) {
            editor.onDidChangeModelContent(() => {
                onChange?.(editor.getValue())
            })
        }
    }, [])

    useEffect(() => {
        if (editor) {
            editor.updateOptions({
                readOnly: disabled
            })
        }
    }, [disabled, editor])

    return (
        <div
            style={{
                width: width,
                height: height,
                borderColor: token.colorBorder
                // 设置disabled
                // pointerEvents: disabled ? 'none' : 'auto'
            }}
            className="editorInput"
            ref={monacoEl}
        />
    )
}
