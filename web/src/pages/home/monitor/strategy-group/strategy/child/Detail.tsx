import { FC, useEffect, useState } from 'react'
import { Modal } from 'antd'
import { StrategyItemType } from '@/pages/home/monitor/strategy-group/strategy/type.ts'

export interface DetailProps {
    open: boolean
    onClose: () => void
    id: string
}

const defaultData: StrategyItemType[] = [{}]

export const Detail: FC<DetailProps> = (props) => {
    const { open, onClose } = props

    const [detail, setDetail] = useState<StrategyItemType>({})

    const fetchDetail = async () => {
        // const res = await getDetail(id)
        // setDetail(res)
        setDetail(defaultData[1])
    }

    useEffect(() => {
        fetchDetail().then((r) => r)
        console.log(detail)
    }, [])

    return (
        <Modal
            title="规则详情"
            open={open}
            onCancel={onClose}
            onOk={onClose}
            width={800}
            destroyOnClose={true}
        >
            detail
        </Modal>
    )
}
