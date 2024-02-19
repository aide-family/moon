import { PageReqType, PageRes, PageResType, Status } from '@/apis/types'

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
    status: Status
    remark: string
    endpoint: string
}

interface AppendEndpointRequest {
    name: string
    endpoint: string
    remark: string
}

interface deleteEndpointRequest {}

interface ListEndpointRequest {
    page: PageReqType
    keyword?: string
    status?: Status
}

interface ListEndpointResponse {
    list?: PrometheusServerItem[]
    page: PageResType
}

interface SelectEndpointRequest {
    page: PageReqType
    keyword?: string
    status?: Status
}

interface SelectEndpointResponse {
    list: PrometheusServerSelectItem[]
    page: PageRes
}

interface UpdateEndpointRequest {
    id: number
    name: string
    endpoint: string
    remark: string
}

export const defaultListEndpointRequest: ListEndpointRequest = {
    page: {
        curr: 1,
        size: 10
    }
}

export type {
    PrometheusServerItem,
    AppendEndpointRequest,
    ListEndpointRequest,
    deleteEndpointRequest,
    ListEndpointResponse,
    SelectEndpointResponse,
    SelectEndpointRequest,
    PrometheusServerSelectItem,
    UpdateEndpointRequest
}
