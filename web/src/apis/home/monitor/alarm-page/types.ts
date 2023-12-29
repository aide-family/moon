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
    status: number
}

interface SelectAlarmPageReply {
    list: AlarmPageItem[]
    page: PageResType
}

interface BatchChangeAlarmPageStatusRequest {
    ids: number[]
    status: Status
}

interface BatchChangeAlarmPageStatusReply {
    ids: number[]
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
    ListAlarmPageRequest
}
