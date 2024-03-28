import { PageReqType, PageResType, Status } from '@/apis/types'
import { UserSelectItem } from '../../system/user/types'
import { ChatGroupSelectItem } from '../chat-group/types'

interface ListAlarmGroupRequest {
    page: PageReqType
    keyword?: string
}

interface ListAlarmGroupResponse {
    page: PageResType
    list?: AlarmGroupItem[]
}

interface AlarmNotifyMember {
    memberId: number
    notifyType: number
    id: number
}

interface CreateAlarmGroupRequest {
    name: string
    remark?: string
    members?: AlarmNotifyMember[]
    chatGroups?: number[]
}

interface CreateAlarmGroupResponse {
    id: number
}

interface NotifyMemberItem {
    memberId: number
    notifyType: number
    user?: UserSelectItem
    status: number
    id: number
}

interface AlarmGroupItem {
    id: number
    name: string
    remark: string
    status: Status
    members?: NotifyMemberItem[]
    chatGroups?: ChatGroupSelectItem[]
    createdAt: number
    updatedAt: number
    deletedAt: number
}

interface AlarmGroupSelectItem {
    value: number
    label: string
    remark: string
    status: number
}

interface SelectAlarmGroupRequest {
    page: PageReqType
    keyword?: string
}

interface SelectAlarmGroupResponse {
    page: PageResType
    list?: AlarmGroupSelectItem[]
}

interface UpdateAlarmGroupRequest {
    id: number
    name?: string
    remark?: string
    members?: AlarmNotifyMember[]
    chatGroups?: number[]
}

interface UpdateAlarmGroupResponse {
    id: number
}

export const defaultListAlarmGroupRequest: ListAlarmGroupRequest = {
    page: {
        curr: 1,
        size: 10
    }
}

export const defaultSelectAlarmGroupRequest: SelectAlarmGroupRequest = {
    page: {
        curr: 1,
        size: 10
    }
}

export type {
    ListAlarmGroupRequest,
    ListAlarmGroupResponse,
    AlarmGroupItem,
    NotifyMemberItem,
    CreateAlarmGroupRequest,
    CreateAlarmGroupResponse,
    AlarmNotifyMember,
    AlarmGroupSelectItem,
    SelectAlarmGroupRequest,
    SelectAlarmGroupResponse,
    UpdateAlarmGroupRequest,
    UpdateAlarmGroupResponse
}
