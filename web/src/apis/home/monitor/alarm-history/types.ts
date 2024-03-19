import { AlarmStatus, Map, PageReqType, PageResType } from '@/apis/types'
import { DictSelectItem } from '@/apis/home/system/dict/types'
import dayjs from 'dayjs'

/** 报警历史明细 */
interface AlarmHistoryItem {
    id: number
    alarmId: number
    alarmName: string
    alarmLevel: DictSelectItem
    alarmStatus: string
    labels: Map
    annotations: Map
    startAt: number
    endAt: number
    duration: string
    expr: string
    datasource: string
}

/** 报警历史详情请求参数 */
interface AlarmHistoryDetailRequest {
    id: number
}

/** 报警历史详情响应 */
interface AlarmHistoryDetailReply {
    alarmHistory: AlarmHistoryItem
}

/** 报警历史列表请求参数 */
interface AlarmHistoryListRequest {
    page: PageReqType
    keyword?: string
    status?: AlarmStatus
    alarmPages?: number[]
    firingStartAt?: number
    firingEndAt?: number
    resolvedStartAt?: number
    resolvedEndAt?: number
    firingTime?: [string, string]
    resolvedTime?: [string, string]
    duration?: number
}

export const defaultAlarmHistoryListRequest: AlarmHistoryListRequest = {
    page: {
        curr: 1,
        size: 10
    },
    firingStartAt: dayjs().add(-30, 'day').unix(),
    firingEndAt: dayjs().unix(),
    status: AlarmStatus.ALARM_STATUS_UNKNOWN
}

/** 报警历史列表响应 */
interface AlarmHistoryListReply {
    page: PageResType
    list: AlarmHistoryItem[]
}

export type {
    AlarmHistoryItem,
    AlarmHistoryListRequest,
    AlarmHistoryListReply,
    AlarmHistoryDetailReply,
    AlarmHistoryDetailRequest
}
