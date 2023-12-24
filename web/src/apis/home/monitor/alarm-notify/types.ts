import { PageReqType, PageResType, Status } from '@/apis/types'
import { UserSelectItem } from '../../system/user/types'

/** 通知成员 */
interface NotifyMember {
    memberId: number
    notifyTypes: number[]
    user?: UserSelectItem
    status: number
    id?: number
}

/** 群组hook */
interface ChatGroup {
    value: number
    app: number
    label: string
    status: number
}

/** 外部客户hook */
interface ExternalCustomerHook {
    id: number
    hookName: string
    remark: string
    status: number
    customerId: number
    hook: string
    notifyApp: number
    createdAt: number
    updatedAt: number
    deletedAt: number
}

/** 外部客户 */
interface ExternalCustomer {
    id: number
    name: string
    remark: string
    status: number
    addr: string
    contact: string
    phone: string
    email: string
    createdAt: number
    updatedAt: number
    deletedAt: number
    externalCustomerHookList: ExternalCustomerHook[]
}

/** 外部客户通知对象 */
interface EexternalNotifyObj {
    id: number
    name: string
    remark: string
    status: number
    externalCustomerList: ExternalCustomer[]
    externalCustomerHookList: ExternalCustomerHook[]
    createdAt: number
    updatedAt: number
    deletedAt: number
}

/** 通知对象 */
interface NotifyItem {
    id: number
    name: string
    remark: string
    status: number
    members: NotifyMember[]
    chatGroups: ChatGroup[]
    createdAt: number
    updatedAt: number
    deletedAt: number
    externalNotifyObjs: EexternalNotifyObj[]
}

/** 通知对象详情获取请求 */
interface NotifyDetailRequest {
    id: number
}

/** 通知对象详情获取回复 */
interface NotifyDetailReply {
    detail: NotifyItem
}

/** 通知对象保存请求格式 */
interface NotifyMemberSave {
    memberId: number
    notifyTypes: number[]
    id?: number
}

/** 通知对象创建请求参数 */
interface NotifyCreateRequest {
    name: string
    remark: string
    members: NotifyMemberSave[]
    chatGroups: number[]
}

/** 通知对象创建回复 */
interface NotifyCreateReply {
    id: number
}

/** 通知对象删除请求参数 */
interface NotifyDeleteRequest {
    id: number
}

/** 通知对象删除回复 */
interface NotifyDeleteReply {
    id: number
}

/** 通知对象更新请求参数 */
interface NotifyUpdateRequest {
    id: number
    name: string
    remark: string
    members: NotifyMemberSave[]
    chatGroups: number[]
    status: Status
}

/** 通知对象更新回复 */
interface NotifyUpdateReply {
    id: number
}

/** 通知对象列表请求参数 */
interface NotifyListRequest {
    page: PageReqType
    keyword?: string
    status?: Status[]
}

/** 通知对象列表回复 */
interface NotifyListReply {
    list: NotifyItem[]
    page: PageResType
}

interface NotifySelectItem {
    value: number
    label: string
    remark: string
    status: number
}

interface NotifySelectRequest {
    page: PageReqType
    keyword?: string
    status?: Status[]
}

interface NotifySelectReply {
    list: NotifySelectItem[]
    page: PageResType
}

export type {
    NotifyItem,
    NotifyCreateRequest,
    NotifyCreateReply,
    NotifyUpdateRequest,
    NotifyUpdateReply,
    NotifyListRequest,
    NotifyListReply,
    NotifyDetailRequest,
    NotifyDetailReply,
    NotifyDeleteRequest,
    NotifyDeleteReply,
    NotifySelectRequest,
    NotifySelectReply,
    NotifySelectItem,
    EexternalNotifyObj,
    ExternalCustomer,
    NotifyMember
}
