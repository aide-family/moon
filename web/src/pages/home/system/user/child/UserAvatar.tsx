import { UserListItem } from '@/apis/home/system/user/types'
import { randomColor } from '@/utils/random'
import { Avatar, Image } from 'antd'
import React from 'react'

export const UserAvatar: React.FC<UserListItem> = (props: UserListItem) => {
    const { nickname, username, avatar } = props
    if (!avatar) {
        return (
            <Avatar
                size={40}
                style={{
                    backgroundColor: randomColor(),
                    fontSize: 14,
                    lineHeight: '40px',
                    textAlign: 'center'
                }}
                shape="square"
            >
                {nickname || username}
            </Avatar>
        )
    }
    return <Image width={40} height={40} src={avatar} />
}
