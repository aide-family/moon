import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, Space } from 'antd'
import React, { useEffect, useState } from 'react'
import { addChatGroupItems, updateChatGroupItems } from '../options'
import {
    ChatGroupItem,
    CreateChatGroupRequest,
    UpdateChatGroupRequest
} from '@/apis/home/monitor/chat-group/types'
import chatGroupApi from '@/apis/home/monitor/chat-group'
import { useWatch } from 'antd/es/form/Form'
import { NotifyApp } from '@/apis/types'
import { ActionKey } from "@/apis/data.tsx";

export interface EditChatGroupProps {
    chatGroupId?: number
    open: boolean
    action: string
    onClose: () => void
    onOk: () => void
}

const getItems = (app: NotifyApp, action?: string) => {
    if (action==ActionKey.EDIT) return updateChatGroupItems
    return addChatGroupItems(app)
}

const EditChatGroupModal: React.FC<EditChatGroupProps> = (props) => {
    const { chatGroupId, onOk, onClose, open,action } = props

    const [form] = Form.useForm<
        CreateChatGroupRequest | UpdateChatGroupRequest
    >()
    const [detail, setDetail] = useState<ChatGroupItem>()
    const [loading, setLoading] = useState(false)
    const [disable, setDisable] = useState(false)
    const app = useWatch('app', form)

    const handleGetChatGroupDetail = () => {
        if (!chatGroupId) return
        setLoading(true)
        chatGroupApi
            .getChatGroupDetail({ id: chatGroupId })
            .then((res) => {
                const item = res.detail
                setDetail(item)
                form.setFieldsValue(item)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    const handleCreateChatGroup = (item: CreateChatGroupRequest) => {
        setLoading(true)
        chatGroupApi
            .createChatGroup(item)
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    const handleUpdateChatGroup = (item: UpdateChatGroupRequest) => {
        setLoading(true)
        chatGroupApi
            .updateChatGroup(item)
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    const handleSububmit = (
        values: CreateChatGroupRequest | UpdateChatGroupRequest
    ) => {
        if (chatGroupId) {
            handleUpdateChatGroup({
                ...values,
                id: chatGroupId
            } as UpdateChatGroupRequest)
        } else {
            handleCreateChatGroup(values as CreateChatGroupRequest)
        }
    }

    const handleOnOk = () => {
        form.validateFields().then(handleSububmit)
    }

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

    const handeRestForm = () => {
        if (chatGroupId && detail) {
            form.setFieldsValue(detail)
        } else {
            form.resetFields()
        }
    }

    const Footer = () => {
        return (
            <Space size={8}>
                <Button onClick={onClose} loading={loading}>
                    取消
                </Button>
                {
                    action!=ActionKey.DETAIL?
                        <>
                            <Button onClick={handeRestForm} type="dashed" loading={loading}>
                                重置
                            </Button>
                            <Button type="primary" onClick={handleOnOk} loading={loading}>
                                保存
                            </Button>
                        </>
                        :null
                }
            </Space>
        )
    }

    useEffect(() => {
        form.resetFields()
        if (!open) return
        handleGetChatGroupDetail()
    }, [open])

    return (
        <Modal
            title={<Title />}
            width="60%"
            open={open}
            onCancel={onClose}
            footer={<Footer />}
        >
            <DataForm
                items={getItems(app, action)}
                form={form}
                formProps={{
                    layout: 'vertical',
                    disabled: disable
                }}
            />
        </Modal>
    )
}

export default EditChatGroupModal
