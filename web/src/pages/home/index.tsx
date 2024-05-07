import React, {FC, useContext, useEffect, useState} from 'react'

import {HeightLine, PaddingLine} from '@/components/HeightLine'
import RouteBreadcrumb from '@/components/PromLayout/RouteBreadcrumb'
import {Card, Col, Collapse, Layout, Row, Space, Tabs, Tag, theme} from 'antd'
import {DataOption} from '@/components/Data'
import {leftOptions, rightOptions} from './options'
import {ChartItem, DashboardConfigItem, defaultListDashboardRequest} from '@/apis/home/dashboard/types'
import dashboardApi from '@/apis/home/dashboard'
import {GlobalContext} from '@/context'
import {ActionKey} from '@/apis/data'
import {ConfigDashboardChartModal} from './child/ConfigDashboardChart'
import {ChartType} from "@/apis/types.ts";

const {Content} = Layout
const {useToken} = theme

const Box: React.FC<ChartItem> = (props) => {
    const {title, remark, url, chartType, width = '540px', height = '350px'} = props
    const {token} = useToken()
    const {sysTheme} = useContext(GlobalContext)
    const collapseItem = {
        key: '1',
        label: title,
        children: <p>{remark}</p>
    }
    let _width_ = width
    let _height_ = height
    switch (chartType) {
        case ChartType.ChartTypeFull:
            _width_ = '100%'
            _height_ = document.documentElement.clientHeight - document.getElementById('footer')!.clientHeight - 100 + 'px'
            return <Content
                style={{borderColor: token.colorBorderBg, background: token.colorBgContainer, padding: 0, margin: 0}}>
                <Card hoverable style={{
                    width: _width_,
                    height: _height_
                }}>
                    <iframe
                        src={`${url}&theme=${sysTheme}`}
                        width={_width_}
                        height={_height_}
                        frameBorder="0"
                    ></iframe>
                    <Collapse size="small" items={[collapseItem]} bordered={false}/>
                </Card>
            </Content>
        case ChartType.ChartTypeRow:
            _width_ = '100%'
            return <Row>
                <Col span={24} style={{
                    width: _width_,
                    height: _height_
                }}>
                    <Card hoverable>
                        <iframe
                            src={`${url}&theme=${sysTheme}`}
                            width={_width_}
                            height={_height_}
                            frameBorder="0"
                        ></iframe>
                        <Collapse size="small" items={[collapseItem]} bordered={false}/>
                    </Card>
                </Col>
            </Row>
    }
    return (
        <Card hoverable>
            <iframe
                src={`${url}&theme=${sysTheme}`}
                width={width}
                height={height}
                frameBorder="0"
            ></iframe>
            <Collapse size="small" items={[collapseItem]} bordered={false}/>
        </Card>
    )
}

const renderBox = (item: ChartItem, key: any): React.ReactNode => {
    // 如果是数组
    if (Array.isArray(item)) {
        return (
            <Space size={[8, 8]} wrap>
                {item.map((item, index) => renderBox(item, `${key}-${index}`))}
            </Space>
        )
    }

    return <Box {...item} key={key}/>
}

let autoRefreshTimer: NodeJS.Timeout | null = null

const Home: FC = () => {
    const {size, autoRefresh, setAutoRefresh} = useContext(GlobalContext)
    const [dashboards, setDashboards] = useState<ChartItem[]>([])
    const [dashboardList, setDashboardList] = useState<DashboardConfigItem[]>(
        []
    )
    const [activeKey, setActiveKey] = useState<string | number>(0)
    const [refresh, setRefresh] = useState<boolean>(false)
    const [oepnConfigModal, setOpenConfigModal] = useState<boolean>(false)

    const handleOpenConfigModal = () => {
        setOpenConfigModal(true)
    }

    const handleCloseConfigModal = () => {
        setOpenConfigModal(false)
    }

    const handleOnOk = () => {
        handleCloseConfigModal()
        handleRefresh()
    }

    const handleGetDashboardDetail = (dashboardId: string | number) => {
        if (!dashboardId) return
        dashboardApi.getDashboardDetail(+dashboardId).then(({detail}) => {
            if (!detail) return
            setDashboards(detail.charts || [])
        })
    }

    const handleRefresh = () => {
        setRefresh(!refresh)
    }

    const handleGetDashboards = () => {
        dashboardApi
            .getDashboardList(defaultListDashboardRequest)
            .then(({list}) => {
                if (!list) return
                setDashboardList(list)
                if (!activeKey) {
                    setActiveKey(`${list[0]?.id}`)
                }
            })
    }

    const buildTabsItems = () => {
        return dashboardList.map((item, index) => {
            const {title, id, color} = item
            return {
                label: (
                    <Tag color={color || '#1677ff'}>
                        {title || `报警页面${index}`}
                    </Tag>
                ),
                key: `${id}`
            }
        })
    }

    const handleAutoRefresh = () => {
        const isAutoRefresh = !!autoRefresh
        if (autoRefreshTimer) {
            clearInterval(autoRefreshTimer)
        }
        if (isAutoRefresh) {
            autoRefreshTimer = setInterval(() => {
                handleRefresh()
                // 1min
            }, 1000 * 10 * 6)
        }
    }

    const handleAction = (action: ActionKey) => {
        switch (action) {
            case ActionKey.REFRESH:
                handleRefresh()
                break
            case ActionKey.CONFIG_DASHBOARD_CHART:
                handleOpenConfigModal()
                break
            case ActionKey.AUTO_REFRESH:
                setAutoRefresh?.(!autoRefresh)
                handleAutoRefresh()
                break
        }
    }

    useEffect(() => {
        if (activeKey) {
            handleGetDashboardDetail(+activeKey)
        }
    }, [activeKey])

    useEffect(() => {
        handleGetDashboards()
        handleGetDashboardDetail(+activeKey)
    }, [refresh])

    useEffect(() => {
        handleGetDashboards()
        handleAutoRefresh()
    }, [])

    return (
        <div style={{height: '100%', overflowY: 'auto'}}>
            <ConfigDashboardChartModal
                dashboardId={+activeKey}
                open={oepnConfigModal}
                onCancel={handleCloseConfigModal}
                onOk={handleOnOk}
            />
            <RouteBreadcrumb/>
            <HeightLine/>
            <DataOption
                showAdd={false}
                showClear={false}
                action={handleAction}
                // showSegmented={false}
                rightOptions={rightOptions(autoRefresh)}
                leftOptions={leftOptions(handleRefresh)}
            />
            <PaddingLine/>
            <Tabs
                items={buildTabsItems()}
                onChange={setActiveKey}
                activeKey={`${activeKey}`}
                tabBarGutter={8}
                tabBarStyle={{
                    boxShadow: '0px 0px 10px 0px rgba(0,0,0,0.1)'
                }}
                defaultActiveKey={`${dashboardList[0]?.id}`}
                size={size}
            />
            <div id='dashboard'>
                {dashboards.filter((item) => item.chartType === ChartType.ChartTypeFull).map((item, index: number) => {
                    return renderBox(item, index)
                })}
                {dashboards.filter((item) => item.chartType === ChartType.ChartTypeRow).map((item, index: number) => {
                    return renderBox(item, index)
                })}
                <Space size={[8, 8]} style={{width: '100%'}} wrap>
                    {dashboards.filter((item) => item.chartType !== ChartType.ChartTypeFull && item.chartType !== ChartType.ChartTypeRow).map((item, index: number) => {
                        return renderBox(item, index)
                    })}
                </Space>
            </div>
        </div>
    )
}

export default Home
