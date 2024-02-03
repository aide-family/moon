import React, { useEffect, useRef, useState } from 'react'
import { Button, Form, Modal, message } from 'antd'
import { SearchForm, DataTable } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import Detail from './child/Detail'
import EditModal from './child/EditModal'
import roleOptions, { columns } from './options'
import roleApi from '@/apis/home/system/role'
import { ExclamationCircleFilled } from '@ant-design/icons'
import { RoleListItem, RoleListReq } from '@/apis/home/system/role/types'
import AuthConfigModal from './child/AuthConfig'
import { ActionKey } from '@/apis/data'

const { confirm } = Modal
const { roleDelete, roleList } = roleApi
const { searchItems, operationItems } = roleOptions()

const defaultPadding = 12

let timer: NodeJS.Timeout

/**
 * 角色管理
 */
const Role: React.FC = () => {
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<RoleListItem[]>([])

    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [openEdit, setOpenEdit] = useState<boolean>(false)
    const [editId, setEditId] = useState<number | undefined>()
    const [roleId, setRoleId] = useState<number>(0)
    const [search, setSearch] = useState<RoleListReq>({
        page: {
            curr: 1,
            size: 10
        },
        keyword: ''
    })
    const [tableSelectedRows, setTableSelectedRows] = useState<RoleListItem[]>(
        []
    )
    const [authVisible, setAuthVisible] = useState<boolean>(false)

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
    const handlerCloseAuthConfig = () => {
        setAuthVisible(false)
    }
    const handleAuthConfig = () => {
        setAuthVisible(false)
        handlerRefresh()
    }

    // 获取数据
    const handlerGetData = () => {
        setLoading(true)
        roleList({ ...search })
            .then((res) => {
                setDataSource(res.list)
                setTotal(res.page.total)
                console.log('res', res)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    useEffect(() => {
        handlerGetData()
    }, [refresh, search])

    // 刷新
    const handlerRefresh = () => {
        setRefresh((prev) => !prev)
    }

    // 分页变化
    const handlerTablePageChange = (page: number, pageSize?: number) => {
        console.log(page, pageSize)
        setSearch({
            ...search,
            page: {
                curr: page,
                size: pageSize || 10
            }
        })
    }

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: React.Key[],
        selectedRows: RoleListItem[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
        setTableSelectedRows(selectedRows)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: RoleListItem) => {
        console.log(key, record)
        switch (key) {
            case ActionKey.DETAIL:
                // handlerOpenDetail()
                setOpenDetail(true)
                setRoleId(record.id)

                break
            case ActionKey.EDIT:
                setOpenEdit(true)
                setEditId(record.id)
                break
            case ActionKey.DELETE:
                confirm({
                    title: `请确认是否删除该用户?`,
                    icon: <ExclamationCircleFilled />,
                    content: '',
                    onOk() {
                        roleDelete({
                            id: record.id
                        }).then(() => {
                            message.success('删除成功')
                            handlerRefresh()
                        })
                    },
                    onCancel() {
                        message.info('取消操作')
                    }
                })
                break
            case ActionKey.ASSIGN_AUTH:
                setAuthVisible(true)
                setRoleId(record.id)
                break
        }
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (
        changedValues: any, // TODO 不要any
        allValues: any
    ) => {
        timer && clearTimeout(timer)
        timer = setTimeout(() => {
            setSearch({
                ...search,
                ...changedValues
            })
            console.log(changedValues, allValues)
        }, 500)
    }

    const leftOptions: DataOptionItem[] = [
        // {
        //     key: ActionKey.BATCH_IMPORT,
        //     label: (
        //         <Button type="primary" loading={loading}>
        //             批量导入
        //         </Button>
        //     )
        // }
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
    //操作栏按钮
    const handleOptionClick = (val: ActionKey) => {
        switch (val) {
            case ActionKey.ADD:
                setOpenEdit(true)
                setEditId(0)
                break
            case ActionKey.RESET:
                setSearch({
                    keyword: '',
                    page: {
                        curr: 1,
                        size: 10
                    }
                })
                break
        }
    }

    return (
        <div className="bodyContent">
            <Detail
                open={openDetail}
                onClose={handlerCloseDetail}
                roleId={roleId}
            />
            <EditModal
                open={openEdit}
                onClose={handlerCloseEdit}
                id={editId}
                onOk={handleEditOnOk}
            />
            <AuthConfigModal
                open={authVisible}
                onClose={handlerCloseAuthConfig}
                onOk={handleAuthConfig}
                roleId={roleId}
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
                operationItems={operationItems}
                pageOnChange={handlerTablePageChange}
                pageSize={search?.page?.size}
                current={search?.page?.curr}
                rowSelection={{
                    onChange: handlerBatchData,
                    selectedRowKeys: tableSelectedRows.map((item) => item.id)
                }}
                action={handlerTableAction}
            />
        </div>
    )
}

export default Role
