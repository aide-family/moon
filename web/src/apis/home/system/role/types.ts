import { PageReqType, PageResType, Status } from '@/apis/types'
import { UserSelectItem } from '../user/types'
import { ApiAuthSelectItem } from '../auth/types'

/**角色下拉配置项 */
interface RoleSelectItem {
    value: number
    label: string
    status: number
    remark: string
}

/**创建角色参数 */
interface RoleCreateReq {
    name: string
    remark: string
}
/**根据id查询角色参数 */
interface RoleByIdType {
    id: number
}
/**角色列表项 */
interface RoleListItem {
    id: number
    name: string
    status: Status
    remark: string
    createdAt: number
    updatedAt: number
    deletedAt: number
    users?: UserSelectItem[]
    apis?: ApiAuthSelectItem[]
}
/**角色列表参数 */
interface RoleListReq {
    page: PageReqType
    keyword?: string
}
/**角色详情 */
interface RoleDetailRes {
    detail: RoleListItem
}
/**角色列表 */
interface RoleListRes {
    page: PageResType
    list: RoleListItem[]
}
/**角色关联api */
interface RoleRelateApiParams {
    id: number
    apiIds: number[]
}
/**角色下拉列表获取参数 */
interface RoleSelectReq extends RoleListReq {
    userId?: number
}
/**角色下拉列表 */
interface RoleSelectRes {
    page: PageResType
    list?: RoleSelectItem[]
}
/**修改角色参数 */
interface RoleUpdateReq {
    id: number
    name: string
    remark: string
    status: number
}

export const defaultRoleSelectReq: RoleSelectReq = {
    page: {
        curr: 1,
        size: 10
    }
}

export type {
    RoleSelectItem,
    RoleCreateReq,
    RoleByIdType,
    RoleListItem,
    RoleListReq,
    RoleDetailRes,
    RoleListRes,
    RoleRelateApiParams,
    RoleSelectReq,
    RoleSelectRes,
    RoleUpdateReq
}
