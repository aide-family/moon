import { PageReqType, PageResType } from '@/apis/types'
import { DictSelectItem } from '@/apis/home/system/dict/types'

/** 报警历史明细 */
interface AlarmHistoryItem {
    id: number
    alarmId: number
    alarmName: string
    alarmLevel: DictSelectItem
    alarmStatus: string
    labels: Map<string, string>
    annotations: Map<string, string>
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
