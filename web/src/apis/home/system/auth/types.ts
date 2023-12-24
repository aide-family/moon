/**创建Api权限 */

import {
    DomainType,
    ModuleType,
    PageReqType,
    PageResType,
    Status
} from '@/apis/types'

interface CreateApiAuth {
    name: string
    path: string
    method: string
    remark: string
    module: ModuleType
    domain: DomainType
}
interface ApiAuthById {
    id: number
}

interface ApiAuthUpdate {
    id: number
    name: string
    path: string
    method: string
    status: number
    remark: string
    module: ModuleType
    domain: DomainType
}

interface ApiAuthDetailRes {
    detail: ApiAuthListItem
}

interface ApiAuthListItem {
    id: number
    name: string
    path: string
    method: string
    status: Status
    remark: string
    createdAt: number
    updatedAt: number
    deletedAt: number
    module: ModuleType
    domain: DomainType
}

interface ApiAuthListRes {
    list: ApiAuthListItem[]
    page: PageResType
}
interface ApiAuthListReq {
    page: PageReqType
    keyword?: string
}

interface ApiAuthSelectRes {
    list: ApiAuthSelectItem[]
    page: PageResType
}

interface ApiAuthSelectItem {
    value: number
    label: string
    status: number
    remark: string
    module: ModuleType
    domain: DomainType
}

export type {
    CreateApiAuth,
    ApiAuthById,
    ApiAuthUpdate,
    ApiAuthDetailRes,
    ApiAuthListItem,
    ApiAuthListRes,
    ApiAuthListReq,
    ApiAuthSelectRes,
    ApiAuthSelectItem
}
