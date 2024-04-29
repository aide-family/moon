import React from 'react'
import { useRef, useState, useEffect } from 'react'

import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import { theme } from 'antd'
import './userWorker'

import './style.css'

export interface WechatTemplateEditorProps {
    value?: string
    defaultValue?: string
    onChange?: (value: string) => void
    width?: number | string
    height?: number | string
}

const { useToken } = theme

const WechatNotifyTemplate = 'json'
const WechatNotifyTemplateTheme = 'wechatNotifyTemplateTheme'

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

const tplMarkdown = `# Moon监控系统告警通知
> 告警状态: {{ .Status }}
> 机器实例: {{ .Labels.instance }}
> 规则名称: {{ .Labels.alertname }}
> 告警描述: {{ .Annotations.summary }}
> 告警详情: {{ .Annotations.description }}
> 告警时间: {{ .StartsAt }}
> 恢复时间: {{ .EndsAt }}
> 链接地址: {{ .GeneratorURL }}
> 告警指纹: {{ .Fingerprint }}
> 当前值: {{ .Value }}`

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
        },
        {
            label: 'tplMarkdown',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: tplMarkdown,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        }
    ]
}

const textTemplate = `{
    "msgtype": "text",
    "text": {
        "content": "\${1:content}",
        "mentioned_list": ["\${2:@all}"],
        "mentioned_mobile_list": ["\${3:@all}"]
    }
}`

const markdownTemplate = `{
    "msgtype": "markdown",
    "markdown": {
        "content": "\${1:content}"
    }
}`

const imageTemplate = `{
    "msgtype": "image",
    "image": {
        "base64": "\${1:base64}",
        "md5": "\${2:md5}"
    }
}`

const newsTemplate = `{
    "msgtype": "news",
    "news": {
        "articles": [
            {
                "title": "\${1:title}",
                "description": "\${2:description}",
                "url": "\${3:url}",
                "picurl": "\${4:picurl}"
            }
            \${5:others}
        ]
    }
}`

const fileTemplate = `{
    "msgtype": "file",
    "file": {
        "media_id": "\${1:media_id}"
    }
}`

const templateCardTemplate = `{
    "msgtype": "template_card",
    "template_card": {
        "card_type": "text_notice",
        "source": {
            "icon_url": "\${1:icon_url}",
            "desc": "\${2:desc}",
            "desc_color": \${3:desc_color}
        },
        "main_title": {
            "title": "\${4:欢迎使用企业微信}",
            "desc": "\${5:您的好友正在邀请您加入企业微信}"
        },
        "emphasis_content": {
            "title": "\${6:欢迎使用企业微信}",
            "desc": "\${7:数据含义}"
        },
        "quote_area": {
            "type": 1,
            "url": "\${8:https://work.weixin.qq.com/wework_admin/frame#index}",
            "appid": "\${9:APPID}",
            "pagepath": "\${10:pagepath}",
            "title": "\${11:引用文本标题}",
            "quote_text": "\${12:Jack: 企业微信真的很好用~, Balian: 超级好的一款软件！}"
        },
        "sub_title_text": "\${13:下载企业微信还能抢红包！}",
        "horizontal_content_list": [
            {
                "keyname": "\${14:邀请人}",
                "value": "\${15:张三}"
            },
            {
                "keyname": "\${16:企微官网}",
                "value": "\${17:点击访问}",
                "type": 1,
                "url": "\${18:https://work.weixin.qq.com/?from=openApi}"
            },
            {
                "keyname": "\${19:企微下载}",
                "value": "\${20:企业微信.apk}",
                "type": 2,
                "media_id": "\${21:MEDIAID}"
            }
        ],
        "jump_list": [
            {
                "type": 1,
                "url": "\${22:https://work.weixin.qq.com/?from=openApi}",
                "title": "\${23:企业微信官网}"
            },
            {
                "type": 2,
                "appid": "\${24:APPID}",
                "pagepath": "\${25:PAGEPATH}",
                "title": "\${26:跳转小程序}"
            }
        ],
        "card_action": {
            "type": 1,
            "url": "\${27:https://work.weixin.qq.com/?from=openApi}",
            "appid": "\${28:APPID}",
            "pagepath": "\${29:PAGEPATH}"
        }
    }
}`

function wechatJsonTemplateProposals(range: monaco.IRange) {
    return [
        {
            label: 'tplText',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: textTemplate,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplMarkdown',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: markdownTemplate,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplImage',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: imageTemplate,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplNews',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: newsTemplate,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplFile',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: fileTemplate,
            insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            range: range
        },
        {
            label: 'tplTemplateCard',
            kind: monaco.languages.CompletionItemKind.Snippet,
            insertText: templateCardTemplate,
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
            suggestions: wechatJsonTemplateProposals(range)
        }
    }

    return {
        suggestions: createDependencyProposals(range)
    }
}

const modelUri = monaco.Uri.parse('./json/wechat.json')

const model = monaco.editor.createModel('', WechatNotifyTemplate, modelUri)

const init = () => {
    monaco.languages.setMonarchTokensProvider(WechatNotifyTemplate, {
        tokenizer: {
            root: [[/\{\{[ ]*\.[ ]*[^}]*[ ]*\}\}/, 'keyword']]
        }
    })

    monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
        validate: false,
        schemas: [
            {
                uri: './json/wechat.json', // id of the first schema
                fileMatch: [modelUri.toString()], // associate with our model
                schema: {
                    type: 'object',
                    properties: {
                        msgtype: {
                            enum: [
                                'text',
                                'markdown',
                                'image',
                                'news',
                                'file',
                                'template_card'
                            ]
                        },
                        text: {
                            type: 'object',
                            properties: {
                                content: {
                                    type: 'string'
                                },
                                mentioned_list: {
                                    type: 'array'
                                },
                                mentioned_mobile_list: {
                                    type: 'array'
                                }
                            }
                        },
                        markdown: {
                            type: 'object',
                            properties: {
                                content: {
                                    type: 'string'
                                }
                            }
                        },
                        image: {
                            type: 'object',
                            properties: {
                                base64: {
                                    type: 'string'
                                },
                                md5: {
                                    type: 'string'
                                }
                            }
                        },
                        news: {
                            type: 'object',
                            properties: {
                                articles: {
                                    type: 'array',
                                    items: {
                                        type: 'object',
                                        properties: {
                                            title: {
                                                type: 'string'
                                            },
                                            description: {
                                                type: 'string'
                                            },
                                            url: {
                                                type: 'string'
                                            },
                                            picurl: {
                                                type: 'string'
                                            }
                                        }
                                    }
                                }
                            }
                        },
                        file: {
                            type: 'object',
                            properties: {
                                media_id: {
                                    type: 'string'
                                }
                            }
                        },
                        template_card: {
                            type: 'object',
                            properties: {
                                card_type: {
                                    type: 'string',
                                    enum: ['text_notice']
                                },
                                source: {
                                    type: 'object',
                                    properties: {
                                        icon_url: {
                                            type: 'string'
                                        },
                                        desc_color: {
                                            type: 'number'
                                        },
                                        desc: {
                                            type: 'string'
                                        }
                                    }
                                },
                                main_title: {
                                    type: 'object',
                                    properties: {
                                        title: {
                                            type: 'string'
                                        },
                                        desc: {
                                            type: 'string'
                                        }
                                    }
                                },
                                emphasis_content: {
                                    type: 'object',
                                    properties: {
                                        title: {
                                            type: 'string'
                                        },
                                        desc: {
                                            type: 'string'
                                        }
                                    }
                                },
                                quote_area: {
                                    type: 'object',
                                    properties: {
                                        type: {
                                            type: 'string'
                                        },
                                        url: {
                                            type: 'string'
                                        },
                                        appid: {
                                            type: 'string'
                                        },
                                        title: {
                                            type: 'string'
                                        },
                                        quote_text: {
                                            type: 'string'
                                        }
                                    }
                                },
                                sub_title_text: {
                                    type: 'string'
                                },
                                horizontal_content_list: {
                                    type: 'array',
                                    items: {
                                        type: 'object',
                                        properties: {
                                            keyname: {
                                                type: 'string'
                                            },
                                            type: {
                                                type: 'number'
                                            },
                                            url: {
                                                type: 'string'
                                            },
                                            value: {
                                                type: 'string'
                                            },
                                            media_id: {
                                                type: 'string'
                                            }
                                        }
                                    }
                                },
                                jump_list: {
                                    type: 'array',
                                    items: {
                                        type: 'object',
                                        properties: {
                                            title: {
                                                type: 'string'
                                            },
                                            type: {
                                                type: 'number'
                                            },
                                            url: {
                                                type: 'string'
                                            },
                                            appid: {
                                                type: 'string'
                                            },
                                            pagepath: {
                                                type: 'string'
                                            }
                                        }
                                    }
                                },
                                card_action: {
                                    type: 'object',
                                    properties: {
                                        type: {
                                            type: 'number'
                                        },
                                        url: {
                                            type: 'string'
                                        },
                                        appid: {
                                            type: 'string'
                                        },
                                        pagepath: {
                                            type: 'string'
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
    monaco.editor.defineTheme(WechatNotifyTemplateTheme, {
        base: 'vs',
        inherit: false,
        rules: [{ token: 'keyword', foreground: 'F55D04', fontStyle: 'bold' }],
        colors: {
            'editor.foreground': '#000000'
        }
    })

    monaco.languages.registerCompletionItemProvider(WechatNotifyTemplate, {
        provideCompletionItems: provideCompletionItems
    })
}

export const WechatTemplateEditor: React.FC<WechatTemplateEditorProps> = (
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
                theme: WechatNotifyTemplateTheme,
                language: WechatNotifyTemplate,
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
