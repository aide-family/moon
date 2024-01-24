import { PageReqType, PageResType, Status } from '@/apis/types'

interface BacthDeleteAlarmPageRequest {
    ids: number[]
}

interface BacthDeleteAlarmPageReply {
    ids: number[]
}

interface CreateAlarmPageRequest {
    name: string
    icon?: string
    color?: string
    remark?: string
}

interface CreateAlarmPageReply {
    id: number
}

interface DeleteAlarmPageRequest {
    id: number
}

interface DeleteAlarmPageReply {
    id: number
}

interface UpdateAlarmPageRequest {
    id: number
    name: string
    icon?: string
    color?: string
    remark?: string
}

interface UpdateAlarmPageReply {
    id: number
}

interface GetAlarmPageRequest {
    id: number
}

interface AlarmPageItem {
    id: number
    name: string
    icon: string
    color: string
    status: number
    remark: string
    createdAt: number
    updatedAt: number
    deletedAt: number
}

interface AlarmPageSelectItem {
    value: number
    label: string
    icon: string
    color: string
    status: number
    remark: string
}

interface GetAlarmPageReply {
    alarmPage: AlarmPageItem
}

interface ListAlarmPageRequest {
    page: PageReqType
    keyword: string
    status: number
}

interface ListAlarmPageReply {
    list: AlarmPageItem[]
    page: PageResType
}

interface SelectAlarmPageRequest {
    page: PageReqType
    keyword: string
    status?: Status
}

interface SelectAlarmPageReply {
    list: AlarmPageSelectItem[]
    page: PageResType
}

interface BatchChangeAlarmPageStatusRequest {
    ids: number[]
    status: Status
}

interface BatchChangeAlarmPageStatusReply {
    ids: number[]
}

interface CountAlarmPageRequest {
    ids: number[]
}

interface CountAlarmPageReply {
    alarmCount: { [key: number]: number | string }
}

export type {
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
    AlarmPageItem,
    ListAlarmPageReply,
    ListAlarmPageRequest,
    AlarmPageSelectItem,
    CountAlarmPageReply,
    CountAlarmPageRequest
}
