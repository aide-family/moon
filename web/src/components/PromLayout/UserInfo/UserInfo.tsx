import { useContext, FC } from 'react'

import type { MenuProps } from 'antd'

import { Avatar, Dropdown } from 'antd'
import { useNavigate } from 'react-router-dom'
import { IconFont } from '@/components/IconFont/IconFont'
import { GlobalContext, removeToken } from '@/context'

const UserInfo: FC = () => {
    const { user, intervalId } = useContext(GlobalContext)
    const navigate = useNavigate()

    const items: MenuProps['items'] = [
        {
            key: 'self-setting',
            label: '个人设置',
            icon: <IconFont type="icon-user_role" />
        },
        {
            key: 'self-change-password',
            label: '修改密码',
            icon: <IconFont type="icon-Password" />
        },
        {
            key: 'self-change-role',
            label: '切换角色',
            icon: <IconFont type="icon-role1" />
        },
        {
            type: 'divider'
        },
        {
            key: 'self-logout',
            label: '退出登录',
            icon: <IconFont type="icon-logout-" />
        }
    ]

    const handleMenuOnClick: MenuProps['onClick'] = ({ key }) => {
        switch (key) {
            case 'self-logout':
                clearInterval(intervalId)
                removeToken()
                navigate('/login')
                break
        }
    }

    return (
        <>
            <Dropdown menu={{ items, onClick: handleMenuOnClick }}>
                <Avatar
                    src={<img src={user?.avatar} alt="avatar" />}
                    shape="square"
                >
                    {user?.username}
                </Avatar>
            </Dropdown>
        </>
    )
}

export { UserInfo }
