import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { Form } from 'antd'
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

export interface AlarmHistoryProps {}

const getTimeRange = (params: AlarmHistoryListRequest) => {
    return params && params.startAt && params.endAt
        ? [dayjs(+params?.startAt * 1000), dayjs(+params?.endAt * 1000)]
        : undefined
}

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
        const val = {
            ...reqParams,
            ...values,
            startAt: values.time_range && values.time_range[0].unix(),
            endAt: values.time_range && values.time_range[1].unix()
        }
        delete val['time_range']
        console.log('val', val)
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

    const handlerTableAction = (key: ActionKey, record?: AlarmHistoryItem) => {
        console.log('record', record)
        switch (key) {
            case ActionKey.EDIT:
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
        getHistory()
    }, [reqParams])

    return (
        <div>
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
