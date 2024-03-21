import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { FC, useEffect, useRef, useState } from 'react'
import { Form, Table } from 'antd'
import {
    AlarmGroupItem,
    ListAlarmGroupRequest,
    defaultListAlarmGroupRequest
} from '@/apis/home/monitor/alarm-group/types'
import { ActionKey } from '@/apis/data'
import {
    chartGroupCoumns,
    columns,
    defaultPadding,
    leftOptions,
    rightOptions,
    searchItems,
    tableOperationItems
} from './options'
import alarmGroupApi from '@/apis/home/monitor/alarm-group'
import EditAlarmGroupModal from './child/EditAlarmGroupModal'
import { ModuleType } from '@/apis/types'
import { SysLogDetail } from '../../child/SysLogDetail'

let timer: NodeJS.Timeout | null = null
const AlarmGroup: FC = () => {
    const [queryForm] = Form.useForm()
    const operationRef = useRef<HTMLDivElement>(null)
    const [searchRequest, setSearchRequest] = useState<ListAlarmGroupRequest>(
        defaultListAlarmGroupRequest
    )
    const [refresh, setRefresh] = useState<boolean>(false)
    const [loading, setLoading] = useState<boolean>(false)
    const [dataSource, setDataSource] = useState<AlarmGroupItem[]>([])
    const [total, setTotal] = useState<number>(0)
    const [oprateAlarmGroupId, setOperateAlarmGroupId] = useState<number>()
    const [openEditAlarmGroupModal, setOpenEditAlarmGroupModal] =
        useState<boolean>(false)
    const [actionKey, setActionKey] = useState<ActionKey>(ActionKey.EDIT)
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

    const handleOpenEditAlarmGroupModal = (key: ActionKey, id?: number) => {
        setOperateAlarmGroupId(id)
        setOpenEditAlarmGroupModal(true)
        setActionKey(key)
    }

    const handleCancelEditAlarmGroupModal = () => {
        setOperateAlarmGroupId(undefined)
        setOpenEditAlarmGroupModal(false)
    }

    const handleRefresh = () => {
        setRefresh((p) => !p)
    }

    const handOnOkEditAlarmGroupModal = () => {
        handleCancelEditAlarmGroupModal()
        handleRefresh()
    }

    const handleGetAlarmGroupList = () => {
        if (timer) {
            clearTimeout(timer)
        }
        timer = setTimeout(() => {
            setLoading(true)
            alarmGroupApi
                .list(searchRequest)
                .then((data) => {
                    setDataSource(data.list || [])
                    setTotal(data.page.total)
                    return data
                })
                .finally(() => {
                    setLoading(false)
                })
        }, 500)
    }

    const handlerSearchFormValuesChange = (_: any, allValues: any) => {
        setSearchRequest({
            ...searchRequest,
            ...allValues
        })
    }

    const handlerDataOptionAction = (key: ActionKey) => {
        switch (key) {
            case ActionKey.REFRESH:
                setRefresh(!refresh)
                break
            case ActionKey.RESET:
                setSearchRequest(defaultListAlarmGroupRequest)
                break
            case ActionKey.EXPORT:
                break
            case ActionKey.BATCH_IMPORT:
                break
            case ActionKey.BATCH_EXPORT:
                break
            case ActionKey.ADD:
                handleOpenEditAlarmGroupModal(key)
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

    const handlerTableAction = (key: ActionKey, item: AlarmGroupItem) => {
        switch (key) {
            case ActionKey.EDIT:
                handleOpenEditAlarmGroupModal(key, item.id)
                break
            case ActionKey.DETAIL:
                handleOpenEditAlarmGroupModal(key, item.id)
                break
            case ActionKey.DELETE:
                break
            case ActionKey.ENABLE:
                // handleChangeStatus([item.id], Status.STATUS_ENABLED)
                break
            case ActionKey.DISABLE:
                // handleChangeStatus([item.id], Status.STATUS_DISABLED)
                break
            case ActionKey.OPERATION_LOG:
                openLogDetail(item.id)
                break
            default:
                break
        }
    }

    useEffect(() => {
        handleRefresh()
    }, [searchRequest])

    useEffect(() => {
        handleGetAlarmGroupList()
    }, [refresh])

    return (
        <>
            <SysLogDetail
                module={ModuleType.ModuleAlarmNotifyGroup}
                moduleId={logDataId}
                open={logOpen}
                width={600}
                onClose={closeLogDetail}
            />
            <EditAlarmGroupModal
                open={openEditAlarmGroupModal}
                alarmGroupId={oprateAlarmGroupId}
                onClose={handleCancelEditAlarmGroupModal}
                onOk={handOnOkEditAlarmGroupModal}
                disabled={
                    actionKey !== ActionKey.EDIT && actionKey !== ActionKey.ADD
                }
            />
            <div>
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
                    expandable={{
                        expandedRowRender: (record: AlarmGroupItem) => {
                            return (
                                <Table
                                    columns={chartGroupCoumns}
                                    dataSource={record.chatGroups}
                                    size="small"
                                    pagination={false}
                                    rowKey="id"
                                />
                            )
                        }
                    }}
                />
            </div>
        </>
    )
}

export default AlarmGroup
