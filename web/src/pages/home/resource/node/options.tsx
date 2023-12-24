import { Map } from '@/apis/types'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import { operationItems } from '@/components/Data/DataOption/option'
import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, DescriptionsProps } from 'antd'
import { ColumnGroupType, ColumnType } from 'antd/lib/table'
import { NodeItemType } from './type'
import { DataFormItem } from '@/components/Data'
import dayjs from 'dayjs'
import { ActionKey } from '@/apis/data'

export const rightOptions: DataOptionItem[] = [
    {
        key: ActionKey.REFRESH,
        label: <Button type="primary">刷新</Button>
    }
]

export const tableOperationItems: DataOptionItem[] = [
    {
        key: ActionKey.IKUAI,
        label: (
            <Button
                type="link"
                size="small"
                icon={<IconFont type="icon-linkedin-fill" />}
            >
                Ikuai
            </Button>
        )
    },
    ...(operationItems as any)
]

export const columns: (
    | ColumnGroupType<NodeItemType>
    | ColumnType<NodeItemType>
)[] = [
    {
        title: '节点名称',
        dataIndex: 'name',
        width: 200
    },
    {
        title: '中文名称',
        dataIndex: 'cname',
        width: 200
    },
    {
        title: '节点状态',
        dataIndex: 'status',
        width: 160,
        render: (status: Map) => {
            return status ? (
                <Badge status={status.color} text={status.name} />
            ) : (
                '-'
            )
        }
    },
    {
        title: '节点类型',
        dataIndex: 'type',
        width: 160,
        render: (type: Map) => {
            return type ? type.name : '-'
        }
    },
    {
        title: '带宽(Gbps)',
        dataIndex: 'band_width',
        width: 160,
        render: (bandWidth: number) => {
            // bit to Gbps
            return bandWidth ? `${bandWidth / 1000000000}Gbps` : '-'
        }
    },
    {
        title: '节点用途',
        dataIndex: 'purpose',
        width: 160,
        render: (purpose: Map) => {
            return purpose ? purpose.name : '-'
        }
    },
    {
        title: '跑量客户',
        dataIndex: 'customer',
        width: 220,
        render: (customer: Map) => {
            return customer ? customer.name : '-'
        }
    },
    {
        title: '资源供应商',
        dataIndex: 'supplier',
        width: 220
    },
    {
        title: '对接监控',
        dataIndex: 'is_monitor',
        width: 160,
        render: (isMonitor: boolean) => {
            return isMonitor ? '是' : '否'
        }
    },
    {
        title: '接堡垒机',
        dataIndex: 'is_jump',
        width: 160,
        render: (isJump: boolean) => {
            return isJump ? '是' : '否'
        }
    }
]

export const searchItems: DataFormItem[] = [
    {
        name: 'name',
        label: '节点名称'
    },
    {
        name: 'cname',
        label: '中文名称'
    },
    {
        name: 'type',
        label: '节点类型'
    },
    {
        name: 'status',
        label: '节点状态'
    },
    {
        name: 'is_jump',
        label: '接堡垒机'
    }
]

export const buildDetailItems = (
    data?: NodeItemType
): DescriptionsProps['items'] => {
    let count = 0
    return [
        {
            label: '节点ID',
            key: count++,
            children: data?.id || '-'
        },
        {
            label: '节点名称',
            key: count++,
            children: data?.name || '-'
        },
        {
            label: '节点类型',
            key: count++,
            children: data?.type?.name || '-'
        },
        {
            label: '节点状态',
            key: count++,
            children: data?.status?.name || '-'
        },

        {
            label: '创建人',
            key: count++,
            children: data?.created_by || '-'
        },
        {
            label: '创建时间',
            key: count++,
            children: data?.created_at
                ? dayjs(data?.created_at).format('YYYY-MM-DD HH:mm:ss')
                : '-'
        },
        {
            label: '更新人',
            key: count++,
            children: data?.updated_by || '-'
        },
        {
            label: '更新时间',
            key: count++,
            children: data?.updated_at
                ? dayjs(data?.updated_at).format('YYYY-MM-DD HH:mm:ss')
                : '-'
        }
    ]
}

export const editFormItems: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'name',
            label: '节点名称',
            rules: [
                {
                    required: true
                }
            ]
        },
        {
            name: 'cname',
            label: '中文名称'
        }
    ],
    [
        {
            name: 'type_id',
            label: '节点类型',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择节点类型'
                }
            }
        },
        {
            name: 'status_id',
            label: '节点状态',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择节点状态'
                }
            }
        }
    ],
    [
        {
            name: 'purpose_id',
            label: '节点用途',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择节点用途'
                }
            }
        },
        {
            name: '客户',
            label: 'customer_id',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择节点所属客户'
                }
            }
        }
    ],
    [
        {
            name: 'supplier_id',
            label: '资源供应商',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择节点供应商'
                }
            }
        },
        {
            name: 'band_width',
            label: '带宽(Gbps)'
        }
    ],
    [
        {
            name: 'is_monitor',
            label: '对接监控',
            dataProps: {
                type: 'radio-group',
                parentProps: {
                    options: [
                        {
                            label: '是',
                            value: 1
                        },
                        {
                            label: '否',
                            value: 0
                        }
                    ]
                }
            }
        },
        {
            name: 'is_jump',
            label: '接堡垒机',
            dataProps: {
                type: 'radio-group',
                parentProps: {
                    options: [
                        {
                            label: '是',
                            value: 1
                        },
                        {
                            label: '否',
                            value: 0
                        }
                    ]
                }
            }
        }
    ],
    [
        {
            label: '代理沟通群',
            name: 'agent_group'
        },
        {
            name: 'ikuai',
            label: 'ikuai'
        }
    ],
    {
        name: 'remark',
        label: '备注',
        dataProps: {
            type: 'textarea',
            parentProps: {
                placeholder: '请输入节点备注',
                maxLength: 200,
                showCount: true,
                allowClear: true
            }
        }
    }
]
