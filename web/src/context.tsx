import { createContext } from 'react'
import { SizeType } from 'antd/es/config-provider/SizeContext'
import type { ItemType } from 'antd/es/menu/hooks/useItems'
import type { SpaceType } from './pages'
import { UserListItem } from './apis/home/system/user/types'
import { breadcrumbNameType } from './pages/menus'

export type ThemeType = 'light' | 'dark'

export type GlobalContextType = {
    lang?: string
    setLang?: (value: string) => void
    sysTheme?: ThemeType
    setSysTheme?: (value: ThemeType) => void
    size?: SizeType
    setSize?: (size?: SizeType) => void
    user?: UserListItem
    setUser?: (user: UserListItem) => void
    layoutContentElement?: HTMLElement | null
    setLayoutContentElement?: (element: HTMLElement | null) => void
    menus?: ItemType[]
    breadcrumbNameMap?: Record<string, breadcrumbNameType>
    spaces?: SpaceType[]
    setSpaces?: (spaces: SpaceType[]) => void
    setAuthToken?: (token: string) => void
    autToken?: string
    removeAuthToken?: () => void
    intervalId?: any
    setIntervalId?: (intervalId: any) => void
    // 自动刷新
    autoRefresh?: boolean
    setAutoRefresh?: (autoRefresh: boolean) => void
    redirectPathName?: string
    setRedirectPathName?: (redirectPathName: string) => void
    ws?: WebSocket
    reltimeAlarmShowRowColor?: boolean
    setReltimeAlarmShowRowColor?: (reltimeAlarmShowRowColor: boolean) => void
}

export const GlobalContext = createContext<GlobalContextType>({
    lang: 'zh-CN',
    setLang: () => {},
    sysTheme: 'light',
    setSysTheme: () => {},
    size: 'middle',
    setSize: () => {}
})
