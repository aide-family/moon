import { FC } from 'react'

import type { MenuProps } from 'antd/es/menu'
import { Button, ButtonProps, Dropdown } from 'antd'
import { ActionKey } from '@/apis/data'

export type MoreMenuProps = {
    items: MenuProps['items']
    onClick?: (key: ActionKey) => void
    type?: 'a' | 'button'
    buttonProps?: ButtonProps
    children?: React.ReactNode
    trigger?: 'click' | 'hover'
}

const MoreMenu: FC<MoreMenuProps> = (props) => {
    const {
        items,
        trigger = 'click',
        onClick,
        type = 'a',
        buttonProps,
        children = '更多'
    } = props
    return (
        <Dropdown
            menu={{ items, onClick: ({ key }) => onClick?.(key as ActionKey) }}
            trigger={[trigger]}
        >
            {type === 'a' ? (
                <a onClick={(e) => e.preventDefault()}>{children}</a>
            ) : (
                <Button {...buttonProps}>{children}</Button>
            )}
        </Dropdown>
    )
}

export default MoreMenu
