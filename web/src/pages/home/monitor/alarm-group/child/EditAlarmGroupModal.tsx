import DataForm from '@/components/Data/DataForm/DataForm'
import { Badge, Button, Form, Modal, Space, Table } from 'antd'
import React, { useEffect, useState } from 'react'
import { ChartGroupTableColumnType, editorItems } from '../options'
import {
    AlarmGroupItem,
    CreateAlarmGroupRequest,
    UpdateAlarmGroupRequest
} from '@/apis/home/monitor/alarm-group/types'
import alarmGroupApi from '@/apis/home/monitor/alarm-group'
import { HeightLine } from '@/components/HeightLine'
import { NotifyApp, Status, StatusMap } from '@/apis/types'
import { NotifyAppData } from '@/apis/data'
import { IconFont } from '@/components/IconFont/IconFont'
import FetchSelect from '@/components/Data/FetchSelect'
import { DefaultOptionType } from 'antd/es/select'
import {
    ChatGroupSelectItem,
    defaultSelectChatGroupReques
} from '@/apis/home/monitor/chat-group/types'
import chatGroupApi from '@/apis/home/monitor/chat-group'

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
    const [chatSelectFrom] = Form.useForm<{ groups: DefaultOptionType[] }>()
    const [loading, setLoading] = useState<boolean>(false)
    const [detail, setDetail] = useState<AlarmGroupItem>()
    const [chatGroups, setChatGroups] = useState<ChatGroupSelectItem[]>([])

    const handeResetForm = (item?: AlarmGroupItem) => {
        form.setFieldsValue({
            name: item?.name,
            remark: item?.remark
        })
    }

    const removeDetailChatGroup = (chatItem: ChatGroupSelectItem) => {
        if (chatItem && detail) {
            const chatGroupsTmp = chatGroups?.filter(
                (item) => item.value !== chatItem.value
            )
            setChatGroups(chatGroupsTmp)
        }
    }

    const appendChatGroup = () => {
        chatSelectFrom.validateFields().then((values) => {
            const chatGroupsTmp = chatGroups
            values.groups.map((item) => {
                chatGroupsTmp.push({
                    value: item.value as number,
                    label: item.label as string,
                    app: item.title as number,
                    status: Status.STATUS_ENABLED
                })
            })

            // 去重
            setChatGroups(
                chatGroupsTmp?.filter((item, index, arr) => {
                    return (
                        arr.findIndex(
                            (itemTmp) => itemTmp.value === item.value
                        ) === index
                    )
                })
            )
            chatSelectFrom.resetFields()
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
            chatGroups: chatGroups?.map((item) => item.value)
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
        return alarmGroupId ? '编辑告警组' : '添加告警组'
    }

    const chartGroupCoumns: ChartGroupTableColumnType[] = [
        {
            title: '名称',
            dataIndex: 'label',
            key: 'label',
            width: 200
        },
        {
            title: '所属APP',
            dataIndex: 'app',
            key: 'app',
            width: 160,
            render: (app: NotifyApp) => {
                return NotifyAppData[app]
            }
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            align: 'center',
            width: 120,
            render: (status: Status) => {
                const { color, text } = StatusMap[status]
                return <Badge color={color} text={text} />
            }
        },
        {
            title: '操作',
            dataIndex: 'action',
            key: 'action',
            align: 'center',
            width: 80,
            render: (_: any, record: ChatGroupSelectItem) => {
                return (
                    <Button
                        size="small"
                        danger
                        type="link"
                        onClick={() => removeDetailChatGroup(record)}
                        disabled={disabled}
                        icon={
                            <IconFont
                                type="icon-shanchu-copy"
                                style={{ color: 'red' }}
                            />
                        }
                    >
                        删除
                    </Button>
                )
            }
        }
    ]

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

    const getChatGroupList = (
        keyword?: string
    ): Promise<DefaultOptionType[]> => {
        return chatGroupApi
            .getChatGroupSelect({
                ...defaultSelectChatGroupReques,
                keyword
            })
            .then((data) => {
                if (!data || !data.list || data.list.length === 0) return []
                return data.list?.map((item) => {
                    const option: DefaultOptionType = {
                        label: item.label,
                        app: item.app,
                        status: item.status,
                        value: item.value,
                        title: item.app
                    }
                    return option
                })
            })
    }

    useEffect(() => {
        if (!open) return
        form.resetFields()
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
                    <Form
                        layout="inline"
                        form={chatSelectFrom}
                        onFinish={appendChatGroup}
                    >
                        <Form.Item name="groups" label={<b>机器人</b>}>
                            {!disabled && (
                                <FetchSelect
                                    handleFetch={getChatGroupList}
                                    width={400}
                                    defaultOptions={[]}
                                    selectProps={{
                                        placeholder: '请选择机器人',
                                        mode: 'multiple',
                                        labelInValue: true,
                                        disabled: disabled
                                    }}
                                />
                            )}
                        </Form.Item>
                        {!disabled && (
                            <Button
                                type="primary"
                                htmlType="submit"
                                disabled={disabled}
                            >
                                添加
                            </Button>
                        )}
                    </Form>
                    <Table
                        columns={chartGroupCoumns}
                        dataSource={chatGroups}
                        size="small"
                        pagination={false}
                        rowKey={(record) => record.value}
                    />
                    <HeightLine />
                </Space>
            </Modal>
        </>
    )
}

export default EditAlarmGroupModal
