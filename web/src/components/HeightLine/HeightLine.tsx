import { FC, ReactNode } from 'react'

export type HeightLineProps = {
    height?: number | string
    children?: ReactNode
}

const HeightLine: FC<HeightLineProps> = (props) => {
    const { height = 8, children } = props
    return <div style={{ height, width: '100%' }}>{children}</div>
}

export { HeightLine }
