/**用户详情 */
import { FC, useEffect, useState } from 'react'
import { Descriptions, DescriptionsProps, Modal, Button, Badge } from 'antd'

import authApi from '@/apis/home/system/auth'
import type { ApiAuthListItem } from '@/apis/home/system/auth/types'
import { StatusMap } from '@/apis/types'
import dayjs from 'dayjs'
const { authApiDetail } = authApi

export type DetailProps = {
    authId: number
    open: boolean
    onClose: () => void
}
/** 用户详情组件
 * @param props type: DetailProps
 * @type authId : number   API权限 id
 * @type open : boolean    是否显示
 * @type onClose : () => void 关闭回调
 */
const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, authId } = props

    const [authApiDetailData, setAuthApiDetailData] = useState<ApiAuthListItem>(
        {} as ApiAuthListItem
    )
    const { name, createdAt, updatedAt } = authApiDetailData

    const fetchUserDetail = async () => {
        const res = await authApiDetail({ id: authId })
        setAuthApiDetailData(res.detail)
    }

    const getEnabledText = () => {
        const { color, text } = StatusMap[authApiDetailData.status || 0]
        return <Badge color={color} text={text} />
    }

    const buildDetailItems = (): DescriptionsProps['items'] => {
        const { remark, method, path } = authApiDetailData
        return [
            {
                key: '1',
                label: '状态',
                children: getEnabledText()
            },
            {
                key: '2',
                label: '请求方式',
                children: method
            },
            {
                key: '3',
                label: '接口路径',
                children: path,
                span: 2
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
                label: '备注',
                children: remark
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
        }
    }, [open, authId])

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
                title={<DescriptionsTitle />}
                items={buildDetailItems()}
            />
        </Modal>
    )
}

export default Detail
