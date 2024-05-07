import {ChartType, PageReqType, PageResType, Status} from '@/apis/types'

interface ChartItem {
    title: string
    remark: string
    url: string
    id: number
    status: number
    chartType: ChartType
    width: string
    height: string
}

interface DashboardConfigItem {
    id: number
    title: string
    remark: string
    createdAt: number | string
    updatedAt: number | string
    deletedAt: number | string
    color: string
    charts: ChartItem[]
}

interface DashboardConfigOptionItem {
    value: number
    label: string
    color: string
}

interface CreateDashboardRequest {
    title: string
    color: string
    remark: string
    chartIds: number[]
}

interface UpdateDashboardRequest {
    id: number
    title: string
    color: string
    remark: string
    chartIds: number[]
}

interface SelectDashboardOptionRequest {
    page: PageReqType
    status?: Status
    keyword?: string
}

interface SelectDashboardOptionResponse {
    list?: DashboardConfigOptionItem[]
    page: PageResType
}

interface ListDashboardRequest {
    page: PageReqType
    status?: Status
    keyword?: string
}

interface ListDashboardResponse {
    list?: DashboardConfigItem[]
    page: PageResType
}

interface GetDashboardResponse {
    detail?: DashboardConfigItem
}

interface CreateChartRequest {
    title: string
    remark: string
    url: string
}

interface UpdateChartRequest {
    id: number
    title: string
    remark: string
    url: string
}

interface GetChartResponse {
    detail?: ChartItem
}

interface ListChartRequest {
    page: PageReqType
    status?: Status
    keyword?: string
}

interface ListChartResponse {
    list?: ChartItem[]
    page: PageResType
}

export const defaultSelectDashboardOptionRequest: SelectDashboardOptionRequest =
    {
        page: {
            curr: 1,
            size: 10
        },
        status: Status.STATUS_ENABLED
    }

export const defaultListDashboardRequest: ListDashboardRequest = {
    page: {
        curr: 1,
        size: 10
    },
    status: Status.STATUS_ENABLED
}

export const defaultListChartRequest: ListChartRequest = {
    page: {
        curr: 1,
        size: 200
    },
    status: Status.STATUS_ENABLED
}

export type {
    ChartItem,
    DashboardConfigItem,
    DashboardConfigOptionItem,
    CreateDashboardRequest,
    UpdateDashboardRequest,
    SelectDashboardOptionRequest,
    ListDashboardRequest,
    CreateChartRequest,
    UpdateChartRequest,
    GetChartResponse,
    ListChartRequest,
    ListChartResponse,
    GetDashboardResponse,
    SelectDashboardOptionResponse,
    ListDashboardResponse
}
