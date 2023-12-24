import { useContext, FC, useState, useEffect } from 'react'

import type { ItemType } from 'antd/es/menu/hooks/useItems'
import type { MenuProps } from 'antd'

import { Menu } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { GlobalContext } from '@/context'

import styles from './style/index.module.less'

export type SiderMenuProps = {
    items?: ItemType[]
}

const SiderMenu: FC<SiderMenuProps> = (props) => {
    const { menus } = useContext(GlobalContext)
    const { items = menus } = props

    const [openKeys, setOpenKeys] = useState<string[]>([])
    const [selectedKeys, setSelectedKeys] = useState<string[]>([])

    const navigate = useNavigate()
    const location = useLocation()

    const onClick: MenuProps['onClick'] = (e) => {
        navigate(e.key)
        setOpenKeys(e.keyPath)
    }

    useEffect(() => {
        setSelectedKeys([location.pathname])
    }, [location.pathname])

    return (
        <Menu
            className={styles.SiderMenu}
            onClick={onClick}
            mode="inline"
            items={items}
            defaultOpenKeys={openKeys}
            defaultSelectedKeys={selectedKeys}
            openKeys={openKeys}
            selectedKeys={selectedKeys}
            onOpenChange={(keys) => setOpenKeys(keys)}
        />
    )
}

export { SiderMenu }
