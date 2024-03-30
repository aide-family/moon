import { FC } from 'react'

import type { MenuProps } from 'antd/es/menu'
import { Dropdown } from 'antd'

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
            <a onClick={(e) => e.preventDefault()}>更多</a>
        </Dropdown>
    )
}

export default MoreMenu
