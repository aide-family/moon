import { FC, useContext, useEffect, useState } from 'react'
import { DownOutlined } from '@ant-design/icons'
import { Dropdown, type MenuProps, Space, Typography } from 'antd'
import { GlobalContext } from '@/context'

import styles from './style/index.module.less'

const SpaceInfo: FC = () => {
    const { user, setSpaceInfo, spaces, spaceInfo } = useContext(GlobalContext)

    const [selectedKeys, setSelectedKeys] = useState<string[]>([])
    const [selectedKey, setSelectedKey] = useState<string>(
        spaceInfo?.name || ''
    )

    const [menuItems, setMenuItems] = useState<MenuProps['items']>([])

    useEffect(() => {
        if (spaces) {
            const self = spaces.filter((item) => item.is_team === 0)
            const team = spaces.filter((item) => item.is_team === 1)
            setMenuItems([
                {
                    key: 'self-space',
                    label: '个人空间',
                    type: 'group',
                    children: self.map((item) => {
                        return {
                            key: item.id,
                            label: item.name
                        }
                    })
                },
                {
                    type: 'divider'
                },
                {
                    key: 'team-space',
                    label: '团队空间',
                    type: 'group',
                    children: team.map((item) => {
                        return {
                            key: item.id,
                            label: item.name
                        }
                    })
                }
            ])
        }
    }, [spaces])

    return (
        <>
            <Dropdown
                menu={{
                    items: menuItems,
                    selectable: true,
                    defaultSelectedKeys: [user?.id ? `${user?.id}` : ''],
                    selectedKeys: selectedKeys,
                    mode: 'vertical',
                    onSelect: (info) => {
                        setSelectedKeys(info.selectedKeys)
                        const spanceInfo = spaces?.find(
                            (item) => item?.id === info.key
                        )
                        setSelectedKey(spanceInfo?.name || '无可用空间')
                        if (spanceInfo) {
                            setSpaceInfo?.({
                                id: spanceInfo.id,
                                name: spanceInfo.name,
                                is_team: 0
                            })
                        }
                    }
                }}
            >
                <Typography.Link>
                    <Space className={styles.spaceInfoSpace}>
                        {selectedKey ?? '选择空间'}
                        <DownOutlined />
                    </Space>
                </Typography.Link>
            </Dropdown>
        </>
    )
}

export { SpaceInfo }
