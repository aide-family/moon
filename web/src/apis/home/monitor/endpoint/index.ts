import { POST } from '@/apis/request'
import {
    AppendEndpointRequest,
    deleteEndpointRequest,
    ListEndpointRequest,
    SelectEndpointRequest,
    SelectEndpointResponse
} from '@/apis/home/monitor/endpoint/types.ts'

enum URL {
    APPEND = '/api/v1/endpoint/append',
    DELETE = '/api/v1/endpoint/delete',
    LIST = '/api/v1/endpoint/list',
    SELECT = '/api/v1/endpoint/select'
}

/** 增加数据源 */
const appendEndpoint = (data: AppendEndpointRequest) => {
    return POST(URL.APPEND, data)
}

/** 删除数据源 */
const deleteEndpoint = (data: deleteEndpointRequest) => {
    return POST(URL.DELETE, data)
}

/** 获取数据源列表 */
const listEndpoint = (data: ListEndpointRequest) => {
    return POST(URL.LIST, data)
}

const selectEndpoint = (data: SelectEndpointRequest) => {
    return POST<SelectEndpointResponse>(URL.SELECT, data)
}

/** 数据源端点接口导出 */
const endpointApi = {
    appendEndpoint,
    deleteEndpoint,
    listEndpoint,
    selectEndpoint
}

export default endpointApi
