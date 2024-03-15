import React, { useEffect } from 'react'
import { Alert, Button, List, Modal, Space, Tabs } from 'antd'
import dayjs from 'dayjs'
import SearchForm, { DateDataType } from '@/components/Prom/SearchForm'
import AreaStackGradient from '@/components/Prom/charts/area-stack-gradient/AreaStackGradient'
import { promData } from '@/components/Prom/charts/area-stack-gradient/option'
import { UnorderedListOutlined, AreaChartOutlined } from '@ant-design/icons'
import BaseText from '../Text/BaseText'
import { Duration } from '@/apis/types'

import weekday from 'dayjs/plugin/weekday'
import localeData from 'dayjs/plugin/localeData'

dayjs.extend(weekday)
dayjs.extend(localeData)

export interface PromValueModalProps {
    visible: boolean
    onCancel: () => void
    pathPrefix: string
    apiPath?: string
    expr?: string
    height?: number | string
    eventAt?: number
    duration?: Duration
    endAt?: number
}

export interface PromValue {
    metric: {
        __name__: string
        instance: string
        [key: string]: string
    }
    value?: [number, string]
    values?: [number, string][]
}

const PromValueModal: React.FC<PromValueModalProps> = (props) => {
    const {
        visible,
        onCancel,
        pathPrefix,
        apiPath = 'api/v1',
        expr,
        height = 400,
        endAt,
        eventAt = dayjs().unix(),
        duration = {
            value: 1,
            unit: 'h'
        }
    } = props

    const [startTime, setStartTime] = React.useState<number>(
        eventAt || dayjs().unix()
    )
    const [endTime, setEndTime] = React.useState<number>(
        endAt || eventAt || dayjs().unix()
    )
    const [resolution, setResolution] = React.useState<number>(14)
    const [data, setData] = React.useState<PromValue[]>([])
    const [tabKey, setTabKey] = React.useState<string>('table')
    const [loading, setLoading] = React.useState<boolean>(false)
    const [dateType, setDateType] = React.useState<DateDataType>('date')
    const [err, setErr] = React.useState<string | undefined>()
    const [datasource, setDatasource] = React.useState<promData[]>([])

    const handleTabChange = (key: string) => {
        setTabKey(key)
        switch (key) {
            case 'graph':
                setEndTime(endAt ?? dayjs().unix())
                setStartTime(
                    endAt
                        ? eventAt
                        : dayjs(eventAt * 1000)
                              .subtract(
                                  duration?.value ? +duration?.value : 1,
                                  duration.unit as any
                              )
                              .unix()
                )
                setDateType('range')
                break
            case 'table':
                setEndTime(eventAt)
                setStartTime(eventAt)
                setDateType('date')
                break
        }
    }

    const fetchValues = async (tabKey: string) => {
        if (!expr) return
        let path = ''
        const abortController = new AbortController()
        const params: URLSearchParams = new URLSearchParams({
            query: expr
        })
        switch (tabKey) {
            case 'graph':
                path = 'query_range'
                params.append('start', startTime.toString())
                params.append('end', endTime.toString())
                params.append('step', resolution.toString())
                break
            case 'table':
                path = 'query'
                params.append('time', endTime.toString())
                break
            default:
                throw new Error('Invalid panel type "' + tabKey + '"')
        }

        setLoading(true)
        const query = await fetch(
            `${pathPrefix}/${apiPath}/${path}?${params}`,
            {
                cache: 'no-store',
                credentials: 'same-origin',
                signal: abortController.signal
            }
        )
            .then((resp) => resp.json())
            .catch((err) => {
                console.log('err', err)
            })
            .finally(() => setLoading(false))

        if (query.status !== 'success') {
            setErr(query.error)
            return
        }

        setDatasource(query.data.result)
        setErr(undefined)

        if (query.data) {
            const { result } = query.data
            setData([...result])
        }
    }

    const handleSearch = (
        type: DateDataType,
        values: number | [number, number],
        step?: number
    ) => {
        switch (type) {
            case 'date':
                setEndTime(values as number)
                break
            case 'range':
                setStartTime((values as [number, number])[0])
                setEndTime((values as [number, number])[1])
                setResolution(step || 14)
                break
        }
    }

    useEffect(() => {
        if (!visible || !expr) return
        fetchValues(tabKey).then()
    }, [expr, visible, tabKey, startTime, endTime, resolution])

    const getValues = (val: PromValue) => {
        if (val.value && !Array.isArray(val.value[1])) {
            return val.value[1]
        }

        if (
            val.values &&
            val.values.length > 0 &&
            Array.isArray(val.values[0]) &&
            val.values[0].length === 2 &&
            !Array.isArray(val.values[0][1])
        ) {
            return val.values[0][1]
        }

        return ''
    }

    const renderGraph = (metricValues?: promData[]) => {
        if (!metricValues) return null
        return (
            <AreaStackGradient
                height={height}
                data={metricValues}
                id="prom-metric-data-chart"
            />
        )
    }

    return (
        <Modal
            open={visible}
            onCancel={onCancel}
            width="80%"
            footer={null}
            destroyOnClose
        >
            <SearchForm
                type={dateType}
                onSearch={handleSearch}
                eventAt={eventAt}
                duration={duration}
            />
            <Tabs
                direction="ltr"
                onChange={handleTabChange}
                defaultActiveKey="table"
                items={[
                    {
                        key: 'table',
                        label: (
                            <Button type="text" disabled={false}>
                                <UnorderedListOutlined />
                                指标列表
                            </Button>
                        ),
                        children: (
                            <>
                                {err && (
                                    <Alert
                                        closable
                                        type="error"
                                        message={err}
                                    />
                                )}
                                <List
                                    loading={loading}
                                    style={{
                                        height: height,
                                        overflowY: 'auto'
                                    }}
                                    dataSource={data}
                                    renderItem={(
                                        item: PromValue,
                                        index: React.Key
                                    ) => {
                                        return (
                                            <List.Item
                                                key={index}
                                                id={`list-${index}`}
                                            >
                                                <Space
                                                    direction="horizontal"
                                                    style={{ width: '100%' }}
                                                >
                                                    <BaseText
                                                        maxLine={2}
                                                        showTooltip
                                                        copy
                                                    >
                                                        {item?.metric
                                                            ? `${
                                                                  item?.metric
                                                                      ?.__name__ ||
                                                                  ''
                                                              }{${Object.keys(
                                                                  item.metric
                                                              )
                                                                  .filter(
                                                                      (key) =>
                                                                          key !==
                                                                              '__name__' &&
                                                                          key !==
                                                                              'id'
                                                                  )
                                                                  .map(
                                                                      (key) =>
                                                                          `${key}="${item.metric[key]}"`
                                                                  )
                                                                  .join(', ')}}`
                                                            : expr || ''}
                                                    </BaseText>
                                                    <div
                                                        style={{
                                                            float: 'right'
                                                        }}
                                                    >
                                                        {getValues(item)}
                                                    </div>
                                                </Space>
                                            </List.Item>
                                        )
                                    }}
                                />
                            </>
                        )
                    },
                    {
                        key: 'graph',
                        label: (
                            <Button type="text" disabled={false}>
                                <AreaChartOutlined />
                                指标图表
                            </Button>
                        ),
                        children: (
                            <div
                                style={{
                                    height: height,
                                    overflowY: 'auto',
                                    overflowX: 'hidden'
                                }}
                            >
                                {err ? (
                                    <Alert
                                        closable
                                        type="error"
                                        message={err}
                                    />
                                ) : (
                                    renderGraph([...datasource])
                                )}
                            </div>
                        )
                    }
                ]}
            />
        </Modal>
    )
}

export default PromValueModal
