import { UserListItem } from '@/apis/home/system/user/types'
import { Gender } from '@/apis/types'
import { ManOutlined, WomanOutlined } from '@ant-design/icons'
import { Button } from 'antd'
import React from 'react'

export const Username: React.FC<UserListItem> = (props: UserListItem) => {
    const { username, gender } = props
    return (
        <Button
            type="text"
            icon={
                gender === Gender.Gender_MALE ? (
                    <ManOutlined style={{ color: '#1890ff' }} />
                ) : gender === Gender.Gender_FEMALE ? (
                    <WomanOutlined style={{ color: '#f759ab' }} />
                ) : null
            }
        >
            {username}
        </Button>
    )
}
