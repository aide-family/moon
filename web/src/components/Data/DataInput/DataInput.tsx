import { FC } from 'react'

import type {
    CheckboxProps,
    ColorPickerProps,
    DatePickerProps,
    InputProps,
    RadioGroupProps,
    RadioProps,
    SegmentedProps,
    SelectProps
} from 'antd'
import {
    Input,
    Select,
    Radio,
    Checkbox,
    DatePicker,
    ColorPicker,
    Segmented
} from 'antd'
import { PasswordProps, TextAreaProps } from 'antd/lib/input'
import FetchSelect, { FetchSelectProps } from '../FetchSelect'
import TimeUintInput, { TimeUintInputProps } from '../TimeValue'
import { RangePickerProps } from 'antd/es/date-picker'
import {
    TemplateAutoComplete,
    TemplateAutoCompleteProps
} from '../TemplateAutoComplete'
import {
    EmailTemplateEditor,
    EmailTemplateEditorProps
} from './child/EmailTemplateEditor'

export type DataInputProps =
    | (
          | {
                type: 'input'
                parentProps?: InputProps
            }
          | {
                type: 'time-range'
                parentProps?: RangePickerProps
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
                type: 'segmented'
                parentProps?: SegmentedProps
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
          | {
                type: 'color'
                parentProps?: ColorPickerProps
            }
          | {
                type: 'template-auto-complete'
                parentProps?: TemplateAutoCompleteProps
            }
          | {
                type: 'email-emplate-editor'
                parentProps?: EmailTemplateEditorProps
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
            case 'color':
                return (
                    <ColorPicker
                        allowClear
                        value={value}
                        defaultValue={defaultValue}
                        onChange={(val) => onChange?.(val.toHexString())}
                        {...parentProps}
                    />
                )
            case 'time-range':
                return (
                    <DatePicker.RangePicker
                        style={{ width }}
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            case 'template-auto-complete':
                return (
                    <TemplateAutoComplete
                        style={{ width }}
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        {...parentProps}
                    />
                )
            case 'segmented':
                return (
                    <Segmented
                        {...parentProps}
                        value={value}
                        defaultValue={defaultValue}
                        onChange={onChange}
                        options={parentProps?.options || []}
                    />
                )
            case 'email-emplate-editor':
                return <EmailTemplateEditor {...props} {...parentProps} />
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
