import { Modal } from 'antd'
import { FC, useState } from 'react'

export interface EditStrategyModalProps {
    open: boolean
    onCancel: () => void
    onOk: (values: any) => void
    id?: number
}

export const EditStrategyModal: FC<EditStrategyModalProps> = (props) => {
    const { open, onCancel, onOk, id } = props
    const [values, setValues] = useState({})
    console.log(setValues)
    const handleOk = () => {
        onOk(values)
    }

    const handleCancel = () => {
        onCancel()
    }

    const Title = () => {
        if (id) {
            return <div>编辑规则</div>
        } else {
            return <div>新增规则</div>
        }
    }

    return (
        <Modal
            open={open}
            onOk={handleOk}
            onCancel={handleCancel}
            title={<Title />}
            width="80%"
        ></Modal>
    )
}
