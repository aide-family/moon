import { POST } from '@/apis/request'
import type {
    BindMyAlarmPagesRequest,
    CountAlarmPageReply,
    CountAlarmPageRequest,
    CreateDict,
    DictBatchUpdateStatusType,
    DictById,
    DictDetailRes,
    DictListReq,
    DictListRes,
    DictSelectReq,
    DictSelectRes,
    MyAlarmPageListResponse,
    UpdateDict,
    dictBatchDeleteType
} from './types'

/** URL枚举 */
enum URL {
    /** 创建字典 */
    dictCreate = '/api/v1/dict/create',
    /** 字典详情 */
    dictDetail = '/api/v1/dict/get',
    /** 字典列表 */
    dictList = '/api/v1/dict/list',
    /** 字典删除 */
    dictDelete = '/api/v1/dict/delete',
    /** 字典更新 */
    dictUpdate = '/api/v1/dict/update',
    /** 字典下拉列表 */
    dictSelect = '/api/v1/dict/select',
    /** 批量删除字典 */
    dictBatchDelete = '/api/v1/dict/batch/delete',
    /** 批量更新字典状态 */
    dictBatchUpdateStatus = 'api/v1/dict/status/update/batch',
    COUNT_ALARM_PAGE = '/api/v1/alarm_page/alarm/count',
    /** 我的告警页面列表 */
    LIST_MY_ALARM_PAGE = '/api/v1/alarm_page/my/list',
    /** 我的告警页面列表配置 */
    CONFIG_MY_ALARM_PAGE = '/api/v1/alarm_page/my/list/config'
}

/** 创建字典 */
const dictCreate = (data: CreateDict) => {
    return POST<DictById>(URL.dictCreate, data)
}
/** 字典详情 */
const dictDetail = (data: DictById) => {
    return POST<DictDetailRes>(URL.dictDetail, data)
}
/** 字典列表 */
const dictList = (data: DictListReq) => {
    return POST<DictListRes>(URL.dictList, data)
}
/** 字典删除 */
const dictDelete = (data: DictById) => {
    return POST<DictById>(URL.dictDelete, data)
}

/** 字典更新 */
const dictUpdate = (data: UpdateDict) => {
    return POST<DictById>(URL.dictUpdate, data)
}
/** 字典下拉列表 */
const dictSelect = (data: DictSelectReq) => {
    return POST<DictSelectRes>(URL.dictSelect, data)
}
const dictBatchDelete = (data: dictBatchDeleteType) => {
    return POST<dictBatchDeleteType>(URL.dictBatchDelete, data)
}
/** 批量更新字典状态 */
const dictBatchUpdateStatus = (data: DictBatchUpdateStatusType) => {
    return POST<DictBatchUpdateStatusType>(URL.dictBatchUpdateStatus, data)
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

/** 字典接口 */
export const dictApi = {
    /** 创建字典 */
    dictCreate,
    /** 字典详情 */
    dictDetail,
    /** 字典列表 */
    dictList,
    /** 字典删除 */
    dictDelete,
    /** 字典更新 */
    dictUpdate,
    /** 字典下拉列表 */
    dictSelect,
    /** 批量删除字典 */
    dictBatchDelete,
    /** 批量更新字典状态 */
    dictBatchUpdateStatus,
    countAlarmPage,
    myAlarmPageList,
    myAlarmPageConfig
}
export default dictApi
