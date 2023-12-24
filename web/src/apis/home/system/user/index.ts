import { POST } from '@/apis/request'
import type {
    UesrByIdParams,
    UserCreateParams,
    UserDetailRes,
    UserListParams,
    UserListRes,
    UserPasswordEditParams,
    UserRolesRelateParams,
    UserSelectRes,
    UserUpdateParams,
    userStatusEditParams,
    userStatusEditRes
} from './types'

/**URL枚举 */
enum URL {
    /**创建用户 */
    userCreate = '/api/v1/user/create',
    /**获取用户列表 */
    userList = '/api/v1/user/list',
    /**获取用户详情 */
    userDetail = '/api/v1/user/get',
    /**更新用户 */
    userUpdate = '/api/v1/user/update',
    /**删除用户 */
    userDelete = '/api/v1/user/delete',
    /**修改密码 */
    userPasswordEdit = '//api/v1/user/password/edit',
    /**用户关联角色 */
    userRolesRelate = 'api/v1/user/roles/relate',
    /**获取用户下拉列表 */
    userSelect = '/api/v1/user/select',
    /**修改状态 */
    userStatusEdit = '/api/v1/user/status/edit'
}

/**创建用户 */
const userCreate = (data: UserCreateParams) => {
    return POST<UesrByIdParams>(URL.userCreate, data)
}

/**获取用户列表 */
const userList = (data: UserListParams) => {
    return POST<UserListRes>(URL.userList, data)
}

/**获取用户详情 */
const userDetail = (data: UesrByIdParams) => {
    return POST<UserDetailRes>(URL.userDetail, data)
}

/**更新用户 */
const userUpdate = (data: UserUpdateParams) => {
    return POST<UesrByIdParams>(URL.userUpdate, data)
}

/**删除用户 */
const userDelete = (data: UesrByIdParams) => {
    return POST<UesrByIdParams>(URL.userDelete, data)
}

/**修改密码 */
const userPasswordEdit = (data: UserPasswordEditParams) => {
    return POST<UesrByIdParams>(URL.userPasswordEdit, data)
}

/**用户关联角色 */
const userRolesRelate = (data: UserRolesRelateParams) => {
    return POST<UesrByIdParams>(URL.userRolesRelate, data)
}

/**获取用户下拉列表 */
const userSelect = (data: UserListParams) => {
    return POST<UserSelectRes>(URL.userSelect, data)
}

/**修改状态 */
const userStatusEdit = (data: userStatusEditParams) => {
    return POST<userStatusEditRes>(URL.userStatusEdit, data)
}
/**用户接口导出
 * @module userApi
 */
export const userApi = {
    /**创建用户*/
    userCreate,
    /**获取用户列表 */
    userList,
    /**获取用户详情 */
    userDetail,
    /**更新用户 */
    userUpdate,
    /**删除用户 */
    userDelete,
    /**修改密码 */
    userPasswordEdit,
    /**用户关联角色 */
    userRolesRelate,
    /**获取用户下拉列表 */
    userSelect,
    /**修改状态 */
    userStatusEdit
}
export default userApi
