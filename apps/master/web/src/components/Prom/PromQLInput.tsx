import React, {useContext, useEffect, useRef, useState} from "react";

import {EditorView, highlightSpecialChars, keymap, placeholder, ViewUpdate} from "@codemirror/view";
import {Compartment, EditorState, Prec} from "@codemirror/state";
import {bracketMatching, indentOnInput, syntaxHighlighting,} from "@codemirror/language";
import {defaultKeymap, history, historyKeymap, insertNewlineAndIndent,} from "@codemirror/commands";

import {highlightSelectionMatches} from "@codemirror/search";
import {lintKeymap} from "@codemirror/lint";
import {autocompletion, closeBrackets, closeBracketsKeymap, completionKeymap,} from "@codemirror/autocomplete";
import {PromQLExtension} from "@prometheus-io/codemirror-promql";
import {newCompleteStrategy} from "@prometheus-io/codemirror-promql/dist/esm/complete";
import {baseTheme, darkPromqlHighlighter, darkTheme, lightTheme, promqlHighlighter,} from "./CMTheme";
import {GlobalContext} from "@/pages/context";
import {HistoryCompleteStrategy} from "@/components/Prom/HistoryCompleteStrategy";
import {Button, Input} from "@arco-design/web-react";
import "./style/index.less";
import {IconThunderbolt} from "@arco-design/web-react/icon";
import PromValueModal from "@/components/Prom/PromValueModal";

export type PromValidate = {
    help: string;
    validateStatus: any;
};

export interface PromQLInputProps {
    pathPrefix: string;
    onChange?: (expression?: string) => void;
    formatExpression?: boolean;
    setPromValidate?: (promValidate?: PromValidate) => void;
    btnDisabled?: boolean;
    placeholderString?: string;
    value?: string;
    defaultValue?: string;
    disabled?: boolean
}

const promqlExtension = new PromQLExtension();
const dynamicConfigCompartment = new Compartment();

export const formatExpressionFunc = (pathPrefix: string, doc?: string) => {
    if (!doc) {
        return Promise.reject("empty expression");
    }
    return fetch(
        `${pathPrefix}/api/v1/format_query?${new URLSearchParams({
            query: doc || "",
        })}`,
        {
            cache: "no-store",
            credentials: "same-origin",
        }
    )
        .then((resp) => {
            if (!resp.ok && resp.status !== 400) {
                return Promise.reject(`format HTTP request failed: ${resp.statusText}`);
            }

            return resp.json();
        })
        .then((json) => {
            if (json.status !== "success") {
                return Promise.reject(json.error || "invalid response JSON");
            }
        })
        .catch((err) => {
            return Promise.reject(err.toString());
        });
};

const PromQLInput: React.FC<PromQLInputProps> = (props) => {
    const {
        pathPrefix,
        onChange,
        formatExpression,
        setPromValidate,
        btnDisabled,
        placeholderString = "Please input your PromQL",
        value,
        defaultValue,
        disabled,
    } = props
    const {theme} = useContext(GlobalContext);
    const containerRef = useRef<HTMLDivElement>(null);
    const viewRef = useRef<EditorView | null>(null);
    const [doc, setDoc] = useState<string | undefined>(value || defaultValue);
    const onExpressionChange = (expression: string) => {
        setDoc(expression);
    };

    const [isFormatLoading, setIsFormatLoading] = useState(false);
    const [isModalVisible, setIsModalVisible] = useState<boolean>(false);

    const handleOnCancelModal = () => {
        setIsModalVisible(false);
    };

    const handleFormatExpression = () => {
        setIsFormatLoading(true);
        formatExpressionFunc(pathPrefix, doc)
            .then(() => {
                setPromValidate?.({
                    validateStatus: "success",
                    help: "Your PromQL is formatted correctly.",
                });
                setIsModalVisible(true);
            })
            .catch((err) => {
                setPromValidate?.({
                    validateStatus: "error",
                    help: err,
                });
            })
            .finally(() => {
                setIsFormatLoading(false);
            });
    };

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
            });

        let highlighter = syntaxHighlighting(
            theme === "dark" ? darkPromqlHighlighter : promqlHighlighter
        );
        if (theme === "dark") {
            highlighter = syntaxHighlighting(darkPromqlHighlighter);
        }

        const dynamicConfig = [
            highlighter,
            promqlExtension.asExtension(),
            theme === "dark" ? darkTheme : lightTheme,
        ];

        const view = viewRef.current;
        if (view === null) {
            // If the editor does not exist yet, create it.
            if (!containerRef.current) {
                throw new Error("expected CodeMirror container element to exist");
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
                    placeholder(placeholderString),
                    dynamicConfigCompartment.of(dynamicConfig),
                    keymap.of([
                        {
                            key: "Escape",
                            run: (v: EditorView): boolean => {
                                v.contentDOM.blur();
                                return false;
                            },
                        },
                    ]),
                    Prec.highest(
                        keymap.of([
                            {
                                key: "Enter",
                                run: (_: EditorView): boolean => {
                                    // executeQuery()
                                    return true;
                                },
                            },
                            {
                                key: "Shift-Enter",
                                run: insertNewlineAndIndent,
                            },
                        ])
                    ),
                    EditorView.updateListener.of((update: ViewUpdate): void => {
                        if (update.docChanged) {
                            onExpressionChange(update.state.doc.toString());
                        }
                    }),
                ],
            });
            const view = new EditorView({
                state: startState,
                parent: containerRef.current,
            });

            viewRef.current = view;
            view.focus();
        } else {
            view.dispatch(
                view.state.update({
                    effects: dynamicConfigCompartment.reconfigure(dynamicConfig),
                })
            );
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [containerRef]);

    useEffect(() => {
        onChange && onChange(doc);
    }, [doc, onChange]);

    return (
        <div
            style={{
                width: "100%",
                display: "flex",
                flexDirection: "row",
                alignItems: "end",
                justifyContent: "space-between",
            }}
        >
            <PromValueModal
                visible={isModalVisible}
                onCancel={handleOnCancelModal}
                pathPrefix={pathPrefix}
                expr={doc}
                height={600}
            />
            {
                disabled &&
                <Input
                    size="large"
                    onChange={onChange}
                    defaultValue={defaultValue}
                    value={value}
                    disabled={disabled}
                    placeholder={placeholderString}
                    className="cm-expression-input arco-input arco-input-size-default"
                />

            }
            <div
                style={{display: disabled ?"none" : ""}}
                ref={containerRef}
                className="cm-expression-input arco-input arco-input-size-default"
            />

            {formatExpression && (
                <Button
                    onClick={handleFormatExpression}
                    type="primary"
                    id="format-btn"
                    size="large"
                    disabled={!doc || btnDisabled}
                    loading={isFormatLoading}
                >
                    <IconThunderbolt/>
                </Button>
            )}
        </div>
    );
};

export default PromQLInput;
