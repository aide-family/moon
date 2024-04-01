import { HeightLine, PaddingLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { Alert, Badge, Form, Tabs, Tag, message } from 'antd'
import { FC, useContext, useEffect, useState } from 'react'
import {
    columns,
    defaultAlarmRealtimeListRequest,
    operationItems,
    rightOptions,
    searchFormItems
} from './options'
import { DataOption, DataTable, SearchForm } from '@/components/Data'
import {
    AlarmRealtimeItem,
    AlarmRealtimeListRequest
} from '@/apis/home/monitor/alarm-realtime/types'
import alarmRealtimeApi from '@/apis/home/monitor/alarm-realtime'
import { ActionKey } from '@/apis/data'
import { GlobalContext } from '@/context'
import PromValueModal from '@/components/Prom/PromValueModal'
import dayjs from 'dayjs'
import alarmHistoryApi from '@/apis/home/monitor/alarm-history'
import EditDictModal from '../../system/dict/child/EditModal'
import { DictListItem } from '@/apis/home/system/dict/types'
import dictApi from '@/apis/home/system/dict'

const article = '默认展示告警时间前一小时到告警恢复时间段内的数据'

let fetchTimer: NodeJS.Timeout | null = null
const AlarmRealtime: FC = () => {
    const [queryForm] = Form.useForm()
    const { size, reltimeAlarmShowRowColor, setReltimeAlarmShowRowColor } =
        useContext(GlobalContext)

    const [alarmPageList, setAlarmPageList] = useState<DictListItem[]>([])
    const [dataSource, setDataSource] = useState<AlarmRealtimeItem[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [queryParams, setQueryParams] = useState<AlarmRealtimeListRequest>(
        defaultAlarmRealtimeListRequest
    )
    const [alarmPageIds, setAlarmPageIds] = useState<number[]>([])
    const [alarmCountMap, setAlarmCountMap] = useState<{
        [key: number]: number | string
    }>()
    const [openEditModal, setOpenEditModal] = useState<boolean>(false)
    const [openAlarmRealtimeValue, setOpenAlarmRealtimeValue] =
        useState<boolean>(false)
    const [alarmRealtimeItem, setAlarmRealtimeItem] =
        useState<AlarmRealtimeItem>()

    const handleCloseAlarmRealtimeValue = () => {
        setOpenAlarmRealtimeValue(false)
    }

    const handleOpenAlarmRealtimeValue = (record: AlarmRealtimeItem) => {
        setOpenAlarmRealtimeValue(true)
        setAlarmRealtimeItem(record)
    }

    const handleRefresh = () => {
        setRefresh((pre) => !pre)
    }

    const handleOpenEditModal = () => {
        setOpenEditModal(true)
    }

    const handleCancelEditModal = () => {
        setOpenEditModal(false)
    }

    const handleEditModelOnOk = () => {
        handleCancelEditModal()
        handleRefresh()
    }

    const handleGetAlarmPageList = () => {
        dictApi.myAlarmPageList().then((res) => {
            setAlarmPageList(res.list)
            setAlarmPageIds(res.list.map((item) => item.id))
        })
    }

    const handleCountAlarmByPageIds = () => {
        if (alarmPageIds.length === 0) {
            return
        }
        dictApi
            .countAlarmPage({
                ids: alarmPageIds
            })
            .then((res) => {
                setAlarmCountMap(res.alarmCount)
            })
    }

    const handleGetAlarmRealtime = () => {
        if (alarmPageIds.length === 0) {
            return
        }

        if (fetchTimer) {
            clearTimeout(fetchTimer)
        }
        fetchTimer = setTimeout(() => {
            setLoading(true)
            alarmRealtimeApi
                .getAlarmRealtimeList({ ...queryParams })
                .then((res) => {
                    setDataSource(res.list)
                    setTotal(res.page.total)
                    return res
                })
                .then(() => {
                    handleCountAlarmByPageIds()
                })
                .finally(() => {
                    setLoading(false)
                })
        }, 1000)
    }

    const handleOnChangeTabs = (key: string) => {
        setLoading(true)
        setQueryParams({
            ...queryParams,
            alarmPageId: +key
        })
    }

    const buildTabsItems = () => {
        return alarmPageList.map((item, index) => {
            const { name, id, color } = item
            return {
                label: (
                    <Badge
                        count={alarmCountMap?.[id] || 0}
                        overflowCount={999}
                        size="small"
                    >
                        <Tag
                            color={color || ''}
                            style={{ textAlign: 'center' }}
                        >
                            {name || `报警页面${index}`}
                        </Tag>
                    </Badge>
                ),
                key: `${id}`
            }
        })
    }

    const handleOptionClick = (action: ActionKey) => {
        switch (action) {
            case ActionKey.REFRESH:
                handleRefresh()
                break
            case ActionKey.BIND_MY_ALARM_PAGES:
                break
            case ActionKey.ADD:
                handleOpenEditModal()
                break
            case ActionKey.RESET:
                setQueryParams(defaultAlarmRealtimeListRequest)
                break
            case ActionKey.ALARM_ROW_COLOR:
                setReltimeAlarmShowRowColor?.(!reltimeAlarmShowRowColor)
                break
            default:
                break
        }
    }

    const handlerTableAction = (
        action: ActionKey,
        record: AlarmRealtimeItem
    ) => {
        switch (action) {
            case ActionKey.ALARM_EVENT_CHART:
                if (!record || !record.historyId) {
                    message.warning('请先选择告警记录')
                    return
                }
                // 查询告警历史
                alarmHistoryApi
                    .getAlarmHistoryDetail({ id: record?.historyId })
                    .then(({ alarmHistory }) => {
                        if (!alarmHistory.expr || !alarmHistory.datasource) {
                            message.warning('无数据源可查看')
                            return
                        }
                        handleOpenAlarmRealtimeValue({
                            ...record,
                            expr: alarmHistory.expr,
                            datasource: alarmHistory.datasource
                        })
                    })
                break
            case ActionKey.ALARM_INTERVENTION:
                console.log('告警介入')
                break
            case ActionKey.ALARM_UPGRADE:
                console.log('告警升级')
                break
            case ActionKey.ALARM_MARK:
                console.log('告警标记')
                break
            case ActionKey.EDIT:
                console.log('编辑')
                break
            case ActionKey.DETAIL:
                console.log('详情', record)
                break
        }
    }

    // 处理搜索表单的值变化
    const handlerSearFormValuesChange = (_: any, allValues: any) => {
        delete allValues.startAt
        delete allValues.endAt
        let requestValues: AlarmRealtimeListRequest = {
            ...queryParams,
            ...allValues,
            startAt: allValues.eventAt && dayjs(allValues.eventAt[0]).unix(),
            endAt: allValues.eventAt && dayjs(allValues.eventAt[1]).unix(),
            page: defaultAlarmRealtimeListRequest.page
        }
        delete requestValues.eventAt
        setQueryParams(requestValues)
    }

    const onRow = (record?: AlarmRealtimeItem) => {
        if (!record || !record.level || !reltimeAlarmShowRowColor) return {}
        const {
            level: { color }
        } = record
        return {
            style: {
                background: color || ''
            }
        }
    }

    useEffect(() => {
        handleRefresh()
    }, [queryParams])

    useEffect(() => {
        handleCountAlarmByPageIds()
        const params = queryParams
        if (!params.alarmPageId) {
            params.alarmPageId = alarmPageIds[0]
            setQueryParams(params)
        }
    }, [alarmPageIds])

    useEffect(() => {
        handleGetAlarmRealtime()
    }, [refresh])

    useEffect(() => {
        handleGetAlarmPageList()
        handleRefresh()
        let interval = setInterval(() => {
            handleRefresh()
        }, 60000)
        return () => {
            clearInterval(interval)
        }
    }, [])

    return (
        <div>
            <PromValueModal
                visible={openAlarmRealtimeValue}
                onCancel={handleCloseAlarmRealtimeValue}
                pathPrefix={alarmRealtimeItem?.datasource || ''}
                expr={alarmRealtimeItem?.expr}
                height={400}
                eventAt={alarmRealtimeItem?.eventAt}
                alert={
                    <Alert
                        style={{ width: '96%' }}
                        message={article}
                        type="info"
                        showIcon
                    />
                }
                title={`${alarmRealtimeItem?.strategy?.alert}: ${dayjs(
                    +(alarmRealtimeItem?.eventAt || 0) * 1000
                ).format('YYYY-MM-DD HH:mm:ss')}`}
            />
            <EditDictModal
                open={openEditModal}
                onClose={handleCancelEditModal}
                onOk={handleEditModelOnOk}
            />
            <RouteBreadcrumb />
            <HeightLine />
            <SearchForm
                form={queryForm}
                items={searchFormItems}
                formProps={{
                    onValuesChange: handlerSearFormValuesChange
                }}
            />
            <HeightLine />
            <DataOption
                queryForm={queryForm}
                rightOptions={rightOptions({
                    refresh: handleGetAlarmPageList,
                    reltimeAlarmShowRowColor: reltimeAlarmShowRowColor
                })}
                action={handleOptionClick}
                // showAdd={false}
            />
            <PaddingLine padding={12} height={1} borderRadius={4} />
            <Tabs
                items={buildTabsItems()}
                onChange={handleOnChangeTabs}
                tabBarStyle={{
                    boxShadow: '0px 0px 10px 0px rgba(0,0,0,0.1)'
                }}
                size={size}
            />
            {alarmPageList.length > 0 && (
                <DataTable
                    showIndex={false}
                    // showOperation={false}
                    columns={columns({})}
                    dataSource={dataSource}
                    total={total}
                    loading={loading}
                    operationItems={operationItems}
                    action={handlerTableAction}
                    onRow={onRow}
                    pageSize={queryParams?.page?.size}
                    current={queryParams?.page?.curr}
                />
            )}
        </div>
    )
}

export default AlarmRealtime
