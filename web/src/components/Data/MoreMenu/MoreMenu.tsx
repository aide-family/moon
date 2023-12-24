import { FC } from 'react'

import type { MenuProps } from 'antd/es/menu'
import { Dropdown } from 'antd'

import { IconFont } from '@/components/IconFont/IconFont'

export type MoreMenuProps = {
    items: MenuProps['items']
    onClick?: (key: string) => void
}

const MoreMenu: FC<MoreMenuProps> = (props) => {
    const { items, onClick } = props
    return (
        <Dropdown
            menu={{ items, onClick: ({ key }) => onClick?.(key) }}
            trigger={['click']}
        >
            <a onClick={(e) => e.preventDefault()}>
                <IconFont type="icon-more" />
            </a>
        </Dropdown>
    )
}

export default MoreMenu
