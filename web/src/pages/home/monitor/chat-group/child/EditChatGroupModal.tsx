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

export interface EditChatGroupProps {
    chatGroupId?: number
    open: boolean
    onClose: () => void
    onOk: () => void
}

const getItems = (id?: number) => {
    if (id) return updateChatGroupItems
    return addChatGroupItems
}

const EditChatGroupModal: React.FC<EditChatGroupProps> = (props) => {
    const { chatGroupId, onOk, onClose, open } = props

    const [form] = Form.useForm<
        CreateChatGroupRequest | UpdateChatGroupRequest
    >()
    const [detail, setDetail] = useState<ChatGroupItem>()
    const [loading, setLoading] = useState(false)

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
            handleUpdateChatGroup(values as UpdateChatGroupRequest)
        } else {
            handleCreateChatGroup(values as CreateChatGroupRequest)
        }
    }

    const handleOnOk = () => {
        form.validateFields().then(handleSububmit)
    }

    const Title = () => {
        return chatGroupId ? '编辑群组' : '添加群组'
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
                <Button onClick={handeRestForm} type="dashed" loading={loading}>
                    重置
                </Button>
                <Button type="primary" onClick={handleOnOk} loading={loading}>
                    保存
                </Button>
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
                items={getItems(chatGroupId)}
                form={form}
                formProps={{
                    layout: 'vertical'
                }}
            />
        </Modal>
    )
}

export default EditChatGroupModal
