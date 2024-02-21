import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { FC, useEffect, useRef, useState } from 'react'
import { Form } from 'antd'
import {
    AlarmGroupItem,
    ListAlarmGroupRequest,
    defaultListAlarmGroupRequest
} from '@/apis/home/monitor/alarm-group/types'
import { ActionKey } from '@/apis/data'
import {
    columns,
    defaultPadding,
    leftOptions,
    rightOptions,
    searchItems,
    tableOperationItems
} from './options'
import alarmGroupApi from '@/apis/home/monitor/alarm-group'
import EditAlarmGroupModal from './child/EditAlarmGroupModal'

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

    const handleOpenEditAlarmGroupModal = (id?: number) => {
        setOperateAlarmGroupId(id)
        setOpenEditAlarmGroupModal(true)
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
                handleOpenEditAlarmGroupModal()
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
                handleOpenEditAlarmGroupModal(item.id)
                break
            case ActionKey.DELETE:
                break
            case ActionKey.ENABLE:
                // handleChangeStatus([item.id], Status.STATUS_ENABLED)
                break
            case ActionKey.DISABLE:
                // handleChangeStatus([item.id], Status.STATUS_DISABLED)
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
            <EditAlarmGroupModal
                open={openEditAlarmGroupModal}
                alarmGroupId={oprateAlarmGroupId}
                onClose={handleCancelEditAlarmGroupModal}
                onOk={handOnOkEditAlarmGroupModal}
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
                />
            </div>
        </>
    )
}

export default AlarmGroup
