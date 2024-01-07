import { InputNumber, Select, Space } from 'antd'
import { DefaultOptionType } from 'antd/es/select'
import { SelectProps } from 'antd/lib'
import React, { useEffect } from 'react'

export interface TimeUintInputProps {
    value?: string
    onChange?: (v: string) => void
    placeholder?: [string, string]
    disabled?: boolean
    width?: number | string
    selectProps?: SelectProps
    defaultValue?: string
    style?: React.CSSProperties
    unitOptions?: DefaultOptionType[]
}

type timeUintType = [string?, string?]

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
    const [timeValue, setTimeValue] = React.useState<string>()
    const [unitValue, setUnitValue] = React.useState<string>()

    const handleTimeValueOnChange = (v?: string | null) => {
        setTimeValue(v || undefined)
        if (unitValue && unitValue) {
            onChange?.(v + unitValue)
            return
        }
        if (v) {
            onChange?.(v)
            return
        }
        if (unitValue) {
            onChange?.(unitValue)
            return
        }
    }

    const handleUnitValueOnChange = (v?: string) => {
        setUnitValue(v)
        if (timeValue && v) {
            onChange?.(timeValue + v)
            return
        }
        if (timeValue) {
            onChange?.(timeValue)
            return
        }
        if (v) {
            onChange?.(v)
            return
        }
    }

    const buildTimeAndUnit = (v?: string | number): [string?, string?] => {
        let res: timeUintType = [undefined, undefined]
        if (!v) {
            return res
        }
        if (typeof v !== 'string') {
            v = v + ''
        }
        const u = v?.charAt(v.length - 1)
        const index = unitOptions?.findIndex((item) => item.value === u) ?? -1
        if (index !== -1) {
            res[1] = u
        }
        if (index !== -1) {
            res[0] = v.slice(0, -1)
        }
        if (index === -1 && v.length > 1) {
            res[0] = v
        }

        return res
    }

    useEffect(() => {
        if (!value) {
            return
        }
        const [time, unit] = buildTimeAndUnit(value)
        setTimeValue(time)
        setUnitValue(unit)
    }, [value])

    useEffect(() => {
        const [time, unit] = buildTimeAndUnit(defaultValue)
        setTimeValue(time)
        setUnitValue(unit)
    }, [defaultValue])

    return (
        <Space.Compact style={{ ...style, width }} defaultValue={defaultValue}>
            <InputNumber
                placeholder={placeholder?.[0]}
                style={{ width: '80%' }}
                disabled={disabled}
                value={timeValue}
                onChange={handleTimeValueOnChange}
            />
            <Select
                {...selectProps}
                options={unitOptions}
                style={{ width: '20%', minWidth: 80 }}
                disabled={disabled}
                value={unitValue}
                onChange={handleUnitValueOnChange}
                placeholder={placeholder?.[1]}
                allowClear
            />
        </Space.Compact>
    )
}

export default TimeUintInput
