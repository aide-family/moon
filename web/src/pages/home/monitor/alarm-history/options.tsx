import { ActionKey } from '@/apis/data'
import { AlarmHistoryItem } from '@/apis/home/monitor/alarm-history/types'
import { DictSelectItem } from '@/apis/home/system/dict/types'
import { Status, StatusMap } from '@/apis/types'
import { SearchFormItem } from '@/components/Data/SearchForm/SearchForm'
import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps } from 'antd'
import { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { CodeBox } from './child/Detail'

export const searchFormItems: SearchFormItem[] = [
    {
        name: 'keyword',
        label: '模糊查询'
    },
    {
        name: 'time_range',
        label: '时间范围',
        dataProps: {
            type: 'time-range',
            parentProps: {
                format: 'YYYY-MM-DD HH:mm:ss'
            }
        }
    }
]

export const rightOptions = [
    {
        key: ActionKey.REFRESH,
        label: <Button type="primary">刷新</Button>
    }
]

export const columns: ColumnsType<AlarmHistoryItem> = [
    {
        title: 'ID',
        dataIndex: 'id',
        key: 'id',
        align: 'start',
        width: 80
    },
    {
        title: '主机名',
        dataIndex: 'endpoint',
        key: 'endpoint',
        align: 'center',
        width: 200,
        render: (_: string, record: AlarmHistoryItem) => {
            return (
                <Button type="link" onClick={() => {}}>
                    {record.labels['endpoint'] || '-'}
                </Button>
            )
        }
    },
    {
        title: '实例名',
        dataIndex: 'instance',
        key: 'instance',
        align: 'center',
        width: 200,
        render: (_: string, record: AlarmHistoryItem) => {
            return (
                <Button type="link" onClick={() => {}}>
                    {record.labels['instance'] || '-'}
                </Button>
            )
        }
    },
    {
        title: '所属策略',
        dataIndex: 'alarmName',
        key: 'alarmName',
        width: 200,
        align: 'center',
        render: (alert: string, record: AlarmHistoryItem) => {
            return (
                <Button
                    type="link"
                    onClick={() => {
                        window.open(
                            `/#/home/monitor/strategy?strategyId=${record.alarmId}`
                        )
                    }}
                >
                    {alert || '-'}
                </Button>
            )
        }
    },
    {
        title: '告警级别',
        dataIndex: 'alarmLevel',
        key: 'alarmLevel',
        width: 100,
        align: 'center',
        render: (alarmLevel: DictSelectItem) => {
            const { color, label, value } = alarmLevel
            return <Badge color={color} key={value} text={label} />
        }
    },
    {
        title: '告警状态',
        dataIndex: 'alarmStatus',
        key: 'alarmStatus',
        width: 100,
        align: 'center',
        render: (alarmStatus: string) => {
            const isResolved =
                alarmStatus === 'resolved'
                    ? Status.STATUS_ENABLED
                    : Status.STATUS_DISABLED
            const { color } = StatusMap[isResolved]
            return <Badge color={color} key={alarmStatus} text={alarmStatus} />
        }
    },
    {
        title: '告警内容',
        dataIndex: 'annotations',
        key: 'annotations',
        // ellipsis: true,
        width: 600,
        render: (annotations: { [key: string]: string }) => {
            return <CodeBox code={annotations} />
        }
    },
    {
        title: '持续时间',
        dataIndex: 'duration',
        key: 'duration',
        align: 'center',
        width: 120,
        render: (_, record: AlarmHistoryItem) => {
            const s = dayjs(+record.startAt * 1000).diff(
                dayjs(+record.endAt * 1000),
                's'
            )
            // 将毫秒转换为时、分、秒
            const duration = dayjs.unix(s)
            // 格式化输出为 "x时x分x秒"
            const formattedDuration = `${duration.hour()}h${duration.minute()}m${duration.second()}s`
            return formattedDuration
        }
    },
    {
        title: '开始时间',
        dataIndex: 'startAt',
        key: 'startAt',
        width: 180,
        render: (startAt: string) => {
            return dayjs(+startAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    // endAt
    {
        title: '结束时间',
        dataIndex: 'endAt',
        key: 'endAt',
        width: 180,
        render: (startAt: string) => {
            return dayjs(+startAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    }
]

export const tableOperationItems = (
    item?: AlarmHistoryItem
): MenuProps['items'] => [
    {
        // 告警标记
        key: ActionKey.ALARM_MARK,
        label: (
            <Button
                type="primary"
                size="small"
                icon={<IconFont type="icon-mark" />}
            >
                告警标记
            </Button>
        ),
        disabled: !item
    },
    {
        //  告警详情
        key: ActionKey.DETAIL,
        label: (
            <Button
                type="primary"
                size="small"
                icon={<IconFont type="icon-detail" />}
            >
                告警详情
            </Button>
        )
    }
]
