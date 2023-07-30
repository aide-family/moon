import React, {useContext, useEffect, useRef, useState} from 'react'
import {Button, Input} from '@arco-design/web-react';


import {EditorView, highlightSpecialChars, keymap, placeholder, ViewUpdate,} from '@codemirror/view'
import {Compartment, EditorState, Prec} from '@codemirror/state'
import {bracketMatching, indentOnInput, syntaxHighlighting,} from '@codemirror/language'
import {defaultKeymap, history, historyKeymap, insertNewlineAndIndent,} from '@codemirror/commands'

import {highlightSelectionMatches} from '@codemirror/search'
import {lintKeymap} from '@codemirror/lint'
import {autocompletion, closeBrackets, closeBracketsKeymap, completionKeymap,} from '@codemirror/autocomplete'
import {PromQLExtension,} from '@prometheus-io/codemirror-promql'
import {newCompleteStrategy} from '@prometheus-io/codemirror-promql/dist/esm/complete'
import {baseTheme, darkPromqlHighlighter, darkTheme, lightTheme, promqlHighlighter,} from './CMTheme'
import {GlobalContext} from '@/pages/context'
import {HistoryCompleteStrategy} from "@/components/Prom/HistoryCompleteStrategy";

const InputSearch = Input.Search;
const promqlExtension = new PromQLExtension()
const dynamicConfigCompartment = new Compartment()

const pathPrefix = 'https://prometheus.boruixingyunvip1.com'

const Monitor = () => {
    const {theme} = useContext(GlobalContext)
    const containerRef = useRef<HTMLDivElement>(null)
    const viewRef = useRef<EditorView | null>(null)
    const [doc, setDoc] = useState('')
    const onExpressionChange = (expression: string) => {
        setDoc(expression)
    }

    const [formatError, setFormatError] = useState<string | null>(null)
    const [isFormatting, setIsFormatting] = useState<boolean>(false)
    const [exprFormatted, setExprFormatted] = useState<boolean>(false)

    useEffect(() => {
        // Build the dynamic part of the config.
        promqlExtension
            .activateCompletion(true)
            .activateLinter(true)
            .setComplete({
                completeStrategy: new HistoryCompleteStrategy(
                    newCompleteStrategy({
                        remote: {
                            url: pathPrefix,
                        },
                    }),
                    []
                ),
            })

        let highlighter = syntaxHighlighting(
            theme === 'dark' ? darkPromqlHighlighter : promqlHighlighter
        )
        if (theme === 'dark') {
            highlighter = syntaxHighlighting(darkPromqlHighlighter)
        }

        const dynamicConfig = [
            highlighter,
            promqlExtension.asExtension(),
            theme === 'dark' ? darkTheme : lightTheme,
        ]

        const view = viewRef.current
        if (view === null) {
            // If the editor does not exist yet, create it.
            if (!containerRef.current) {
                throw new Error('expected CodeMirror container element to exist')
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
                        ...lintKeymap,
                    ]),
                    placeholder('Expression (press Shift+Enter for newlines)'),
                    dynamicConfigCompartment.of(dynamicConfig),
                    keymap.of([
                        {
                            key: 'Escape',
                            run: (v: EditorView): boolean => {
                                v.contentDOM.blur()
                                return false
                            },
                        },
                    ]),
                    Prec.highest(
                        keymap.of([
                            {
                                key: 'Enter',
                                run: (v: EditorView): boolean => {
                                    // executeQuery()
                                    return true
                                },
                            },
                            {
                                key: 'Shift-Enter',
                                run: insertNewlineAndIndent,
                            },
                        ])
                    ),
                    EditorView.updateListener.of((update: ViewUpdate): void => {
                        if (update.docChanged) {
                            onExpressionChange(update.state.doc.toString())
                            setExprFormatted(false)
                        }
                    }),
                ],
            })
            const view = new EditorView({
                state: startState,
                parent: containerRef.current,
            })

            viewRef.current = view
            view.focus()
        } else {
            view.dispatch(
                view.state.update({
                    effects: dynamicConfigCompartment.reconfigure(dynamicConfig),
                })
            )
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [containerRef])

    const formatExpression = () => {
        setFormatError(null)
        setIsFormatting(true)

        fetch(
            `${pathPrefix}/api/v1/format_query?${new URLSearchParams({
                query: doc,
            })}`,
            {
                cache: 'no-store',
                credentials: 'same-origin',
            }
        )
            .then((resp) => {
                if (!resp.ok && resp.status !== 400) {
                    throw new Error(`format HTTP request failed: ${resp.statusText}`)
                }

                return resp.json()
            })
            .then((json) => {
                if (json.status !== 'success') {
                    throw new Error(json.error || 'invalid response JSON')
                }

                const view = viewRef.current
                if (view === null) {
                    return
                }

                view.dispatch(
                    view.state.update({
                        changes: {from: 0, to: view.state.doc.length, insert: json.data},
                    })
                )
                setExprFormatted(true)
            })
            .catch((err) => {
                setFormatError(err.message)
            })
            .finally(() => {
                setIsFormatting(false)
            })
    }

    return (
        <>
            <div className='arco-input-inner-wrapper arco-input-inner-wrapper-default' style={{width: "600px"}}>
                <div
                    ref={containerRef}
                    className='cm-expression-input arco-input arco-input-size-default'
                    // style={{width: '600px', border: '1px solid #d9d9d9', color: "rgb(var(--arcoblue-6))"}}
                    style={{width: '600px'}}
                ></div>


            </div>
            <Button onClick={formatExpression} type='primary'>
                Format
            </Button>
        </>
    )
}

export default Monitor
