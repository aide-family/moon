import { Form, InputNumber, Select, Space } from 'antd'
import { DefaultOptionType } from 'antd/es/select'
import { SelectProps } from 'antd/lib'
import React from 'react'

export interface TimeUintInputProps {
    name: string
    placeholder?: [string, string]
    disabled?: boolean
    width?: number | string
    selectProps?: SelectProps
    style?: React.CSSProperties
    unitOptions?: DefaultOptionType[]
}

export const TimeUintInput: React.FC<TimeUintInputProps> = (props) => {
    const {
        name,
        placeholder,
        disabled,
        width,
        selectProps,
        style,
        unitOptions
    } = props

    return (
        <Space.Compact style={{ ...style, width }}>
            <Form.Item name={[name, 'value']} noStyle>
                <InputNumber
                    placeholder={placeholder?.[0]}
                    style={{ width: '80%' }}
                    disabled={disabled}
                />
            </Form.Item>

            <Form.Item name={[name, 'unit']} noStyle>
                <Select
                    {...selectProps}
                    options={unitOptions}
                    style={{ width: '20%', minWidth: 80 }}
                    disabled={disabled}
                    placeholder={placeholder?.[1]}
                    allowClear
                />
            </Form.Item>
        </Space.Compact>
    )
}

export default TimeUintInput
