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
import FetchSelect, { FetchSelectProps } from '../FetchSelect'
import TimeUintInput, { TimeUintInputProps } from '../TimeValue'

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
                type: 'select-fetch'
                parentProps?: FetchSelectProps
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
                type: 'time-value'
                parentProps: TimeUintInputProps
            }
          | {
                type: 'textarea'
                parentProps?: TextAreaProps
            }
      ) & {
          width?: number | string
          value?: any
          onChange?: (value: any) => void
          defaultValue?: any
      }

const DataInput: FC<DataInputProps> = (props) => {
    const { type, parentProps, width, value, defaultValue, onChange } = props

    const renderInput = () => {
        switch (type) {
            case 'select':
                return (
                    <Select
                        allowClear
                        style={{ width }}
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            case 'select-fetch':
                return <FetchSelect {...props} {...parentProps} />
            case 'radio':
                return (
                    <Radio
                        {...parentProps}
                        style={{ width }}
                        defaultChecked={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            case 'checkbox':
                return (
                    <Checkbox
                        {...parentProps}
                        style={{ width }}
                        defaultChecked={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            case 'date':
                return (
                    <DatePicker
                        allowClear
                        style={{ width }}
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            case 'time-value':
                return <TimeUintInput {...parentProps} />
            case 'radio-group':
                return (
                    <Radio.Group
                        style={{ width }}
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            case 'password':
                return (
                    <Input.Password
                        autoComplete="off"
                        allowClear
                        style={{ width }}
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            case 'textarea':
                return (
                    <Input.TextArea
                        autoComplete="off"
                        allowClear
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            default:
                return (
                    <Input
                        autoComplete="off"
                        allowClear
                        style={{ width }}
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
        }
    }

    return <>{renderInput()}</>
}

export default DataInput
