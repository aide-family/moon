import { UserListItem } from '@/apis/home/system/user/types'
import { randomColor } from '@/utils/random'
import { Avatar, AvatarProps, Image, Tooltip, TooltipProps } from 'antd'
import React, { useState } from 'react'

export interface UserAvatarProps extends UserListItem, AvatarProps {
    toolTip?: boolean
    preview?: boolean
}

export interface MyTooltip {
    show?: boolean
    children?: React.ReactNode | string
    tooltipProps?: TooltipProps
    title?: React.ReactNode | string
}

export const MyTooltip: React.FC<MyTooltip> = (props) => {
    const { show, tooltipProps, children, title } = props
    if (show) {
        return (
            <Tooltip {...tooltipProps} title={title || tooltipProps?.title}>
                {children || tooltipProps?.children}
            </Tooltip>
        )
    }
    return children || tooltipProps?.children
}

export const UserAvatar: React.FC<UserAvatarProps> = (
    props: UserAvatarProps
) => {
    const [backgroundColor] = useState(randomColor())
    const {
        nickname,
        username,
        avatar,
        toolTip,
        shape = 'square',
        preview = true,
        style = {
            backgroundColor: backgroundColor,
            fontSize: 14,
            lineHeight: '40px',
            textAlign: 'center'
        },
        size = 40
    } = props
    return (
        <MyTooltip title={nickname || username} show={toolTip}>
            {!avatar ? (
                <Avatar {...props} size={size} style={style} shape={shape}>
                    {nickname || username}
                </Avatar>
            ) : (
                <Image width={40} height={40} src={avatar} preview={preview} />
            )}
        </MyTooltip>
    )
}
