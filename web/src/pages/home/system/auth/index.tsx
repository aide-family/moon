import React, { useEffect, useRef, useState, FC } from 'react'
import { Button, Form, Modal, message } from 'antd'
import { SearchForm, DataTable } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import Detail from './child/Detail'
import EditModal from './child/EditModal'
import authOptions, { columns } from './options'
import authApi from '@/apis/home/system/auth'
import { ExclamationCircleFilled } from '@ant-design/icons'
import { ApiAuthListItem, ApiAuthListReq } from '@/apis/home/system/auth/types'
import { ActionKey } from '@/apis/data'
import { SysLogDetail } from '../../child/SysLogDetail'
import { ModuleType } from '@/apis/types'

const { confirm } = Modal
const { authApiList, authApiDelete } = authApi
const { searchItems, operationItems } = authOptions()

const defaultPadding = 12

let timer: any

/**
 * 角色管理
 */
const Auth: FC = () => {
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<ApiAuthListItem[]>([])

    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [openEdit, setOpenEdit] = useState<boolean>(false)
    const [editId, setEditId] = useState<number | undefined>()
    const [authId, setRoleId] = useState<number>(0)
    const [search, setSearch] = useState<ApiAuthListReq>({
        page: {
            curr: 1,
            size: 10
        },
        keyword: ''
    })
    const [tableSelectedRows, setTableSelectedRows] = useState<
        ApiAuthListItem[]
    >([])
    const [logOpen, setLogOpen] = useState<boolean>(false)
    const [logDataId, setLogDataId] = useState<number | undefined>()

    const openLogDetail = (id: number) => {
        setLogOpen(true)
        setLogDataId(id)
    }

    const closeLogDetail = () => {
        setLogOpen(false)
        setLogDataId(undefined)
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

    // 获取数据
    const handlerGetData = () => {
        setLoading(true)
        authApiList(search)
            .then((res) => {
                setDataSource(res.list)
                setTotal(res.page.total)
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
        _: React.Key[],
        selectedRows: ApiAuthListItem[]
    ) => {
        setTableSelectedRows(selectedRows)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: ApiAuthListItem) => {
        switch (key) {
            case ActionKey.OPERATION_LOG:
                openLogDetail(record.id)
                break
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
                    title: `请确认是否删除该权限?`,
                    icon: <ExclamationCircleFilled />,
                    content: '删除后不可恢复，请谨慎操作',
                    onOk() {
                        authApiDelete({
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
        }
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (changedValues: any) => {
        timer && clearTimeout(timer)
        timer = setTimeout(() => {
            setSearch({
                ...search,
                ...changedValues,
                page: {
                    curr: 1,
                    size: 10
                }
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
    //操作栏按钮
    const handleOptionClick = (val: ActionKey) => {
        switch (val) {
            case ActionKey.ADD:
                setOpenEdit(true)
                setEditId(undefined)
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
        <div>
            <SysLogDetail
                module={ModuleType.ModelTypeApi}
                moduleId={logDataId}
                open={logOpen}
                width={600}
                onClose={closeLogDetail}
            />
            <Detail
                open={openDetail}
                onClose={handlerCloseDetail}
                authId={authId}
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

export default Auth
