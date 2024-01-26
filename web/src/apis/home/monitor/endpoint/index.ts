import { POST } from '@/apis/request'
import {
    AppendEndpointRequest,
    deleteEndpointRequest,
    ListEndpointRequest,
    ListEndpointResponse,
    PrometheusServerItem,
    SelectEndpointRequest,
    SelectEndpointResponse,
    UpdateEndpointRequest
} from '@/apis/home/monitor/endpoint/types.ts'
import { Status } from '@/apis/types'

enum URL {
    APPEND = '/api/v1/endpoint/append',
    DELETE = '/api/v1/endpoint/delete',
    LIST = '/api/v1/endpoint/list',
    SELECT = '/api/v1/endpoint/select',
    DETAIL = '/api/v1/endpoint/detail',
    EDIT = '/api/v1/endpoint/edit',
    BATCH_CHANGE_STATUS = '/api/v1/endpoint/batch/status'
}

/** 增加数据源 */
const appendEndpoint = (data: AppendEndpointRequest) => {
    return POST<{ id: number }>(URL.APPEND, data)
}

/** 删除数据源 */
const deleteEndpoint = (data: deleteEndpointRequest) => {
    return POST<{ id: number }>(URL.DELETE, data)
}

/** 获取数据源列表 */
const listEndpoint = (data: ListEndpointRequest) => {
    return POST<ListEndpointResponse>(URL.LIST, data)
}

const selectEndpoint = (data: SelectEndpointRequest) => {
    return POST<SelectEndpointResponse>(URL.SELECT, data)
}

const detailEndpoint = (id: number) => {
    return POST<{ detail: PrometheusServerItem }>(URL.DETAIL, { id })
}

const editEndpoint = (data: UpdateEndpointRequest) => {
    return POST<{ id: number }>(URL.EDIT, data)
}

const batchChangeStatus = (ids: number[], status: Status) => {
    return POST<{ ids: number[] }>(URL.BATCH_CHANGE_STATUS, {
        ids,
        status
    })
}

/** 数据源端点接口导出 */
const endpointApi = {
    appendEndpoint,
    deleteEndpoint,
    listEndpoint,
    selectEndpoint,
    detailEndpoint,
    editEndpoint,
    batchChangeStatus
}

export default endpointApi
