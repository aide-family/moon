import { POST } from '@/apis/request'
import {
    AlarmGroupItem,
    CreateAlarmGroupRequest,
    CreateAlarmGroupResponse,
    ListAlarmGroupRequest,
    ListAlarmGroupResponse,
    SelectAlarmGroupRequest,
    SelectAlarmGroupResponse,
    UpdateAlarmGroupRequest,
    UpdateAlarmGroupResponse
} from './types'

enum URL {
    LIST = '/api/v1/prom/notify/list',
    DETAIL = '/api/v1/prom/notify/get',
    CREATE = '/api/v1/prom/notify/create',
    UPDATE = '/api/v1/prom/notify/update',
    DELETE = '/api/v1/prom/notify/delete',
    SELECT = '/api/v1/prom/notify/select'
}

const create = (params: CreateAlarmGroupRequest) => {
    return POST<CreateAlarmGroupResponse>(URL.CREATE, params)
}

const update = (params: UpdateAlarmGroupRequest) => {
    return POST<UpdateAlarmGroupResponse>(URL.UPDATE, params)
}

const detail = (id: number) => {
    return POST<{ detail: AlarmGroupItem }>(URL.DETAIL, { id })
}

const deleteById = (id: number) => {
    return POST<{ id: number }>(URL.DELETE, { id })
}

const list = (params: ListAlarmGroupRequest) => {
    return POST<ListAlarmGroupResponse>(URL.LIST, params)
}

const select = (params: SelectAlarmGroupRequest) => {
    return POST<SelectAlarmGroupResponse>(URL.SELECT, params)
}

const alarmGroupApi = {
    create,
    update,
    detail,
    deleteById,
    list,
    select
}

export default alarmGroupApi
