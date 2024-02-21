import { createContext } from 'react'
import { SizeType } from 'antd/es/config-provider/SizeContext'
import type { ItemType } from 'antd/es/menu/hooks/useItems'
import type { SpaceType } from './pages'
import { UserListItem } from './apis/home/system/user/types'

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
    spaceInfo?: SpaceType
    setSpaceInfo?: (spaceInfo: SpaceType) => void
    layoutContentElement?: HTMLElement | null
    setLayoutContentElement?: (element: HTMLElement | null) => void
    menus?: ItemType[]
    breadcrumbNameMap?: Record<string, string>
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
}

export const GlobalContext = createContext<GlobalContextType>({
    lang: 'zh-CN',
    setLang: () => {},
    sysTheme: 'light',
    setSysTheme: () => {},
    size: 'middle',
    setSize: () => {}
})

// const tokenKey = 'token'

// export function getToken() {
//     return localStorage.getItem(tokenKey) || ''
// }

// export function setToken(token: string) {
//     localStorage.setItem(tokenKey, token)
// }

// export function removeToken() {
//     localStorage.removeItem(tokenKey)
// }

// export function getSpaceID() {
//     return localStorage.getItem('spaceId') || ''
// }

// export function setSpaceID(spaceId: string) {
//     localStorage.setItem('spaceId', spaceId)
// }

// export function setUserInfo(userInfo: UserType) {
//     localStorage.setItem('user', JSON.stringify(userInfo))
// }

// export function setSpaces(spaces: SpaceType[]) {
//     localStorage.setItem('spaces', JSON.stringify(spaces))
// }
