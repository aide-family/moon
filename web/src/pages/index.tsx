import React, { Suspense, useState } from 'react'
import { App, ConfigProvider } from 'antd'
import { createHashRouter, RouterProvider } from 'react-router-dom'
import zhCN from 'antd/locale/zh_CN'

import Loading from '@/components/Loading'
import { GlobalContext, GlobalContextType } from '@/context'
import useStorage from '@/utils/storage'
import { SizeType } from 'antd/es/config-provider/SizeContext'
import { breadcrumbNameMap, defaultMenuItems } from './menus'
import { routers } from './router'
import Logo from '@/assets/logo.svg'

import styles from './style/index.module.less'
import { UserListItem } from '@/apis/home/system/user/types'

export type UserType = {
    user_name: string
    nick_name?: string
    gender?: number
    avatar?: string
    id?: string
}

export type SpaceType = {
    id: string
    name: string
    logo?: string
    is_team: number
}

const defaultSize: SizeType =
    (localStorage.getItem('size') as SizeType) || 'middle'
const defaultUser: UserListItem = {
    username: '未登录',
    id: 0,
    email: '',
    phone: '',
    status: 0,
    remark: '',
    avatar: '',
    createdAt: 0,
    updatedAt: 0,
    deletedAt: 0,
    roles: [],
    nickname: '',
    gender: 0
}
const defaultSpaceInfo: SpaceType = {
    id: '1',
    name: 'AideDevOps',
    logo: Logo,
    is_team: 1
}

const Index: React.FC = () => {
    const [size, setSize] = useStorage<SizeType>('size', defaultSize)
    const [user, setUser] = useStorage<UserListItem>('user', defaultUser)
    const [spaceInfo, setSpaceInfo] = useStorage<SpaceType>(
        'spaceInfo',
        defaultSpaceInfo
    )
    const [spaces, setSpaces] = useStorage<SpaceType[]>('spaces', [])
    const [layoutContentElement, setLayoutContentElement] =
        useState<HTMLElement | null>(null)
    const [token, setToken, removeToken] = useStorage<string>('token', '')
    const [intervalId, setIntervalId] = useState<any>()

    const contextValue: GlobalContextType = {
        size: size,
        setSize: setSize,
        setUser: setUser,
        user: user,
        spaceInfo: spaceInfo,
        setSpaceInfo: setSpaceInfo,
        layoutContentElement: layoutContentElement,
        setLayoutContentElement: setLayoutContentElement,
        menus: defaultMenuItems,
        breadcrumbNameMap: breadcrumbNameMap,
        spaces: spaces,
        setSpaces: setSpaces,
        intervalId: intervalId,
        setIntervalId: setIntervalId,
        token: token,
        setToken: setToken,
        removeToken: removeToken
    }

    return (
        <div className={styles.App}>
            <ConfigProvider locale={zhCN}>
                <GlobalContext.Provider value={contextValue}>
                    <App className={styles.widthHight100}>
                        <Suspense fallback={<Loading />}>
                            <RouterProvider
                                router={createHashRouter(routers)}
                            />
                        </Suspense>
                    </App>
                </GlobalContext.Provider>
            </ConfigProvider>
        </div>
    )
}

export default Index
