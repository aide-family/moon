import { useContext, FC, useState, useEffect } from 'react'

import { MenuProps } from 'antd'

import { Avatar, Dropdown } from 'antd'
import { useLocation, useNavigate } from 'react-router-dom'
import { IconFont } from '@/components/IconFont/IconFont'
import Logo from '@/assets/logo.svg'
import { GlobalContext } from '@/context'
import { ActionKey } from '@/apis/data'
import { ChangePass } from './ChangePass'

const UserInfo: FC = () => {
    const navigate = useNavigate()
    const local = useLocation()

    const { user, intervalId, removeAuthToken, setRedirectPathName } =
        useContext(GlobalContext)

    const [pathname, setPathname] = useState<string>('/')
    const [openChangePassModal, setOpenChangePassModal] =
        useState<boolean>(false)

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
            icon: <IconFont type="icon-mima" />
        },
        {
            key: ActionKey.SWITCH_ROLE,
            label: '切换角色',
            icon: <IconFont type="icon-kehuguanli1" />
        },
        {
            type: 'divider'
        },
        {
            key: ActionKey.LOGOUT,
            label: '退出登录',
            icon: <IconFont type="icon-logout-" />,
            danger: true
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
            case ActionKey.CHANGE_PASSWORD:
                setOpenChangePassModal(true)
                break
            case ActionKey.SWITCH_ROLE:
                // TODO
                break
            default:
                break
        }
    }

    return (
        <>
            <ChangePass
                title="修改密码"
                open={openChangePassModal}
                onOk={() => setOpenChangePassModal(false)}
                onCancel={() => setOpenChangePassModal(false)}
            />
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
