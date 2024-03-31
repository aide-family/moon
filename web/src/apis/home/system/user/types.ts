import { Gender, PageReqType, PageResType, Status } from '@/apis/types'
import type { RoleSelectItem } from '../role/types'

/**创建用户参数 */
interface UserCreateParams {
    username: string
    password: string
    email: string
    phone: string
    nickname: string
    gender: Gender
}

/**更新用户参数 */
interface UserUpdateParams {
    id: number
    nickname: string
    remark: string
    avatar: string
    status: number
    gender: Gender
}

/**根据id查询用户参数 */
interface UesrByIdParams {
    id: number
}

/**查询用户列表参数 */
interface UserListParams {
    page: PageReqType
    keyword?: string
}

/**查询用户列表返回 */
interface UserListRes {
    page: PageResType
    list: UserListItem[]
}

/**用户列表项 */
interface UserListItem {
    id: number
    username: string
    email: string
    phone: string
    /**UNKNOWN 未知 ENABLED 启用 DISABLED 禁用 */
    status: Status
    remark: string
    avatar: string
    createdAt: number | string
    updatedAt: number | string
    deletedAt: number | string
    roles?: RoleSelectItem[]
    nickname: string
    gender: Gender
}

/**用户详情返回 */
interface UserDetailRes {
    detail: UserListItem
}

/**修改密码参数 */
interface UserPasswordEditParams {
    oldPassword: string
    newPassword: string
    code: string
    captchaId: string
}

/**关联角色参数 */
interface UserRolesRelateParams {
    id: number
    roleIds: number[]
}

/**查询用户下拉列表返回 */
interface UserSelectRes {
    page: PageResType
    list?: UserSelectItem[]
}

/**用户下拉列表项 */
interface UserSelectItem {
    value: number
    label: string
    status: number
    avatar: string
    nickname: string
}

/**批量修改状态参数 */
interface userStatusEditParams {
    ids: number[]
    status: number
}
/**批量修改状态返回 */
interface userStatusEditRes {
    ids: number[]
}

export type {
    UserCreateParams,
    UserUpdateParams,
    UesrByIdParams,
    UserListParams,
    UserListRes,
    UserListItem,
    UserDetailRes,
    UserPasswordEditParams,
    UserRolesRelateParams,
    UserSelectRes,
    UserSelectItem,
    userStatusEditParams,
    userStatusEditRes
}
