import { FC, useEffect, useState } from 'react'
import { StrategyGroupItemType } from '@/pages/home/monitor/strategy-group/type.ts'
import { Modal } from 'antd'

export interface DetailProps {
    open: boolean
    onClose: () => void
    id: string
}

const defaultData: StrategyGroupItemType[] = [{}]

export const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, id } = props

    const [detail, setDetail] = useState<StrategyGroupItemType>({})

    const fetchDetail = async () => {
        console.log(id, detail)
        // const res = await getDetail(id)
        // setDetail(res)
        setDetail(defaultData[1])
    }

    useEffect(() => {
        fetchDetail().then((r) => r)
    }, [])

    return (
        <Modal
            title="详情"
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
