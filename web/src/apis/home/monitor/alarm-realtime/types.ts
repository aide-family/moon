import { PageReqType, PageResType, AlarmStatus } from '@/apis/types'
import { DictSelectItem } from '../../system/dict/types'
import { UserSelectItem } from '../../system/user/types'

interface InterveneItem {
    intervenedUser: {
        value: number
        label: string
        status: number
        avatar: string
        nickname: string
        gender: number
    }
    intervenedAt: number
    remark: string
    id: number
}

interface BeenNotifyMember {
    memberId: number
    notifyTypes: number[]
    user?: UserSelectItem
    status: number
    id: number
}

interface UpgradeItem {
    upgradedUser?: UserSelectItem
    upgradedAt: number
    remark: string
    id: number
}

interface SuppressItem {
    suppressedUser?: UserSelectItem
    suppressedAt: number
    remark: string
    duration: number
    id: number
}

interface NotifiedChatGroup {
    value: number
    app: number
    label: string
    status: number
}

interface AlarmRealtimeItem {
    id: number
    instance: string
    note: string
    levelId: number
    eventAt: number
    status: number
    pageIds: number[]
    intervenedUser: InterveneItem[]
    beenNotifyMembers: BeenNotifyMember[]
    notifiedAt: number
    historyId: number
    upgradedUser: UpgradeItem
    suppressedUser: SuppressItem
    strategyId: number
    notifiedChatGroups: NotifiedChatGroup[]
    createdAt: number
    updatedAt: number
    level: DictSelectItem
}

interface AlarmRealtimeDetailRequest {
    id: number
}

interface AlarmRealtimeDetailResponse {
    detail: AlarmRealtimeItem
}

interface AlarmRealtimeInterveneRequest {
    id: number
    remark: string
}

interface AlarmRealtimeInterveneResponse {}

interface AlarmRealtimeListRequest {
    page: PageReqType
    keyword: string
    alarmPages: number[]
    startAt: number
    endAt: number
}

interface AlarmRealtimeListResponse {
    list: AlarmRealtimeItem[]
    page: PageResType
}

interface AlarmRealtimeSuppressRequest {
    id: number
    remark: string
    duration: number
}

interface AlarmRealtimeSuppressResponse {}

interface AlarmRealtimeUpgradeRequest {
    id: number
    remark: string
}

interface AlarmRealtimeUpgradeResponse {}

export type {
    AlarmRealtimeItem,
    AlarmRealtimeDetailRequest,
    AlarmRealtimeDetailResponse,
    AlarmRealtimeInterveneRequest,
    AlarmRealtimeInterveneResponse,
    AlarmRealtimeListResponse,
    AlarmRealtimeListRequest,
    AlarmRealtimeSuppressRequest,
    AlarmRealtimeSuppressResponse,
    AlarmRealtimeUpgradeRequest,
    AlarmRealtimeUpgradeResponse
}
