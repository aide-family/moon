import { Button, MenuProps, Space, Tag, Tooltip } from 'antd'
import { IconFont } from '@/components/IconFont/IconFont.tsx'
import { operationItems } from '@/components/Data/DataOption/option.tsx'
import { DataFormItem } from '@/components/Data'
import { StrategyItemType } from '@/apis/home/monitor/strategy/types'
import { ActionKey } from '@/apis/data'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
    Category,
    Duration,
    PageReqType,
    Status,
    StatusMap
} from '@/apis/types'
import { NotifyItem } from '@/apis/home/monitor/alarm-notify/types'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import endpointApi from '@/apis/home/monitor/endpoint'
import strategyGroupApi from '@/apis/home/monitor/strategy-group'
import dictApi from '@/apis/home/system/dict'
import alarmPageApi from '@/apis/home/monitor/alarm-page'
import strategyApi from '@/apis/home/monitor/strategy'
import { DictSelectItem } from '@/apis/home/system/dict/types'
import { PrometheusServerSelectItem } from '@/apis/home/monitor/endpoint/types'
import { StrategyGroupSelectItemType } from '@/apis/home/monitor/strategy-group/types'
import { checkDuration } from '@/components/Data/TimeValue'

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

export const defaultPageReq: PageReqType = {
    curr: 1,
    size: 10
}

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
        title: 'ID',
        dataIndex: 'id',
        key: 'id',
        align: 'start',
        width: 100
    },
    {
        title: '数据源',
        dataIndex: 'dataSource',
        key: 'dataSource',
        align: 'start',
        width: 200,
        render: (dataSource?: PrometheusServerSelectItem) => {
            if (!dataSource) return '-'
            const { label, value, endpoint, status } = dataSource
            return (
                <Tooltip
                    title={status !== Status.STATUS_ENABLED ? '' : endpoint}
                >
                    <Tag
                        key={value}
                        color={
                            status !== Status.STATUS_ENABLED
                                ? '#CD9A6A'
                                : '#1677ff'
                        }
                    >
                        {label}
                    </Tag>
                </Tooltip>
            )
        }
    },
    {
        title: '名称',
        dataIndex: 'alert',
        key: 'alert',
        width: 200,
        render: (alert: string) => {
            return alert
        }
    },
    {
        title: '持续时间',
        dataIndex: 'duration',
        key: 'duration',
        width: 120,
        render: (duration: Duration) => {
            return duration.value + '' + duration.unit
        }
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        width: 80,
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return (
                <Tag key={text} color={color}>
                    {text}
                </Tag>
            )
        }
    },
    {
        title: '策略等级',
        dataIndex: 'alarmLevelInfo',
        key: 'alarmLevelInfo',
        width: 160,
        render: (alarmLevelInfo: DictSelectItem) => {
            if (!alarmLevelInfo) return '-'
            const { color, label, value } = alarmLevelInfo
            return (
                <Tag key={value} color={color}>
                    {label}
                </Tag>
            )
        }
    },
    {
        title: '策略类型',
        dataIndex: 'categoryInfo',
        key: 'categoryInfo',
        width: 300,
        render: (_: number, record: StrategyItemType) => {
            if (!record.categoryInfo || !record.categoryInfo.length) return '-'
            const categyList = record.categoryInfo
            return (
                <Space size={[8, 16]} wrap>
                    {categyList.map((item) => {
                        return (
                            <Tag key={item.value} color={item.color}>
                                {item.label}
                            </Tag>
                        )
                    })}
                </Space>
            )
        }
    },
    {
        title: '告警页面',
        dataIndex: 'alarmPageInfo',
        key: 'alarmPageInfo',
        width: 300,
        render: (_: number, record: StrategyItemType) => {
            if (!record.alarmPageInfo || !record.alarmPageInfo.length)
                return '-'
            const alarmPageInfoList = record.alarmPageInfo
            return (
                <Space size={[8, 16]} wrap>
                    {alarmPageInfoList.map((item) => {
                        return (
                            <Tag key={item.value} color={item.color}>
                                {item.label}
                            </Tag>
                        )
                    })}
                </Space>
            )
        }
    },
    {
        title: '告警恢复通知',
        dataIndex: 'sendRecover',
        key: 'sendRecover',
        width: 160,
        align: 'center',
        render: (sendRecover: boolean) => {
            return sendRecover ? '是' : '否'
        }
    },
    {
        title: '策略组',
        dataIndex: 'groupInfo',
        key: 'groupInfo',
        width: 160,
        align: 'center',
        render: (groupInfo?: StrategyGroupSelectItemType) => {
            if (!groupInfo) return '-'
            const { label, value, status, color, remark } = groupInfo
            return (
                <Tooltip title={remark}>
                    <Button
                        key={value}
                        type="link"
                        color={color}
                        disabled={status !== Status.STATUS_ENABLED}
                    >
                        {label}
                    </Button>
                </Tooltip>
            )
        }
    },
    // {
    //     title: '创建时间',
    //     dataIndex: 'createdAt',
    //     key: 'createdAt',
    //     width: 160,
    //     render: (createdAt: string) => {
    //         return dayjs(+createdAt * 1000).format('YYYY-MM-DD HH:mm:ss')
    //     }
    // },
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

export type NotifyObjectTableColumnType =
    | ColumnGroupType<NotifyItem>
    | ColumnType<NotifyItem>

export const notifyObjectTableColumns: NotifyObjectTableColumnType[] = [
    {
        title: '名称',
        dataIndex: 'name',
        key: 'name',
        width: 200,
        render: (name: string) => {
            return name
        }
    },
    {
        title: '备注',
        dataIndex: 'remark',
        key: 'remark',
        // width: '30%',
        ellipsis: true,
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

export const getEndponts = (keyword: string) => {
    return endpointApi
        .selectEndpoint({
            keyword,
            page: defaultPageReq,
            status: Status.STATUS_ENABLED
        })
        .then((items) => {
            return items.list.map((item) => {
                const { value, label, endpoint } = item
                return {
                    value: value,
                    label: label,
                    title: endpoint
                }
            })
        })
}

export const getStrategyGroups = (status?: Status) => {
    const selectFetch = (keyword: string) => {
        return strategyGroupApi
            .getStrategyGroupSelect({ keyword, page: defaultPageReq, status })
            .then((items) => {
                return items.list.map((item) => {
                    const { value, label } = item
                    return {
                        value: value,
                        label: <Tag color="blue">{label}</Tag>
                    }
                })
            })
    }
    return selectFetch
}

export const getLevels = (keyword: string) => {
    return dictApi
        .dictSelect({
            keyword,
            page: defaultPageReq,
            category: Category.CATEGORY_ALARM_LEVEL
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

export const getCategories = (keyword: string) => {
    return dictApi
        .dictSelect({
            keyword,
            page: defaultPageReq,
            category: Category.CATEGORY_PROM_STRATEGY
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

export const getAlarmPages = (keyword: string) => {
    return alarmPageApi
        .getAlarmPageSelect({
            keyword,
            page: defaultPageReq,
            status: Status.STATUS_ENABLED
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

export const getRestrain = (keyword: string) => {
    return strategyApi
        .getStrategySelectList({ keyword, page: defaultPageReq })
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

export const strategyEditOptions: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'dataSource',
            label: '数据源',
            dataProps: {
                type: 'select-fetch',
                parentProps: {
                    selectProps: {
                        placeholder: '请选择数据源',
                        labelInValue: true
                    },
                    width: '100%',
                    handleFetch: getEndponts,
                    defaultOptions: []
                }
            },
            formItemProps: {
                tooltip: <p>请选择Prometheus数据源, 目前仅支持Prometheus</p>,
                dependencies: ['expr'],
                rules: [
                    {
                        required: true,
                        message: '请选择Prometheus数据源'
                    }
                ]
            }
        },
        {
            name: 'groupId',
            label: '策略组',
            dataProps: {
                type: 'select-fetch',
                parentProps: {
                    selectProps: {
                        placeholder: '请选择策略分组'
                    },
                    handleFetch: getStrategyGroups(Status.STATUS_ENABLED),
                    defaultOptions: []
                }
            },
            formItemProps: {
                tooltip: <p>把当前规则归类到不同的策略组, 便于业务关联</p>
            },
            rules: [
                {
                    required: true,
                    message: '请选择策略组'
                }
            ]
        }
    ],
    [
        {
            name: 'alert',
            label: '告警名称',
            formItemProps: {
                tooltip: (
                    <p>请输入策略名称, 策略名称必须唯一, 例如: 'cpu_usage'</p>
                )
            },
            rules: [
                {
                    required: true,
                    message: '请输入策略名称'
                }
            ]
        },
        {
            name: 'duration',
            label: '持续时间',
            dataProps: {
                type: 'time-value',
                parentProps: {
                    name: 'duration',
                    placeholder: ['请输入持续时间', '选择单位'],
                    unitOptions: durationOptions
                }
            },
            formItemProps: {
                tooltip: (
                    <p>
                        持续时间是下面PromQL规则连续匹配,
                        建议为此规则采集周期的整数倍, 例如采集周期为15s,
                        持续时间为30s, 则表示连续2个周期匹配
                    </p>
                ),
                required: true
            },
            rules: [
                {
                    validator: checkDuration('持续时间', true)
                }
            ]
        }
    ],
    [
        {
            name: 'alarmLevelId',
            label: '策略等级',
            dataProps: {
                type: 'select-fetch',
                parentProps: {
                    selectProps: {
                        placeholder: '请选择告警级别'
                    },
                    handleFetch: getLevels,
                    defaultOptions: []
                }
            },
            rules: [
                {
                    required: true,
                    message: '请选择策略等级'
                }
            ]
        },
        {
            name: 'categoryIds',
            label: '策略类型',
            dataProps: {
                type: 'select-fetch',
                parentProps: {
                    selectProps: {
                        placeholder: '请选择策略类型',
                        mode: 'multiple'
                    },
                    handleFetch: getCategories,
                    defaultOptions: []
                }
            },
            rules: [
                {
                    required: true,
                    message: '请选择策略类型',
                    validator(_, value) {
                        if (value?.length === 0) {
                            return Promise.reject('请选择策略类型')
                        }
                        return Promise.resolve()
                    }
                }
            ]
        }
    ],
    {
        name: 'alarmPageIds',
        label: '告警页面',
        dataProps: {
            type: 'select-fetch',
            parentProps: {
                selectProps: {
                    placeholder: '请选择告警页面',
                    mode: 'multiple'
                },
                handleFetch: getAlarmPages,
                defaultOptions: []
            }
        },
        formItemProps: {
            tooltip: <p>报警页面: 当该规则触发时, 页面将跳转到报警页面</p>
        },
        rules: [
            {
                required: true,
                message: '请选择报警页面'
            }
        ]
    },
    [
        {
            name: 'maxSuppress',
            label: '抑制策略',
            dataProps: {
                type: 'time-value',
                parentProps: {
                    name: 'maxSuppress',
                    placeholder: ['请输入最大抑制时间', '选择单位'],
                    unitOptions: durationOptions
                }
            },
            formItemProps: {
                tooltip: (
                    <p>
                        抑制时常: 报警发生时, 开启抑制后,
                        从开始告警时间加抑制时长,如果在抑制周期内,
                        则不再发送告警
                    </p>
                )
            },
            rules: [
                {
                    validator: checkDuration('抑制时间')
                }
            ]
        },
        {
            name: 'sendInterval',
            label: '告警通知间隔',
            dataProps: {
                type: 'time-value',
                parentProps: {
                    name: 'sendInterval',
                    placeholder: ['请输入通知间隔时间', '选择单位'],
                    unitOptions: durationOptions
                }
            },
            formItemProps: {
                tooltip: (
                    <p>
                        告警通知间隔: 告警通知间隔, 在一定时间内没有消警,
                        则再次触发告警通知的时间
                    </p>
                )
            },
            rules: [
                {
                    validator: checkDuration('告警通知间隔时间')
                }
            ]
        },
        {
            name: 'sendRecover',
            label: '告警恢复通知',
            dataProps: {
                type: 'checkbox',
                parentProps: {
                    children: '发送告警恢复通知'
                }
            },
            formItemProps: {
                valuePropName: 'checked',
                tooltip: (
                    <p>
                        发送告警恢复通知: 开启该选项, 告警恢复后,
                        发送告警恢复通知的时间
                    </p>
                )
            }
        }
    ],
    // {
    //     name: 'restrain',
    //     label: '抑制对象',
    //     formItemProps: {
    //         tooltip: <p>抑制对象: 当该规则触发时, 此列表对象的告警将会被抑制</p>
    //     }
    // },
    {
        name: 'remark',
        label: '备注',
        formItemProps: {
            tooltip: <p>请输入备注</p>
        },
        dataProps: {
            type: 'textarea',
            parentProps: {
                autoSize: { minRows: 2, maxRows: 6 },
                maxLength: 200,
                showCount: true,
                placeholder: '请输入备注'
            }
        }
    }
]

export const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '策略名称'
    },
    {
        name: 'groupId',
        label: '策略组',
        dataProps: {
            type: 'select-fetch',
            parentProps: {
                handleFetch: getStrategyGroups(),
                selectProps: {
                    placeholder: '请选择策略组'
                },
                defaultOptions: []
            }
        }
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
