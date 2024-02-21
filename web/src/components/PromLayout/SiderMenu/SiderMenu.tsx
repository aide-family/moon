import { useContext, FC, useState, useEffect } from 'react'

import type { ItemType } from 'antd/es/menu/hooks/useItems'

import { Menu } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { GlobalContext } from '@/context'

import styles from './style/index.module.less'

export type SiderMenuProps = {
    items?: ItemType[]
    inlineCollapsed?: boolean
}

const SiderMenu: FC<SiderMenuProps> = (props) => {
    const navigate = useNavigate()
    const location = useLocation()

    const { menus } = useContext(GlobalContext)
    const { items = menus, inlineCollapsed } = props

    const [openKeys, setOpenKeys] = useState<string[]>([])
    const [selectedKeys, setSelectedKeys] = useState<string[]>([])
    const [locationPath, setLocationPath] = useState<string>(location.pathname)

    const handleMenuOpenChange = (keys: string[]) => {
        let openKeyList: string[] = keys
        if (openKeyList.length === 0) {
            openKeyList = locationPath.split('/').slice(1)
            // 去掉最后一级
            openKeyList.pop()
            openKeyList = ['/' + openKeyList.join('/')]
        }
        setOpenKeys(openKeyList)
    }

    const handleOnSelect = (key: string, keyPath: string[]) => {
        setSelectedKeys(keyPath)
        handleMenuOpenChange(keyPath)
        navigate(key)
        setOpenKeys(keyPath)
    }

    useEffect(() => {
        setSelectedKeys([location.pathname])

        const openKey = location.pathname.split('/').slice(1)
        // 去掉最后一级
        openKey.pop()
        setOpenKeys(['/' + openKey.join('/')])
        setLocationPath(location.pathname)
    }, [location.pathname])

    return (
        <Menu
            className={styles.SiderMenu}
            mode="inline"
            items={items}
            style={{
                borderInlineEnd: 'none'
            }}
            openKeys={inlineCollapsed ? [] : openKeys}
            onSelect={({ keyPath, key }) => handleOnSelect(key, keyPath)}
            selectedKeys={selectedKeys}
            onOpenChange={handleMenuOpenChange}
            forceSubMenuRender
        />
    )
}

export { SiderMenu }
