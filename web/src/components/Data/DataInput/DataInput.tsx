import { FC } from 'react'

import type {
    CheckboxProps,
    DatePickerProps,
    InputProps,
    RadioGroupProps,
    RadioProps,
    SelectProps
} from 'antd'
import { Input, Select, Radio, Checkbox, DatePicker } from 'antd'
import { PasswordProps, TextAreaProps } from 'antd/lib/input'

export type DataInputProps =
    | (
          | {
                type: 'input'
                parentProps?: InputProps
            }
          | {
                type: 'password'
                parentProps?: PasswordProps
            }
          | {
                type: 'select'
                parentProps?: SelectProps
            }
          | {
                type: 'radio'
                parentProps?: RadioProps
            }
          | {
                type: 'radio-group'
                parentProps?: RadioGroupProps
            }
          | {
                type: 'checkbox'
                parentProps?: CheckboxProps
            }
          | {
                type: 'date'
                parentProps?: DatePickerProps
            }
          | {
                type: 'textarea'
                parentProps?: TextAreaProps
            }
      ) & {
          width?: number | string
          value?: any
          onChange?: (value: any) => void
      }

const DataInput: FC<DataInputProps> = (props) => {
    const { type, parentProps, width, value, onChange } = props

    const renderInput = () => {
        switch (type) {
            case 'select':
                return (
                    <Select
                        allowClear
                        {...parentProps}
                        style={{ width }}
                        value={value}
                        onChange={onChange}
                    />
                )
            case 'radio':
                return (
                    <Radio
                        {...parentProps}
                        style={{ width }}
                        value={value}
                        onChange={onChange}
                    />
                )
            case 'checkbox':
                return (
                    <Checkbox
                        {...parentProps}
                        style={{ width }}
                        value={value}
                        onChange={onChange}
                    />
                )
            case 'date':
                return (
                    <DatePicker
                        allowClear
                        {...parentProps}
                        style={{ width }}
                        value={value}
                        onChange={onChange}
                    />
                )
            case 'radio-group':
                return (
                    <Radio.Group
                        {...parentProps}
                        style={{ width }}
                        value={value}
                        onChange={onChange}
                    />
                )
            case 'password':
                return (
                    <Input.Password
                        autoComplete="off"
                        allowClear
                        {...parentProps}
                        style={{ width }}
                        value={value}
                        onChange={onChange}
                    />
                )
            case 'textarea':
                return (
                    <Input.TextArea
                        autoComplete="off"
                        allowClear
                        {...parentProps}
                        value={value}
                        onChange={onChange}
                    />
                )
            default:
                return (
                    <Input
                        autoComplete="off"
                        allowClear
                        {...parentProps}
                        style={{ width }}
                        value={value}
                        onChange={onChange}
                    />
                )
        }
    }

    return <>{renderInput()}</>
}

export default DataInput
