import { FC, ReactNode } from 'react'

import type { FormInstance } from 'antd'
import type { FormProps, Rule } from 'antd/es/form'
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
            dataProps = {
                type: 'input',
                parentProps: {
                    placeholder: `请输入${label}`
                }
            }
        } = item
        return (
            <Form.Item
                name={name}
                label={label}
                rules={rules}
                key={name}
                className={styles.Item}
            >
                {props[name] || <DataInput {...dataProps} />}
            </Form.Item>
        )
    }

    const renderFormItems = (items: SearchFormItem[]) => {
        return items.map((item) => {
            return (
                <Col xs={12} sm={12} md={8} lg={8} xl={6} xxl={4}>
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
