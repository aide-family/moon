import { AutoComplete, Input, Tag } from 'antd'
import React, { useRef, useState } from 'react'
import { TextAreaProps } from 'antd/es/input'
import { TextAreaRef } from 'antd/es/input/TextArea'
import { autoCompleteOption } from './options'

const { TextArea } = Input

export interface TemplateAutoCompleteProps {
    autoCompleteProps?: TextAreaProps
    onChange?: (value?: string) => void
    value?: string
    defaultValue?: string
    placeholder?: string
    style?: React.CSSProperties
}

export const TemplateAutoComplete: React.FC<TemplateAutoCompleteProps> = (
    props
) => {
    const { value, defaultValue, onChange, placeholder, autoCompleteProps } =
        props
    const inputRef = useRef<TextAreaRef>(null)
    const [cursorPosition, setCursorPosition] = useState(0)

    const [options, setOptions] = useState<{ value: string }[]>([])
    const [currStr, setCurrStr] = useState('')

    const handleSearch = (value: string) => {
        onChange?.(value)
        let curr =
            inputRef.current?.resizableTextArea?.textArea.selectionStart || 0
        let last = value.slice(0, curr)
        let currStrList = last.split(' ')
        last = currStrList[currStrList.length - 1]

        currStrList = last.split('\n')
        last = currStrList.at(-1) || ''

        currStrList = last.split('}')
        last = currStrList.at(-1) || ''

        setCurrStr(last)
        setCursorPosition(curr || 0)
        if (last === '') {
            setOptions([])
            return
        }
        setOptions(
            !value
                ? []
                : autoCompleteOption
                      .filter((item) => item.value.includes(last))
                      .map((item) => ({
                          label: (
                              <Tag color="red">
                                  {item.value.slice(4, item.value.length - 3)}
                              </Tag>
                          ),
                          value: item.value
                      }))
        )
    }

    const onSelect = (selectVal: string) => {
        const prefix =
            value?.slice(0, cursorPosition - currStr.length) + selectVal
        onChange?.(prefix + value?.slice(cursorPosition))
        setCursorPosition(prefix.length)
        setCurrStr('')
        setOptions([])
    }

    return (
        <AutoComplete
            options={options}
            onSelect={onSelect}
            onSearch={handleSearch}
            defaultActiveFirstOption
            value={value}
            defaultValue={defaultValue}
        >
            <TextArea
                {...autoCompleteProps}
                ref={inputRef}
                placeholder={placeholder}
            />
        </AutoComplete>
    )
}
