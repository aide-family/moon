import { DataTable } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { Alert, Checkbox, Col, Modal, Row, Space } from 'antd'
import { FC, useEffect, useState } from 'react'
import { notifyObjectTableColumns } from '../options'
import {
    NotifyItem,
    NotifyMember
} from '@/apis/home/monitor/alarm-notify/types'

export interface BindNotifyObjectProps {
    open?: boolean
    onClose?: () => void
    strategyId?: number
}

interface NotifyMemberListProps {
    members: NotifyMember[]
}

const NotifyMemberList: FC<NotifyMemberListProps> = (props) => {
    const { members } = props
    return (
        <>
            <Row gutter={[16, 16]}>
                {members.map((member, index) => {
                    const [email, sms, phone] = member.notifyTypes
                    return (
                        <Col
                            span={12}
                            style={{
                                marginBottom:
                                    index < members.length - 1 ? 16 : 0
                            }}
                            key={member.user?.value || member.id}
                        >
                            <span
                                style={{ paddingRight: defaultPadding }}
                            >{`${member.user?.label}(${member.user?.nickname})`}</span>
                            <span>
                                <Checkbox checked={!!email}>邮箱</Checkbox>
                                <Checkbox checked={!!sms}>短信</Checkbox>
                                <Checkbox checked={!!phone}>电话</Checkbox>
                            </span>
                        </Col>
                    )
                })}
            </Row>
        </>
    )
}

const defaultPadding = 12

export const BindNotifyObject: FC<BindNotifyObjectProps> = (props) => {
    const { open, onClose, strategyId } = props

    const [data, setData] = useState<NotifyItem[]>([])

    const handleCancel = () => {
        onClose?.()
    }

    const handleOnOk = () => {
        onClose?.()
    }
    const AlertMessage = () => {
        return '当告警发生时候, 会通过下面配置的方式发送给对应的接收者, 如果告警发生升级, 也能发送给升级后的对象'
    }

    useEffect(() => {
        if (strategyId) {
            setData([
                {
                    id: 1,
                    name: '监控值班组',
                    remark: '日常监控值班同学',
                    status: 1,
                    members: [
                        {
                            memberId: 1,
                            notifyTypes: [1, 1, 1],
                            user: {
                                value: 1,
                                label: 'admin',
                                status: 1,
                                avatar: '',
                                nickname: '梧桐'
                            },
                            status: 1
                        },
                        {
                            memberId: 1,
                            notifyTypes: [1, 1, 1],
                            user: {
                                value: 2,
                                label: 'test-1',
                                status: 1,
                                avatar: '',
                                nickname: '梧桐-1'
                            },
                            status: 1
                        },
                        {
                            memberId: 1,
                            notifyTypes: [1, 1, 1],
                            user: {
                                value: 3,
                                label: 'test-2',
                                status: 1,
                                avatar: '',
                                nickname: '梧桐-2'
                            },
                            status: 1
                        },
                        {
                            memberId: 1,
                            notifyTypes: [1, 1, 1],
                            user: {
                                value: 4,
                                label: 'test-4',
                                status: 1,
                                avatar: '',
                                nickname: '梧桐-4'
                            },
                            status: 1
                        }
                    ],
                    chatGroups: [],
                    createdAt: 1643004763,
                    updatedAt: 1643004763,
                    deletedAt: 0,
                    externalNotifyObjs: []
                }
            ])
        }
    }, [strategyId])

    return (
        <Modal
            title="绑定通知对象"
            open={open}
            onCancel={handleCancel}
            onOk={handleOnOk}
            width="60%"
        >
            <Alert message={<AlertMessage />} type="info" showIcon />
            <HeightLine />
            <b>告警接收者</b>
            <PaddingLine padding={defaultPadding} height={1} borderRadius={4} />
            <DataTable
                size="small"
                dataSource={data}
                columns={notifyObjectTableColumns}
                expandable={{
                    expandedRowRender: (record: NotifyItem) => (
                        <>
                            <NotifyMemberList members={record.members} />
                        </>
                    )
                }}
            />
            <HeightLine />
            <b>告警升级接收者</b>
            <PaddingLine padding={defaultPadding} height={1} borderRadius={4} />
        </Modal>
    )
}
