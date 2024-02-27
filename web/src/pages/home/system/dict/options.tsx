import { ActionKey, categoryData } from '@/apis/data'
import { Category, Status, StatusMap } from '@/apis/types'
import { DataFormItem } from '@/components/Data'

import { IconFont } from '@/components/IconFont/IconFont'
import { Badge, Button, MenuProps } from 'antd'
import { DictListItem } from '@/apis/home/system/dict/types.ts'
import { ColumnGroupType, ColumnType } from 'antd/es/table'
import dayjs from 'dayjs'

const searchItems: DataFormItem[] = [
    {
        name: 'keyword',
        label: '关键词'
    },
    {
        name: 'category',
        label: '字典类别',
        dataProps: {
            type: 'select',
            parentProps: {
                placeholder: '请选择字典类别',
                options: Object.entries(categoryData).map(([key, value]) => ({
                    label: value,
                    value: Number(key)
                }))
            }
        }
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
            label: '字典名称',
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
            name: 'category',
            label: '字典类别',
            rules: [
                {
                    required: true
                }
            ],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择字典类别',
                    options: Object.entries(categoryData).map(
                        ([key, value]) => ({ label: value, value: Number(key) })
                    )
                }
            }
        },
        {
            name: 'color',
            label: '字典颜色',
            dataProps: {
                type: 'color',
                parentProps: {
                    showText: true,
                    defaultFormat: 'hex'
                }
            }
        }
    ],
    [
        {
            name: 'remark',
            label: '字典备注',
            rules: [
                {
                    max: 255,
                    message: '备注最多255个字符'
                }
            ],
            dataProps: {
                type: 'textarea',
                parentProps: {
                    maxLength: 255,
                    showCount: true,
                    rows: 4,
                    placeholder: '请输入字典备注'
                }
            }
        }
    ]
]
const editFormItems: (DataFormItem | DataFormItem[])[] = [
    [
        {
            name: 'name',
            label: '字典名称',
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
            name: 'category',
            label: '字典类别',
            rules: [
                {
                    required: true
                }
            ],
            dataProps: {
                type: 'select',
                parentProps: {
                    placeholder: '请选择字典类别',
                    options: Object.entries(categoryData).map(
                        ([key, value]) => ({ label: value, value: Number(key) })
                    )
                }
            }
        }
    ],
    [
        {
            name: 'color',
            label: '字典颜色',
            dataProps: {
                type: 'color',
                parentProps: {
                    showText: true,
                    defaultFormat: 'hex'
                }
            }
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
            label: '字典备注',
            rules: [
                {
                    max: 255,
                    message: '备注最多255个字符'
                }
            ],
            dataProps: {
                type: 'textarea',
                parentProps: {
                    maxLength: 255,
                    showCount: true,
                    rows: 4,
                    placeholder: '请输入字典备注'
                }
            }
        }
    ]
]

const operationItems = (_: DictListItem): MenuProps['items'] => [
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

export type DictColumnType =
    | ColumnGroupType<DictListItem>
    | ColumnType<DictListItem>

export const columns: DictColumnType[] = [
    {
        title: '字典名称',
        dataIndex: 'name',
        key: 'name',
        width: 220
    },
    {
        title: '字典类型',
        dataIndex: 'category',
        key: 'category',
        width: 100,
        render: (category: Category) => {
            return categoryData[category]
        }
    },
    {
        title: '字典颜色',
        dataIndex: 'color',
        key: 'color',
        align: 'center',
        width: 220,
        render: (color: string) => {
            return (
                <Badge
                    color={color}
                    text={color}
                    style={{
                        backgroundColor: color,
                        color: '#fff',
                        width: '60%',
                        textAlign: 'center'
                    }}
                />
            )
        }
    },
    {
        title: '字典状态',
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
        // TODO 两行溢出显示省略号
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

export default function dictOptions(): typeof options {
    return options
}
