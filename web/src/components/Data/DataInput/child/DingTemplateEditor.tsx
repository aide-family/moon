import React from 'react'
import { useRef, useState, useEffect } from 'react'

import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import { theme } from 'antd'
import './userWorker'

import './style.css'

export interface DingTemplateEditorProps {
    value?: string
    defaultValue?: string
    onChange?: (value: string) => void
    width?: number | string
    height?: number | string
}

const { useToken } = theme

const DingTemplate = 'json'
const DingTemplateTheme = 'DingTemplateTheme'

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

const tplText = `{
    "at": {
        "atMobiles":[
            \${1:atMobile}
        ],
        "atUserIds":[
            \${2:atUserId}
        ],
        "isAtAll": \${3:false}
    },
    "text": {
        "content":"\${4:content}"
    },
    "msgtype":"text"
}`

const tplLink = `{
    "msgtype":"link",
    "link": {
        "text": "\${1:text}",
        "title": "\${2:title}",
        "picUrl": "\${3:picUrl}",
        "messageUrl": "\${4:messageUrl}"
    }
}`

const tplMarkdown = `{
    "msgtype":"markdown",
    "markdown": {
        "title": "\${1:title}",
        "text": "\${2:text}"
    }
}`

const tplActionCard = `{
    "msgtype":"actionCard",
    "actionCard": {
        "title": "\${1:title}",
        "btnOrientation": "0",
        "singleTitle": "\${2:singleTitle}",
        "singleURL": "\${3:singleURL}",
        "btns": [
            {
                "title": "\${4:title}",
                "actionURL": "\${5:actionURL}"
            },
            {
                "title": "\${6:title}",
                "actionURL": "\${7:actionURL}"
            }
        ]
    }
}`

const tplFeedCard = `{
    "msgtype":"feedCard",
    "feedCard": {
        "links": [
            {
                "title": "\${1:title}",
                "messageURL": "\${2:messageURL}",
                "picURL": "\${3:picURL}"
            },
            {
                "title": "\${4:title}",
                "messageURL": "\${5:messageURL}",
                "picURL": "\${6:picURL}"
            }
        ]
    }
}`

function dingJsonTemplateProposals(range: monaco.IRange) {
    return [
        {
            label: 'tplText',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tplText,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplLink',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tplLink,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplMarkdown',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tplMarkdown,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplActionCard',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tplActionCard,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplFeedCard',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tplFeedCard,
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
    const extUntilPosition = model.getValueInRange({
        startLineNumber: 1,
        startColumn: 1,
        endLineNumber: position.lineNumber,
        endColumn: position.column
    })

    // 匹配json格式
    const reg = /\{\s*|\s*\}/
    const match = extUntilPosition.match(reg)
    const word = model.getWordUntilPosition(position)
    const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn
    }
    if (!match) {
        return {
            suggestions: dingJsonTemplateProposals(range)
        }
    }

    return {
        suggestions: createDependencyProposals(range)
    }
}

const modelUri = monaco.Uri.parse('./json/ding.json')

const model = monaco.editor.createModel('', DingTemplate, modelUri)

const init = () => {
    monaco.languages.setMonarchTokensProvider(DingTemplate, {
        tokenizer: {
            root: [[/\{\{[ ]*\.[ ]*[^}]*[ ]*\}\}/, 'keyword']]
        }
    })

    monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
        validate: false,
        schemas: [
            {
                uri: './json/ding.json', // id of the first schema
                fileMatch: [modelUri.toString()], // associate with our model
                schema: {
                    type: 'object',
                    properties: {
                        msgtype: {
                            type: 'string',
                            enum: [
                                'text',
                                'link',
                                'markdown',
                                'actionCard',
                                'feedCard'
                            ]
                        },
                        text: {
                            type: 'object',
                            properties: {
                                content: {
                                    type: 'string'
                                }
                            }
                        },
                        at: {
                            type: 'object',
                            properties: {
                                atMobiles: {
                                    type: 'array',
                                    items: {
                                        type: 'string'
                                    }
                                },
                                atUserIds: {
                                    type: 'array',
                                    items: {
                                        type: 'string'
                                    }
                                },
                                isAtAll: {
                                    type: 'boolean',
                                    enum: [true, false]
                                }
                            }
                        },
                        link: {
                            type: 'object',
                            properties: {
                                title: {
                                    type: 'string'
                                },
                                text: {
                                    type: 'string'
                                },
                                picUrl: {
                                    type: 'string'
                                },
                                messageUrl: {
                                    type: 'string'
                                }
                            }
                        },
                        markdown: {
                            type: 'object',
                            properties: {
                                title: {
                                    type: 'string'
                                },
                                text: {
                                    type: 'string'
                                }
                            }
                        },
                        actionCard: {
                            type: 'object',
                            properties: {
                                title: {
                                    type: 'string'
                                },
                                text: {
                                    type: 'string'
                                },
                                btnOrientation: {
                                    type: 'string'
                                },
                                singleTitle: {
                                    type: 'string'
                                },
                                singleURL: {
                                    type: 'string'
                                },
                                btns: {
                                    type: 'array',
                                    items: {
                                        type: 'object',
                                        properties: {
                                            title: {
                                                type: 'string'
                                            },
                                            actionURL: {
                                                type: 'string'
                                            }
                                        }
                                    }
                                }
                            }
                        },
                        feedCard: {
                            type: 'object',
                            properties: {
                                links: {
                                    type: 'array',
                                    items: {
                                        type: 'object',
                                        properties: {
                                            title: {
                                                type: 'string'
                                            },
                                            messageURL: {
                                                type: 'string'
                                            },
                                            picURL: {
                                                type: 'string'
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        ]
    })

    // Define a new theme that contains only rules that match this language
    monaco.editor.defineTheme(DingTemplateTheme, {
        base: 'vs',
        inherit: false,
        rules: [{ token: 'keyword', foreground: 'F55D04', fontStyle: 'bold' }],
        colors: {
            'editor.foreground': '#000000'
        }
    })

    monaco.languages.registerCompletionItemProvider(DingTemplate, {
        provideCompletionItems: provideCompletionItems
    })
}

export const DingTemplateEditor: React.FC<DingTemplateEditorProps> = (
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
                theme: DingTemplateTheme,
                language: DingTemplate,
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
        init()
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
