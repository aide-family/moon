import { CSSProperties, useState, FC } from 'react'
import { CopyOutlined } from '@ant-design/icons'
import { Button, Space, Tooltip, App } from 'antd'

export type BaseTextProps = {
    children: string
    maxLine?: number
    width?: number | string
    showTooltip?: boolean
    style?: CSSProperties
    placeholder?: string
    copy?: boolean
    copyFunc?: (text: string) => Promise<any>
}

const BaseText: FC<BaseTextProps> = (props) => {
    const {
        children,
        maxLine = 1,
        width = '100%',
        showTooltip,
        style,
        placeholder,
        copy,
        copyFunc
    } = props

    const { message } = App.useApp()

    const Width = typeof width === 'number' ? `${width}px` : width
    const Style: CSSProperties = {
        // 超出显示省略号, 最多两行
        display: '-webkit-box',
        WebkitBoxOrient: 'vertical',
        WebkitLineClamp: maxLine,
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        wordBreak: 'break-all',
        wordWrap: 'break-word',
        width: Width
    }

    const [coppied, setCoppied] = useState(false)

    const handleCopy = () => {
        if (coppied) return
        setCoppied(true)
        const callCopy = copyFunc
            ? copyFunc(children)
            : navigator?.clipboard?.writeText(children)
        callCopy.then((resp) => {
            message?.success({ content: '复制成功' })
            return resp
        })
        setTimeout(() => {
            setCoppied(false)
        }, 1000)
    }

    return (
        <Space>
            {copy && (
                <Button
                    type="primary"
                    icon={<CopyOutlined />}
                    onClick={handleCopy}
                    disabled={coppied}
                    size="small"
                />
            )}
            <Tooltip
                destroyTooltipOnHide
                title={showTooltip ? children : ''}
                autoAdjustOverflow
                overlayInnerStyle={{ width: '100%' }}
                trigger="click"
            >
                <div
                    style={{
                        ...style,
                        ...Style
                    }}
                >
                    {children || placeholder}
                </div>
            </Tooltip>
        </Space>
    )
}

export default BaseText
