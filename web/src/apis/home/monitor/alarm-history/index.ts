import { POST } from '@/apis/request'
import type {
    AlarmHistoryListRequest,
    AlarmHistoryListReply,
    AlarmHistoryDetailReply,
    AlarmHistoryDetailRequest
} from './types'

enum URL {
    /** 获取报警历史列表 */
    LIST = '/api/v1/alarm/history/list',
    /** 获取报警历史详情 */
    DETAIL = '/api/v1/alarm/history/get'
}

/** 报警历史详情获取 */
const getAlarmHistoryDetail = (params: AlarmHistoryDetailRequest) => {
    return POST<AlarmHistoryDetailReply>(URL.DETAIL, params)
}

/** 报警历史列表获取 */
const getAlarmHistoryList = (params: AlarmHistoryListRequest) => {
    return POST<AlarmHistoryListReply>(URL.LIST, params)
}

/** 报警历史接口 */
const alarmHistoryApi = {
    getAlarmHistoryDetail,
    getAlarmHistoryList
}

export default alarmHistoryApi
