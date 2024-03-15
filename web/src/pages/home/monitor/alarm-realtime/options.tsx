import { ActionKey } from '@/apis/data'
import {
    AlarmRealtimeItem,
    AlarmRealtimeListRequest
} from '@/apis/home/monitor/alarm-realtime/types'
import { StrategyItemType } from '@/apis/home/monitor/strategy/types'
import { DictSelectItem } from '@/apis/home/system/dict/types'
import { Map } from '@/apis/types'
import { DataFormItem } from '@/components/Data'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
// import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps, Tooltip } from 'antd'
import { ColumnGroupProps } from 'antd/es/table/ColumnGroup'
import dayjs from 'dayjs'
import { SelectAalrmPageModal } from './child/SelectAlarmPageModal'
import { ColumnType } from 'antd/es/table'

export const defaultAlarmRealtimeListRequest: AlarmRealtimeListRequest = {
    page: {
        size: 200,
        curr: 1
    },
    alarmPageId: 1,
    keyword: '',
    startAt: 0,
    endAt: 0
}

export type ColumnsType<T = AlarmRealtimeItem> =
    | ColumnGroupProps<T>
    | ColumnType<T>

export const columns = (hiddenMap: Map): ColumnsType[] => [
    {
        title: '告警时间',
        dataIndex: 'eventAt',
        key: 'eventAt',
        width: 200,
        hidden: hiddenMap['eventAt'],
        render: (eventAt: number | string) => {
            return dayjs(+eventAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    {
        title: '持续时间',
        dataIndex: 'duration',
        key: 'duration',
        align: 'center',
        width: 120,
        hidden: hiddenMap['duration']
    },
    {
        title: '策略名称',
        dataIndex: 'strategy',
        key: 'strategy',
        width: 200,
        hidden: hiddenMap['strategy'],
        render: (strategy?: StrategyItemType) => {
            return (
                <Button
                    type="link"
                    href={`/#/home/monitor/strategy?strategyId=${
                        strategy?.id || ''
                    }`}
                >
                    {strategy?.alert || '-'}
                </Button>
            )
        }
    },
    {
        title: '告警等级',
        dataIndex: 'level',
        key: 'level',
        width: 160,
        hidden: hiddenMap['level'],
        render: (level?: DictSelectItem) => {
            if (!level) return '-'
            const { color, label, value } = level
            return <Badge color={color} key={value} text={label || '-'} />
        }
    },
    {
        title: '主机名',
        dataIndex: 'instance',
        key: 'instance',
        width: 220,
        hidden: hiddenMap['instance']
    },
    {
        title: '告警内容',
        dataIndex: 'note',
        hidden: hiddenMap['note'],
        ellipsis: true,
        render: (note: string) => {
            let result: any = null
            try {
                const annotations = JSON.parse(note)
                result = Object.keys(annotations).map((key) => {
                    const value = annotations[key]
                    return (
                        <div key={key}>
                            <span style={{ color: 'orangered' }}>{key}: </span>
                            <span>{value}</span>
                        </div>
                    )
                })
            } catch (error) {
                result = note
            }

            return <>{<Tooltip title={result}>{result}</Tooltip>}</>
        }
    }
]

export const searchFormItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '模糊查询'
    }
]

export const rightOptions = (refresh: () => void): DataOptionItem[] => [
    {
        key: ActionKey.REFRESH,
        label: <Button type="primary">刷新</Button>
    },
    {
        key: ActionKey.BIND_MY_ALARM_PAGES,
        label: <SelectAalrmPageModal refresh={refresh} />
    }
]

export const operationItems = (_: AlarmRealtimeItem): MenuProps['items'] => [
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
        key: ActionKey.ALARM_INTERVENTION,
        label: (
            <Button
                type="link"
                size="small"
                // icon={<IconFont type="icon-detail" />}
            >
                告警介入
            </Button>
        )
    },
    {
        key: ActionKey.ALARM_UPGRADE,
        label: (
            <Button
                type="link"
                size="small"
                // icon={<IconFont type="icon-upgrade" />}
            >
                告警升级
            </Button>
        )
    },
    {
        key: ActionKey.ALARM_MARK,
        label: (
            <Button
                type="link"
                size="small"
                // icon={<IconFont type="icon-mark" />}
            >
                告警标记
            </Button>
        )
    }
]
