import { ActionKey } from '@/apis/data'
import { DataFormItem } from '@/components/Data'

import { IconFont } from '@/components/IconFont/IconFont'
import { ManOutlined, WomanOutlined } from '@ant-design/icons'
import { Button, MenuProps } from 'antd'

const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '关键词'
    }
]

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
                            value: 0
                        }
                    ]
                }
            }
        }
    ]
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
            name: 'mobile_phone',
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
                            label: '男',
                            value: 1
                        },
                        {
                            label: '女',
                            value: 0
                        }
                    ]
                }
            }
        },
        {
            name: 'enabled',
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
            name: 'avatars',
            label: '用户头像'
        }
    ]
]

const operationItems: MenuProps['items'] = [
    {
        key: ActionKey.CHANGE_STATUS,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-edit" />}
            >
                状态修改
            </Button>
        )
    },
    {
        key: ActionKey.DELETE,
        label: (
            <Button
                size="small"
                danger
                type="link"
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

export const options = {
    searchItems,
    addFormItems,
    editFormItems,
    operationItems
}

export default function userOptions(): typeof options {
    return options
}
