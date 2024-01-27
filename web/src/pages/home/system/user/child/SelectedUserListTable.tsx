import { UserListItem } from '@/apis/home/system/user/types'
import React from 'react'
import { Button, Modal, Table } from 'antd'
import { UserAvatar } from './UserAvatar'
import { Username } from './Username'
import { Status, StatusMap } from '@/apis/types'
import { UserColumnType } from '../options'

export interface SelectedUserListTableProps {
    setTableSelectedRows?: React.Dispatch<React.SetStateAction<UserListItem[]>>
    tableSelectedRows?: UserListItem[]
    open?: boolean
    closeModal?: () => void
    onOk?: () => void
    status: Status
}

const defaultColumns: UserColumnType[] = [
    {
        title: '头像',
        dataIndex: 'avatar',
        key: 'avatar',
        width: 100,
        render: (_: string, item: UserListItem) => {
            return <UserAvatar {...item} />
        }
    },
    {
        title: '姓名',
        dataIndex: 'username',
        key: 'username',
        width: 120,
        render: (_: string, record: UserListItem) => {
            return <Username {...record} />
        }
    },
    {
        title: '昵称',
        dataIndex: 'nickname',
        key: 'nickname',
        width: 200,
        ellipsis: true,
        render: (text: String) => {
            return <>{text || '-'}</>
        }
    }
]

export const SelectedUserListTable: React.FC<SelectedUserListTableProps> = (
    props
) => {
    const {
        setTableSelectedRows,
        closeModal,
        onOk,
        tableSelectedRows,
        status,
        open
    } = props

    const Title = () => {
        const s =
            status === Status.STATUS_DISABLED
                ? StatusMap[Status.STATUS_DISABLED]
                : StatusMap[Status.STATUS_ENABLED]
        return (
            <span>
                请确认是否批量
                <span style={{ color: s.color }}>{s.text}</span>
                以下用户?
            </span>
        )
    }

    const hendleRemove = (record: UserListItem) => {
        setTableSelectedRows?.((prev) => [
            ...prev.filter((item) => item.id !== record.id)
        ])
    }

    const handleOnOk = () => {
        onOk?.()
    }

    const handleOnCancel = () => {
        closeModal?.()
    }

    const RemoveButton: React.FC<UserListItem> = (props) => {
        return (
            <Button
                type="primary"
                danger
                size="small"
                onClick={() => hendleRemove(props)}
            >
                移除
            </Button>
        )
    }

    const actionColumn = {
        title: '操作',
        key: 'action',
        width: 100,
        render: (_: string, record: UserListItem) => {
            return <RemoveButton {...record} />
        }
    }

    const okButtonProps = {
        disabled: tableSelectedRows?.length === 0
    }

    return (
        <>
            <Modal
                title={<Title />}
                open={open}
                onCancel={handleOnCancel}
                onOk={handleOnOk}
                width="50%"
                okButtonProps={okButtonProps}
            >
                <Table
                    size="small"
                    pagination={false}
                    dataSource={tableSelectedRows}
                    columns={[...defaultColumns, actionColumn]}
                />
            </Modal>
        </>
    )
}
