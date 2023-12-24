import { POST } from '@/apis/request'
import type {
    SelectAlarmPageRequest,
    SelectAlarmPageReply,
    BatchChangeAlarmPageStatusRequest,
    BatchChangeAlarmPageStatusReply,
    CreateAlarmPageReply,
    CreateAlarmPageRequest,
    UpdateAlarmPageRequest,
    UpdateAlarmPageReply,
    DeleteAlarmPageRequest,
    DeleteAlarmPageReply,
    GetAlarmPageReply,
    GetAlarmPageRequest,
    BacthDeleteAlarmPageReply,
    BacthDeleteAlarmPageRequest,
    ListAlarmPageReply,
    ListAlarmPageRequest
} from './types'

enum URL {
    BATCH_DELETE = '/api/v1/alarm_page/batch/delete',
    BATCH_STATUS = '/api/v1/alarm_page/status/batch/update',
    CREATE = '/api/v1/alarm_page/create',
    DELETE = '/api/v1/alarm_page/delete',
    GET = '/api/v1/alarm_page/get',
    UPDATE = '/api/v1/alarm_page/update',
    LIST = '/api/v1/alarm_page/list',
    SELECT = '/api/v1/alarm_page/select'
}

const getAlarmPageDetail = (params: GetAlarmPageRequest) => {
    return POST<GetAlarmPageReply>(URL.GET, params)
}

const getAlarmPageList = (params: ListAlarmPageRequest) => {
    return POST<ListAlarmPageReply>(URL.LIST, params)
}

const getAlarmPageSelect = (params: SelectAlarmPageRequest) => {
    return POST<SelectAlarmPageReply>(URL.SELECT, params)
}

const createAlarmPage = (params: CreateAlarmPageRequest) => {
    return POST<CreateAlarmPageReply>(URL.CREATE, params)
}

const updateAlarmPage = (params: UpdateAlarmPageRequest) => {
    return POST<UpdateAlarmPageReply>(URL.UPDATE, params)
}

const deleteAlarmPage = (params: DeleteAlarmPageRequest) => {
    return POST<DeleteAlarmPageReply>(URL.DELETE, params)
}

const batchDeleteAlarmPage = (params: BacthDeleteAlarmPageRequest) => {
    return POST<BacthDeleteAlarmPageReply>(URL.BATCH_DELETE, params)
}

const batchUpdateAlarmPageStatus = (
    params: BatchChangeAlarmPageStatusRequest
) => {
    return POST<BatchChangeAlarmPageStatusReply>(URL.BATCH_STATUS, params)
}

const alarmPageApi = {
    batchDeleteAlarmPage,
    batchUpdateAlarmPageStatus,
    getAlarmPageDetail,
    getAlarmPageList,
    createAlarmPage,
    updateAlarmPage,
    deleteAlarmPage,
    getAlarmPageSelect
}

export default alarmPageApi
