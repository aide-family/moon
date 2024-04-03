import endpointApi from '@/apis/home/monitor/endpoint'
import {
    AppendEndpointRequest,
    PrometheusServerItem,
    UpdateEndpointRequest
} from '@/apis/home/monitor/endpoint/types'
import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, Space } from 'antd'
import React, { useEffect, useState } from 'react'
import { editModalFormItems } from '../options'
import {ActionKey} from '@/apis/data.tsx'

export interface EditEndpointModalProps {
    endpointId?: number
    open: boolean
    action:string
    onClose: () => void
    onOk: () => void
}

const EditEndpointModal: React.FC<EditEndpointModalProps> = (props) => {
    const { endpointId, open, onClose, onOk,action } = props

    const [form] = Form.useForm<AppendEndpointRequest>()
    const [detail, setDetail] = useState<PrometheusServerItem>()
    const [loading, setLoading] = useState(false)
    const [disable, setDisable] = useState(false)

    const Title = () => {
        switch (action) {
            case ActionKey.ADD:
                setDisable(false)
                return "新增"
            case ActionKey.EDIT:
                setDisable(false)
                return "编辑"
            default:
                setDisable(true)
                return "详情"
        }
    }

    const handleGetDetail = () => {
        if (!endpointId) {
            return
        }
        setLoading(true)
        endpointApi
            .detailEndpoint(endpointId)
            .then((data) => {
                const item = data.detail
                setDetail(item)
                form.setFieldsValue(item)
            })
            .finally(() => setLoading(false))
    }

    const createEndpoint = (data: AppendEndpointRequest) => {
        setLoading(true)
        endpointApi
            .appendEndpoint(data)
            .then(onOk)
            .finally(() => setLoading(false))
    }

    const updateEndpoint = (data: UpdateEndpointRequest) => {
        setLoading(true)
        endpointApi
            .editEndpoint(data)
            .then(onOk)
            .finally(() => setLoading(false))
    }

    const handleSubmit = () => {
        form.validateFields().then((values) => {
            if (endpointId) {
                updateEndpoint({ ...values, id: endpointId })
            } else {
                createEndpoint(values)
            }
        })
    }

    const handleRestForm = () => {
        if (!detail) {
            form.resetFields()
        } else {
            form?.setFieldsValue(detail)
        }
    }

    const Footer = () => {
        return (
            <Space size={8}>
                <Button onClick={onClose}>取消</Button>
                {
                    action!=ActionKey.DETAIL?
                        <>
                            <Button onClick={handleRestForm}>重置</Button>
                            <Button type="primary" onClick={handleSubmit} loading={loading}>
                                确定
                            </Button>
                        </>
                    :null
                }
            </Space>
        )
    }

    useEffect(() => {
        form?.resetFields()
        if (!open) return
        handleGetDetail()
    }, [open])

    return (
        <Modal
            title={<Title />}
            open={open}
            width="50%"
            onCancel={onClose}
            footer={<Footer />}
        >
            <DataForm
                form={form}
                items={editModalFormItems}
                formProps={{ layout: 'vertical', disabled:disable}}
            />
        </Modal>
    )
}

export default EditEndpointModal
