import { SelectProps, Spin } from 'antd'
import Select, { DefaultOptionType } from 'antd/es/select'
import React, { useEffect } from 'react'

export interface FetchSelectProps {
    width?: number | string
    defaultOptions?: DefaultOptionType[]
    handleFetch?: (value: string) => Promise<DefaultOptionType[]>
    selectProps?: SelectProps
    value?: any
    onChange?: (
        value: any,
        option: DefaultOptionType | DefaultOptionType[]
    ) => void
}

let fetchTimeout: NodeJS.Timeout
const FetchSelect: React.FC<FetchSelectProps> = (props) => {
    const {
        value,
        onChange,
        handleFetch,
        width = 400,
        selectProps,
        defaultOptions = []
    } = props
    const [options, setOptions] =
        React.useState<DefaultOptionType[]>(defaultOptions)
    const [loading, setLoading] = React.useState(false)
    const [defaultGroupSelectOpen, setDefaultGroupSelectOpen] =
        React.useState(false)

    const debounceFetcher = (keyword: string) => {
        if (fetchTimeout) {
            clearTimeout(fetchTimeout)
        }
        setLoading(true)
        setDefaultGroupSelectOpen(true)
        fetchTimeout = setTimeout(() => {
            handleFetch?.(keyword)
                .then((items) => {
                    setOptions(items)
                })
                .finally(() => {
                    setLoading(false)
                })
        }, 300)
    }

    useEffect(() => {
        // 判断defaultOptions和option是否相同, 如果不同则更新
        if (defaultOptions.length !== options.length) {
            setOptions(defaultOptions)
        }
    }, [defaultOptions])

    return (
        <>
            <Select
                style={{ width: width }}
                filterOption={false}
                onSearch={debounceFetcher}
                allowClear
                showSearch
                notFoundContent={loading ? <Spin size="small" /> : null}
                onSelect={(v, e) => {
                    setDefaultGroupSelectOpen(false)
                    selectProps?.onSelect?.(v, e)
                }}
                value={value}
                onChange={onChange}
                defaultOpen={defaultGroupSelectOpen}
                loading={loading}
                options={options}
                {...selectProps}
            />
        </>
    )
}

export default FetchSelect
