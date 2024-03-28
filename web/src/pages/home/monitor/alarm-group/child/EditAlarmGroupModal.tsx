import DataForm from '@/components/Data/DataForm/DataForm'
import { Button, Form, Modal, Space } from 'antd'
import React, { useEffect, useState } from 'react'
import { editorItems } from '../options'
import {
    AlarmGroupItem,
    AlarmNotifyMember,
    CreateAlarmGroupRequest,
    NotifyMemberItem,
    UpdateAlarmGroupRequest
} from '@/apis/home/monitor/alarm-group/types'
import alarmGroupApi from '@/apis/home/monitor/alarm-group'
import { HeightLine } from '@/components/HeightLine'
import { ChatGroupSelectItem } from '@/apis/home/monitor/chat-group/types'
import { BindHook } from './BindHook'
import { BindMember } from './BindMember'

export interface EditAlarmGroupModalProps {
    alarmGroupId?: number
    open: boolean
    onClose: () => void
    onOk: () => void
    disabled?: boolean
}

const EditAlarmGroupModal: React.FC<EditAlarmGroupModalProps> = (props) => {
    const { alarmGroupId, open, onClose, onOk, disabled } = props
    const [form] = Form.useForm<CreateAlarmGroupRequest>()
    const [loading, setLoading] = useState<boolean>(false)
    const [detail, setDetail] = useState<AlarmGroupItem>()
    const [chatGroups, setChatGroups] = useState<ChatGroupSelectItem[]>([])
    const [members, setMembers] = useState<NotifyMemberItem[]>([])

    const handeResetForm = (item?: AlarmGroupItem) => {
        form.setFieldsValue({
            name: item?.name,
            remark: item?.remark
        })
    }

    const getAlarmGroupDetail = () => {
        if (!alarmGroupId) return
        setLoading(true)
        alarmGroupApi
            .detail(alarmGroupId)
            .then((res) => {
                const item = res.detail
                setDetail(item)
                setChatGroups(item?.chatGroups || [])
                setMembers(item?.members || [])
                handeResetForm(item)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    const createAlarmGroup = (data: CreateAlarmGroupRequest) => {
        setLoading(true)
        alarmGroupApi
            .create(data)
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    const updateAlarmGroup = (data: UpdateAlarmGroupRequest) => {
        setLoading(true)
        alarmGroupApi
            .update(data)
            .then(onOk)
            .finally(() => {
                setLoading(false)
            })
    }

    const handleSubmit = (data: CreateAlarmGroupRequest) => {
        const newData: CreateAlarmGroupRequest = {
            ...data,
            chatGroups: chatGroups?.map((item) => item.value),
            members: members?.map((item) => {
                const alarmNotifyMember: AlarmNotifyMember = {
                    memberId: item.memberId,
                    notifyType: item.notifyType,
                    id: item.id
                }
                return alarmNotifyMember
            })
        }
        if (alarmGroupId) {
            updateAlarmGroup({ ...newData, id: alarmGroupId })
        } else {
            createAlarmGroup(newData)
        }
    }

    const handleOnOk = () => {
        form.validateFields().then((values) => {
            handleSubmit(values)
        })
    }

    const Title = () => {
        return alarmGroupId
            ? disabled
                ? '告警组详情'
                : '编辑告警组'
            : '添加告警组'
    }

    const Footer = () => {
        return (
            <>
                <Space size={12}>
                    <Button onClick={onClose} loading={loading}>
                        取消
                    </Button>
                    {disabled ? null : (
                        <>
                            <Button
                                onClick={() => handeResetForm(detail)}
                                loading={loading}
                                type="dashed"
                            >
                                重置
                            </Button>
                            <Button
                                type="primary"
                                onClick={handleOnOk}
                                loading={loading}
                            >
                                提交
                            </Button>
                        </>
                    )}
                </Space>
            </>
        )
    }

    useEffect(() => {
        if (!open) {
            setDetail(undefined)
            setMembers([])
            setChatGroups([])
            return
        }
        form.resetFields()
        setChatGroups([])
        getAlarmGroupDetail()
    }, [open])

    return (
        <>
            <Modal
                open={open}
                onCancel={onClose}
                title={<Title />}
                width="60%"
                footer={<Footer />}
            >
                <DataForm
                    form={form}
                    items={editorItems}
                    formProps={{ layout: 'vertical', disabled: disabled }}
                />
                <Space direction="vertical" style={{ width: '100%' }}>
                    <BindHook
                        value={chatGroups}
                        defaultValue={chatGroups}
                        onChange={(v) => setChatGroups(v || [])}
                        disabled={disabled}
                    />
                    <HeightLine />
                    <BindMember
                        value={members}
                        defaultValue={members}
                        onChange={(v) => setMembers(v || [])}
                        disabled={disabled}
                    />
                </Space>
            </Modal>
        </>
    )
}

export default EditAlarmGroupModal
