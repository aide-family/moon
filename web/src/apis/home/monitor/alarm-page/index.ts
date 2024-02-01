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
    ListAlarmPageRequest,
    CountAlarmPageReply,
    CountAlarmPageRequest,
    MyAlarmPageListResponse,
    BindMyAlarmPagesRequest
} from './types'

enum URL {
    BATCH_DELETE = '/api/v1/alarm_page/batch/delete',
    BATCH_STATUS = '/api/v1/alarm_page/status/batch/update',
    CREATE = '/api/v1/alarm_page/create',
    DELETE = '/api/v1/alarm_page/delete',
    GET = '/api/v1/alarm_page/get',
    UPDATE = '/api/v1/alarm_page/update',
    LIST = '/api/v1/alarm_page/list',
    SELECT = '/api/v1/alarm_page/select',
    COUNT_ALARM_PAGE = '/api/v1/alarm_page/alarm/count',
    /** 我的告警页面列表 */
    LIST_MY_ALARM_PAGE = '/api/v1/alarm_page/my/list',
    /** 我的告警页面列表配置 */
    CONFIG_MY_ALARM_PAGE = '/api/v1/alarm_page/my/list/config'
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

const countAlarmPage = (params: CountAlarmPageRequest) => {
    return POST<CountAlarmPageReply>(URL.COUNT_ALARM_PAGE, params)
}

const myAlarmPageList = () => {
    return POST<MyAlarmPageListResponse>(URL.LIST_MY_ALARM_PAGE, {})
}

const myAlarmPageConfig = (params: BindMyAlarmPagesRequest) => {
    return POST<{}>(URL.CONFIG_MY_ALARM_PAGE, params)
}

const alarmPageApi = {
    batchDeleteAlarmPage,
    batchUpdateAlarmPageStatus,
    getAlarmPageDetail,
    getAlarmPageList,
    createAlarmPage,
    updateAlarmPage,
    deleteAlarmPage,
    getAlarmPageSelect,
    countAlarmPage,
    myAlarmPageList,
    myAlarmPageConfig
}

export default alarmPageApi
