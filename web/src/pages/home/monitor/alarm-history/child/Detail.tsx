import alarmHistoryApi from '@/apis/home/monitor/alarm-history'
import { AlarmHistoryItem } from '@/apis/home/monitor/alarm-history/types'
import { Map, Status, StatusMap } from '@/apis/types'
import {
    Badge,
    Button,
    Descriptions,
    DescriptionsProps,
    Drawer,
    DrawerProps,
    Typography
} from 'antd'
import dayjs from 'dayjs'
import React, { useEffect, useState } from 'react'

import { Detail } from '../../strategy/child/Detail'
import { ActionKey } from '@/apis/data'

const { Paragraph, Text } = Typography

export interface HistoryDetailProps extends DrawerProps {
    historyId?: number
}

export interface YmlCodeBoxProps {
    code: Map
    width?: number | string
}

export const keyValues: Map = {
    // __alert_id__: '策略ID',
    // __group_id__: '策略组ID',
    __group_name__: '策略组',
    // __level_id__: '告警等级ID',
    __name__: '指标',
    alertname: '策略',
    endpoint: '主机',
    env: '环境',
    instance: '实例',
    job: 'Job',
    sverity: '告警等级',
    description: '告警内容',
    summary: '告警摘要'
}

export const getKeyValue = (key: string) => {
    return keyValues[key] || null
}

export const CodeBox: React.FC<YmlCodeBoxProps> = ({ code, width }) => {
    return (
        <div style={{ width }}>
            {Object.keys(code).map((key) => {
                const value = code[key]
                const keyV = getKeyValue(key)
                if (!keyV) return null
                return (
                    <Paragraph
                        key={keyV}
                        ellipsis={{
                            rows: 2,
                            expandable: true,
                            tooltip: value
                            // symbol: 'more'
                        }}
                        style={{ width }}
                    >
                        <Text style={{ color: 'orangered' }}>{keyV}: </Text>
                        {value}
                    </Paragraph>
                )
            })}
        </div>
    )
}

const buildItems = (
    item?: AlarmHistoryItem,
    handler?: () => void
): DescriptionsProps['items'] => {
    if (!item) return []
    const {
        alarmName,
        alarmLevel,
        startAt,
        endAt,
        alarmStatus,
        annotations,
        labels,
        duration,
        id
    } = item
    const isResolved =
        alarmStatus === 'resolved'
            ? Status.STATUS_ENABLED
            : Status.STATUS_DISABLED
    const { color } = StatusMap[isResolved]
    let note: React.ReactNode = '-'
    const annotationsKeys = Object.keys(annotations)
    if (annotationsKeys.length > 0) {
        note = <CodeBox key={id + 'annotations'} code={annotations} />
    }
    let labelsBox: React.ReactNode = '-'
    const labelsKeys = Object.keys(labels)
    if (labelsKeys.length > 0) {
        labelsBox = <CodeBox key={id + 'labels'} code={labels} width={600} />
    }

    return [
        {
            label: '策略名称',
            span: { xs: 1, sm: 1, md: 1, lg: 1, xl: 3, xxl: 2 },
            key: 'alarmName',
            children: (
                <Button onClick={handler} type="link">
                    {alarmName}
                </Button>
            )
        },
        {
            label: '策略等级',
            span: { xs: 1, sm: 1, md: 1, lg: 1, xl: 1, xxl: 2 },
            key: 'alarmLevel',
            children: (
                <Badge
                    color={alarmLevel.color}
                    key={alarmLevel.value}
                    text={alarmLevel.label || '-'}
                />
            )
        },
        {
            label: '开始时间',
            key: 'startAt',
            span: { xs: 1, sm: 1, md: 1, lg: 1, xl: 2, xxl: 1 },
            children: dayjs(+startAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        },
        {
            label: '结束时间',
            key: 'endAt',
            span: { xs: 1, sm: 1, md: 1, lg: 1, xl: 2, xxl: 1 },
            children: dayjs(+endAt * 1000).format('YYYY-MM-DD HH:mm:ss')
        },
        {
            label: '持续时间',
            key: 'duration',
            span: { xs: 1, sm: 1, md: 1, lg: 1, xl: 2, xxl: 1 },
            children: duration
        },
        {
            label: '状态',
            key: 'alarmStatus',
            span: { xs: 1, sm: 1, md: 1, lg: 1, xl: 2, xxl: 1 },
            children: (
                <Badge color={color} key={alarmStatus} text={alarmStatus} />
            )
        },
        {
            label: '告警标签',
            key: 'labels',
            span: { xs: 1, sm: 1, md: 2, lg: 2, xl: 4, xxl: 4 },
            children: labelsBox
        },
        {
            label: '告警内容',
            key: 'annotations',
            span: { xs: 1, sm: 1, md: 2, lg: 2, xl: 4, xxl: 4 },
            children: note
        }
    ]
}

export const HistoryDetail: React.FC<HistoryDetailProps> = (props) => {
    const { historyId, open, placement = 'left', width = '80%' } = props

    const [detail, setDetail] = useState<AlarmHistoryItem>()
    const [openStrategyDetail, setOpenStrategyDetail] = useState(false)

    const fetchAlarmDetail = () => {
        if (!historyId) return
        alarmHistoryApi
            .getAlarmHistoryDetail({ id: historyId })
            .then(({ alarmHistory }) => {
                setDetail(alarmHistory)
            })
    }

    useEffect(() => {
        if (!open) return
        fetchAlarmDetail()
    }, [historyId, open])

    return (
        <>
            <Detail
                open={openStrategyDetail}
                disabled
                id={detail?.alarmId}
                onClose={() => setOpenStrategyDetail(false)}
                actionKey={ActionKey.DETAIL}
            />
            <Drawer
                title={`${detail?.alarmName} 历史告警详情`}
                {...props}
                open={open}
                placement={placement}
                width={width}
                closeIcon={null}
            >
                <Descriptions
                    bordered
                    column={{ xs: 1, sm: 1, md: 2, lg: 2, xl: 4, xxl: 4 }}
                    items={buildItems(detail, () =>
                        setOpenStrategyDetail(true)
                    )}
                />
            </Drawer>
        </>
    )
}
