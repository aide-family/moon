import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { Form } from 'antd'
import { FC, useEffect, useRef, useState } from 'react'
import {
    columns,
    defaultListAlarmPageRequest,
    defaultPadding,
    leftOptions,
    rightOptions,
    searchItems,
    tableOperationItems
} from './options'
import { ActionKey } from '@/apis/data'
import {
    AlarmPageItem,
    ListAlarmPageRequest
} from '@/apis/home/monitor/alarm-page/types'
import alarmPageApi from '@/apis/home/monitor/alarm-page'
import EditAlarmPageModal from './child/EditAlarmPageModal'
import { ModuleType, Status } from '@/apis/types'
import { SysLogDetail } from '../../child/SysLogDetail'

let timer: NodeJS.Timeout

const AlarmPage: FC = () => {
    const operationRef = useRef<HTMLDivElement>(null)
    const [queryForm] = Form.useForm()
    const [openEditModal, setOpenEditModal] = useState<boolean>(false)
    const [editDetail, setEditDetail] = useState<AlarmPageItem | undefined>()
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [loading, setLoading] = useState<boolean>(false)
    const [dataSource, setDataSource] = useState<AlarmPageItem[]>([])
    const [searchRequest, setSearchRequest] = useState<ListAlarmPageRequest>(
        defaultListAlarmPageRequest
    )
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

    const handleRefresh = () => {
        setRefresh(!refresh)
    }

    const handleOpenEditModal = (item?: AlarmPageItem) => {
        setEditDetail(item)
        setOpenEditModal(true)
    }

    const handleCancelEditModal = () => {
        setOpenEditModal(false)
        setEditDetail(undefined)
    }

    const handleEditModelOnOk = () => {
        handleCancelEditModal()
        handleRefresh()
    }

    const handlerSearchFormValuesChange = (_: any, allValues: any) => {
        if (timer) {
            clearTimeout(timer)
        }

        timer = setTimeout(() => {
            setSearchRequest({
                ...searchRequest,
                ...allValues,
                page: defaultListAlarmPageRequest.page
            })
        }, 500)
    }

    const handleChangeStatus = (ids: number[], status: Status) => {
        return alarmPageApi
            .batchUpdateAlarmPageStatus({
                ids,
                status
            })
            .then(() => {
                handleRefresh()
            })
    }

    const handlerDataOptionAction = (key: ActionKey) => {
        switch (key) {
            case ActionKey.REFRESH:
                setRefresh(!refresh)
                break
            case ActionKey.RESET:
                setSearchRequest(defaultListAlarmPageRequest)
                break
            case ActionKey.EXPORT:
                break
            case ActionKey.BATCH_IMPORT:
                break
            case ActionKey.BATCH_EXPORT:
                break
            case ActionKey.ADD:
                handleOpenEditModal()
                break
            default:
                break
        }
    }

    const handlerTableAction = (key: ActionKey, item: AlarmPageItem) => {
        switch (key) {
            case ActionKey.EDIT:
                handleOpenEditModal(item)
                break
            case ActionKey.DELETE:
                break
            case ActionKey.ENABLE:
                handleChangeStatus([item.id], Status.STATUS_ENABLED)
                break
            case ActionKey.DISABLE:
                handleChangeStatus([item.id], Status.STATUS_DISABLED)
                break
            case ActionKey.OPERATION_LOG:
                openLogDetail(item.id)
                break
            default:
                break
        }
    }

    const handlerTablePageChange = (page: number, size: number) => {
        setSearchRequest({
            ...searchRequest,
            page: {
                curr: page,
                size: size
            }
        })
    }

    const getAlarmPageList = () => {
        setLoading(true)
        alarmPageApi
            .getAlarmPageList(searchRequest)
            .then((res) => {
                setTotal(res.page.total)
                setDataSource(res.list)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    useEffect(() => {
        getAlarmPageList()
    }, [refresh, searchRequest])

    return (
        <div>
            <SysLogDetail
                module={ModuleType.ModuleAlarmPage}
                moduleId={logDataId}
                open={logOpen}
                width={600}
                onClose={closeLogDetail}
            />
            <EditAlarmPageModal
                open={openEditModal}
                onCancel={handleCancelEditModal}
                onOk={handleEditModelOnOk}
                id={editDetail?.id}
            />
            <div ref={operationRef}>
                <RouteBreadcrumb />
                <HeightLine />
                <SearchForm
                    form={queryForm}
                    items={searchItems}
                    formProps={{
                        onValuesChange: handlerSearchFormValuesChange
                    }}
                />
                <HeightLine />
                <DataOption
                    queryForm={queryForm}
                    rightOptions={rightOptions}
                    leftOptions={leftOptions}
                    action={handlerDataOptionAction}
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
                total={+total}
                loading={loading}
                operationItems={tableOperationItems}
                pageOnChange={handlerTablePageChange}
                action={handlerTableAction}
                pageSize={searchRequest?.page?.size}
                current={searchRequest?.page?.curr}
            />
        </div>
    )
}

export default AlarmPage
