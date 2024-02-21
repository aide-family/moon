import { ThemeType } from '@/context'
import { theme } from 'antd'

export const getUseTheme = (t?: ThemeType) => {
    return t === 'dark' ? theme.darkAlgorithm : undefined
}
