/**用户详情 */
import { FC, useEffect, useState } from 'react'
import { Descriptions, DescriptionsProps, Modal, Button, Badge } from 'antd'

import dict from '@/apis/home/system/dict'
import type { DictListItem } from '@/apis/home/system/dict/types'
import { Status, StatusMap } from '@/apis/types'
import { categoryData } from '@/apis/data'
import dayjs from 'dayjs'
const { dictDetail } = dict
// TODO 完善字典详情
export type DetailProps = {
    dictId: number
    open: boolean
    onClose: () => void
}
/** 用户详情组件
 * @param props type: DetailProps
 * @type dictId : number   // API权限 id
 * @type open : boolean    // 是否显示
 * @type onClose : () => void // 关闭回调
 *
 */
const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, dictId } = props

    const [dictDetailData, setAuthApiDetailData] = useState<DictListItem>(
        {} as DictListItem
    )
    const { name, status, createdAt, updatedAt } = dictDetailData

    const fetchUserDetail = async () => {
        const res = await dictDetail({ id: dictId })
        setAuthApiDetailData(res.promDict)
    }

    const getEnabledText = () => {
        if (status === Status.STATUS_UNKNOWN) return '-'

        const { color, text } = StatusMap[dictDetailData.status || 0]
        return <Badge color={color} text={text} />
    }

    const buildDetailItems = (): DescriptionsProps['items'] => {
        const { remark, category, color } = dictDetailData
        return [
            {
                key: '1',
                label: '状态',
                children: getEnabledText()
            },
            {
                key: '2',
                label: '字典颜色',
                children: color
            },
            {
                key: '3',
                label: '字典分类',
                children: categoryData[category]
            },
            {
                key: 'createAt',
                label: '创建时间',
                children: dayjs(createdAt * 1000).format('YYYY-MM-DD HH:mm:ss'),
                span: 3
            },
            {
                key: 'updateAt',
                label: '更新时间',
                children: dayjs(updatedAt * 1000).format('YYYY-MM-DD HH:mm:ss'),
                span: 3
            },
            {
                key: '4',
                label: '备注',
                children: remark || '-',
                span: 3
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
    }, [open, dictId])

    return (
        <Modal
            open={open}
            onCancel={onClose}
            centered
            footer={null}
            keyboard={false}
        >
            <Descriptions
                column={3}
                title={<DescriptionsTitle />}
                items={buildDetailItems()}
            />
        </Modal>
    )
}

export default Detail
