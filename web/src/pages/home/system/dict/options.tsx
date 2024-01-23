import { ActionKey, categoryData } from '@/apis/data'
import { Status, StatusMap } from '@/apis/types'
import { DataFormItem } from '@/components/Data'

import { IconFont } from '@/components/IconFont/IconFont'
import { Button, MenuProps } from 'antd'
import { DictListItem } from '@/apis/home/system/dict/types.ts'

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
            rules: [
                //正则校验css颜色
                {
                    pattern:
                        /^(#([0-9a-fA-F]{3}){1,2}|(rgb|hsl)a?\((-?\d+%?[,\s]+){2,3}\s*[\d\.]+%?\))$/,
                    message: '请输入正确的颜色'
                }
            ]
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
            rules: [
                //正则校验css颜色
                {
                    pattern:
                        /^(#([0-9a-fA-F]{3}){1,2}|(rgb|hsl)a?\((-?\d+%?[,\s]+){2,3}\s*[\d\.]+%?\))$/,
                    message: '请输入正确的颜色'
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
                            label: StatusMap[Status['STATUS_ENABLED']].text,
                            value: Status['STATUS_ENABLED']
                        },
                        {
                            label: StatusMap[Status['STATUS_DISABLED']].text,
                            value: Status['STATUS_DISABLED']
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

export default function dictOptions(): typeof options {
    return options
}
