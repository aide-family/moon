import { ActionKey } from '@/apis/data'
import { Status, StatusMap } from '@/apis/types'
import { DataFormItem } from '@/components/Data'

import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps } from 'antd'
import { RoleListItem } from '@/apis/home/system/role/types.ts'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import dayjs from 'dayjs'

const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '关键词'
    },
    {
        name: 'status',
        label: '状态',
        dataProps: {
            type: 'radio-group',
            parentProps: {
                optionType: 'button',
                options: [
                    {
                        label: '全部',
                        value: Status.STATUS_UNKNOWN
                    },
                    {
                        label: '启用',
                        value: Status.STATUS_ENABLED
                    },
                    {
                        label: '禁用',
                        value: Status.STATUS_DISABLED
                    }
                ]
            }
        }
    }
]

const addFormItems: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'name',
            label: '角色名称',
            rules: [
                {
                    required: true
                },
                {
                    min: 3
                }
            ]
        }
    ],
    [
        {
            name: 'remark',
            label: '备注',
            rules: [
                // 不允许出现特殊字符
                {
                    pattern: /^[\u4e00-\u9fa5\w]+$/
                }
            ],
            dataProps: {
                type: 'textarea',
                parentProps: {
                    rows: 4,
                    showCount: true,
                    maxLength: 200,
                    placeholder: '请输入备注...'
                }
            }
        }
    ]
]
const editFormItems: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'name',
            label: '角色名称',
            rules: [
                {
                    required: true
                },
                {
                    min: 3
                }
            ]
        },
        {
            name: 'status',
            label: '是否启用',
            dataProps: {
                type: 'radio-group',
                parentProps: {
                    options: [
                        {
                            label: StatusMap[Status.STATUS_ENABLED].text,
                            value: Status.STATUS_ENABLED
                        },
                        {
                            label: StatusMap[Status.STATUS_DISABLED].text,
                            value: Status.STATUS_DISABLED
                        }
                    ]
                }
            }
        }
    ],
    [
        {
            name: 'remark',
            label: '备注',
            rules: [
                // 不允许出现特殊字符
                {
                    pattern: /^[\u4e00-\u9fa5\w]+$/
                }
            ],
            dataProps: {
                type: 'textarea',
                parentProps: {
                    rows: 4,
                    showCount: true,
                    maxLength: 200,
                    placeholder: '请输入备注...'
                }
            }
        }
    ]
]

const ADMIN_ROLE_ID = 1

const operationItems = (item: RoleListItem): MenuProps['items'] => [
    {
        key: ActionKey.EDIT,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-edit" />}
                disabled={item.id === ADMIN_ROLE_ID}
            >
                编辑
            </Button>
        )
    },
    {
        key: ActionKey.ASSIGN_AUTH,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-configure" />}
                disabled={item.id === ADMIN_ROLE_ID}
            >
                分配权限
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
    {
        type: 'divider'
    },
    {
        key: ActionKey.DELETE,
        label: (
            <Button
                size="small"
                danger
                type="link"
                disabled={item.id === ADMIN_ROLE_ID}
                icon={
                    <IconFont
                        type="icon-shanchu-copy"
                        style={{ color: 'red' }}
                    />
                }
            >
                删除
            </Button>
        )
    }
]

export type RoleColumnType =
    | ColumnType<RoleListItem>
    | ColumnGroupType<RoleListItem>

export const columns: RoleColumnType[] = [
    {
        title: '角色名称',
        dataIndex: 'name',
        key: 'name',
        width: 220
    },
    {
        title: '角色状态',
        dataIndex: 'status',
        key: 'status',
        width: 100,
        render: (status: Status) => {
            return (
                <Badge
                    color={StatusMap[status].color}
                    text={StatusMap[status].text}
                />
            )
        }
    },
    {
        title: '备注',
        dataIndex: 'remark',
        key: 'remark',
        // width: 200,
        ellipsis: true
    },
    {
        title: '创建时间',
        dataIndex: 'createdAt',
        key: 'createdAt',
        width: 180,
        render: (createdAt: number | string) => {
            return dayjs(+createdAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    },
    {
        title: '更新时间',
        dataIndex: 'updatedAt',
        key: 'updatedAt',
        width: 180,
        render: (updatedAt: number | string) => {
            return dayjs(+updatedAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        }
    }
]

export const options = {
    /**搜索角色配置 */
    searchItems,
    /**添加角色配置 */
    addFormItems,
    /**编辑角色配置 */
    editFormItems,
    /**操作角色配置 */
    operationItems
}

export default function roleOptions(): typeof options {
    return options
}
