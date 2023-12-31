import { Button, MenuProps, Space, Tag } from 'antd'
import { IconFont } from '@/components/IconFont/IconFont.tsx'
import { operationItems } from '@/components/Data/DataOption/option.tsx'
import { DataFormItem } from '@/components/Data'
import { StrategyItemType } from '@/apis/home/monitor/strategy/types'
import { ActionKey } from '@/apis/data'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import dayjs from 'dayjs'
import { Status, StatusMap } from '@/apis/types'
import { NotifyItem } from '@/apis/home/monitor/alarm-notify/types'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'

export const OP_KEY_STRATEGY_GROUP_LIST = 'strategy-group-list'

export const tableOperationItems = (
    item: StrategyItemType
): MenuProps['items'] => [
    {
        key: ActionKey.STRATEGY_GROUP_LIST,
        label: (
            <Button
                type="link"
                size="small"
                icon={<IconFont type="icon-linkedin-fill" />}
            >
                策略组列表
            </Button>
        )
    },
    item.status === Status.STATUS_DISABLED
        ? {
              key: ActionKey.ENABLE,
              label: (
                  <Button
                      type="link"
                      size="small"
                      icon={<IconFont type="icon-Enable" />}
                  >
                      启用
                  </Button>
              )
          }
        : {
              key: ActionKey.DISABLE,
              label: (
                  <Button
                      type="link"
                      size="small"
                      icon={<IconFont type="icon-disable4" />}
                      danger
                  >
                      禁用
                  </Button>
              )
          },
    {
        key: ActionKey.STRATEGY_NOTIFY_OBJECT,
        label: (
            <Button
                type="link"
                size="small"
                icon={<IconFont type="icon-email1" />}
            >
                通知对象
            </Button>
        )
    },
    ...(operationItems(item) as [])
]

export const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '策略名称'
    },
    {
        name: 'groupId',
        label: '策略组'
    },
    {
        name: 'status',
        label: '策略状态',
        dataProps: {
            type: 'radio-group',
            parentProps: {
                optionType: 'button',
                defaultValue: 0,
                options: [
                    {
                        label: '全部',
                        value: 0
                    },
                    {
                        label: '启用',
                        value: 1
                    },
                    {
                        label: '禁用',
                        value: 2
                    }
                ]
            }
        }
    }
]

export const durationOptions = [
    {
        label: '秒',
        value: 's'
    },
    {
        label: '分钟',
        value: 'm'
    },
    {
        label: '小时',
        value: 'h'
    },
    {
        label: '天',
        value: 'd'
    }
]

export const columns: (
    | ColumnGroupType<StrategyItemType>
    | ColumnType<StrategyItemType>
)[] = [
    {
        title: '名称',
        dataIndex: 'alert',
        key: 'alert',
        // width: 160,
        render: (alert: string) => {
            return alert
        }
    },
    {
        title: '持续时间',
        dataIndex: 'duration',
        key: 'duration',
        width: 160,
        render: (duration: string) => {
            return duration
        }
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        width: 160,
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return <Tag color={color}>{text}</Tag>
        }
    },
    {
        title: '策略等级',
        dataIndex: 'level',
        key: 'level',
        width: 160,
        render: (_: number, record: StrategyItemType) => {
            if (!record.alarmLevelInfo) return '-'
            const { color, label } = record.alarmLevelInfo
            return <Tag color={color}>{label}</Tag>
        }
    },
    {
        title: '策略类型',
        dataIndex: 'level',
        key: 'level',
        width: 260,
        render: (_: number, record: StrategyItemType) => {
            if (!record.categoryInfo || !record.categoryInfo.length) return '-'
            const categyList = record.categoryInfo
            return (
                <Space direction="horizontal">
                    {categyList.map((item) => {
                        return <Tag color={item.color}>{item.label}</Tag>
                    })}
                </Space>
            )
        }
    },
    {
        title: '创建时间',
        dataIndex: 'createdAt',
        key: 'createdAt',
        width: 160,
        render: (createdAt: string) => {
            return dayjs(+createdAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    {
        title: '更新时间',
        dataIndex: 'updatedAt',
        key: 'updatedAt',
        width: 160,
        render: (updatedAt: string) => {
            return dayjs(+updatedAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    }
]

// TODO 获取数据源
export const endpoIntOptions = [
    {
        label: 'Prometheus',
        value: 'http://124.223.104.203:9090'
    },
    {
        label: 'Localhost',
        value: 'http://localhost:9090'
    }
]

// TODO 获取策略组列表
export const strategyGroupOptions = [
    {
        label: 'Default',
        value: 1
    },
    {
        label: '网络',
        value: 2
    },
    {
        label: '存储',
        value: 3
    }
]

export const alarmPageOptions = [
    {
        label: '实时告警',
        value: 1
    },
    {
        label: '测试页面',
        value: 2
    },
    {
        label: '值班页面',
        value: 3
    }
]

export const categoryOptions = [
    {
        label: '业务监控',
        value: 1
    },
    {
        label: '系统监控',
        value: 2
    },
    {
        label: '业务日志',
        value: 3
    },
    {
        label: '业务告警',
        value: 4
    },
    {
        label: '系统告警',
        value: 5
    }
]

export const sverityOptions = [
    {
        label: 'warning',
        value: '1'
    },
    {
        label: 'critical',
        value: '2'
    },
    {
        label: 'info',
        value: '3'
    }
]

export const restrainOptions = [
    {
        label: '策略1',
        value: 1
    }
]

export const maxSuppressUnitOptions = [
    {
        label: '秒',
        value: 's'
    },
    {
        label: '分钟',
        value: 'm'
    },
    {
        label: '小时',
        value: 'h'
    },
    {
        label: '天',
        value: 'd'
    }
]

export const notifyObjectTableColumns: (
    | ColumnGroupType<NotifyItem>
    | ColumnType<NotifyItem>
)[] = [
    {
        title: '名称',
        dataIndex: 'name',
        key: 'name',
        width: 160,
        render: (name: string) => {
            return name
        }
    },
    {
        title: '备注',
        dataIndex: 'remark',
        key: 'remark',
        // width: 160,
        render: (remark: string) => {
            return remark
        }
    }
]

export const leftOptions = (loading: boolean): DataOptionItem[] => [
    {
        key: ActionKey.BATCH_IMPORT,
        label: (
            <Button type="primary" loading={loading}>
                批量导入
            </Button>
        )
    }
]

export const rightOptions = (loading: boolean): DataOptionItem[] => [
    {
        key: ActionKey.REFRESH,
        label: (
            <Button type="primary" loading={loading}>
                刷新
            </Button>
        )
    }
]
