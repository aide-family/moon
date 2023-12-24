import { useContext, useEffect, useState, useRef, FC } from 'react'

import { Form, message } from 'antd'

import { SearchForm, DataTable } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import type { AccountItemType } from './type'
import { columns, rightOptions, searchItems } from './options'
import { Map, PageReq, defaultPageReqInfo } from '@/apis/types'
import { useSearchParams } from 'react-router-dom'
import QueryString from 'qs'
import { operationItems } from '@/components/Data/DataOption/option'
import { GlobalContext } from '@/context'
import DetailModal from './child/DetailModal'
import EditModal from './child/EditModal'
import { AccountListReq } from '@/apis/home/resource/account/types'
import { DeleteAccount, GetAccountList } from '@/apis/home/resource/account'
import { ActionKey } from '@/apis/data'

const defaultPadding = 12

let timer: NodeJS.Timeout | null = null

const Account: FC = () => {
    const { spaceInfo } = useContext(GlobalContext)
    const oprationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()
    const [urlSearchParams, setUrlSearchParams] = useSearchParams()
    const formData = QueryString.parse(urlSearchParams.toString(), {
        ignoreQueryPrefix: true
    })

    const [dataSource, setDataSource] = useState<AccountItemType[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [pgInfo, setPgInfo] = useState<PageReq>(defaultPageReqInfo)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [openEdit, setOpenEdit] = useState<boolean>(false)
    const [editId, setEditId] = useState<string | undefined>()
    const [detailId, setDetailId] = useState<string | undefined>()

    const [searchParams, setSearchParams] = useState<AccountListReq>({
        ...pgInfo,
        ...formData
    })

    const handlerOpenDetail = (id?: string) => {
        setDetailId(id)
        setOpenDetail(true)
    }

    const handlerOpenEdit = (id?: string) => {
        setEditId(id)
        setOpenEdit(true)
    }

    const handlerCloseDetail = () => {
        setOpenDetail(false)
    }

    const handlerCloseEdit = () => {
        setOpenEdit(false)
    }

    // 分页变化
    const handlerTablePageChange = (page: number, pageSize?: number) => {
        let pg: PageReq = {
            current: page,
            size: pageSize || defaultPageReqInfo.size
        }
        setPgInfo(pg)
        setSearchParams({
            ...searchParams,
            ...pg
        })
    }

    // 刷新
    const handlerRefresh = () => {
        setRefresh((prev) => !prev)
    }

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: React.Key[],
        selectedRows: AccountItemType[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (changedValues: Map) => {
        // 如果有值变化，就重置分页
        if (Object.keys(changedValues).length > 0) {
            setPgInfo(defaultPageReqInfo)
        }
        setSearchParams({
            ...searchParams,
            ...changedValues,
            ...defaultPageReqInfo
        })
    }

    const handleDelete = (id: string) => {
        DeleteAccount(id)
            .then(() => handlerRefresh())
            .catch((e) => message.error(e))
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: AccountItemType) => {
        console.log(key, record)
        switch (key) {
            case ActionKey.DETAIL:
                handlerOpenDetail(record.id)
                break
            case ActionKey.EDIT:
                handlerOpenEdit(record.id)
                break
            case ActionKey.DELETE:
                record.id && handleDelete(record.id)
                break
        }
    }

    // 处理操作栏的点击事件
    const actionHandle = (key: ActionKey) => {
        switch (key) {
            case ActionKey.REFRESH:
                handlerRefresh()
                break
            case ActionKey.ADD:
                handlerOpenEdit()
                break
            case ActionKey.RESET:
                setSearchParams({ ...defaultPageReqInfo })
                handlerRefresh()
                break
        }
    }

    // 获取数据
    const handlerGetData = () => {
        GetAccountList(
            { ...searchParams },
            {
                setLoading: setLoading,
                OK(res) {
                    setDataSource(res.records || [])
                    setTotal(res.total)
                    setPgInfo({
                        size: res.size,
                        current: res.current
                    })
                }
            }
        ).finally(() => setLoading(false))
    }

    const handleEditOnOk = () => {
        handlerCloseEdit()
        handlerRefresh()
    }

    useEffect(() => {
        queryForm.setFieldsValue(formData)
    }, [])

    useEffect(() => {
        setUrlSearchParams({
            ...urlSearchParams,
            ...searchParams
        })
    }, [searchParams])

    useEffect(() => {
        queryForm.resetFields()
        setSearchParams({
            ...defaultPageReqInfo
        })

        handlerRefresh()
    }, [spaceInfo])

    useEffect(() => {
        timer && clearTimeout(timer)
        timer = setTimeout(() => {
            handlerGetData()
        }, 500)
    }, [refresh, urlSearchParams])

    return (
        <div className="bodyContent">
            <DetailModal
                open={openDetail}
                onClose={handlerCloseDetail}
                id={detailId}
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
                    action={actionHandle}
                    rightOptions={rightOptions}
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
                oprationRef={oprationRef}
                total={total}
                loading={loading}
                operationItems={operationItems}
                pageOnChange={handlerTablePageChange}
                pageSize={pgInfo.size}
                current={pgInfo.current}
                rowSelection={{
                    onChange: handlerBatchData
                }}
                action={handlerTableAction}
                expandable={{
                    expandedRowRender: (record) => (
                        <p style={{ margin: 0 }}>{record.remark}</p>
                    )
                }}
            />
        </div>
    )
}

export default Account
