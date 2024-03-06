import { FC, ReactNode } from 'react'

import type { FormInstance } from 'antd'
import type { FormProps, Rule } from 'antd/es/form'
import type { DataInputProps } from '../DataInput/DataInput'

import { Form, Space } from 'antd'
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

const inputWidth = 400

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
                {props[name] || <DataInput {...dataProps} width={inputWidth} />}
            </Form.Item>
        )
    }

    const renderFormItems = (items: SearchFormItem[]) => {
        return items.map((item) => {
            return renderFormItem(item)
        })
    }

    return (
        <>
            <Form
                {...formProps}
                layout="horizontal"
                form={form}
                className={styles.SearchForm}
            >
                <Space wrap size={18}>
                    {renderFormItems(items)}
                </Space>
            </Form>
        </>
    )
}

export default SearchForm
