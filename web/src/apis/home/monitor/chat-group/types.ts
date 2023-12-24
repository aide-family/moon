import {
    AlarmApp,
    AlarmStatus,
    PageReqType,
    PageResType,
    Status
} from '@/apis/types'

interface ChatGroupItem {
    id: number
    name: string
    remark: string
    createdAt: number
    updatedAt: number
    hook: string
    status: number
    app: number
    hookName: string
}

interface GetChatGroupDetailRequest {
    id: number
}

interface GetChatGroupDetailResponse {
    detail: ChatGroupItem
}

interface DeleteChatGroupRequest {
    id: number
}

interface DeleteChatGroupResponse {
    id: number
}

interface UpdateChatGroupRequest {
    id: number
    name: string
    remark: string
    hook: string
    app: number
    hookName: string
    status: Status
}

interface UpdateChatGroupResponse {
    id: number
}

interface ListChatGroupRequest {
    page: PageReqType
    keyword?: string
    status?: AlarmStatus
}

interface ListChatGroupResponse {
    list: ChatGroupItem[]
    page: PageResType
}

interface SelectChatGroupRequest {
    page: PageReqType
    keyword?: string
    status?: AlarmStatus
}

interface SelectChatGroupResponse {
    list: ChatGroupItem[]
    page: PageResType
}

interface CreateChatGroupRequest {
    name: string
    remark: string
    hook: string
    app: AlarmApp
    hookName: string
}

interface CreateChatGroupResponse {
    id: number
}

export type {
    ListChatGroupRequest,
    ListChatGroupResponse,
    SelectChatGroupRequest,
    SelectChatGroupResponse,
    UpdateChatGroupRequest,
    UpdateChatGroupResponse,
    DeleteChatGroupRequest,
    DeleteChatGroupResponse,
    GetChatGroupDetailRequest,
    GetChatGroupDetailResponse,
    CreateChatGroupRequest,
    CreateChatGroupResponse,
    ChatGroupItem
}
