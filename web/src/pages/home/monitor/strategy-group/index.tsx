import React, { Key, useEffect, useRef, useState } from 'react'
import { Form } from 'antd'
import { useNavigate } from 'react-router-dom'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import { DataOption, DataTable, SearchForm } from '@/components/Data'
import {
    defaultStrategyGroupListRequest,
    leftOptions,
    rightOptions,
    searchItems,
    tableOperationItems
} from '@/pages/home/monitor/strategy-group/options.tsx'
import { ActionKey } from '@/apis/data.ts'
import { columns } from './options'
import strategyGroupApi from '@/apis/home/monitor/strategy-group'
import {
    StrategyGroupItemType,
    StrategyGroupListRequest
} from '@/apis/home/monitor/strategy-group/types'
import { Detail } from './child/detail'
import EditGroupModal from './child/EditGroupModal'
import { Status } from '@/apis/types'

const defaultPadding = 12

let timer: NodeJS.Timeout
const StrategyGroup: React.FC = () => {
    const navigate = useNavigate()
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()

    const [dataSource, setDataSource] = useState<StrategyGroupItemType[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState<boolean>(false)
    const [search, setSearch] = useState<StrategyGroupListRequest>(
        defaultStrategyGroupListRequest
    )
    const [openEditModal, setOpenEditModal] = useState<boolean>(false)
    const [editItem, setEditItem] = useState<
        StrategyGroupItemType | undefined
    >()

    // 刷新
    const handlerRefresh = () => {
        setRefresh(!refresh)
    }

    const handlerOpenDetail = () => {
        setOpenDetail(true)
    }

    const handlerCloseDetail = () => {
        setOpenDetail(false)
    }

    const handleOpenEditModal = (record?: StrategyGroupItemType) => {
        setEditItem(record)
        setOpenEditModal(true)
    }

    const handleCloseEditModal = () => {
        setOpenEditModal(false)
        setEditItem(undefined)
    }

    const handleEditModalOnOK = () => {
        handleCloseEditModal()
        handlerRefresh()
    }

    // 获取数据
    const handlerGetData = () => {
        setLoading(true)
        strategyGroupApi
            .getStrategyGroupList(search)
            .then((res) => {
                setDataSource(res.list)
                setTotal(res.page.total)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    // 分页变化
    const handlerTablePageChange = (page: number, pageSize?: number) => {
        console.log(page, pageSize)
    }

    // 可以批量操作的数据
    const handlerBatchData = (
        selectedRowKeys: Key[],
        selectedRows: StrategyGroupItemType[]
    ) => {
        console.log(selectedRowKeys, selectedRows)
    }

    const toStrategyListPage = (record: StrategyGroupItemType) => {
        console.log(record)
        navigate(`/home/monitor/strategy?groupId=${record.id}`)
    }

    const handlebatchChangeStatus = (ids: number[], status: Status) => {
        return strategyGroupApi.batchChangeStatus({ ids, status })
    }

    // 处理表格操作栏的点击事件
    const handlerTableAction = (key: string, record: StrategyGroupItemType) => {
        switch (key) {
            case ActionKey.OP_KEY_STRATEGY_LIST:
                toStrategyListPage(record)
                break
            case ActionKey.DETAIL:
                handlerOpenDetail()
                break
            case ActionKey.EDIT:
                handleOpenEditModal(record)
                break
            case ActionKey.DISABLE:
                handlebatchChangeStatus(
                    [record.id],
                    Status.STATUS_DISABLED
                ).then(handlerRefresh)
                break
            case ActionKey.ENABLE:
                handlebatchChangeStatus(
                    [record.id],
                    Status.STATUS_ENABLED
                ).then(handlerRefresh)
                break
        }
    }

    const hendleDataAction = (key: ActionKey) => {
        switch (key) {
            case ActionKey.REFRESH:
                handlerRefresh()
                break
            case ActionKey.ADD:
                handleOpenEditModal()
                break
            case ActionKey.BATCH_IMPORT:
                break
            default:
                break
        }
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (_: any, allValues: any) => {
        if (timer) {
            clearTimeout(timer)
        }
        timer = setTimeout(() => {
            setSearch({
                ...search,
                ...allValues
            })
            handlerRefresh()
        }, 500)
    }

    useEffect(() => {
        handlerGetData()
    }, [refresh])

    return (
        <div className="bodyContent">
            <Detail open={openDetail} onClose={handlerCloseDetail} id="1" />
            <EditGroupModal
                open={openEditModal}
                onCancel={handleCloseEditModal}
                onOk={handleEditModalOnOK}
                groupId={editItem?.id}
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
                    rightOptions={rightOptions(loading)}
                    leftOptions={leftOptions(loading)}
                    action={hendleDataAction}
                />
                <PaddingLine
                    padding={defaultPadding}
                    height={1}
                    borderRadius={4}
                />
                DE
            </div>

            <DataTable
                rowKey="id"
                dataSource={dataSource}
                columns={columns}
                total={total}
                loading={loading}
                operationItems={tableOperationItems}
                pageOnChange={handlerTablePageChange}
                rowSelection={{
                    onChange: handlerBatchData
                }}
                action={handlerTableAction}
            />
        </div>
    )
}

export default StrategyGroup
