import {
    Avatar,
    Badge,
    Button,
    Image,
    MenuProps,
    Tag,
    Tooltip,
    TourProps
} from 'antd'
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
    Map,
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
import { SizeType } from 'antd/es/config-provider/SizeContext'
import { AvatarSize } from 'antd/es/avatar/AvatarContext'
import { MutableRefObject } from 'react'

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
    {
        key: ActionKey.OPERATION_LOG,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-wj-rz" />}
            >
                操作日志
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

export type ColumnsType<T = StrategyItemType> =
    | ColumnGroupType<T>
    | ColumnType<T>

export const columns = (size: SizeType, hiddenMap: Map): ColumnsType[] => [
    {
        title: 'ID',
        dataIndex: 'id',
        key: 'id',
        align: 'start',
        width: 100,
        hidden: hiddenMap['id']
    },
    {
        title: '名称',
        dataIndex: 'alert',
        key: 'alert',
        width: 200,
        ellipsis: true,
        render: (alert: string) => {
            return <Tooltip title={alert}>{alert}</Tooltip>
        }
    },
    {
        title: '数据源',
        dataIndex: 'dataSource',
        key: 'dataSource',
        width: 200,
        align: 'center',
        hidden: hiddenMap['dataSource'],
        render: (dataSource?: PrometheusServerSelectItem) => {
            if (!dataSource) return '-'
            const { label, value, endpoint, status } = dataSource
            return (
                <Tooltip
                    title={status !== Status.STATUS_ENABLED ? '' : endpoint}
                >
                    <Button
                        type="link"
                        key={value}
                        disabled={status !== Status.STATUS_ENABLED}
                        href={`/#/home/monitor/endpoint?id=${value || ''}`}
                    >
                        {label}
                    </Button>
                </Tooltip>
            )
        }
    },
    {
        title: '策略组',
        dataIndex: 'groupInfo',
        key: 'groupInfo',
        width: 200,
        align: 'center',
        hidden: hiddenMap['groupInfo'],
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
                        href={`/#/home/monitor/strategy-group?id=${
                            value || ''
                        }`}
                    >
                        {label}
                    </Button>
                </Tooltip>
            )
        }
    },
    {
        title: '持续时间',
        dataIndex: 'duration',
        key: 'duration',
        width: 120,
        align: 'center',
        hidden: hiddenMap['duration'],
        render: (duration: Duration) => {
            return duration.value + '' + duration.unit
        }
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        width: 80,
        hidden: hiddenMap['status'],
        render: (status: Status) => {
            const { color, text } = StatusMap[status]
            return <Badge color={color} text={text} />
        }
    },
    {
        title: '策略等级',
        dataIndex: 'alarmLevelInfo',
        key: 'alarmLevelInfo',
        width: 160,
        hidden: hiddenMap['alarmLevelInfo'],
        render: (alarmLevelInfo: DictSelectItem) => {
            if (!alarmLevelInfo) return '-'
            const { color, label } = alarmLevelInfo
            return <Badge color={color} text={label} />
        }
    },
    {
        title: '策略类型',
        dataIndex: 'categoryInfo',
        key: 'categoryInfo',
        width: 120,
        hidden: hiddenMap['categoryInfo'],
        render: (_: number, record: StrategyItemType) => {
            if (!record.categoryInfo || !record.categoryInfo.length) return '-'
            const categyList = record.categoryInfo
            if (categyList.length === 1) {
                const { label, value, color } = categyList[0]
                return (
                    <Tag key={value} color={color}>
                        {label}
                    </Tag>
                )
            }
            return (
                <Avatar.Group
                    maxCount={2}
                    shape="square"
                    size={size as AvatarSize}
                >
                    {categyList.map((item, index) => {
                        return (
                            <Tooltip title={item.label} key={index}>
                                <Avatar
                                    key={item.value}
                                    style={{ backgroundColor: item.color }}
                                >
                                    {item.label}
                                </Avatar>
                            </Tooltip>
                        )
                    })}
                </Avatar.Group>
            )
        }
    },
    {
        title: '告警页面',
        dataIndex: 'alarmPageInfo',
        key: 'alarmPageInfo',
        width: 120,
        hidden: hiddenMap['alarmPageInfo'],
        render: (_: number, record: StrategyItemType) => {
            if (!record.alarmPageInfo || !record.alarmPageInfo.length)
                return '-'
            const alarmPageInfoList = record.alarmPageInfo
            if (alarmPageInfoList.length === 1) {
                const { label, value, color } = alarmPageInfoList[0]
                return (
                    <Tag key={value} color={color}>
                        {label}
                    </Tag>
                )
            }
            return (
                <Avatar.Group
                    maxCount={2}
                    shape="square"
                    size={size as AvatarSize}
                >
                    {alarmPageInfoList.map((item, index) => {
                        return (
                            <Tooltip title={item.label} key={index}>
                                <Avatar
                                    key={item.value}
                                    style={{ backgroundColor: item.color }}
                                >
                                    {item.label}
                                </Avatar>
                            </Tooltip>
                        )
                    })}
                </Avatar.Group>
            )
        }
    },
    {
        title: '告警恢复通知',
        dataIndex: 'sendRecover',
        key: 'sendRecover',
        width: 160,
        align: 'center',
        hidden: hiddenMap['sendRecover'],
        render: (sendRecover: boolean) => {
            return sendRecover ? '是' : '否'
        }
    },
    {
        title: '创建时间',
        dataIndex: 'createdAt',
        key: 'createdAt',
        width: 180,
        hidden: hiddenMap['createdAt'],
        render: (createdAt: string | number) => {
            return dayjs(+createdAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    {
        title: '更新时间',
        dataIndex: 'updatedAt',
        key: 'updatedAt',
        width: 180,
        hidden: hiddenMap['updatedAt'],
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
            id: 'dataSource',
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
            id: 'groupId',
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
            id: 'alert',
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
            id: 'duration',
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
            id: 'alarmLevelId',
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
            id: 'categoryIds',
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
                    message: '请选择策略类型'
                }
            ]
        }
    ],
    {
        name: 'alarmPageIds',
        label: '告警页面',
        id: 'alarmPageIds',
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
    // [
    //     {
    //         name: 'maxSuppress',
    //         label: '抑制策略',
    //         id: 'maxSuppress',
    //         dataProps: {
    //             type: 'time-value',
    //             parentProps: {
    //                 name: 'maxSuppress',
    //                 placeholder: ['请输入最大抑制时间', '选择单位'],
    //                 unitOptions: durationOptions
    //             }
    //         },
    //         formItemProps: {
    //             tooltip: (
    //                 <p>
    //                     抑制时常: 报警发生时, 开启抑制后,
    //                     从开始告警时间加抑制时长,如果在抑制周期内,
    //                     则不再发送告警
    //                 </p>
    //             )
    //         },
    //         rules: [
    //             {
    //                 validator: checkDuration('抑制时间')
    //             }
    //         ]
    //     },
    //     {
    //         name: 'sendInterval',
    //         label: '告警通知间隔',
    //         id: 'sendInterval',
    //         dataProps: {
    //             type: 'time-value',
    //             parentProps: {
    //                 name: 'sendInterval',
    //                 placeholder: ['请输入通知间隔时间', '选择单位'],
    //                 unitOptions: durationOptions
    //             }
    //         },
    //         formItemProps: {
    //             tooltip: (
    //                 <p>
    //                     告警通知间隔: 告警通知间隔, 在一定时间内没有消警,
    //                     则再次触发告警通知的时间
    //                 </p>
    //             )
    //         },
    //         rules: [
    //             {
    //                 validator: checkDuration('告警通知间隔时间')
    //             }
    //         ]
    //     },
    //     {
    //         name: 'sendRecover',
    //         label: '告警恢复通知',
    //         id: 'sendRecover',
    //         dataProps: {
    //             type: 'checkbox',
    //             parentProps: {
    //                 children: '发送告警恢复通知'
    //             }
    //         },
    //         formItemProps: {
    //             valuePropName: 'checked',
    //             tooltip: (
    //                 <p>
    //                     发送告警恢复通知: 开启该选项, 告警恢复后,
    //                     发送告警恢复通知的时间
    //                 </p>
    //             )
    //         }
    //     }
    // ],
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
        id: 'remark',
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

export const tourSteps = (refs: {
    [key: string]: MutableRefObject<any>
}): TourProps['steps'] => [
    {
        title: 'prometheus告警配置引导',
        description: '你可以跟着我的指引来完成prometheus告警配置',
        cover: (
            <div>
                <Image
                    preview={false}
                    src="https://img1.baidu.com/it/u=1602553681,3417380941&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=500"
                />
            </div>
        ),
        target: null
    },
    {
        title: '数据源选择',
        description:
            '数据源是报警规则配置的开始, 你只有选择了正确的数据源, 才能配置出有效的报警规则, 如果你还没有数据源, 请先去数据源菜单页面创建并开启它',
        placement: 'bottomRight',
        target: () => document.getElementById('dataSource')!
    },
    {
        title: '告警策略组选择',
        description:
            '告警策略组是报警规则的分组, 可以方便你管理你的报警规则, 如果你没有对应的策略组, 你可以先去策略组菜单页面创建并开启它',
        placement: 'bottomRight',
        target: () => document.getElementById('groupId')!
    },
    {
        title: '告警名称配置',
        description:
            '告警名称是报警规则的名称, 你可以在这里配置你的报警规则名称, 你的名称必须满足prometheus的格式, 不能是中文或者特殊字符',
        placement: 'bottomRight',
        target: () => document.getElementById('alert')!
    },
    {
        title: '告警策略持续时间配置',
        description:
            '告警策略持续时间是报警规则的持续时间, 也就是Prometheus规则中的for属性, 这里拆分成了数值+单位的形式',
        placement: 'bottomRight',
        target: () => document.getElementById('duration')!
    },
    {
        title: '策略等级配置',
        description:
            '策略等级是报警规则的等级, 这里是平台给策略划分的告警等级, 不同等级的策略具备不同的告警级别, 同时展示出来的告警颜色也有所不同, 如果没有策略等级, 需要你到字典管理菜单页面创建你的策略等级字典',
        placement: 'bottomRight',
        target: () => document.getElementById('alarmLevelId')!
    },
    {
        title: '策略类型配置（多选）',
        description:
            '策略类型是报警策略的类型, 你可以在这里选择你的报警策略类型, 该属性是用于业务层面给策略划分类型的, 方便我们统一管理不同业务场景的策略, 如果没有策略类型, 需要你到字典管理菜单页面创建你的策略类型字典, 注意区分策略类型和策略组类型',
        placement: 'bottomRight',
        target: () => document.getElementById('categoryIds')!
    },
    {
        title: '告警页面配置（多选）',
        description:
            '告警页面是报警规则的分组, 你可以在这里选择你的报警页面, 我们会按照你选择的报警页面在实时告警页面归类你的告警数据, 方便我们处理各种业务场景的实时告警, 如果没有告警页面, 需要你到告警页面管理菜单页面创建你的告警页面',
        placement: 'bottomRight',
        target: () => document.getElementById('alarmPageIds')!
    },
    {
        title: '抑制策略配置',
        description:
            '抑制策略: 当该规则触发时, 可以抑制该策略产生的告警, 从而避免产生大量无效告警的场景',
        placement: 'bottomRight',
        target: () => document.getElementById('maxSuppress')!
    },
    {
        title: '告警通知间隔配置',
        description:
            '告警发生时候会立即告警, 当持续通知间隔时长时, 会再次告警, 该时间默认时2小时',
        placement: 'bottomRight',
        target: () => document.getElementById('sendInterval')!
    },
    {
        title: '告警恢复通知配置',
        description: '告警恢复后是否发送告警恢复通知的开关, 默认是开启的',
        placement: 'bottomRight',
        target: () => document.getElementById('sendRecover')!
    },
    {
        title: '备注配置',
        description: '告警策略的辅助说明信息',
        placement: 'bottomRight',
        target: () => document.getElementById('remark')!
    },
    {
        title: 'PromQL规则编辑',
        description:
            'PromQL规则编辑是报警规则的配置入口, 你可以在这里配置你的报警规则, 这里支持语句智能提示, 语法校验等, 能有效帮助你写出正确的策略语句',
        placement: 'bottomRight',
        target: refs['promQLRef'].current
    },
    {
        title: 'PromQL验证',
        description:
            '这个闪电按钮是报警规则的验证入口,  你可以在这里验证你的报警规则是否正确, 策略语句正确的时候, 我会变成蓝色, 点开后能看到指标对应的数据和对应的图表展示',
        placement: 'bottomRight',
        target: refs['promQLButtonRef'].current
    },
    {
        title: '标签配置',
        description: '这里你可以添加自定义的label字段',
        placement: 'bottomRight',
        target: refs['labelsRef'].current
    },
    {
        title: '注释配置',
        description: '这里你可以添加自定义的 annotations 字段',
        placement: 'topRight',
        target: refs['annotationsRef'].current
    },
    {
        title: '注释标题配置',
        description:
            '这里你可以输入 annotations.title 字段, 支持例如 {{ $labels.xxx }} {{ $value }} 的模板',
        placement: 'topRight',
        target: refs['annotationsTitleRef'].current
    },
    {
        title: '注释明细配置',
        description:
            '这里你可以输入 annotations.description 字段, 支持例如 {{ $labels.xxx }} {{ $value }} 的模板',
        placement: 'topRight',
        target: refs['annotationsDescriptionRef'].current
    },
    {
        title: '你已经完成了prometheus 策略编辑的学习, 真是太棒了',
        description: '你现在可以去配置你的策略了',
        cover: (
            <div>
                <Image
                    preview={false}
                    src="https://img2.baidu.com/it/u=2874698497,1325088475&fm=253&fmt=auto&app=138&f=GIF?w=440&h=335"
                />
            </div>
        ),
        target: null
    }
]
