import { InputNumber, Select, Space } from 'antd'
import { DefaultOptionType } from 'antd/es/select'
import { SelectProps } from 'antd/lib'
import React, { useEffect } from 'react'

export interface TimeUintInputProps {
    value?: string
    onChange?: (v: string) => void
    onBlur?: (v: string) => void
    onFocus?: (v: string) => void
    placeholder?: [string, string]
    disabled?: boolean
    width?: number | string
    selectProps?: SelectProps
    defaultValue?: string
    style?: React.CSSProperties
    unitOptions?: DefaultOptionType[]
}

export const TimeUintInput: React.FC<TimeUintInputProps> = (props) => {
    const {
        value,
        onChange,
        placeholder,
        disabled,
        width,
        selectProps,
        defaultValue,
        style,
        unitOptions
    } = props
    const [timeValue, setTimeValue] = React.useState<string>('')
    const [unitValue, setUnitValue] = React.useState<string>('')

    const handleTimeValueOnChange = (v: string | null) => {
        if (!v) {
            setTimeValue('')
            return
        }
        setTimeValue(v)
        onChange?.(v + unitValue)
    }

    const handleUnitValueOnChange = (v: string) => {
        setUnitValue(v)
        onChange?.(timeValue + v)
    }

    const buildTimeAndUnit = (v?: string): [string, string] => {
        if (!v) {
            return ['', '']
        }
        const unit =
            unitOptions?.find(
                (item) => item.value === v?.substring(v.length - 1)
            )?.value || ''
        let time = value?.substring(0, value.length - 1) || ''
        if (!unit) {
            time = v
        }
        return [time.toString(), unit.toString()]
    }

    useEffect(() => {
        const [time, unit] = buildTimeAndUnit(value)
        setTimeValue(time.toString())
        setUnitValue(unit.toString())
    }, [value])

    useEffect(() => {
        const [time, unit] = buildTimeAndUnit(defaultValue)
        setTimeValue(time.toString())
        setUnitValue(unit.toString())
    }, [defaultValue])

    return (
        <Space.Compact style={{ ...style, width }} defaultValue={defaultValue}>
            <InputNumber
                placeholder={placeholder?.[0]}
                style={{ width: '80%' }}
                disabled={disabled}
                value={timeValue}
                onChange={handleTimeValueOnChange}
                onBlur={() => {
                    props.onBlur?.(timeValue + unitValue)
                }}
                onFocus={() => {
                    props.onFocus?.(timeValue + unitValue)
                }}
            />
            <Select
                {...selectProps}
                options={unitOptions}
                style={{ width: '20%', minWidth: 80 }}
                disabled={disabled}
                value={unitValue}
                onChange={handleUnitValueOnChange}
                onBlur={() => {
                    props.onBlur?.(timeValue + unitValue)
                }}
                onFocus={() => {
                    props.onFocus?.(timeValue + unitValue)
                }}
                placeholder={placeholder?.[1]}
            />
        </Space.Compact>
    )
}

export default TimeUintInput
