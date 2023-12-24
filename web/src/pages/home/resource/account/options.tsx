import { ColumnGroupType, ColumnType } from 'antd/lib/table'
import { AccountItemType } from './type'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import { Badge, Button, DescriptionsProps } from 'antd'
import { DataFormItem } from '@/components/Data'
import { Map } from '@/apis/types'
import dayjs from 'dayjs'
import { ActionKey } from '@/apis/data'

export const columns: (
    | ColumnGroupType<AccountItemType>
    | ColumnType<AccountItemType>
)[] = [
    {
        dataIndex: 'name',
        key: 'name',
        title: '账号名称',
        width: 200
    },
    {
        dataIndex: 'password',
        key: 'password',
        title: '密码',
        width: 200,
        render: (text) => {
            return text ? '******' : '-'
        }
    },
    {
        dataIndex: 'status',
        key: 'status',
        title: '状态',
        width: 200,
        render: (status: Map) => {
            return status ? (
                <Badge color={status?.color || 'default'} text={status?.name} />
            ) : (
                '-'
            )
        }
    },
    {
        dataIndex: 'type',
        key: 'type',
        title: '账号类型',
        width: 200,
        render: (type: Map) => {
            return type ? type.name : '-'
        }
    },
    {
        dataIndex: 'vlan_type',
        key: 'vlan_type',
        title: 'VLAN类型',
        width: 200,
        render: (vlan_type: Map) => {
            return vlan_type ? vlan_type.name : '-'
        }
    },
    {
        dataIndex: 'vlan_id',
        key: 'vlan_id',
        title: 'VLAN ID',
        width: 200
    },
    {
        dataIndex: 'bandwidth',
        key: 'bandwidth',
        title: '带宽(Mbps)',
        width: 200,
        render: (bandwidth: number) => {
            return bandwidth ? bandwidth : '-'
        }
    },
    {
        dataIndex: 'ipv4',
        key: 'ipv4',
        title: 'IPv4',
        width: 200
    },
    {
        dataIndex: 'ipv6',
        key: 'ipv6',
        title: 'IPv6',
        width: 200
    },
    {
        dataIndex: 'node',
        key: 'node',
        title: '所属节点',
        width: 200,
        render: (node: Map) => {
            return node ? node.name : '-'
        }
    },
    {
        dataIndex: 'host',
        key: 'host',
        title: '所属主机',
        width: 200,
        render: (host: Map) => {
            return host ? host.name : '-'
        }
    },
    {
        dataIndex: '__switch__',
        key: '__switch__',
        title: '交换机',
        width: 200,
        render: (_: string, row: AccountItemType) => {
            return (
                <div>
                    <span>名称: {row.switch_device?.name}</span>
                    <span>端口: {row.switch_port}</span>
                </div>
            )
        }
    }
]

export const rightOptions: DataOptionItem[] = [
    {
        key: ActionKey.REFRESH,
        label: <Button type="primary">刷新</Button>
    }
]

export const searchItems: DataFormItem[] = [
    {
        name: 'name',
        label: '账号名称'
    },
    {
        name: 'status_id',
        label: '账号状态'
    },
    {
        name: 'type_id',
        label: '账号类型'
    },
    {
        name: 'vlan_type_id',
        label: 'VLAN类型'
    }
]

export const buildDetailItems = (
    data?: AccountItemType
): DescriptionsProps['items'] => {
    let count = 0
    return [
        {
            label: '账号ID',
            key: count++,
            children: data?.id || '-'
        },
        {
            label: '账号名称',
            key: count++,
            children: data?.name || '-'
        },
        {
            label: '账号类型',
            key: count++,
            children: data?.type?.name || '-'
        },
        {
            label: '账号状态',
            key: count++,
            children: data?.status?.name || '-'
        },
        {
            label: 'IPV4',
            key: count++,
            children: data?.ipv4 || '-'
        },
        {
            label: 'IPV6',
            key: count++,
            children: data?.ipv6 || '-'
        },
        {
            label: '所属节点',
            key: count++,
            children: data?.node?.name || '-'
        },
        {
            label: 'VLAN类型',
            key: count++,
            children: data?.vlan_type?.name || '-'
        },
        {
            label: 'VLAN ID',
            key: count++,
            children: data?.vlan_id || '-'
        },
        {
            label: '所属主机',
            key: count++,
            children: data?.host?.name || '-'
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
    {
        name: 'name',
        label: '账号名称',
        dataProps: {
            type: 'input',
            parentProps: {
                placeholder: '请输入账号名称',
                autoComplete: 'off'
            }
        },
        rules: [{ required: true, message: '请输入账号名称' }]
    },
    [
        {
            name: 'password',
            label: '账号密码',
            dataProps: {
                type: 'password',
                parentProps: {
                    placeholder: '请输入账号密码',
                    autoComplete: 'new-password',
                    autoCapitalize: 'off'
                }
            }
        },
        {
            name: 'type_id',
            label: '账号类型',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择账号类型'
                }
            }
        }
    ],
    [
        {
            name: 'vlan_type_id',
            label: 'VLAN类型',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择VLAN类型'
                }
            }
        },
        {
            name: 'vlan_id',
            label: 'VLAN ID'
        }
    ],
    [
        {
            name: 'ipv4',
            label: 'IPv4'
        },
        {
            name: 'ipv6',
            label: 'IPv6'
        }
    ],
    [
        {
            name: 'node_id',
            label: '所属节点',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择所属节点'
                }
            }
        },
        {
            name: 'host_id',
            label: '所属主机',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择所属主机'
                }
            }
        }
    ],
    [
        {
            name: 'status_id',
            label: '账号状态',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择账号状态'
                }
            }
        },
        {
            name: 'switch_device_id',
            label: '交换机',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择交换机'
                }
            }
        }
    ],
    [
        {
            name: 'bandwidth',
            label: '带宽(Mbps)'
        },
        {
            name: 'switch_port',
            label: '交换机端口'
        }
    ]
]
