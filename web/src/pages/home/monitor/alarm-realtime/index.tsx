import alarmPageApi from '@/apis/home/monitor/alarm-page'
import { AlarmPageSelectItem } from '@/apis/home/monitor/alarm-page/types'
import { HeightLine, PaddingLine } from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import { Button, Form, Tabs } from 'antd'
import { FC, useEffect, useState } from 'react'
import {
    columns,
    defaultAlarmPageSelectReq,
    defaultAlarmRealtimeListRequest,
    operationItems,
    rightOptions,
    searchFormItems
} from './options'
import { IconFont } from '@/components/IconFont/IconFont'
import { DataOption, DataTable, SearchForm } from '@/components/Data'
import { AlarmRealtimeItem } from '@/apis/home/monitor/alarm-realtime/types'
import alarmRealtimeApi from '@/apis/home/monitor/alarm-realtime'
import { ActionKey } from '@/apis/data'

const AlarmRealtime: FC = () => {
    const [queryForm] = Form.useForm()

    const [alarmPageList, setAlarmPageList] = useState<AlarmPageSelectItem[]>(
        []
    )
    const [dataSource, setDataSource] = useState<AlarmRealtimeItem[]>([])
    const [loading, setLoading] = useState<boolean>(false)
    const [total, setTotal] = useState<number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)

    const handleRefresh = () => {
        setRefresh(!refresh)
    }

    const handleGetAlarmPageList = () => {
        alarmPageApi
            .getAlarmPageSelect(defaultAlarmPageSelectReq)
            .then((res) => {
                setAlarmPageList(res.list)
            })
    }

    const handleGetAlarmRealtime = () => {
        setLoading(true)
        alarmRealtimeApi
            .getAlarmRealtimeList(defaultAlarmRealtimeListRequest)
            .then((res) => {
                setDataSource(res.list)
                setTotal(res.page.total)
                return res
            })
            .finally(() => {
                setLoading(false)
            })
    }

    const handleOnChangeTabs = (key: string) => {
        console.log(key)
    }

    const buildTabsItems = () => {
        return alarmPageList.map((item) => {
            const { label, value, color, icon } = item
            return {
                label: label,
                key: `${value}`,
                icon: (
                    <Button
                        type="link"
                        icon={<IconFont type={icon} style={{ color: color }} />}
                    />
                )
            }
        })
    }

    const handleOptionClick = (action: ActionKey) => {
        switch (action) {
            case ActionKey.REFRESH:
                handleRefresh()
                break
        }
    }

    const handlerTableAction = (
        atcion: ActionKey,
        record: AlarmRealtimeItem
    ) => {
        switch (atcion) {
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

    useEffect(() => {
        handleGetAlarmRealtime()
    }, [refresh])

    useEffect(() => {
        handleGetAlarmPageList()
        handleRefresh()
    }, [])

    return (
        <div className="bodyContent">
            <RouteBreadcrumb />
            <HeightLine />
            <SearchForm form={queryForm} items={searchFormItems} />
            <HeightLine />
            <DataOption
                queryForm={queryForm}
                rightOptions={rightOptions}
                action={handleOptionClick}
            />
            <PaddingLine padding={12} height={1} borderRadius={4} />
            <Tabs items={buildTabsItems()} onChange={handleOnChangeTabs} />
            <DataTable
                showIndex={false}
                // showOperation={false}
                columns={columns}
                dataSource={dataSource}
                total={total}
                loading={loading}
                operationItems={operationItems}
                action={handlerTableAction}
                onRow={(record: AlarmRealtimeItem) => {
                    return {
                        style: {
                            background: record.level.color
                        }
                    }
                }}
            />
        </div>
    )
}

export default AlarmRealtime
