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
    DETAIL = '/api/v1/alarm/realtime/detail',
    LIST = '/api/v1/alarm/realtime/list',
    INTERVENE = '/api/v1/alarm/realtime/intervene',
    SUPPRESS = '/api/v1/alarm/realtime/suppress',
    UPGRADE = '/api/v1/alarm/realtime/upgrade'
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
