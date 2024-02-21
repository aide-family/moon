import { GlobalContext, ThemeType } from '@/context'
import { MoonOutlined, SunOutlined } from '@ant-design/icons'
import React, { useContext } from 'react'

export interface ThemeButtonProps {}

export const ThemeButton: React.FC<ThemeButtonProps> = () => {
    const { sysTheme, setSysTheme } = useContext(GlobalContext)

    const handleThemeChange = () => {
        const t: ThemeType = sysTheme === 'light' ? 'dark' : 'light'
        setSysTheme?.(t)
    }

    const ThemeIcon = () => {
        return sysTheme === 'light' ? (
            <MoonOutlined onClick={handleThemeChange} />
        ) : (
            <SunOutlined onClick={handleThemeChange} />
        )
    }

    return <ThemeIcon />
}
