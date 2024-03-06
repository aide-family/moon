import React, { FC, useEffect, useRef, useState } from 'react'

import { Button, Form, message, Modal } from 'antd'
import { DataOption, DataTable, SearchForm } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import Detail from './child/Detail'
import EditModal from './child/EditModal'
import userOptions, { columns } from './options'
import userApi from '@/apis/home/system/user'
import type {
    UserListItem,
    UserListParams
} from '@/apis/home/system/user/types'
import { ModuleType, Status, StatusMap } from '@/apis/types'
import { ActionKey } from '@/apis/data'
import { SelectedUserListTable } from './child/SelectedUserListTable'
import { SysLogDetail } from '../../child/SysLogDetail'

const { confirm } = Modal
const { userList, userStatusEdit, userDelete } = userApi
const { searchItems, operationItems } = userOptions()

const defaultPadding = 12

let timer: NodeJS.Timeout

const defaultSearchParams = {
    page: {
        curr: 1,
        size: 10
    },
    keyword: ''
}

const Customer: FC = () => {
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<UserListItem[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [openEdit, setOpenEdit] = useState<boolean>(false)
    const [logOpen, setLogOpen] = useState<boolean>(false)
    const [logDataId, setLogDataId] = useState<number | undefined>()
    const [editId, setEditId] = useState<number | undefined>()
    const [userId, setUserId] = useState<number>(0)
    const [search, setSearch] = useState<UserListParams>(defaultSearchParams)
    const [tableSelectedRows, setTableSelectedRows] = useState<UserListItem[]>(
        []
    )
    const [confirmOpen, setConfirmOpen] = useState<boolean>(false)
    const [batchChangeStatus, setBatchChangeStatus] = useState<Status>(
        Status.STATUS_UNKNOWN
    )

    /** 刷新 */
    const handlerRefresh = () => {
        setRefresh((prev) => !prev)
    }

    const isSelected = () => {
        return tableSelectedRows.length == 0
    }

    const handlerCloseEdit = () => {
        setOpenEdit(false)
    }

    const handleEditOnOk = () => {
        handlerCloseEdit()
        handlerRefresh()
    }

    const handlerCloseDetail = () => {
        setOpenDetail(false)
    }

    const openLogDetail = (id: number) => {
        setLogOpen(true)
        setLogDataId(id)
    }

    const closeLogDetail = () => {
        setLogOpen(false)
        setLogDataId(undefined)
    }

    /** 获取数据 */
    const handlerGetData = () => {
        setLoading(true)
        userList({ ...search })
            .then((res) => {
                setDataSource(res.list)
                setTotal(res.page.total)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    /** 分页变化 */
    const handlerTablePageChange = (page: number, pageSize?: number) => {
        setSearch({
            ...search,
            page: {
                curr: page,
                size: pageSize || 10
            }
        })
    }

    /** 可以批量操作的数据 */
    const handlerBatchData = (_: React.Key[], selectedRows: UserListItem[]) => {
        setTableSelectedRows(selectedRows)
    }

    const EditConfirmTitle: React.FC<UserListItem> = (props) => {
        const { status } = props
        const s =
            status === Status.STATUS_ENABLED
                ? StatusMap[Status.STATUS_DISABLED]
                : StatusMap[Status.STATUS_ENABLED]
        return (
            <span>
                请确认是否修改状态为
                <span style={{ color: s.color }}>{s.text}</span>?
            </span>
        )
    }

    const getStatusValue = (status: Status) => {
        return status === Status.STATUS_ENABLED
            ? Status.STATUS_DISABLED
            : Status.STATUS_ENABLED
    }

    const cancelAction = () => {
        return message.info('取消操作')
    }

    const handleBatchChangeUserStatus = (status: Status, ids: number[]) => {
        userStatusEdit({
            ids: ids,
            status: status
        }).then(() => {
            setTableSelectedRows([])
            setConfirmOpen(false)
            message.success(`操作成功`).then(handlerRefresh)
        })
    }

    const openChangeStatusConfirm = (record: UserListItem) => {
        confirm({
            title: <EditConfirmTitle {...record} />,
            onOk: () =>
                handleBatchChangeUserStatus(getStatusValue(record.status), [
                    record.id
                ]),
            onCancel: cancelAction
        })
    }

    const handleDeleteUserInfo = (record: UserListItem) => {
        userDelete({
            id: record.id
        }).then(() => {
            message.success('删除成功').then(handlerRefresh)
        })
    }

    const openDeleteConfirm = (record: UserListItem) => {
        confirm({
            title: `请确认是否删除 ${record.nickname || record.username} ?`,
            content: '操作不可逆, 请谨慎操作!',
            okText: '确认删除',
            okButtonProps: { danger: true },
            type: 'error',
            onOk: () => handleDeleteUserInfo(record),
            onCancel: cancelAction
        })
    }

    /** 处理表格操作栏的点击事件 */
    const handlerTableAction = (key: ActionKey, record: UserListItem) => {
        switch (key) {
            case ActionKey.DETAIL:
                setOpenDetail(true)
                setUserId(record.id)
                break
            case ActionKey.EDIT:
                setOpenEdit(true)
                setEditId(record.id)
                break
            case ActionKey.CHANGE_STATUS:
                openChangeStatusConfirm(record)
                break
            case ActionKey.DELETE:
                openDeleteConfirm(record)
                break
            case ActionKey.OPERATION_LOG:
                openLogDetail(record.id)
                break
            default:
                break
        }
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (
        changedValues: any,
        allValues: any
    ) => {
        timer && clearTimeout(timer)
        timer = setTimeout(() => {
            setSearch({
                ...search,
                ...changedValues,
                ...allValues
            })
        }, 500)
    }

    const leftOptions: DataOptionItem[] = [
        {
            key: ActionKey.BATCH_IMPORT,
            label: (
                <Button type="primary" loading={loading}>
                    批量导入
                </Button>
            )
        },
        {
            key: ActionKey.BATCH_ENABLE,
            label: (
                <Button
                    type="primary"
                    loading={loading}
                    disabled={isSelected()}
                >
                    批量启用
                </Button>
            )
        },
        {
            key: ActionKey.BATCH_DISABLE,
            label: (
                <Button
                    type="primary"
                    danger
                    loading={loading}
                    disabled={isSelected()}
                >
                    批量禁用
                </Button>
            )
        }
    ]

    const rightOptions: DataOptionItem[] = [
        {
            key: ActionKey.REFRESH,
            label: (
                <Button
                    type="primary"
                    loading={loading}
                    onClick={handlerRefresh}
                >
                    刷新
                </Button>
            )
        }
    ]

    const getBatchIds = () => {
        return tableSelectedRows.map((item) => item.id)
    }

    const openBatchChangeStatusConfirm = (status: Status) => {
        if (!tableSelectedRows.length) {
            return message.error('请选择要操作的数据')
        }
        setBatchChangeStatus(status)
        setConfirmOpen(true)
    }

    // 操作栏按钮
    const handleOptionClick = (val: ActionKey) => {
        switch (val) {
            case ActionKey.ADD:
                setOpenEdit(true)
                setEditId(undefined)
                break
            case ActionKey.RESET:
                setSearch(defaultSearchParams)
                break
            case ActionKey.BATCH_ENABLE:
                return openBatchChangeStatusConfirm(Status.STATUS_ENABLED)
            case ActionKey.BATCH_DISABLE:
                return openBatchChangeStatusConfirm(Status.STATUS_DISABLED)
        }
    }

    const handleSelectedUserListTableOnOk = () => {
        handleBatchChangeUserStatus(batchChangeStatus, getBatchIds())
    }

    const handleSelectedUserListTableOnCancel = () => {
        setConfirmOpen(false)
        return message.info('取消操作')
    }

    useEffect(() => {
        handlerGetData()
    }, [refresh, search])

    return (
        <div>
            <SysLogDetail
                module={ModuleType.ModelTypeUser}
                moduleId={logDataId}
                open={logOpen}
                width={600}
                onClose={closeLogDetail}
            />
            <SelectedUserListTable
                tableSelectedRows={tableSelectedRows}
                setTableSelectedRows={setTableSelectedRows}
                status={batchChangeStatus}
                onOk={handleSelectedUserListTableOnOk}
                open={confirmOpen}
                closeModal={handleSelectedUserListTableOnCancel}
            />
            <Detail
                open={openDetail}
                onClose={handlerCloseDetail}
                userId={userId}
            />
            <EditModal
                open={openEdit}
                onClose={handlerCloseEdit}
                id={editId}
                onOk={handleEditOnOk}
            />
            <div ref={operationRef}>
                <RouteBreadcrumb />
                <HeightLine />
                <SearchForm
                    form={queryForm}
                    items={searchItems}
                    formProps={{
                        onValuesChange: handlerSearFormValuesChange
                    }}
                />
                <HeightLine />
                <DataOption
                    queryForm={queryForm}
                    rightOptions={rightOptions}
                    leftOptions={leftOptions}
                    action={handleOptionClick}
                />
                <PaddingLine
                    padding={defaultPadding}
                    height={1}
                    borderRadius={4}
                />
            </div>
            <DataTable
                dataSource={dataSource}
                columns={columns}
                operationRef={operationRef}
                total={total}
                loading={loading}
                pageSize={search?.page?.size}
                current={search?.page?.curr}
                operationItems={operationItems}
                pageOnChange={handlerTablePageChange}
                rowSelection={{
                    onChange: handlerBatchData,
                    selectedRowKeys: tableSelectedRows.map((item) => item.id)
                }}
                action={handlerTableAction}
            />
        </div>
    )
}

export default Customer
