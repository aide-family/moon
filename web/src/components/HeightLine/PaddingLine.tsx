import { FC } from 'react'

export type PaddingLineProps = {
    padding?: number
    color?: string
    height?: number | string
    borderRadius?: number
}

const PaddingLine: FC<PaddingLineProps> = (props) => {
    const {
        padding = 12,
        color = '#f0f0f0',
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
