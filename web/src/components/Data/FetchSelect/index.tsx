import { SelectProps, Spin } from 'antd'
import Select, { DefaultOptionType } from 'antd/es/select'
import React, { useEffect } from 'react'

export interface FetchSelectProps {
    width?: number | string
    defaultOptions?: DefaultOptionType[]
    handleFetch?: (value: string) => Promise<DefaultOptionType[]>
    selectProps?: SelectProps
    value?: any
    defaultValue?: any
    onChange?: (
        value: any,
        option: DefaultOptionType | DefaultOptionType[]
    ) => void
}

const FetchSelect: React.FC<FetchSelectProps> = (props) => {
    let fetchTimeout: NodeJS.Timeout
    const {
        value,
        onChange,
        handleFetch,
        width = '100%',
        selectProps,
        defaultValue,
        defaultOptions = []
    } = props
    const [options, setOptions] =
        React.useState<DefaultOptionType[]>(defaultOptions)
    const [loading, setLoading] = React.useState(false)

    const getOptions = (keyword: string) => {
        setLoading(true)
        handleFetch?.(keyword)
            .then((items) => {
                setOptions(items)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    const debounceFetcher = (keyword: string) => {
        if (fetchTimeout) {
            clearTimeout(fetchTimeout)
        }

        fetchTimeout = setTimeout(() => {
            getOptions(keyword)
        }, 300)
    }

    useEffect(() => {
        console.log('=========')
        // TODO 当没有选中时候, 需要更新options
        // TODO 被执行了两次
        getOptions('')
    }, [])

    useEffect(() => {
        // 判断defaultOptions和option是否相同, 如果不同则更新
        if (defaultOptions.length !== options.length) {
            setOptions(defaultOptions)
        }
    }, [defaultOptions])

    return (
        <>
            <Select
                {...selectProps}
                style={{ width: width }}
                filterOption={false}
                onSearch={debounceFetcher}
                allowClear
                showSearch
                notFoundContent={loading ? <Spin size="small" /> : null}
                // onSelect={(v, e) => {
                //     setDefaultGroupSelectOpen(false)
                //     selectProps?.onSelect?.(v, e)
                // }}
                value={value}
                onChange={onChange}
                defaultValue={defaultValue}
                // autoFocus={defaultGroupSelectOpen}
                // defaultOpen={defaultGroupSelectOpen}
                loading={loading}
                options={options}
            />
        </>
    )
}

export default FetchSelect
