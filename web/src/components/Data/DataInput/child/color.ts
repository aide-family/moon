import { ThemeType } from '@/context'
import { GlobalToken } from 'antd'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

export const foregroundColor = '#fa541c'
export const tokenForegroundColor = '#9254de'

export const defaultTheme = (
    token: GlobalToken,
    theme?: ThemeType
): monaco.editor.IStandaloneThemeData => {
    return {
        base: theme == 'dark' ? 'vs-dark' : 'vs',
        inherit: false,
        rules: [
            {
                token: 'keyword',
                foreground: tokenForegroundColor,
                fontStyle: 'bold'
            }
        ],
        colors: {
            'editor.foreground': foregroundColor,
            'editor.background': token.colorBgContainer
        }
    }
}
