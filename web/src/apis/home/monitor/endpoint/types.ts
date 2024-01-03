import { PageReqType, PageRes } from '@/apis/types'

interface PrometheusServerItem {
    id: number
    name: string
    endpoint: string
    status: number
    remark: string
    createdAt: number
    updatedAt: number
    agentEndpoint: string
    agentCheck: string
    uuid: string
}

interface PrometheusServerSelectItem {
    value: number
    label: string
    status: number
    remark: string
    endpoint: string
}

interface AppendEndpointRequest {
    agentName: string
    endpoints: PrometheusServerItem[]
}

interface deleteEndpointRequest {}

interface ListEndpointRequest {}

interface ListEndpointResponse {
    list: PrometheusServerItem[]
    page: PageReqType
}

interface SelectEndpointRequest {
    page: PageReqType
    keyword?: string
}

interface SelectEndpointResponse {
    list: PrometheusServerSelectItem[]
    page: PageRes
}

export type {
    PrometheusServerItem,
    AppendEndpointRequest,
    ListEndpointRequest,
    deleteEndpointRequest,
    ListEndpointResponse,
    SelectEndpointResponse,
    SelectEndpointRequest,
    PrometheusServerSelectItem
}
