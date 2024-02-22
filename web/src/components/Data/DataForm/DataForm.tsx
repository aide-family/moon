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
    id?: string
}

export type DataFormProps = {
    form?: FormInstance
    items: (DataFormItem[] | DataFormItem)[] | DataFormItem
    formProps?: FormProps
    children?: React.ReactNode
    [key: string]: React.ReactNode | any
}

const renderFormItem = (
    props: DataFormProps,
    item: DataFormItem,
    span: number = 24
) => {
    const {
        name,
        label,
        rules,
        formItemProps,
        id,
        dataProps = {
            type: 'input',
            parentProps: {
                placeholder: `请输入${label}`
            }
        }
    } = item

    return (
        <Col span={span} key={name} id={id}>
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

const renderFormItems = (
    props: DataFormProps,
    items: DataFormItem[],
    span: number = 24
) => {
    return items.map((item) => {
        return renderFormItem(props, item, span)
    })
}

const RenderForm: FC<DataFormProps> = (props) => {
    const { items } = props

    if (Array.isArray(items)) {
        return items.map((item) => {
            const span = 24
            if (Array.isArray(item)) {
                return renderFormItems(props, item, span / item.length)
            }
            return renderFormItem(props, item, span)
        })
    } else {
        return renderFormItem(props, items, 24)
    }
}

const DataForm: FC<DataFormProps> = (props) => {
    const { form, formProps, children } = props

    return (
        <Form {...formProps} form={form}>
            <Row gutter={12}>
                <RenderForm {...props} />
            </Row>
            {children}
        </Form>
    )
}

export default DataForm
