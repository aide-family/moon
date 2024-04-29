import React from 'react'
import { useRef, useState, useEffect } from 'react'

import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import { theme } from 'antd'
import './userWorker'

import './style.css'

export interface FeishuTemplateEditorProps {
    value?: string
    defaultValue?: string
    onChange?: (value: string) => void
    width?: number | string
    height?: number | string
}

const { useToken } = theme

const FeishuTemplate = 'json'
const FeishuTemplateTheme = 'FeishuTemplateTheme'

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
    "msg_type": "text",
    "content": {
        "text": "<at user_id=\"ou_xxx\">Tom</at> 新监控告警提醒\n \${1:alarmContent}"
    }
}`

const tplMarkdown = `{
	"msg_type": "post",
	"content": {
		"post": {
			"zh_cn": {
				"title": "Moon监控告警通知",
				"content": [
					[
                        {
							"tag": "text",
							"text": "\${1:alarmContent}"
						},
						{
							"tag": "a",
							"text": "请查看",
							"href": "\${2:alarmUrl}"
						},
						{
							"tag": "at",
							"user_id": "\${3:userId}"
						}
                        \{$4:alarmExtra}
					]
				]
			}
		}
	}
}`

const tplInteractive = `{
    "msg_type": "interactive",
    "card": {
        "elements": [{
                "tag": "div",
                "text": {
                        "content": "\${1:alarmContent}",
                        "tag": "lark_md"
                }
        }, {
                "actions": [{
                        "tag": "button",
                        "text": {
                                "content": "\${2:进入系统查看}",
                                "tag": "lark_md"
                        },
                        "url": "\${3:alarmUrl}",
                        "type": "default",
                        "value": {}
                }],
                "tag": "action"
        }],
        "header": {
                "title": {
                        "content": "\${4:Moon监控告警通知}",
                        "tag": "plain_text"
                }
        }
    }
}`

function feishuJsonTemplateProposals(range: monaco.IRange) {
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
            label: 'tplMarkdown',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tplMarkdown,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplInteractive',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tplInteractive,
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
            suggestions: feishuJsonTemplateProposals(range)
        }
    }

    return {
        suggestions: createDependencyProposals(range)
    }
}

const modelUri = monaco.Uri.parse('./json/feishu.json')

const model = monaco.editor.createModel('', FeishuTemplate, modelUri)

const i18nJsonSchema = {
    type: 'object',
    properties: {
        title: {
            type: 'string'
        },
        content: {
            type: 'array',
            items: {
                type: 'object',
                properties: {
                    tag: {
                        type: 'string'
                    },
                    text: {
                        type: 'string'
                    },
                    un_escape: {
                        type: 'string'
                    },
                    href: {
                        type: 'string'
                    },
                    user_id: {
                        type: 'string'
                    },
                    user_name: {
                        type: 'string'
                    },
                    image_key: {
                        type: 'string'
                    }
                }
            }
        }
    }
}

const init = () => {
    monaco.languages.setMonarchTokensProvider(FeishuTemplate, {
        tokenizer: {
            root: [[/\{\{[ ]*\.[ ]*[^}]*[ ]*\}\}/, 'keyword']]
        }
    })

    monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
        validate: false,
        schemas: [
            {
                uri: './json/feishu.json', // id of the first schema
                fileMatch: [modelUri.toString()], // associate with our model
                schema: {
                    type: 'object',
                    properties: {
                        msg_type: {
                            enum: [
                                'text',
                                'post',
                                'image',
                                'share_chat',
                                'interactive'
                            ]
                        },
                        content: {
                            type: 'object',
                            properties: {
                                text: {
                                    type: 'string'
                                },
                                share_chat_id: {
                                    type: 'string'
                                },
                                image_key: {
                                    type: 'string'
                                },
                                post: {
                                    type: 'object',
                                    properties: {
                                        zh_cn: {
                                            ...i18nJsonSchema
                                        },
                                        en_us: {
                                            ...i18nJsonSchema
                                        }
                                    }
                                }
                            }
                        },
                        card: {
                            type: 'object',
                            properties: {
                                elements: {
                                    type: 'array',
                                    items: {
                                        type: 'object',
                                        properties: {
                                            tag: {
                                                type: 'string',
                                                enum: [
                                                    'div',
                                                    'at',
                                                    'text',
                                                    'a',
                                                    'img'
                                                ]
                                            },
                                            text: {
                                                type: 'object',
                                                properties: {
                                                    content: {
                                                        type: 'string'
                                                    },
                                                    tag: {
                                                        type: 'string'
                                                    }
                                                }
                                            },
                                            actions: {
                                                type: 'array',
                                                items: {
                                                    type: 'object',
                                                    properties: {
                                                        tag: {
                                                            type: 'string'
                                                        },
                                                        text: {
                                                            type: 'object',
                                                            properties: {
                                                                content: {
                                                                    type: 'string'
                                                                },
                                                                tag: {
                                                                    type: 'string'
                                                                }
                                                            }
                                                        },
                                                        url: {
                                                            type: 'string'
                                                        },
                                                        type: {
                                                            type: 'string',
                                                            enum: ['default']
                                                        },
                                                        value: {
                                                            type: 'object'
                                                        }
                                                    }
                                                }
                                            }
                                        }
                                    }
                                },
                                header: {
                                    type: 'object',
                                    properties: {
                                        title: {
                                            type: 'object',
                                            properties: {
                                                content: {
                                                    type: 'string'
                                                },
                                                tag: {
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
            }
        ]
    })

    // Define a new theme that contains only rules that match this language
    monaco.editor.defineTheme(FeishuTemplateTheme, {
        base: 'vs',
        inherit: false,
        rules: [{ token: 'keyword', foreground: 'F55D04', fontStyle: 'bold' }],
        colors: {
            'editor.foreground': '#000000'
        }
    })

    monaco.languages.registerCompletionItemProvider(FeishuTemplate, {
        provideCompletionItems: provideCompletionItems
    })
}

export const FeishuTemplateEditor: React.FC<FeishuTemplateEditorProps> = (
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
                theme: FeishuTemplateTheme,
                language: FeishuTemplate,
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
