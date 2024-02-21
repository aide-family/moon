import { theme } from 'antd'
import { FC } from 'react'

export type PaddingLineProps = {
    padding?: number
    color?: string
    height?: number | string
    borderRadius?: number
}

const { useToken } = theme

const PaddingLine: FC<PaddingLineProps> = (props) => {
    const { token } = useToken()
    const {
        padding = 12,
        color = token.colorBorder,
        height = 1,
        borderRadius = 0
    } = props
    return (
        <div style={{ padding: `${padding}px 0`, width: '100%' }}>
            <div
                style={{
                    width: '100%',
                    height: height,
                    background: color,
                    borderRadius: borderRadius
                }}
            />
        </div>
    )
}

export { PaddingLine }
