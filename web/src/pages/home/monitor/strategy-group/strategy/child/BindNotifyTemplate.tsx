import {
    CreateTemplateRequest,
    NotifyTemplateItem,
    UpdateTemplateRequest
} from '@/apis/home/monitor/notify-template/types'
import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, ModalProps, Space, message } from 'antd'
import React, { useEffect, useState } from 'react'
import { bindNotifyTemplateDataFormOptions } from '../options'
import { templateApi } from '@/apis/home/monitor/notify-template'
import { NotifyTemplateType, defaultPageReqInfo } from '@/apis/types'
import chatGroupApi from '@/apis/home/monitor/chat-group'
import { TestTemplateResponse } from '@/apis/home/monitor/chat-group/types'

export interface BindNotifyTemplateProps extends ModalProps {
    strategyId?: number
}

let timer: NodeJS.Timeout | null
export const BindNotifyTemplate: React.FC<BindNotifyTemplateProps> = (
    props
) => {
    const { strategyId, open, onOk, onCancel } = props
    const [form] = Form.useForm<CreateTemplateRequest>()
    const notifyTeype =
        Form.useWatch('notifyType', form) ||
        NotifyTemplateType.NotifyTemplateTypeCustom
    const templateConten = Form.useWatch('content', form)

    const [templateList, setTemplateList] = useState<NotifyTemplateItem[]>([])
    const [loading, setLoading] = useState(false)

    const handleOnOk = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
        if (!strategyId) {
            return
        }
        setLoading(true)
        form.validateFields()
            .then((values) => {
                const item = templateList.find(
                    (item) => item.notifyType === notifyTeype
                )
                const params: UpdateTemplateRequest = {
                    ...values,
                    notifyType: notifyTeype,
                    strategyId: strategyId,
                    id: 0
                }
                if (!item) {
                    // 创建
                    return templateApi.createTemplate(params)
                }
                // 更新
                params.id = item.id
                return templateApi.updateTemplate(params)
            })
            .then(() => onOk?.(e))
            .finally(() => setLoading(false))
    }

    const getTemplateList = () => {
        if (!strategyId) {
            return
        }
        setLoading(true)
        templateApi
            .getTemplateList({
                page: defaultPageReqInfo,
                strategyId: strategyId
            })
            .then((items) => {
                setTemplateList(items.list)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    useEffect(() => {
        if (!open || !strategyId) {
            form?.resetFields()
            return
        }
        if (timer) {
            clearTimeout(timer)
        }
        timer = setTimeout(() => {
            getTemplateList()
        }, 500)
    }, [strategyId, open])

    useEffect(() => {
        form?.resetFields()
        const item = templateList.find(
            (item) => item.notifyType === notifyTeype
        ) || { notifyType: notifyTeype, content: '' }
        item.notifyType = notifyTeype
        form?.setFieldsValue(item)
    }, [notifyTeype, templateList])

    const handleOnTest = () => {
        if (!strategyId) {
            return
        }
        form.validateFields().then((values) => {
            setLoading(true)
            chatGroupApi
                .testTemplate({
                    notifyType: notifyTeype,
                    template: values?.content || '',
                    strategyId: strategyId
                })
                .then((res: TestTemplateResponse) => {
                    message.success(res.msg)
                })
                .finally(() => {
                    setLoading(false)
                })
        })
    }

    const Footer = () => {
        return (
            <Space size={8}>
                <Button onClick={onCancel} loading={loading}>
                    取消
                </Button>
                <Button
                    type="default"
                    danger
                    disabled={!templateConten}
                    onClick={handleOnTest}
                    loading={loading}
                >
                    测试
                </Button>
                <Button type="primary" onClick={handleOnOk} loading={loading}>
                    保存
                </Button>
            </Space>
        )
    }

    return (
        <Modal
            {...props}
            onOk={handleOnOk}
            confirmLoading={loading}
            footer={<Footer />}
        >
            <DataForm
                form={form}
                formProps={{ layout: 'vertical' }}
                items={bindNotifyTemplateDataFormOptions(notifyTeype)}
            />
        </Modal>
    )
}
