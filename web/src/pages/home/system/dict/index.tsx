import { useEffect, useRef, useState } from 'react'
import type { ColumnType, ColumnGroupType } from 'antd/es/table'
import { Badge, Button, Form, Modal, message } from 'antd'
import { SearchForm, DataTable } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { DataOptionItem } from '@/components/Data/DataOption/DataOption'
import Detail from './child/Detail'
import EditModal from './child/EditModal'
import dictOptions from './options'
import dictApi from '@/apis/home/system/dict'
import { Category, Status, StatusMap } from '@/apis/types'
import { ExclamationCircleFilled } from '@ant-design/icons'
import { DictListItem, DictListReq } from '@/apis/home/system/dict/types'
import { ActionKey, categoryData } from '@/apis/data'

const { confirm } = Modal
const { dictList, dictDelete, dictBatchUpdateStatus } = dictApi
const { searchItems, operationItems } = dictOptions()

const defaultPadding = 12

let timer: NodeJS.Timeout

type DictColumnType = ColumnGroupType<DictListItem> | ColumnType<DictListItem>

/**
 * 字典管理
 */
const Dict: React.FC = () => {
    const oprationRef = useRef<HTMLDivElement>(null)
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

    const columns: DictColumnType[] = [
        {
            title: '字典名称',
            dataIndex: 'name',
            key: 'name',
            width: 220
        },
        {
            title: '字典类型',
            dataIndex: 'category',
            key: 'category',
            width: 100,
            render: (category: Category) => {
                return categoryData[category]
            }
        },
        {
            title: '字典颜色',
            dataIndex: 'color',
            key: 'color',
            align: 'center',
            width: 220,
            render: (color: string) => {
                return (
                    <Badge
                        color={color}
                        text={color}
                        style={{
                            backgroundColor: color,
                            color: '#fff',
                            width: '60%',
                            textAlign: 'center'
                        }}
                    />
                )
            }
        },
        {
            title: '字典状态',
            dataIndex: 'status',
            key: 'status',
            width: 100,
            render: (status: Status) => {
                return (
                    <Badge
                        color={StatusMap[status].color}
                        text={StatusMap[status].text}
                    />
                )
            }
        },
        {
            // TODO 两行溢出显示省略号
            title: '备注',
            dataIndex: 'remark',
            key: 'remark',
            // width: 200,
            ellipsis: true
        }
    ]

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
        selectedRows: DictListItem[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
        setTableSelectedRows(selectedRows)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: DictListItem) => {
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
                ...changedValues
            })
            console.log(changedValues, allValues)
        }, 500)
    }

    const isDisabled = () => {
        return tableSelectedRows.length === 0
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
        console.log('val----', val)
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
                            setTableSelectedRows([])
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
    ;(function () {
        console.log('a')
    })()

    return (
        <div className="bodyContent">
            <Detail
                open={openDetail}
                onClose={handlerCloseDetail}
                dictId={dictId}
            />
            <EditModal
                open={openEdit}
                onClose={handlerCloseEdit}
                id={editId}
                onOk={handleEditOnOk}
            />
            <div ref={oprationRef}>
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
                operationRef={oprationRef}
                total={total}
                loading={loading}
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

export default Dict
