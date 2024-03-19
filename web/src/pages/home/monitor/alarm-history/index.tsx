import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { Alert, Form, message } from 'antd'
import React, { useContext, useEffect, useState } from 'react'
import {
    columns,
    tableOperationItems,
    rightOptions,
    searchFormItems
} from './options'
import alarmHistoryApi from '@/apis/home/monitor/alarm-history'
import {
    AlarmHistoryItem,
    AlarmHistoryListRequest,
    defaultAlarmHistoryListRequest
} from '@/apis/home/monitor/alarm-history/types'
import { ActionKey } from '@/apis/data'
import SyntaxHighlighter from 'react-syntax-highlighter'
import {
    atomOneDark,
    atomOneLight
} from 'react-syntax-highlighter/dist/esm/styles/hljs'
import { GlobalContext } from '@/context'
import dayjs from 'dayjs'
import { HistoryDetail } from './child/Detail'
import PromValueModal from '@/components/Prom/PromValueModal'

export interface AlarmHistoryProps {}

const getTimeRange = (params: AlarmHistoryListRequest) => {
    return params && params.firingStartAt && params.firingEndAt
        ? [
              dayjs(+params?.firingStartAt * 1000),
              dayjs(+params?.firingEndAt * 1000)
          ]
        : undefined
}
const article =
    '默认展示告警时间前一小时到告警恢复时间段内的数据，如果告警未恢复，则展示告警时间到当前时刻的数据'

const AlarmHistory: React.FC<AlarmHistoryProps> = (props) => {
    const {} = props
    const { sysTheme } = useContext(GlobalContext)

    const [queryForm] = Form.useForm()

    const [reqParams, setReqParams] = useState<AlarmHistoryListRequest>(
        defaultAlarmHistoryListRequest
    )
    const [datasource, setDatasource] = useState<AlarmHistoryItem[]>([])
    const [dataTotal, setDataTotal] = useState<number>(0)
    const [loading, setLoading] = useState<boolean>(false)
    const [openDetail, setOpenDetail] = useState(false)
    const [detailInfo, setDetailInfo] = useState<AlarmHistoryItem>()
    const [openAlarmRealtimeValue, setOpenAlarmRealtimeValue] =
        useState<boolean>(false)
    const handleOpenDetail = (item?: AlarmHistoryItem) => {
        setOpenDetail(true)
        setDetailInfo(item)
    }

    const handleCloseDetail = () => {
        setOpenDetail(false)
        setDetailInfo(undefined)
    }

    const getHistory = () => {
        setLoading(true)
        alarmHistoryApi
            .getAlarmHistoryList(reqParams)
            .then((res) => {
                const {
                    list,
                    page: { total }
                } = res
                setDatasource(list)

                setDataTotal(total)
            })
            .finally(() => {
                setLoading(false)
            })
    }

    const handlerSearFormValuesChange = (_: any, values: any) => {
        delete values.firingStartAt
        delete values.firingEndAt
        delete values.resolvedStartAt
        delete values.resolvedEndAt
        const val: AlarmHistoryListRequest = {
            ...reqParams,
            ...values,
            firingStartAt:
                values.firingTime && dayjs(values.firingTime[0]).unix(),
            firingEndAt:
                values.firingTime && dayjs(values.firingTime[1]).unix(),
            resolvedStartAt:
                values.resolvedTime && dayjs(values.resolvedTime[0]).unix(),
            resolvedEndAt:
                values.resolvedTime && dayjs(values.resolvedTime[1]).unix()
        }
        delete val.firingTime
        delete val.resolvedTime
        setReqParams(val)
    }

    const handleOptionClick = (key: ActionKey) => {
        switch (key) {
            case ActionKey.REFRESH:
                getHistory()
                break
            case ActionKey.RESET:
                setReqParams(defaultAlarmHistoryListRequest)
                queryForm.setFieldsValue({
                    ...defaultAlarmHistoryListRequest,
                    time_range: getTimeRange(defaultAlarmHistoryListRequest)
                })
                break
        }
    }

    const handleOpenAlarmRealtimeValue = (record: AlarmHistoryItem) => {
        setOpenAlarmRealtimeValue(true)
        setDetailInfo(record)
    }

    const handleCloseAlarmRealtimeValue = () => {
        setOpenAlarmRealtimeValue(false)
        setDetailInfo(undefined)
    }

    const handlerTableAction = (key: ActionKey, record?: AlarmHistoryItem) => {
        switch (key) {
            case ActionKey.ALARM_EVENT_CHART:
                if (!record || !record.id) return
                if (!record.expr || !record.datasource) {
                    message.warning('无数据源可查看')
                    return
                }
                handleOpenAlarmRealtimeValue(record)
                break
            case ActionKey.EDIT:
                break
            case ActionKey.DETAIL:
                handleOpenDetail(record)
                break
            case ActionKey.DELETE:
                break
        }
    }

    const handlePageOnChange = (page: number, size: number) => {
        setReqParams({
            ...reqParams,
            page: {
                curr: page,
                size
            }
        })
    }

    useEffect(() => {
        queryForm.setFieldsValue(reqParams)
        getHistory()
    }, [reqParams])

    return (
        <div>
            <PromValueModal
                visible={openAlarmRealtimeValue}
                onCancel={handleCloseAlarmRealtimeValue}
                pathPrefix={detailInfo?.datasource || ''}
                expr={detailInfo?.expr}
                height={400}
                eventAt={detailInfo?.startAt}
                endAt={
                    (detailInfo?.endAt || 0) > 0
                        ? detailInfo?.endAt
                        : dayjs().unix()
                }
                alert={
                    <Alert
                        style={{ width: '96%' }}
                        message={article}
                        type="info"
                        showIcon
                    />
                }
                title={`${detailInfo?.alarmName}: ${dayjs(
                    +(detailInfo?.startAt || 0) * 1000
                ).format('YYYY-MM-DD HH:mm:ss')}`}
            />
            <HistoryDetail
                open={openDetail}
                onClose={handleCloseDetail}
                historyId={detailInfo?.id}
            />
            <RouteBreadcrumb />
            <HeightLine />
            <SearchForm
                form={queryForm}
                items={searchFormItems}
                formProps={{
                    onValuesChange: handlerSearFormValuesChange,
                    initialValues: {
                        ...reqParams,
                        time_range: getTimeRange(reqParams)
                    }
                }}
            />
            <HeightLine />
            <DataOption
                queryForm={queryForm}
                rightOptions={rightOptions}
                action={handleOptionClick}
                showAdd={false}
            />
            <PaddingLine padding={12} height={1} borderRadius={4} />
            <DataTable
                showIndex={false}
                columns={columns}
                dataSource={datasource}
                total={dataTotal}
                loading={loading}
                pageOnChange={handlePageOnChange}
                operationItems={tableOperationItems}
                action={handlerTableAction}
                pageSize={reqParams?.page?.size}
                current={reqParams?.page?.curr}
                expandable={{
                    expandedRowRender: (record: AlarmHistoryItem) => {
                        return (
                            <>
                                <SyntaxHighlighter
                                    key={record.id}
                                    language="json"
                                    style={
                                        sysTheme === 'dark'
                                            ? atomOneDark
                                            : atomOneLight
                                    }
                                >
                                    {JSON.stringify(record, null, 2)}
                                </SyntaxHighlighter>
                            </>
                        )
                    }
                }}
            />
        </div>
    )
}

export default AlarmHistory
