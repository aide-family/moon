import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { Alert, Button, Checkbox, Form, Modal, Table, Tag } from 'antd'
import { FC, useEffect, useState } from 'react'
import {
    NotifyObjectTableColumnType,
    notifyObjectTableColumns
} from '../options'
import {
    ChatGroup,
    NotifyItem,
    NotifyMember
} from '@/apis/home/monitor/alarm-notify/types'
import strategyApi from '@/apis/home/monitor/strategy'
import { NotifyApp, Status, StatusMap } from '@/apis/types'
import { NotifyAppData } from '@/apis/data'
import {
    ChartGroupTableColumnType,
    NotifyMemberTableColumnType
} from '../../alarm-group/options'
import { IconFont } from '@/components/IconFont/IconFont'
import FetchSelect from '@/components/Data/FetchSelect'
import { DefaultOptionType } from 'antd/es/select'
import { getNotifyList } from '../../strategy-group/options'

export interface BindNotifyObjectProps {
    open?: boolean
    onClose?: () => void
    strategyId?: number
}

interface NotifyMemberListProps {
    members: NotifyMember[]
}

interface NotifyChatGroupsProps {
    chatGroups: ChatGroup[]
}

const NotifyMemberList: FC<NotifyMemberListProps> = (props) => {
    const { members } = props
    if (!members || members.length === 0) return null

    const columns: NotifyMemberTableColumnType[] = [
        {
            title: '用户',
            dataIndex: 'user',
            render: (user) => {
                return user?.label || '未知'
            }
        },
        {
            title: '通知方式',
            render: (_, record) => {
                const nType = record?.notifyType
                // const { email, sms, phone } = member
                const email = nType & 2
                const sms = nType & 4
                const phone = nType & 8
                return (
                    <span>
                        <Checkbox checked={!!email}>邮箱</Checkbox>
                        <Checkbox checked={!!sms}>短信</Checkbox>
                        <Checkbox checked={!!phone}>电话</Checkbox>
                    </span>
                )
            }
        }
    ]
    return (
        <>
            <Table
                columns={columns}
                dataSource={members}
                size="small"
                pagination={false}
                rowKey={(record) => record.memberId}
            />
        </>
    )
}

const NotifyChatGroups: FC<NotifyChatGroupsProps> = (props) => {
    const { chatGroups } = props
    if (!chatGroups || chatGroups.length === 0) return null
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
            width: 120,
            render: (status: Status) => {
                const { color, text } = StatusMap[status]
                return <Tag color={color}>{text}</Tag>
            }
        }
    ]
    return (
        <>
            <Table
                columns={chartGroupCoumns}
                dataSource={chatGroups}
                size="small"
                pagination={false}
                rowKey={(record) => record.value}
            />
        </>
    )
}

const defaultPadding = 12

const AlertMessage = () => {
    return '当告警发生时候, 会通过下面配置的方式发送给对应的接收者, 如果告警发生升级, 也能发送给升级后的对象'
}

export const BindNotifyObject: FC<BindNotifyObjectProps> = (props) => {
    const { open, onClose, strategyId } = props
    const [notifySelectFrom] = Form.useForm<{ list: DefaultOptionType[] }>()
    const [notifyData, setNotifyData] = useState<NotifyItem[]>([])
    const [loading, setLoading] = useState(false)

    const handleGetNotify = () => {
        if (!strategyId) return
        setLoading(false)
        strategyApi
            .getNotifyDetail(strategyId)
            .then((data) => {
                const notifyObjectList = data.notifyObjectList

                setNotifyData(notifyObjectList)
            })
            .finally(() => setLoading(false))
    }

    const handleCancel = () => {
        onClose?.()
    }

    const strategyBindNotify = () => {
        if (!strategyId) return
        setLoading(true)
        strategyApi
            .bindNotify({
                id: strategyId,
                notifyObjectIds: notifyData.map((item) => item.id)
            })
            .then(onClose)
            .finally(() => setLoading(false))
    }

    const handleOnOk = () => {
        strategyBindNotify()
    }

    const removeNotifyItem = (notifyItem: NotifyItem) => {
        setNotifyData(notifyData.filter((item) => notifyItem.id !== item.id))
    }

    const optionColumn: NotifyObjectTableColumnType = {
        title: '操作',
        key: 'action',
        width: 100,
        render: (_: any, record: NotifyItem) => {
            return (
                <Button
                    size="small"
                    danger
                    type="link"
                    onClick={() => removeNotifyItem(record)}
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

    const appendNotify = () => {
        notifySelectFrom.validateFields().then((data) => {
            const { list } = data
            if (!list || list.length === 0) return
            const listTmp = notifyData
            list.map((item) => {
                if (listTmp.find((i) => i.id === item.value)) return
                listTmp.push({
                    id: item.value as number,
                    name: item.label as string,
                    remark: item.title as string,
                    status: Status.STATUS_ENABLED,
                    createdAt: 0,
                    updatedAt: 0,
                    deletedAt: 0
                })
            })
            setNotifyData([...listTmp])
            notifySelectFrom.resetFields()
        })
    }

    useEffect(() => {
        if (!open) {
            return
        }
        handleGetNotify()
    }, [open])

    return (
        <Modal
            title="绑定通知对象"
            open={open}
            onCancel={handleCancel}
            onOk={handleOnOk}
            okButtonProps={{ loading }}
            width="60%"
        >
            <Alert message={<AlertMessage />} type="info" showIcon />
            <HeightLine />
            <Form
                layout="inline"
                form={notifySelectFrom}
                onFinish={appendNotify}
            >
                <Form.Item name="list" label={<b>告警接收者</b>}>
                    <FetchSelect
                        handleFetch={getNotifyList}
                        width={400}
                        defaultOptions={[]}
                        selectProps={{
                            placeholder: '请选择告警接收对象',
                            mode: 'multiple',
                            labelInValue: true
                        }}
                    />
                </Form.Item>
                <Button type="primary" htmlType="submit">
                    添加
                </Button>
            </Form>
            <PaddingLine padding={defaultPadding} height={1} borderRadius={4} />
            <Table
                size="small"
                dataSource={notifyData}
                columns={[...notifyObjectTableColumns, optionColumn]}
                pagination={false}
                rowKey={(record) => record.id}
                expandable={{
                    expandedRowRender: (record: NotifyItem) => (
                        <>
                            <NotifyChatGroups
                                chatGroups={record.chatGroups || []}
                            />
                            <HeightLine height={20} />
                            <NotifyMemberList members={record.members || []} />
                        </>
                    )
                }}
            />
        </Modal>
    )
}
