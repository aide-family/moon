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

interface AppendEndpointRequest {
    agentName: string
    endpoints: PrometheusServerItem[]
}

interface deleteEndpointRequest {
}

interface ListEndpointRequest {
}

export type {
    PrometheusServerItem,
    AppendEndpointRequest,
    ListEndpointRequest,
    deleteEndpointRequest
}