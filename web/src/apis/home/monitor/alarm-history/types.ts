import { Map, PageReqType, PageResType } from '@/apis/types'
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
    statusList?: number[]
    alarmPages?: number[]
    startAt?: number
    endAt?: number
}

export const defaultAlarmHistoryListRequest: AlarmHistoryListRequest = {
    page: {
        curr: 1,
        size: 10
    },
    startAt: dayjs().add(-30, 'day').unix(),
    endAt: dayjs().unix()
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
