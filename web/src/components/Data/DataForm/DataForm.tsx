import { FC } from 'react'

import type { FormInstance } from 'antd'
import type { FormItemProps, Rule } from 'antd/es/form'
import type { DataInputProps } from '../DataInput/DataInput'
import type { FormProps } from 'antd/lib'

import { Col, Form, Row } from 'antd'
import { DataInput } from '..'
import React from 'react'

export type DataFormItem = {
    name: string
    label: string
    formItemProps?: FormItemProps
    dataProps?: DataInputProps
    rules?: Rule[]
}

export type DataFormProps = {
    form?: FormInstance
    items: (DataFormItem[] | DataFormItem)[] | DataFormItem
    formProps?: FormProps
    children?: React.ReactNode
    [key: string]: React.ReactNode | any
}

const DataForm: FC<DataFormProps> = (props) => {
    const { form, items, formProps, children } = props

    const renderFormItem = (item: DataFormItem, span: number = 24) => {
        const {
            name,
            label,
            rules,
            formItemProps,
            dataProps = {
                type: 'input',
                parentProps: {
                    placeholder: `请输入${label}`
                }
            }
        } = item

        return (
            <Col span={span} key={name}>
                <Form.Item
                    name={name}
                    label={label}
                    rules={rules}
                    key={name}
                    {...formItemProps}
                >
                    {props[name] || <DataInput {...dataProps} />}
                </Form.Item>
            </Col>
        )
    }

    const renderFormItems = (items: DataFormItem[], span: number = 24) => {
        return items.map((item) => {
            return renderFormItem(item, span)
        })
    }

    const renderForm = () => {
        if (Array.isArray(items)) {
            return items.map((item) => {
                const span = 24
                if (Array.isArray(item)) {
                    return renderFormItems(item, span / item.length)
                }
                return renderFormItem(item, span)
            })
        } else {
            return renderFormItem(items, 24)
        }
    }

    return (
        <Form {...formProps} form={form}>
            <Row gutter={12}>{renderForm()}</Row>
            {children}
        </Form>
    )
}

export default DataForm
