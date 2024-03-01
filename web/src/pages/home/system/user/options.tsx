import { ActionKey } from '@/apis/data'
import { DataFormItem } from '@/components/Data'

import { IconFont } from '@/components/IconFont/IconFont'
import { ManOutlined, WomanOutlined } from '@ant-design/icons'
import { Button, MenuProps } from 'antd'
import { UserListItem } from '@/apis/home/system/user/types.ts'
import { StatusBadge } from './child/StatusBadge'
import { Username } from './child/Username'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import { UserAvatar } from './child/UserAvatar'
import roleApi from '@/apis/home/system/role'
import { defaultRoleSelectReq } from '@/apis/home/system/role/types'
import { DefaultOptionType } from 'antd/es/select'
import { Status } from '@/apis/types'

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

const getRoleSelect = (keyword: string): Promise<DefaultOptionType[]> => {
    return roleApi
        .roleSelect({ ...defaultRoleSelectReq, keyword: keyword })
        .then((data) => {
            if (!data || !data.list) return []
            return data.list.map((item) => ({
                value: item.value,
                label: item.label
            }))
        })
}

const addFormItems: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'username',
            label: '用户名称',
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
            name: 'password',
            label: '密码',
            rules: [
                {
                    required: true
                },
                {
                    min: 6
                }
            ]
        }
    ],
    [
        {
            name: 'nickname',
            label: '用户昵称'
        },
        {
            name: 'phone',
            label: '手机号',
            rules: [
                {
                    required: true
                },
                {
                    pattern: /^1[34578]\d{9}$/
                }
            ]
        }
    ],
    [
        {
            name: 'email',
            label: '邮箱',
            rules: [
                {
                    required: true
                },
                {
                    pattern:
                        /^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$/
                }
            ]
        },
        {
            name: 'gender',
            label: '性别',
            dataProps: {
                type: 'radio-group',
                parentProps: {
                    options: [
                        {
                            label: (
                                <span>
                                    <ManOutlined style={{ color: '#1890ff' }} />
                                    {' 男'}
                                </span>
                            ),
                            value: 1
                        },
                        {
                            label: (
                                <span>
                                    <WomanOutlined
                                        style={{ color: '#f759ab' }}
                                    />
                                    {' 女'}
                                </span>
                            ),
                            value: 2
                        }
                    ]
                }
            }
        }
    ],
    {
        name: 'roleIds',
        label: '用户角色',
        dataProps: {
            type: 'select-fetch',
            parentProps: {
                handleFetch: getRoleSelect,
                defaultOptions: [],
                selectProps: {
                    mode: 'multiple',
                    placeholder: '请选择用户角色'
                }
            }
        }
    }
]
const editFormItems: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'username',
            label: '用户名称',
            rules: [
                {
                    required: true
                }
            ]
        },
        {
            name: 'nickname',
            label: '用户昵称'
        }
    ],
    [
        {
            name: 'phone',
            label: '手机号'
        },
        {
            name: 'email',
            label: '邮箱'
        }
    ],
    [
        {
            name: 'gender',
            label: '性别',
            dataProps: {
                type: 'radio-group',
                parentProps: {
                    options: [
                        {
                            label: (
                                <span>
                                    <ManOutlined style={{ color: '#1890ff' }} />
                                    {' 男'}
                                </span>
                            ),
                            value: 1
                        },
                        {
                            label: (
                                <span>
                                    <WomanOutlined
                                        style={{ color: '#f759ab' }}
                                    />
                                    {' 女'}
                                </span>
                            ),
                            value: 2
                        }
                    ]
                }
            }
        },
        {
            name: 'status',
            label: '是否激活',
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
            name: 'avatar',
            label: '用户头像'
        }
    ],
    {
        name: 'roleIds',
        label: '用户角色',
        dataProps: {
            type: 'select-fetch',
            parentProps: {
                handleFetch: getRoleSelect,
                defaultOptions: [],
                selectProps: {
                    mode: 'multiple',
                    placeholder: '请选择用户角色'
                }
            }
        }
    }
]

const operationItems = (userItem: UserListItem): MenuProps['items'] => [
    {
        key: ActionKey.CHANGE_STATUS,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-edit" />}
                disabled={userItem.id === 1}
            >
                状态修改
            </Button>
        )
    },
    {
        key: ActionKey.EDIT,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-edit" />}
                disabled={userItem.id === 1}
            >
                编辑用户
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
                disabled={userItem.id === 1}
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
                disabled={userItem.id === 1}
                icon={
                    <IconFont
                        type="icon-shanchu-copy"
                        style={{ color: 'red' }}
                    />
                }
            >
                删除用户
            </Button>
        )
    }
]

export type UserColumnType =
    | ColumnType<UserListItem>
    | ColumnGroupType<UserListItem>

export const columns: UserColumnType[] = [
    {
        title: '头像',
        dataIndex: 'avatar',
        key: 'avatar',
        width: 100,
        render: (_: string, item: UserListItem) => {
            return <UserAvatar {...item} />
        }
    },
    {
        title: '姓名',
        dataIndex: 'username',
        key: 'username',
        width: 120,
        render: (_: string, record: UserListItem) => {
            return <Username {...record} />
        }
    },
    {
        title: '昵称',
        dataIndex: 'nickname',
        key: 'nickname',
        width: 200,
        ellipsis: true,
        render: (text: String) => {
            return <>{text || '-'}</>
        }
    },
    {
        title: '状态',
        dataIndex: 'status',
        key: 'status',
        width: 120,
        render: (_: string, record: UserListItem) => {
            return <StatusBadge {...record} />
        }
    },
    {
        title: '手机号',
        dataIndex: 'phone',
        key: 'phone',
        width: 200
    },
    {
        title: '邮箱',
        dataIndex: 'email',
        key: 'email',
        width: 300
    }
]

export const options = {
    searchItems,
    addFormItems,
    editFormItems,
    operationItems
}

export default function userOptions(): typeof options {
    return options
}
