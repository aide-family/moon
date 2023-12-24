import { DataOptionItem } from '@/components/Data/DataOption/DataOption.tsx'
import { Button, Tag } from 'antd'
import { IconFont } from '@/components/IconFont/IconFont.tsx'
import { operationItems } from '@/components/Data/DataOption/option.tsx'
import { DataFormItem } from '@/components/Data'

export const OP_KEY_STRATEGY_LIST = 'strategy-list'

export const tableOperationItems: DataOptionItem[] = [
    {
        key: OP_KEY_STRATEGY_LIST,
        label: (
            <Button
                type="link"
                size="small"
                icon={<IconFont type="icon-linkedin-fill" />}
            >
                策略列表
            </Button>
        )
    },
    ...(operationItems as any)
]

export const searchItems: DataFormItem[] = [
    {
        name: 'name',
        label: '规则组名称'
    },
    {
        name: 'status',
        label: '规则组状态',
        dataProps: {
            type: 'select',
            parentProps: {
                placeholder: '请选择规则状态',
                mode: 'multiple',
                options: [
                    {
                        label: <Tag color="success">启用</Tag>,
                        value: '1'
                    },
                    {
                        label: <Tag color="error">禁用</Tag>,
                        value: '0'
                    }
                ]
            }
        }
    },
    {
        name: 'categories',
        label: '规则分类',
        dataProps: {
            type: 'select',
            parentProps: {
                placeholder: '请选择规则分类',
                options: [
                    {
                        label: '全部',
                        value: ''
                    }
                    // TODO 需要从接口加载
                ]
            }
        }
    },
    {
        name: 'strategy_total',
        label: '规则数量',
        dataProps: {
            type: 'select',
            parentProps: {
                placeholder: '请选择规则数量',
                options: [
                    {
                        label: '全部',
                        value: ''
                    },
                    {
                        label: '0-10',
                        value: '<10'
                    },
                    {
                        label: '10-20',
                        value: '10-20'
                    }
                ]
            }
        }
    }
]
