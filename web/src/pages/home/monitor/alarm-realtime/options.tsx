import { ActionKey } from '@/apis/data'
import { SelectAlarmPageRequest } from '@/apis/home/monitor/alarm-page/types'
import {
    AlarmRealtimeItem,
    AlarmRealtimeListRequest
} from '@/apis/home/monitor/alarm-realtime/types'
import { DictSelectItem } from '@/apis/home/system/dict/types'
import { Status } from '@/apis/types'
import { DataFormItem } from '@/components/Data'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import { IconFont } from '@/components/IconFont/IconFont'
import { Button, MenuProps } from 'antd'
import { ColumnType } from 'antd/es/table'
import { ColumnGroupProps } from 'antd/es/table/ColumnGroup'
import dayjs from 'dayjs'

export const defaultAlarmPageSelectReq: SelectAlarmPageRequest = {
    page: {
        size: 200,
        curr: 1
    },
    keyword: '',
    status: Status.STATUS_ENABLED
}

export const defaultAlarmRealtimeListRequest: AlarmRealtimeListRequest = {
    page: {
        size: 200,
        curr: 1
    },
    alarmPages: [],
    keyword: '',
    startAt: 0,
    endAt: 0
}

export const columns:
    | ColumnGroupProps<AlarmRealtimeItem>[]
    | ColumnType<AlarmRealtimeItem>[] = [
    {
        title: '告警时间',
        dataIndex: 'eventAt',
        key: 'eventAt',
        width: 200,
        render: (eventAt: number | string) => {
            return dayjs(+eventAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    {
        title: '策略ID',
        dataIndex: 'strategyId',
        key: 'strategyId',
        width: 200,
        render: (strategyId: number | string) => {
            return strategyId
        }
    },
    {
        title: '告警等级',
        dataIndex: 'level',
        width: 160,
        render: (level?: DictSelectItem) => {
            return level?.label || '-'
        }
    },
    {
        title: '主机名',
        dataIndex: 'instance',
        key: 'instance',
        width: 220
    },
    {
        title: '告警内容',
        dataIndex: 'note'
    }
]

export const searchFormItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '机器名称'
    },
    {
        name: 'starategyId',
        label: '告警规则',
        dataProps: {
            type: 'select',
            parentProps: {
                placeholder: '请选择告警规则',
                options: [
                    {
                        label: '告警规则1',
                        value: '1'
                    },
                    {
                        label: '告警规则2',
                        value: '2'
                    }
                ]
            }
        }
    },
    {
        name: 'enventAt',
        label: '告警持续时间',
        dataProps: {
            type: 'select',
            parentProps: {
                placeholder: '请选择告警持续时间',
                options: [
                    {
                        label: '全部',
                        value: ''
                    },
                    {
                        label: '1分钟',
                        value: '1m'
                    },
                    {
                        label: '5分钟',
                        value: '5m'
                    },
                    {
                        label: '10分钟',
                        value: '1m'
                    }
                ]
            }
        }
    }
]

export const rightOptions: DataOptionItem[] = [
    {
        key: ActionKey.REFRESH,
        label: <Button type="primary">刷新</Button>
    }
]

export const operationItems = (_: AlarmRealtimeItem): MenuProps['items'] => [
    {
        key: ActionKey.ALARM_INTERVENTION,
        label: (
            <Button
                type="link"
                size="small"
                icon={<IconFont type="icon-detail" />}
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
                icon={<IconFont type="icon-upgrade" />}
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
                icon={<IconFont type="icon-mark" />}
            >
                告警标记
            </Button>
        )
    }
]
