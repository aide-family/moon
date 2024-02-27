import { ActionKey, domainTypeData, moduleTypeData } from '@/apis/data'
import { DomainType, ModuleType, Status, StatusMap } from '@/apis/types'
import { DataFormItem } from '@/components/Data'

import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps } from 'antd'
import { ApiAuthListItem } from '@/apis/home/system/auth/types.ts'
import dayjs from 'dayjs'
import { ColumnGroupType, ColumnType } from 'antd/es/table'

const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', ' HEAD', 'OPTIONS']

const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '关键词'
    }
]

const addFormItems: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'name',
            label: '接口名称',
            rules: [
                {
                    required: true
                },
                {
                    min: 2,
                    max: 20
                }
            ]
        }
    ],
    [
        {
            name: 'domain',
            label: '所属领域',
            rules: [
                {
                    required: true
                }
            ],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择领域',
                    options: Object.entries(domainTypeData).map(
                        ([key, value]) => ({
                            label: value,
                            value: Number(key)
                        })
                    )
                }
            }
        },
        {
            name: 'module',
            label: '所属模块',
            rules: [
                {
                    required: true
                }
            ],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择模块',
                    options: Object.entries(moduleTypeData).map(
                        ([key, value]) => ({
                            label: value,
                            value: Number(key)
                        })
                    )
                }
            }
        }
    ],
    [
        {
            name: 'path',
            label: '接口路径',
            rules: [
                {
                    required: true,
                    message: '请输入接口路径'
                },
                {
                    // 满足uri规范
                    pattern: /^[a-zA-Z0-9\/\-\_]+$/
                }
            ]
        }
    ],
    [
        {
            name: 'method',
            label: '请求方法',
            rules: [
                {
                    required: true
                }
            ],
            dataProps: {
                type: 'radio-group',
                parentProps: {
                    options: methods.map((method) => ({
                        label: method,
                        value: method
                    }))
                }
            }
        }
    ],
    [
        {
            name: 'remark',
            label: '备注',
            dataProps: {
                type: 'textarea',
                parentProps: {
                    rows: 3,
                    showCount: true,
                    maxLength: 255,
                    placeholder: '请输入备注'
                }
            }
        }
    ]
]
const editFormItems: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'name',
            label: '接口名称',
            rules: [
                {
                    required: true
                },
                {
                    min: 2,
                    max: 20
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
            name: 'domain',
            label: '所属领域',
            rules: [
                {
                    required: true
                }
            ],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择领域',
                    options: Object.entries(domainTypeData).map(
                        ([key, value]) => ({
                            label: value,
                            value: Number(key)
                        })
                    )
                }
            }
        },
        {
            name: 'module',
            label: '所属模块',
            rules: [
                {
                    required: true
                }
            ],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择模块',
                    options: Object.entries(moduleTypeData).map(
                        ([key, value]) => ({
                            label: value,
                            value: Number(key)
                        })
                    )
                }
            }
        }
    ],
    [
        {
            name: 'path',
            label: '接口路径',
            rules: [
                {
                    required: true
                },
                {
                    // 满足uri规范
                    pattern: /^[a-zA-Z0-9\/\-\_]+$/
                }
            ]
        }
    ],
    [
        {
            name: 'method',
            label: '请求方法',
            rules: [
                {
                    required: true
                }
            ],
            dataProps: {
                type: 'radio-group',
                parentProps: {
                    options: methods.map((method) => ({
                        label: method,
                        value: method
                    }))
                }
            }
        }
    ],
    [
        {
            name: 'remark',
            label: '备注',
            dataProps: {
                type: 'textarea',
                parentProps: {
                    rows: 3,
                    showCount: true,
                    maxLength: 255,
                    placeholder: '请输入备注'
                }
            }
        }
    ]
]

const operationItems = (_: ApiAuthListItem): MenuProps['items'] => [
    {
        key: ActionKey.EDIT,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-edit" />}
            >
                编辑
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
                删除
            </Button>
        )
    }
]

export const columns: (
    | ColumnGroupType<ApiAuthListItem>
    | ColumnType<ApiAuthListItem>
)[] = [
    {
        title: '接口名称',
        dataIndex: 'name',
        key: 'name',
        width: 220
    },
    {
        title: '接口状态',
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
        title: '所属领域',
        dataIndex: 'domain',
        key: 'domain',
        width: 100,
        render: (domain: DomainType) => {
            return domainTypeData[domain]
        }
    },
    {
        title: '所属模块',
        dataIndex: 'module',
        key: 'module',
        width: 100,
        render: (module: ModuleType) => {
            return moduleTypeData[module]
        }
    },
    {
        // TODO 两行后省略
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

export default function authOptions(): typeof options {
    return options
}
