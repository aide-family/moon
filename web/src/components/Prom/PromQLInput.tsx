import React, { useContext, useEffect, useRef, useState } from 'react'

import {
    EditorView,
    highlightSpecialChars,
    keymap,
    placeholder,
    ViewUpdate
} from '@codemirror/view'
import { Compartment, EditorState, Prec } from '@codemirror/state'
import {
    bracketMatching,
    indentOnInput,
    syntaxHighlighting
} from '@codemirror/language'
import {
    defaultKeymap,
    history,
    historyKeymap,
    insertNewlineAndIndent
} from '@codemirror/commands'

import { highlightSelectionMatches } from '@codemirror/search'
import { lintKeymap } from '@codemirror/lint'
import {
    autocompletion,
    closeBrackets,
    closeBracketsKeymap,
    completionKeymap
} from '@codemirror/autocomplete'
import { PromQLExtension } from '@prometheus-io/codemirror-promql'
import { newCompleteStrategy } from '@prometheus-io/codemirror-promql/dist/esm/complete'
import {
    baseTheme,
    darkPromqlHighlighter,
    darkTheme,
    lightTheme,
    promqlHighlighter
} from './CMTheme'
import { GlobalContext } from '@/context'
import { HistoryCompleteStrategy } from '@/components/Prom/HistoryCompleteStrategy'
import { Button, Form, Input } from 'antd'

import PromValueModal from '@/components/Prom/PromValueModal'
import { ThunderboltOutlined } from '@ant-design/icons'

import type { ValidateStatus } from 'antd/es/form/FormItem'

import styles from './style/index.module.less'

export type PromValidate = {
    help?: string
    validateStatus?: ValidateStatus
}

export interface PromQLInputProps {
    pathPrefix: string
    onChange?: (expression?: string) => void
    formatExpression?: boolean
    placeholderString?: string
    value?: string
    defaultValue?: string
    disabled?: boolean
}

const promqlExtension = new PromQLExtension()
const dynamicConfigCompartment = new Compartment()

const buildPathPrefix = (s: string) => {
    // 去除末尾/
    const promPathPrefix = s?.replace(/\/$/, '')
    return promPathPrefix
}

export const formatExpressionFunc = (pathPrefix: string, doc?: string) => {
    if (!doc || !pathPrefix || pathPrefix === '') {
        return Promise.reject('empty expression')
    }
    return fetch(
        `${buildPathPrefix(
            pathPrefix
        )}/api/v1/format_query?${new URLSearchParams({
            query: doc || ''
        })}`,
        {
            cache: 'no-store',
            credentials: 'same-origin'
        }
    )
        .then((resp) => {
            if (!resp.ok && resp.status !== 400) {
                return Promise.reject(
                    `format HTTP request failed: ${resp.statusText}`
                )
            }

            return resp.json()
        })
        .then(
            (json: {
                data: string
                status: 'success' | 'error'
                error: string
                errorType: string
            }) => {
                if (json.status !== 'success') {
                    return Promise.reject(json.error || 'invalid response JSON')
                }
                return json
            }
        )
        .catch((err) => {
            return Promise.reject(err.toString())
        })
}

const PromQLInput: React.FC<PromQLInputProps> = (props) => {
    const {
        pathPrefix,
        onChange,
        formatExpression,
        placeholderString = 'Please input your PromQL',
        value,
        defaultValue,
        disabled
    } = props

    const { theme } = useContext(GlobalContext)
    const containerRef = useRef<HTMLDivElement>(null)
    const viewRef = useRef<EditorView | null>(null)
    const [doc, setDoc] = useState<string | undefined>(value || defaultValue)
    const onExpressionChange = (expression: string) => {
        setDoc(expression)
    }

    const [isModalVisible, setIsModalVisible] = useState<boolean>(false)
    const { status } = Form.Item.useStatus()

    const handleOnCancelModal = () => {
        setIsModalVisible(false)
    }

    const handleOpenModal = () => {
        setIsModalVisible(true)
    }

    useEffect(() => {
        // if (!pathPrefix) {
        //     return
        // }
        promqlExtension.activateCompletion(true).activateLinter(true)
        if (pathPrefix) {
            promqlExtension.setComplete({
                completeStrategy: new HistoryCompleteStrategy(
                    newCompleteStrategy({
                        remote: {
                            url: pathPrefix
                        }
                    }),
                    []
                )
            })
        }

        let highlighter = syntaxHighlighting(
            theme === 'dark' ? darkPromqlHighlighter : promqlHighlighter
        )
        if (theme === 'dark') {
            highlighter = syntaxHighlighting(darkPromqlHighlighter)
        }

        const dynamicConfig = [
            highlighter,
            promqlExtension.asExtension(),
            theme === 'dark' ? darkTheme : lightTheme
        ]

        const view = viewRef.current
        if (view === null) {
            if (!containerRef.current) {
                throw new Error(
                    'expected CodeMirror container element to exist'
                )
            }
            const startState = EditorState.create({
                doc: doc,
                extensions: [
                    baseTheme,
                    highlightSpecialChars(),
                    history(),
                    EditorState.allowMultipleSelections.of(true),
                    indentOnInput(),
                    bracketMatching(),
                    closeBrackets(),
                    autocompletion(),
                    highlightSelectionMatches(),
                    EditorView.lineWrapping,
                    keymap.of([
                        ...closeBracketsKeymap,
                        ...defaultKeymap,
                        ...historyKeymap,
                        ...completionKeymap,
                        ...lintKeymap
                    ]),
                    placeholder(placeholderString),
                    dynamicConfigCompartment.of(dynamicConfig),
                    keymap.of([
                        {
                            key: 'Escape',
                            run: (v: EditorView): boolean => {
                                v.contentDOM.blur()
                                return false
                            }
                        }
                    ]),
                    Prec.highest(
                        keymap.of([
                            {
                                key: 'Shift-Enter',
                                run: (): boolean => {
                                    return true
                                }
                            },
                            {
                                key: 'Enter',
                                run: insertNewlineAndIndent
                            }
                        ])
                    ),
                    EditorView.updateListener.of((update: ViewUpdate): void => {
                        if (update.docChanged) {
                            onExpressionChange(update.state.doc.toString())
                        }
                    })
                ]
            })
            const view = new EditorView({
                state: startState,
                parent: containerRef.current
            })

            viewRef.current = view
            view?.focus()
        } else {
            view.dispatch(
                view.state.update({
                    effects: dynamicConfigCompartment.reconfigure(dynamicConfig)
                })
            )
        }
    }, [containerRef, pathPrefix])

    useEffect(() => {
        onChange?.(doc)
    }, [doc])

    // if (!pathPrefix) {
    //     return <div>未配置数据源</div>
    // }

    return (
        <>
            <PromValueModal
                visible={isModalVisible}
                onCancel={handleOnCancelModal}
                pathPrefix={pathPrefix}
                expr={doc}
                height={400}
            />
            <div className={styles.promInputContent}>
                {disabled ? (
                    <Input.TextArea
                        ref={containerRef}
                        size="large"
                        defaultValue={defaultValue}
                        value={value}
                        disabled={disabled}
                        placeholder={placeholderString}
                        className="cm-expression-input ant-input"
                    />
                ) : (
                    <div
                        className={'cm-expression-input ' + styles.promInput}
                        style={{
                            borderColor: status === 'error' ? 'red' : ''
                        }}
                        ref={containerRef}
                    />
                )}

                {formatExpression && (
                    <Button
                        onClick={handleOpenModal}
                        type="primary"
                        size="large"
                        style={{
                            borderRadius: '0 6px 6px 0'
                        }}
                        disabled={!doc || status !== 'success'}
                        icon={<ThunderboltOutlined />}
                    />
                )}
            </div>
        </>
    )
}

export default PromQLInput
