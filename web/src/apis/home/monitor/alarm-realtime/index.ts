import { POST } from '@/apis/request'

import type {
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
} from './types'

enum URL {
    DETAIL = '',
    LIST = '',
    INTERVENE = '',
    SUPPRESS = '',
    UPGRADE = ''
}

const getAlarmRealtimeDetail = (params: AlarmRealtimeDetailRequest) => {
    return POST<AlarmRealtimeDetailResponse>(URL.DETAIL, params)
}

const getAlarmRealtimeList = (params: AlarmRealtimeListRequest) => {
    return POST<AlarmRealtimeListResponse>(URL.LIST, params)
}

const alarmRealtimeIntervene = (params: AlarmRealtimeInterveneRequest) => {
    return POST<AlarmRealtimeInterveneResponse>(URL.INTERVENE, params)
}

const alarmRealtimeSuppress = (params: AlarmRealtimeSuppressRequest) => {
    return POST<AlarmRealtimeSuppressResponse>(URL.SUPPRESS, params)
}

const alarmRealtimeUpgrade = (params: AlarmRealtimeUpgradeRequest) => {
    return POST<AlarmRealtimeUpgradeResponse>(URL.UPGRADE, params)
}

const alarmRealtimeApi = {
    getAlarmRealtimeDetail,
    getAlarmRealtimeList,
    alarmRealtimeIntervene,
    alarmRealtimeSuppress,
    alarmRealtimeUpgrade
}

export default alarmRealtimeApi
