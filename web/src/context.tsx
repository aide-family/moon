import { createContext } from 'react'
import { SizeType } from 'antd/es/config-provider/SizeContext'
import type { ItemType } from 'antd/es/menu/hooks/useItems'
import type { SpaceType, UserType } from './pages'
import { UserListItem } from './apis/home/system/user/types'

export type GlobalContextType = {
    lang?: string
    setLang?: (value: string) => void
    theme?: string
    setTheme?: (value: string) => void
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
    setToken?: (token: string) => void
    intervalId?: any
    setIntervalId?: (intervalId: any) => void
}

export const GlobalContext = createContext<GlobalContextType>({
    lang: 'zh-CN',
    setLang: () => {},
    theme: 'light',
    setTheme: () => {},
    size: 'middle',
    setSize: () => {},
    setToken: setToken
})

const tokenKey = 'token'

export function getToken() {
    return localStorage.getItem(tokenKey) || ''
}

export function setToken(token: string) {
    localStorage.setItem(tokenKey, token)
}

export function removeToken() {
    localStorage.removeItem(tokenKey)
}

export function getSpaceID() {
    return localStorage.getItem('spaceId') || ''
}

export function setSpaceID(spaceId: string) {
    localStorage.setItem('spaceId', spaceId)
}

export function setUserInfo(userInfo: UserType) {
    localStorage.setItem('user', JSON.stringify(userInfo))
}

export function setSpaces(spaces: SpaceType[]) {
    localStorage.setItem('spaces', JSON.stringify(spaces))
}
