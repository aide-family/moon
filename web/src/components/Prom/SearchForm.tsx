import React from 'react'
import { DatePicker, Form, InputNumber, TimeRangePickerProps } from 'antd'
import dayjs, { Dayjs } from 'dayjs'

export type DateDataType = 'date' | 'range'

export type SearchFormProps<T = number | [number, number]> = {
    onSearch?: (type: DateDataType, value: T, step?: number) => void
    type?: DateDataType
}

const shortcutsRangePicker: TimeRangePickerProps['presets'] = [
    {
        label: '近一小时',
        value: () => [dayjs().subtract(1, 'hour'), dayjs()]
    },
    {
        label: '近3小时',
        value: () => [dayjs().subtract(3, 'hour'), dayjs()]
    },
    {
        label: '近6小时',
        value: () => [dayjs().subtract(6, 'hour'), dayjs()]
    },
    {
        label: '近12小时',
        value: () => [dayjs().subtract(12, 'hour'), dayjs()]
    },
    {
        label: '最近一天',
        value: () => [dayjs().subtract(1, 'day'), dayjs()]
    },
    {
        label: '今天',
        value: () => [dayjs().startOf('day'), dayjs().endOf('day')]
    },
    {
        label: '昨天',
        value: () => [
            dayjs().subtract(1, 'day').startOf('day'),
            dayjs().subtract(1, 'day').endOf('day')
        ]
    },
    {
        label: '最近三天',
        value: () => [dayjs().subtract(3, 'day'), dayjs()]
    },
    {
        label: '最近一周',
        value: () => [dayjs().subtract(1, 'week'), dayjs()]
    }
]

const shortcutsDatePicker: {
    label: React.ReactNode
    value: Dayjs | (() => Dayjs)
}[] = [
    {
        label: '此刻',
        value: dayjs()
    },
    {
        label: '一小时前',
        value: dayjs().subtract(1, 'hour')
    },
    {
        label: '三小时前',
        value: dayjs().subtract(3, 'hour')
    },
    {
        label: '六小时前',
        value: () => dayjs().subtract(6, 'hour')
    },
    {
        label: '十二小时前',
        value: () => dayjs().subtract(12, 'hour')
    },
    {
        label: '一天前',
        value: () => dayjs().subtract(1, 'day')
    },
    {
        label: '三天前',
        value: () => dayjs().subtract(3, 'day')
    },
    {
        label: '一周前',
        value: () => dayjs().subtract(1, 'week')
    }
]

const SearchForm: React.FC<SearchFormProps> = (props) => {
    const { onSearch, type } = props
    const [form] = Form.useForm()

    const handleOnChang = (value: {
        date_range?: string[]
        date?: string
        step?: number
    }) => {
        switch (type) {
            case 'range':
                if (!value.date_range) return
                onSearch?.(
                    type,
                    [
                        dayjs(value?.date_range[0]).unix(),
                        dayjs(value?.date_range[1]).unix()
                    ],
                    value.step
                )
                break
            case 'date':
                onSearch?.(type, dayjs(value.date).unix())
                break
        }
    }

    return (
        <Form
            layout="inline"
            form={form}
            onFinish={handleOnChang}
            onFieldsChange={form.submit}
        >
            {props.type === 'range' ? (
                <>
                    <Form.Item
                        name="date_range"
                        label="时间范围"
                        initialValue={[dayjs().subtract(1, 'hour'), dayjs()]}
                    >
                        <DatePicker.RangePicker
                            showTime
                            style={{ width: 380 }}
                            presets={shortcutsRangePicker}
                        />
                    </Form.Item>
                    <Form.Item name="step" initialValue={14} label="步长">
                        <InputNumber min={1} max={60} step={1} />
                    </Form.Item>
                </>
            ) : (
                <Form.Item name="date" label="时间" initialValue={dayjs()}>
                    <DatePicker
                        showTime
                        style={{ width: 380 }}
                        presets={shortcutsDatePicker}
                    />
                </Form.Item>
            )}
        </Form>
    )
}

export default SearchForm
