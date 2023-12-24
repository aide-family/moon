import { useContext, useEffect, useState, FC } from 'react'

import type { FormInstance } from 'antd'
import type { FormProps, Rule } from 'antd/es/form'
import type { DataInputProps } from '../DataInput/DataInput'

import { Form, Space } from 'antd'
import DataInput from '../DataInput/DataInput'
import { GlobalContext } from '@/context'

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
}

let timerId: NodeJS.Timeout | null = null
const SearchForm: FC<SearchFormProps> = (props) => {
    const { form, items = [], formProps } = props
    const { layoutContentElement } = useContext(GlobalContext)
    const [inputWidth, setInputWidth] = useState<number>(400)

    const resizeObserver = new ResizeObserver((entries) => {
        // 清除上一次的定时器
        if (timerId) {
            clearTimeout(timerId)
        }

        timerId = setTimeout(() => {
            let gap = 8
            let inputWidth = 300

            for (const entry of entries) {
                const { width } = entry.contentRect
                if (width < 300) {
                    setInputWidth(300)
                    return
                }
                let remainder = width % (inputWidth + gap)
                for (let i = 300; i <= 400; i++) {
                    let remainderTmp = width % (i + gap)
                    if (remainderTmp < remainder) {
                        remainder = remainderTmp
                        inputWidth = i
                    }
                }
            }
            setInputWidth(inputWidth)
        }, 500)
    })

    useEffect(() => {
        if (layoutContentElement) {
            resizeObserver.observe(layoutContentElement)
        }
        return () => {
            resizeObserver.disconnect()
        }
    }, [layoutContentElement])

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
                <DataInput {...dataProps} width={inputWidth} />
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
