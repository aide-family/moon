import {ActionKey} from '@/apis/data'
import {IconFont} from '@/components/IconFont/IconFont'
import {Button, MenuProps} from 'antd'

export const operationItems = (_: any): MenuProps['items'] => [
    {
        key: ActionKey.EDIT,
        label: (
            <Button
                size="small"
                type="link"
                icon={<IconFont type="icon-edit"/>}
            >
                编辑
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
                icon={
                    <IconFont
                        type="icon-shanchu-copy"
                        style={{color: 'red'}}
                    />
                }
            >
                删除
            </Button>
        )
    }
]
