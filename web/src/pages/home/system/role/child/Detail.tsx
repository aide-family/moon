/**用户详情 */
import { FC, useEffect, useState } from 'react'
import {
    Descriptions,
    DescriptionsProps,
    Modal,
    Button,
    Badge,
    Avatar,
    Tooltip,
    Spin
} from 'antd'

import roleApi from '@/apis/home/system/role'
import type { RoleListItem } from '@/apis/home/system/role/types'
import { Status, StatusMap } from '@/apis/types'
import dayjs from 'dayjs'

export type DetailProps = {
    roleId: number
    open: boolean
    onClose: () => void
}

const { roleDetail } = roleApi
const defaultRoleDetail = {} as RoleListItem

/** 用户详情组件
 * @param props type: DetailProps
 * @type roleId : number   // 用户id
 * @type open : boolean    // 是否显示
 * @type onClose : () => void // 关闭回调
 *
 */
const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, roleId } = props

    const [roleDetailData, setUserDetailData] =
        useState<RoleListItem>(defaultRoleDetail)
    const [loading, setLoading] = useState<boolean>(false)
    const { remark, users, createdAt, updatedAt, status, name } = roleDetailData

    const init = () => {
        setLoading(false)
        setUserDetailData(defaultRoleDetail)
    }

    const fetchUserDetail = async () => {
        setLoading(true)
        try {
            const res = await roleDetail({ id: roleId })
            setUserDetailData(res.detail)
        } catch (error) {
            // TODO 错误处理
        }
        setLoading(false)
    }

    const getEnabledText = () => {
        if (status === Status.STATUS_UNKNOWN) return '-'
        const { color, text } = StatusMap[status || 0]
        return <Badge color={color} text={text} />
    }

    const HaveUserList = () => {
        if (!users || users.length === 0) {
            return '该角色暂时没有关联用户'
        }
        return (
            <Avatar.Group shape="square" maxCount={3}>
                {users?.map((user, index) => {
                    return (
                        <Tooltip
                            title={<div>{user.label}</div>}
                            placement="top"
                        >
                            <Avatar
                                key={index}
                                src={user.avatar}
                                style={{
                                    backgroundColor: '#1890ff'
                                }}
                            />
                        </Tooltip>
                    )
                })}
            </Avatar.Group>
        )
    }

    const buildDetailItems = (): DescriptionsProps['items'] => {
        return [
            {
                key: '1',
                label: '状态',
                children: getEnabledText()
            },
            {
                key: '3',
                label: '备注',
                children: remark || '-'
            },
            {
                key: 'createdAt',
                label: '创建时间',
                children: createdAt
                    ? dayjs(createdAt * 1000).format('YYYY-MM-DD HH:mm:ss')
                    : '-'
            },
            {
                key: 'updatedAt',
                label: '更新时间',
                children: updatedAt
                    ? dayjs(updatedAt * 1000).format('YYYY-MM-DD HH:mm:ss')
                    : '-'
            },
            {
                key: '4',
                label: '用户',
                children: <HaveUserList />
            }
        ]
    }
    const DescriptionsTitle = () => {
        return (
            <Button type="text" size="large" style={{ color: '#1890ff' }}>
                {name}
            </Button>
        )
    }

    useEffect(() => {
        if (open) {
            fetchUserDetail()
            return
        }
        init()
    }, [open, roleId])

    return (
        <Modal
            open={open}
            onCancel={onClose}
            centered
            footer={null}
            keyboard={false}
        >
            <Spin spinning={loading}>
                <Descriptions
                    column={2}
                    title={<DescriptionsTitle />}
                    items={buildDetailItems()}
                />
            </Spin>
        </Modal>
    )
}

export default Detail
