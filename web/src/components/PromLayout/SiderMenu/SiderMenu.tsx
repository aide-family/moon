import { useContext, FC, useState, useEffect } from 'react'

import { Button, Menu } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { GlobalContext } from '@/context'

import styles from './style/index.module.less'
import { LeftOutlined, RightOutlined } from '@ant-design/icons'
import { ItemType } from 'antd/es/menu/interface'

export type SiderMenuProps = {
    items?: ItemType[]
    inlineCollapsed?: boolean
    setCollapsed?: React.Dispatch<React.SetStateAction<boolean>>
}

const SiderMenu: FC<SiderMenuProps> = (props) => {
    const navigate = useNavigate()
    const location = useLocation()

    const { menus } = useContext(GlobalContext)
    const { items = menus, inlineCollapsed, setCollapsed } = props

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

    const handleOnSelect = (key: string) => {
        navigate(key)
    }

    useEffect(() => {
        setSelectedKeys([location.pathname])

        const openKey = location.pathname.split('/').slice(1)
        let keys: string[] = []
        let key: string = ''
        openKey.forEach((item) => {
            key += '/' + item
            if (key === '/home') {
                return
            }
            keys.push(key)
        })
        // 去掉最后一级
        openKey.pop()
        setOpenKeys([...keys, '/' + openKey.join('/')])
        setSelectedKeys(keys)
        setLocationPath(location.pathname)
    }, [location.pathname, inlineCollapsed])

    return (
        <>
            <Menu
                className={styles.SiderMenu}
                mode="inline"
                items={items}
                style={{
                    borderInlineEnd: 'none'
                }}
                openKeys={inlineCollapsed ? [] : openKeys}
                defaultOpenKeys={openKeys}
                onSelect={({ key }) => handleOnSelect(key)}
                selectedKeys={selectedKeys}
                onOpenChange={handleMenuOpenChange}
                // forceSubMenuRender
                // inlineCollapsed={inlineCollapsed}
            ></Menu>
            <Button
                type="text"
                style={{
                    // 固定定位到最底部
                    position: 'absolute',
                    bottom: 0,
                    left: 0,
                    width: '100%'
                }}
                onClick={() => setCollapsed?.(!inlineCollapsed)}
            >
                {inlineCollapsed ? <RightOutlined /> : <LeftOutlined />}
            </Button>
        </>
    )
}

export { SiderMenu }
