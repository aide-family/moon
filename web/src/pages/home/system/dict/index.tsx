import { useEffect, FC, useRef, useState } from 'react'
import { Button, Form, Modal, message } from 'antd'
import { SearchForm, DataTable } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import Detail from './child/Detail'
import EditDictModal from './child/EditModal'
import dictOptions, { columns } from './options'
import dictApi from '@/apis/home/system/dict'
import { ModuleType, Status } from '@/apis/types'
import { ExclamationCircleFilled } from '@ant-design/icons'
import { DictListItem, DictListReq } from '@/apis/home/system/dict/types'
import { ActionKey } from '@/apis/data'
import { SysLogDetail } from '../../child/SysLogDetail'

const { confirm } = Modal
const { dictList, dictDelete, dictBatchUpdateStatus } = dictApi
const { searchItems, operationItems } = dictOptions()

const defaultPadding = 12

let timer: NodeJS.Timeout

/**
 * 字典管理
 */
const Dict: FC = () => {
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<DictListItem[]>([])

    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [openEdit, setOpenEdit] = useState<boolean>(false)
    const [editId, setEditId] = useState<number | undefined>()
    const [dictId, setRoleId] = useState<number>(0)
    const [search, setSearch] = useState<DictListReq>({
        page: {
            curr: 1,
            size: 10
        },
        keyword: ''
    })
    const [tableSelectedRows, setTableSelectedRows] = useState<DictListItem[]>(
        []
    )
    const [logOpen, setLogOpen] = useState<boolean>(false)
    const [logDataId, setLogDataId] = useState<number | undefined>()

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
        dictList({ ...search })
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
    const handlerBatchData = (_: React.Key[], selectedRows: DictListItem[]) => {
        setTableSelectedRows(selectedRows)
    }

    const openLogDetail = (id: number) => {
        setLogOpen(true)
        setLogDataId(id)
    }

    const closeLogDetail = () => {
        setLogOpen(false)
        setLogDataId(undefined)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: DictListItem) => {
        switch (key) {
            case ActionKey.OPERATION_LOG:
                openLogDetail(record.id)
                break
            case ActionKey.DETAIL:
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
                    content: '此操作不可逆',
                    onOk() {
                        dictDelete({
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
    const handlerSearFormValuesChange = (
        changedValues: any,
        allValues: any
    ) => {
        timer && clearTimeout(timer)
        timer = setTimeout(() => {
            setSearch({
                ...search,
                ...changedValues,
                ...allValues,
                page: {
                    curr: 1,
                    size: 10
                }
            })
        }, 500)
    }

    const isDisabled = () => {
        return tableSelectedRows.length === 0
    }

    const leftOptions: DataOptionItem[] = [
        {
            key: ActionKey.BATCH_ENABLE,
            label: (
                <Button
                    type="primary"
                    loading={loading}
                    disabled={isDisabled()}
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
                    disabled={isDisabled()}
                    loading={loading}
                >
                    批量禁用
                </Button>
            )
        },
        {
            key: ActionKey.BATCH_DELETE,
            label: (
                <Button
                    type="primary"
                    danger
                    loading={loading}
                    disabled={isDisabled()}
                >
                    批量删除
                </Button>
            )
        }
    ]

    const rightOptions: DataOptionItem[] = [
        {
            key: 'refresh',
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
            case ActionKey.BATCH_ENABLE:
                if (!tableSelectedRows || tableSelectedRows.length === 0) {
                    message.error('请选择要操作的数据')
                    return
                }
                confirm({
                    title: `请确认是否批量启用?`,
                    icon: <ExclamationCircleFilled />,
                    content: '',
                    onOk() {
                        dictBatchUpdateStatus({
                            ids: tableSelectedRows.map((item) => item.id),
                            status: Status['STATUS_ENABLED']
                        }).then(() => {
                            message.success('批量启用成功')
                            handlerRefresh()
                        })
                    },
                    onCancel() {
                        message.info('取消操作')
                    }
                })
                break
            case ActionKey.BATCH_DISABLE:
                if (!tableSelectedRows || tableSelectedRows.length === 0) {
                    message.error('请选择要操作的数据')
                    return
                }
                confirm({
                    title: `请确认是否批量禁用?`,
                    icon: <ExclamationCircleFilled />,
                    content: '',
                    onOk() {
                        dictBatchUpdateStatus({
                            ids: tableSelectedRows.map((item) => item.id),
                            status: Status['STATUS_DISABLED']
                        }).then(() => {
                            message.success('批量禁用成功')
                            handlerRefresh()
                            setTableSelectedRows([])
                        })
                    },
                    onCancel() {
                        message.info('取消操作')
                    }
                })
        }
    }

    return (
        <div>
            <SysLogDetail
                module={ModuleType.ModelTypeDict}
                moduleId={logDataId}
                open={logOpen}
                width={600}
                onClose={closeLogDetail}
            />
            <Detail
                open={openDetail}
                onClose={handlerCloseDetail}
                dictId={dictId}
            />
            <EditDictModal
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
                operationItems={operationItems}
                pageOnChange={handlerTablePageChange}
                pageSize={search?.page?.size}
                current={search?.page?.curr}
                rowSelection={{ onChange: handlerBatchData }}
                action={handlerTableAction}
            />
        </div>
    )
}

export default Dict
