import strategyGroupApi from '@/apis/home/monitor/strategy-group'
import {
    CreateStrategGrouRequest,
    StrategyGroupItemType
} from '@/apis/home/monitor/strategy-group/types'
import DataForm from '@/components/Data/DataForm/DataForm'
import { Form, Modal } from 'antd'
import { FC, useEffect, useState } from 'react'
import { editStrategyGroupDataFormItems } from '../options'

export interface EditGroupModalProps {
    open: boolean
    onCancel: () => void
    onOk: () => void
    groupId?: number
}

const EditGroupModal: FC<EditGroupModalProps> = (props) => {
    const { open, onCancel, onOk, groupId } = props
    const [form] = Form.useForm()
    const [detail, setDetail] = useState<StrategyGroupItemType>()
    const [loading, setLoading] = useState<boolean>(false)

    const handleAddStrategyGroup = (values: CreateStrategGrouRequest) => {
        return strategyGroupApi.createSteategyGroup(values)
    }

    const handleUpdateStrategyGroup = (values: CreateStrategGrouRequest) => {
        if (!groupId) {
            return Promise.reject(new Error('groupId is undefined'))
        }
        return strategyGroupApi.updateSteategyGroup({ ...values, id: groupId })
    }

    const handleOk = () => {
        setLoading(true)
        form.validateFields()
            .then((values) => {
                if (!groupId) {
                    return handleAddStrategyGroup(values)
                }
                return handleUpdateStrategyGroup(values)
            })
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    const hendleGetDetail = () => {
        if (!groupId || !open) {
            return
        }
        return strategyGroupApi
            .getStrategyGroupDetail({ id: groupId })
            .then((res) => {
                setDetail(res.detail)
                form.setFieldsValue(res.detail)
            })
    }

    const handleCancel = () => {
        form.resetFields()
        setDetail(undefined)
        onCancel()
    }

    useEffect(() => {
        hendleGetDetail()
    }, [groupId, open])

    const Title = () => {
        if (groupId) {
            return '编辑分组'
        } else {
            return '新增分组'
        }
    }

    useEffect(() => {
        if (!open) {
            form.resetFields()
            setDetail(undefined)
            return
        }
        form.setFieldsValue(detail)
    }, [open])

    return (
        <Modal
            title={<Title />}
            open={open}
            onCancel={handleCancel}
            onOk={handleOk}
            confirmLoading={loading}
        >
            <DataForm
                formProps={{ layout: 'vertical' }}
                form={form}
                items={editStrategyGroupDataFormItems}
            />
        </Modal>
    )
}

export default EditGroupModal
