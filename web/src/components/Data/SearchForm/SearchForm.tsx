import { FC, ReactNode } from 'react'

import type { FormInstance } from 'antd'
import type { FormItemProps, FormProps, Rule } from 'antd/es/form'
import type { DataInputProps } from '../DataInput/DataInput'

import { Form, Col, Row } from 'antd'
import DataInput from '../DataInput/DataInput'
// import { GlobalContext } from '@/context'

import styles from '../style/data.module.less'

export type SearchFormItem = {
    name: string
    label: string
    dataProps?: DataInputProps
    rules?: Rule[]
    formItemProps?: FormItemProps
    id?: React.Key
}

export type SearchFormProps = {
    form: FormInstance
    items?: SearchFormItem[]
    formProps?: FormProps
    onClear?: () => void
    [key: string]: ReactNode | any
}

// let timerId: NodeJS.Timeout | null = null
const SearchForm: FC<SearchFormProps> = (props) => {
    const { form, items = [], formProps } = props

    const renderFormItem = (item: SearchFormItem) => {
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
            <Form.Item
                {...formItemProps}
                name={name}
                label={label}
                rules={rules}
                key={name + label}
                className={styles.Item}
            >
                {props[name] || <DataInput {...dataProps} />}
            </Form.Item>
        )
    }

    const renderFormItems = (items: SearchFormItem[]) => {
        return items.map((item, index) => {
            return (
                <Col key={index} xs={24} sm={24} md={12} lg={12} xl={8} xxl={6}>
                    {renderFormItem(item)}
                </Col>
            )
        })
    }

    return (
        <>
            <Form
                {...formProps}
                layout="vertical"
                form={form}
                className={styles.SearchForm}
            >
                <Row gutter={16}>{renderFormItems(items)}</Row>
            </Form>
        </>
    )
}

export default SearchForm
