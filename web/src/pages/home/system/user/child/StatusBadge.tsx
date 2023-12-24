import { UserListItem } from '@/apis/home/system/user/types'
import { StatusMap } from '@/apis/types'
import { Badge } from 'antd'
import React from 'react'

export const StatusBadge: React.FC<UserListItem> = (props: UserListItem) => {
    const { status } = props
    return (
        <Badge color={StatusMap[status].color} text={StatusMap[status].text} />
    )
}
