import { useContext, FC, useState, useEffect } from 'react'

import type { MenuProps } from 'antd'

import { Avatar, Dropdown } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { IconFont } from '@/components/IconFont/IconFont'
import Logo from '@/assets/logo.svg'
import { GlobalContext } from '@/context'
import { ActionKey } from '@/apis/data'

const UserInfo: FC = () => {
    const navigate = useNavigate()
    const local = useLocation()

    const { user, intervalId, removeAuthToken, setRedirectPathName } =
        useContext(GlobalContext)

    const [pathname, setPathname] = useState<string>('/')

    useEffect(() => {
        setPathname(local.pathname)
    }, [local.pathname])

    const items: MenuProps['items'] = [
        {
            key: ActionKey.SELF_SETTING,
            label: '个人设置',
            icon: <IconFont type="icon-user_role" />
        },
        {
            key: ActionKey.CHANGE_PASSWORD,
            label: '修改密码',
            icon: <IconFont type="icon-Password" />
        },
        {
            key: ActionKey.SWITCH_ROLE,
            label: '切换角色',
            icon: <IconFont type="icon-role1" />
        },
        {
            type: 'divider'
        },
        {
            key: ActionKey.LOGOUT,
            label: '退出登录',
            icon: <IconFont type="icon-logout-" />
        }
    ]

    const handleMenuOnClick: MenuProps['onClick'] = ({ key }) => {
        switch (key) {
            case ActionKey.LOGOUT:
                setRedirectPathName?.(pathname)
                clearInterval(intervalId)
                removeAuthToken?.()
                navigate('/login')
                break
        }
    }

    return (
        <>
            <Dropdown menu={{ items, onClick: handleMenuOnClick }}>
                <Avatar
                    src={<img src={user?.avatar || Logo} alt="avatar" />}
                    shape="square"
                >
                    {user?.username}
                </Avatar>
            </Dropdown>
        </>
    )
}

export { UserInfo }
