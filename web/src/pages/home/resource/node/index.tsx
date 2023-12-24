import React, { useContext, useEffect, useState } from 'react'

import { Form, message } from 'antd'
import { SearchForm, DataTable } from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { DataOption } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import type { NodeItemType } from './type'

import {
    columns,
    rightOptions,
    searchItems,
} from './options'
import DetailModal from './child/DetailModal'
import { Map, PageReq, defaultPageReqInfo } from '@/apis/types'
import QueryString from 'qs'
import { useSearchParams } from 'react-router-dom'
import EditModal from './child/EditModal'
import { GlobalContext } from '@/context'
import { DeleteNode, GetNodeList } from '@/apis/home/resource/node'
import { NodeListReq } from '@/apis/home/resource/node/types'
import { ActionKey } from '@/apis/data'

const defaultPadding = 12

let timer: NodeJS.Timeout | null = null

const Node: React.FC = () => {
    const { spaceInfo } = useContext(GlobalContext)
    const oprationRef = React.useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()
    const [urlSearchParams, setUrlSearchParams] = useSearchParams()
    const formData = QueryString.parse(urlSearchParams.toString(), {
        ignoreQueryPrefix: true
    })

    const [dataSource, setDataSource] = React.useState<NodeItemType[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [pgInfo, setPgInfo] = useState<PageReq>(defaultPageReqInfo)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [detailId, setDetailId] = useState<string>('')
    const [openEdit, setOpenEdit] = useState<boolean>(false)
    const [editId, setEditId] = useState<string | undefined>()

    const [searchParams, setSearchParams] = useState<NodeListReq>({
        ...pgInfo,
        ...formData
    })

    const handlerOpenDetail = (record: NodeItemType) => {
        setDetailId(record.id || '')
        setOpenDetail(true)
    }

    const handlerCloseDetail = () => {
        setOpenDetail(false)
        setDetailId('')
    }

    // 获取数据
    const handlerGetData = () => {
        GetNodeList(searchParams, {
            setLoading: setLoading,
            OK: (res) => {
                setDataSource(res.records || [])
                setTotal(res.total)
                setPgInfo({
                    size: res.size,
                    current: res.current
                })
            }
        }).finally(() => setLoading(false))
    }

    // 刷新
    const handlerRefresh = () => {
        setRefresh((prev) => !prev)
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

    const handlerOpenEdit = (id?: string) => {
        setEditId(id)
        setOpenEdit(true)
    }

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: React.Key[],
        selectedRows: NodeItemType[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: NodeItemType) => {
        switch (key) {
            case ActionKey.DETAIL:
                handlerOpenDetail(record)
                break
            case ActionKey.IKUAI:
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
    const handleOperationAction = (key: string) => {
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

    const handlerCloseEdit = () => {
        setOpenEdit(false)
    }

    const handleEditOnOk = () => {
        handlerCloseEdit()
        handlerRefresh()
    }

    const handleDelete = (id: string) => {
        DeleteNode(id)
            .then(() => handlerRefresh())
            .catch((e) => message.error(e))
    }

    useEffect(() => {
        queryForm.setFieldsValue(formData)
    }, [])

    useEffect(() => {
        queryForm.resetFields()
        setSearchParams({
            ...defaultPageReqInfo
        })

        handlerRefresh()
    }, [spaceInfo])

    useEffect(() => {
        setUrlSearchParams({
            ...urlSearchParams,
            ...searchParams
        })
    }, [searchParams])

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
                    rightOptions={rightOptions}
                    action={handleOperationAction}
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
                operationItems={[]}
                pageOnChange={handlerTablePageChange}
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

export default Node
