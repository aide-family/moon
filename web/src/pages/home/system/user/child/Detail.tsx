/**用户详情 */
import { FC, useEffect, useState } from 'react'
import {
    Descriptions,
    DescriptionsProps,
    Modal,
    Image,
    Button,
    Badge
} from 'antd'

import userApi from '@/apis/home/system/user'
import type { UserListItem } from '@/apis/home/system/user/types'
import { Gender, StatusMap } from '@/apis/types'
import { ManOutlined, WomanOutlined } from '@ant-design/icons'
const { userDetail } = userApi

export type DetailProps = {
    userId: number
    open: boolean
    onClose: () => void
}
/** 用户详情组件
 * @param props type: DetailProps
 * @type userId : number   // 用户id
 * @type open : boolean    // 是否显示
 * @type onClose : () => void // 关闭回调
 *
 */
const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, userId } = props

    const [userDetailData, setUserDetailData] = useState<UserListItem>(
        {} as UserListItem
    )

    const fetchUserDetail = async () => {
        const res = await userDetail({ id: userId })
        setUserDetailData(res.detail)
    }

    const getEnabledText = () => {
        const { color, text } = StatusMap[userDetailData.status || 0]
        return <Badge color={color} text={text} />
    }

    const buildDetailItems = (): DescriptionsProps['items'] => {
        const { nickname, phone, email, avatar } = userDetailData
        return [
            {
                key: '1',
                label: '昵称',
                children: nickname
            },
            {
                key: '3',
                label: '状态',
                children: getEnabledText()
            },
            {
                key: '2',
                label: '电话号码',
                children: phone
            },
            {
                key: '4',
                label: '邮箱',
                children: email
            },
            {
                key: '5',
                label: '头像',
                children: (
                    <>
                        {avatar ? (
                            <Image width={40} height={40} src={avatar} />
                        ) : (
                            '-'
                        )}
                    </>
                )
            }
        ]
    }

    const getGenderIcon = () => {
        const { gender } = userDetailData
        return gender === Gender.Gender_MALE ? (
            <ManOutlined style={{ color: '#1890ff' }} />
        ) : gender === Gender.Gender_FEMALE ? (
            <WomanOutlined style={{ color: '#f759ab' }} />
        ) : null
    }

    useEffect(() => {
        if (open) {
            fetchUserDetail()
        }
    }, [open, userId])

    return (
        <Modal
            open={open}
            onCancel={onClose}
            centered
            footer={null}
            keyboard={false}
        >
            <Descriptions
                column={2}
                title={
                    <Button type="text" size="large" icon={getGenderIcon()}>
                        {userDetailData.username}
                    </Button>
                }
                items={buildDetailItems()}
            />
        </Modal>
    )
}

export default Detail
