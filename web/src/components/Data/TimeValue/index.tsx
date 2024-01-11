import { Duration } from '@/apis/types'
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

export const checkDuration = (
    title: string = '时间',
    isForce: boolean = false
) => {
    const validator = (_: any, value: Duration) => {
        if (!isForce && (!value || (!value.value && !value.unit))) {
            // 不强制要求必须输入
            return Promise.resolve()
        }

        // 如果输入了, 必须是完整的数据格式
        if (!value) {
            return Promise.reject(new Error(`${title}不能为空`))
        }
        if (!value.value) {
            return Promise.reject(new Error(`${title}不能为空`))
        }
        if (+value.value < 1) {
            return Promise.reject(new Error(`${title}必须大于1`))
        }
        if (+value.value % 1 !== 0) {
            return Promise.reject(new Error(`${title}必须为整数`))
        }
        if (!value.unit) {
            return Promise.reject(new Error(`${title}必须有单位`))
        }
        return Promise.resolve()
    }
    return validator
}

export const TimeUintInput: React.FC<TimeUintInputProps> = (props) => {
    const {
        name,
        placeholder,
        disabled,
        width = '100%',
        selectProps,
        style,
        unitOptions
    } = props

    return (
        <Space.Compact style={{ ...style, width }}>
            <Form.Item name={[name, 'value']} noStyle>
                <InputNumber
                    placeholder={placeholder?.[0]}
                    style={{ width: '70%' }}
                    disabled={disabled}
                />
            </Form.Item>

            <Form.Item name={[name, 'unit']} noStyle>
                <Select
                    {...selectProps}
                    options={unitOptions}
                    style={{ width: '30%', minWidth: 80 }}
                    disabled={disabled}
                    placeholder={placeholder?.[1]}
                    allowClear
                />
            </Form.Item>
        </Space.Compact>
    )
}

export default TimeUintInput
