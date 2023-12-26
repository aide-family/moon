import { Button, MenuProps, Tag } from 'antd'
import { IconFont } from '@/components/IconFont/IconFont.tsx'
import { operationItems } from '@/components/Data/DataOption/option.tsx'
import { DataFormItem } from '@/components/Data'
import { StrategyItemType } from '@/apis/home/monitor/strategy/types'
import { ActionKey } from '@/apis/data'

export const OP_KEY_STRATEGY_GROUP_LIST = 'strategy-group-list'

export const tableOperationItems = (
    item: StrategyItemType
): MenuProps['items'] => [
    {
        key: ActionKey.STRATEGY_GROUP_LIST,
        label: (
            <Button
                type="link"
                size="small"
                icon={<IconFont type="icon-linkedin-fill" />}
            >
                规则组列表
            </Button>
        )
    },
    ...(operationItems(item) as [])
]

export const searchItems: DataFormItem[] = [
    {
        name: 'alert',
        label: '规则名称'
    },
    {
        name: 'status',
        label: '规则状态',
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
        name: 'duration',
        label: '持续时间',
        dataProps: {
            type: 'select',
            parentProps: {
                mode: 'multiple',
                placeholder: '请选择规则持续时间的范围',
                options: [
                    {
                        label: '3m<',
                        value: '3m<'
                    }
                ]
            }
        }
    }
]

export const durationOptions = [
    {
        label: '秒',
        value: 's'
    },
    {
        label: '分钟',
        value: 'm'
    },
    {
        label: '小时',
        value: 'h'
    },
    {
        label: '天',
        value: 'd'
    }
]
