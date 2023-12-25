import React, {useContext, useEffect, useState} from 'react'

import {Form, message} from 'antd'
import {defaultPageReqInfo, type Map, type PageReq} from '@/apis/types'
import type {DeviceItemType} from './type'
import {DataOption, DataTable, SearchForm} from '@/components/Data'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import {HeightLine, PaddingLine} from '@/components/HeightLine'
import {operationItems} from '@/components/Data/DataOption/option'
import {columns, leftOptions, rightOptions, searchItems} from './options'
import {useSearchParams} from 'react-router-dom'
import QueryString from 'qs'
import {GlobalContext} from '@/context'
import EditModal from './child/EditModal'
import DetailModal from './child/DetailModal'
import {EquipmentListReq} from '@/apis/home/resource/device/types'
import {DeleteEquipment, GetEquipmentList} from '@/apis/home/resource/device'
import {ActionKey} from '@/apis/data'

const defaultPadding = 12

let timer: NodeJS.Timeout | null = null

const Device: React.FC = () => {
    const {spaceInfo} = useContext(GlobalContext)

    const operationRef = React.useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()
    const [urlSearchParams, setUrlSearchParams] = useSearchParams()
    const formData = QueryString.parse(urlSearchParams.toString(), {
        ignoreQueryPrefix: true
    })

    const [dataSource, setDataSource] = useState<DeviceItemType[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [pgInfo, setPgInfo] = useState<PageReq>(defaultPageReqInfo)
    const [refresh, setRefresh] = useState<boolean>(false)

    const [searchParams, setSearchParams] = useState<EquipmentListReq>({
        ...pgInfo,
        ...formData
    })
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [openEdit, setOpenEdit] = useState<boolean>(false)
    const [editId, setEditId] = useState<string | undefined>()
    const [detailId, setDetailId] = useState<string | undefined>()

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

    // 获取数据
    const handlerGetData = () => {
        GetEquipmentList(
            {...searchParams},
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

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: React.Key[],
        selectedRows: DeviceItemType[]
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
        DeleteEquipment(id)
            .then(() => handlerRefresh())
            .catch((e) => message.error(e))
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: ActionKey, record: DeviceItemType) => {
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
    const actionHandle = (key: string) => {
        switch (key) {
            case ActionKey.REFRESH:
                handlerRefresh()
                break
            case ActionKey.EDIT:
                handlerOpenEdit()
                break
            case ActionKey.RESET:
                setSearchParams({...defaultPageReqInfo})
                handlerRefresh()
                break
        }
    }

    const handleEditOnOk = () => {
        handlerCloseEdit()
        handlerRefresh()
    }

    // 首次加载页面, 把url中的参数设置到搜索表单中 current=1&size=10&name=12421
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
            <EditModal
                open={openEdit}
                onClose={handlerCloseEdit}
                id={editId}
                onOk={handleEditOnOk}
            />
            <DetailModal
                open={openDetail}
                onClose={handlerCloseDetail}
                id={detailId}
            />
            <div ref={operationRef}>
                <RouteBreadcrumb/>
                <HeightLine/>
                <SearchForm
                    form={queryForm}
                    items={searchItems}
                    formProps={{
                        onValuesChange: handlerSearFormValuesChange
                    }}
                />
                <HeightLine/>
                <DataOption
                    queryForm={queryForm}
                    rightOptions={rightOptions}
                    leftOptions={leftOptions}
                    action={actionHandle}
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
                oprationRef={operationRef}
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
                        <p style={{margin: 0}}>{record.remark}</p>
                    )
                }}
            />
        </div>
    )
}

export default Device
