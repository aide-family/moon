import { DataFormItem } from '@/components/Data'
import dayjs from 'dayjs'
import { DeviceItemType } from './type'
import { Badge, Button, DescriptionsProps } from 'antd'
import { ColumnGroupType, ColumnType } from 'antd/lib/table'
import { Map } from '@/apis/types'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'

export const searchItems: DataFormItem[] = [
    {
        name: 'host_name',
        label: '设备名称'
    },
    {
        name: 'node_id',
        label: '所属节点'
    },

    {
        name: 'type_id',
        label: '设备类型'
    },
    {
        name: 'status_id',
        label: '设备状态'
    },
    {
        name: 'source_id',
        label: '设备来源'
    },
    {
        name: 'sn',
        label: '设备序列号'
    }
]

export const buildDetailItems = (
    data?: DeviceItemType
): DescriptionsProps['items'] => {
    let count = 0
    return [
        {
            label: '设备ID',
            key: count++,
            children: data?.id || '-'
        },
        {
            label: '设备名称',
            key: count++,
            children: data?.host_name || '-'
        },
        {
            label: '设备类型',
            key: count++,
            children: data?.type?.name || '-'
        },
        {
            label: '设备状态',
            key: count++,
            children: data?.status?.name || '-'
        },
        {
            label: '设备来源',
            key: count++,
            children: data?.source?.name || '-'
        },
        {
            label: '设备序列号',
            key: count++,
            children: data?.sn || '-'
        },
        {
            label: '所属节点',
            key: count++,
            children: data?.node?.name || '-'
        },
        {
            label: 'IPMI',
            key: count++,
            children: data?.ipmi || '-'
        },
        {
            label: 'IP',
            key: count++,
            children: data?.manage_ip || '-'
        },
        {
            label: '供应商',
            key: count++,
            children: data?.supplier?.name || '-'
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
            name: 'host_name',
            label: '设备名称',
            rules: [{ required: true, message: '请输入设备名称' }]
        },
        {
            name: 'IPMI',
            label: 'IPMI',
            rules: [{ required: true, message: '请输入IPMI' }]
        }
    ],
    [
        {
            name: 'type_id',
            label: '设备类型',
            rules: [{ required: true, message: '请选择设备类型' }],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择设备类型'
                }
            }
        },
        {
            name: 'source_id',
            label: '设备来源',
            rules: [{ required: true, message: '请选择设备来源' }],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择设备来源'
                }
            }
        }
    ],
    [
        {
            name: 'sn',
            label: '设备序列号',
            rules: [{ required: true, message: '请输入设备序列号' }]
        },
        {
            name: 'supplier_id',
            label: '供应商',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择供应商'
                }
            }
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
            name: 'status_id',
            label: '设备状态',
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择设备状态'
                }
            }
        }
    ],
    [
        {
            name: 'manage_ip',
            label: 'IP'
        },
        {
            name: 'manage_port',
            label: '端口'
        }
    ],
    {
        name: 'remark',
        label: '备注',
        dataProps: {
            type: 'textarea',
            parentProps: {
                placeholder: '请输入备注',
                maxLength: 200,
                showCount: true,
                autoSize: { minRows: 3, maxRows: 5 }
            }
        }
    }
]

export const columns: (
    | ColumnGroupType<DeviceItemType>
    | ColumnType<DeviceItemType>
)[] = [
    {
        dataIndex: 'host_name',
        title: '主机名',
        key: 'host_name',
        width: 230
    },
    {
        dataIndex: 'type',
        title: '设备类型',
        key: 'type',
        width: 160,
        render: (type: Map) => {
            return type ? type.name : '-'
        }
    },
    {
        dataIndex: 'status',
        title: '状态',
        key: 'status',
        width: 160,
        render: (status: Map) => {
            return status ? (
                <Badge color={status?.color || 'default'} text={status?.name} />
            ) : (
                '-'
            )
        }
    },
    {
        dataIndex: 'sn',
        title: '序列号',
        key: 'sn',
        width: 160
    },
    {
        dataIndex: 'supplier',
        title: '供应商',
        key: 'supplier',
        width: 200,
        render: (supplier: Map) => {
            return supplier ? supplier.name : '-'
        }
    },
    {
        dataIndex: 'node',
        title: '节点',
        key: 'node',
        width: 200,
        render: (node: Map) => {
            return node ? node.name : '-'
        }
    },
    {
        dataIndex: 'ipmi',
        title: 'IPMI',
        key: 'ipmi',
        width: 160,
        render: (ipmi?: string) => {
            return ipmi ? ipmi : '-'
        }
    },
    {
        dataIndex: 'manage_ip',
        title: 'IP',
        key: 'manage_ip',
        width: 160,
        render: (ip?: string) => {
            return ip ? ip : '-'
        }
    },
    {
        dataIndex: 'manage_port',
        title: '端口',
        key: 'manage_port',
        width: 160,
        render: (manage_port?: string) => {
            return manage_port ? manage_port : '-'
        }
    },
    {
        dataIndex: 'source',
        title: '来源',
        key: 'source',
        width: 160,
        render: (source: Map) => {
            return source ? source.name : '-'
        }
    }
]

export const refreshActionKey = 'refresh'
export const addDictActionKey = 'add-dict'
export const batchExportActionKey = 'batch-export'
export const batchImportActionKey = 'batch-import'
export const downloadTemplateActionKey = 'download-template'

export const rightOptions: DataOptionItem[] = [
    {
        key: refreshActionKey,
        label: <Button type="primary">刷新</Button>
    },
    {
        key: downloadTemplateActionKey,
        label: <Button type="primary">下载模板</Button>
    }
]

export const leftOptions: DataOptionItem[] = [
    {
        key: addDictActionKey,
        label: <Button type="primary">添加字典</Button>
    },
    {
        key: batchImportActionKey,
        label: <Button type="primary">批量导入</Button>
    },
    {
        key: batchExportActionKey,
        label: <Button type="primary">批量导出</Button>
    }
]
