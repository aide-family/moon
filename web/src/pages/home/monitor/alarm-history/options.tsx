import { ActionKey } from '@/apis/data'
import { AlarmHistoryItem } from '@/apis/home/monitor/alarm-history/types'
import { DictSelectItem } from '@/apis/home/system/dict/types'
import { AlarmStatus, Status, StatusMap } from '@/apis/types'
import { SearchFormItem } from '@/components/Data/SearchForm/SearchForm'
// import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps, Tag } from 'antd'
import { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { CodeBox } from './child/Detail'
import { getLevels } from '../strategy/options'
import strategyApi from '@/apis/home/monitor/strategy'

export const getStrategySelectList = (keyword: string) => {
    return strategyApi
        .getStrategySelectList({
            keyword,
            page: {
                curr: 1,
                size: 10
            }
        })
        .then((items) => {
            return items.list.map((item) => {
                const { color, value, label } = item
                return {
                    value: value,
                    label: <Tag color={color}>{label}</Tag>
                }
            })
        })
}

export const searchFormItems: SearchFormItem[] = [
    {
        name: 'keyword',
        label: '模糊查询',
        formItemProps: {
            tooltip: '根据告警内容和告警实例名称模糊查询'
        }
    },
    {
        name: 'firingTime',
        label: '告警时间',
        formItemProps: {
            initialValue: [dayjs().add(-30, 'day'), dayjs()],
            tooltip: (
                <>
                    即告警开始时刻的时间范围
                    <br />
                    默认近三十天的告警数据
                </>
            )
        },
        dataProps: {
            type: 'time-range',
            parentProps: {
                format: 'YYYY-MM-DD HH:mm:ss'
            }
        }
    },
    {
        name: 'resolvedTime',
        label: '恢复时间',
        formItemProps: {
            tooltip: '即告警结束时刻的时间范围'
        },
        dataProps: {
            type: 'time-range',
            parentProps: {
                format: 'YYYY-MM-DD HH:mm:ss'
            }
        }
    },
    {
        name: 'alarmLevelIds',
        label: '告警等级',
        id: 'alarmLevelIds',
        formItemProps: {
            tooltip: '根据告警等级名称模糊查询，前缀匹配'
        },
        dataProps: {
            type: 'select-fetch',
            parentProps: {
                selectProps: {
                    placeholder: '请选择告警级别',
                    mode: 'multiple'
                },
                handleFetch: getLevels,
                defaultOptions: []
            }
        }
    },
    {
        name: 'strategyIds',
        label: '告警策略',
        id: 'strategyIds',
        formItemProps: {
            tooltip: '根据告警策略名称模糊查询，前缀匹配'
        },
        dataProps: {
            type: 'select-fetch',
            parentProps: {
                selectProps: {
                    placeholder: '请选择告警策略',
                    mode: 'multiple'
                },
                handleFetch: getStrategySelectList,
                defaultOptions: []
            }
        }
    },
    {
        name: 'duration',
        label: '持续时间',
        formItemProps: {
            tooltip: (
                <>
                    告警持续时间, 较大时间不包含前者 <br />
                    例如: <br />
                    5m以内, 实际为 {'1m < duration <= 5m'}
                </>
            ),
            initialValue: 0
        },
        dataProps: {
            type: 'radio-group',
            parentProps: {
                optionType: 'button',
                options: [
                    {
                        label: '全部',
                        value: 0
                    },
                    {
                        label: '1m以内',
                        value: 60
                    },
                    {
                        label: '5m以内',
                        value: 5 * 60
                    },
                    {
                        label: '30m以内',
                        value: 30 * 60
                    },
                    {
                        label: '30m以上',
                        value: 30 * 60 + 1
                    }
                ]
            }
        }
    },
    {
        name: 'status',
        label: '告警状态',
        formItemProps: {
            tooltip:
                '历史告警也包含了正在告警的数据，通过这个条件可以过滤正在告警的事件'
        },
        dataProps: {
            type: 'radio-group',
            parentProps: {
                optionType: 'button',
                options: [
                    {
                        label: '全部',
                        value: AlarmStatus.ALARM_STATUS_UNKNOWN
                    },
                    {
                        label: '告警',
                        value: AlarmStatus.ALARM_STATUS_ALARM
                    },
                    {
                        label: '恢复',
                        value: AlarmStatus.ALARM_STATUS_RESOLVE
                    }
                ]
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
        width: 100
    },
    // {
    //     title: '主机名',
    //     dataIndex: 'endpoint',
    //     key: 'endpoint',
    //     // align: 'center',
    //     width: 200,
    //     render: (_: string, record: AlarmHistoryItem) => {
    //         return record.labels['endpoint'] || '-'
    //     }
    // },
    {
        title: '实例名称',
        key: 'instance',
        // align: 'center',
        width: 300,
        render: (_: string, record: AlarmHistoryItem) => {
            return record.labels['instance'] || '-'
        }
    },
    {
        title: '所属策略',
        dataIndex: 'alarmName',
        key: 'alarmName',
        width: 200,
        // align: 'center',
        render: (alert: string, record: AlarmHistoryItem) => {
            return (
                <a
                    type="link"
                    href={`/#/home/monitor/strategy?strategyId=${record.alarmId}`}
                >
                    {alert || '-'}
                </a>
            )
        }
    },
    {
        title: '告警级别',
        dataIndex: 'alarmLevel',
        key: 'alarmLevel',
        width: 160,
        align: 'center',
        render: (alarmLevel?: DictSelectItem) => {
            if (!alarmLevel) return '-'
            const { color, label, value } = alarmLevel
            return <Badge color={color} key={value} text={label} />
        }
    },
    {
        title: '告警状态',
        dataIndex: 'alarmStatus',
        key: 'alarmStatus',
        width: 120,
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
        render: (annotations: { [key: string]: string }, record) => {
            return <CodeBox key={record.id + 'table'} code={annotations} />
        }
    },
    {
        title: '持续时间',
        dataIndex: 'duration',
        key: 'duration',
        align: 'center',
        width: 120
    },
    {
        title: '告警时间',
        dataIndex: 'startAt',
        key: 'startAt',
        width: 180,
        render: (startAt: string) => {
            return dayjs(+startAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    // endAt
    {
        title: '恢复时间',
        dataIndex: 'endAt',
        key: 'endAt',
        width: 180,
        render: (startAt: string) => {
            return +startAt > 0
                ? dayjs(+startAt * 1000).format('YYYY-MM-DD HH:mm:ss')
                : '持续中'
        }
    }
]

export const tableOperationItems = (
    item?: AlarmHistoryItem
): MenuProps['items'] => [
    {
        key: ActionKey.ALARM_EVENT_CHART,
        label: (
            <Button
                type="link"
                size="small"
                // icon={<IconFont type="icon-detail" />}
            >
                事件图表
            </Button>
        )
    },
    {
        // 告警标记
        key: ActionKey.ALARM_MARK,
        label: (
            <Button
                type="link"
                size="small"
                // icon={<IconFont type="icon-mark" />}
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
                type="link"
                size="small"
                // icon={<IconFont type="icon-detail" />}
            >
                告警详情
            </Button>
        )
    }
]
