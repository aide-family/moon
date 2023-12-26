import { FC, useEffect, useState } from 'react'
import { Form, Modal } from 'antd'
import { StrategyItemType } from '@/apis/home/monitor/strategy/types'
import { StrategyForm } from './StrategyForm'
import { ActionKey } from '@/apis/data'

export interface DetailProps {
    open: boolean
    onClose: () => void
    id?: number
    disabled?: boolean
    actionKey?: ActionKey
}

const defaultData: StrategyItemType[] = []

export const Detail: FC<DetailProps> = (props) => {
    const { open, onClose, disabled, actionKey } = props

    const [form] = Form.useForm()

    const [detail, setDetail] = useState<StrategyItemType>()
    const [loading, setLoading] = useState<boolean>(false)

    const fetchDetail = async () => {
        // const res = await getDetail(id)
        // setDetail(res)
        setDetail(defaultData[1])
    }

    const handleSubmit = () => {
        setLoading(true)
        form.validateFields()
            .then((values) => {
                // TODO
                onClose()
            })
            .finally(() => {
                setLoading(false)
            })
    }

    useEffect(() => {
        fetchDetail().then((r) => r)
        console.log(detail)
    }, [])

    return (
        <Modal
            title="策略详情"
            open={open}
            onCancel={onClose}
            onOk={handleSubmit}
            width="90%"
            destroyOnClose={true}
        >
            <StrategyForm form={form} disabled={disabled} />
        </Modal>
    )
}
