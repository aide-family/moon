import { StrategyType } from '@/apis/types'
import DataForm from '@/components/Data/DataForm/DataForm'
import { Form, Modal, ModalProps } from 'antd'
import React from 'react'
import { prodeStrategyFormOptions } from '../options'

export interface EditProbeModalProps extends ModalProps {
    strategyId?: number
    strategyType?: StrategyType
}

export type EditProbeFormType = {
    // TODO 待补充
}

export const EditProbeModal: React.FC<EditProbeModalProps> = (props) => {
    const { strategyId, onOk } = props

    const [form] = Form.useForm<EditProbeFormType>()
    const probeType = Form.useWatch('probeType', form)
    const httpMethod = Form.useWatch('httpMethod', form)

    const Title = () => {
        return strategyId ? '编辑策略' : '新建策略'
    }

    const handleOnOk = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
        form.validateFields().then((values) => {
            console.log('values', values)
            onOk?.(e)
        })
    }

    return (
        <Modal {...props} title={<Title />} onOk={handleOnOk}>
            <DataForm
                formProps={{ layout: 'vertical' }}
                form={form}
                items={prodeStrategyFormOptions({ probeType, httpMethod })}
            />
        </Modal>
    )
}
