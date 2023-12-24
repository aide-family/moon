import { POST } from '@/apis/request'
import {
    RoleByIdType,
    RoleCreateReq,
    RoleDetailRes,
    RoleListReq,
    RoleListRes,
    RoleRelateApiParams,
    RoleSelectReq,
    RoleSelectRes,
    RoleUpdateReq
} from './types'

/** URL枚举 */
enum URL {
    /** 创建角色 */
    roleCreate = '/api/v1/role/create',
    /** 角色详情 */
    roleDetail = '/api/v1/role/get',
    /** 角色列表 */
    roleList = '/api/v1/role/list',
    /** 修改角色 */
    roleUpdate = '/api/v1/role/update',
    /** 删除角色 */
    roleDelete = '/api/v1/role/delete',
    /** 获取角色下拉列表 */
    roleSelect = '/api/v1/role/select',
    /** 关联角色接口 */
    roleRelateApi = '/api/v1/role/relate/api'
}

/** 创建角色 */
const roleCreate = (data: RoleCreateReq) => {
    return POST<RoleByIdType>(URL.roleCreate, data)
}

/** 角色详情 */
const roleDetail = (data: RoleByIdType) => {
    return POST<RoleDetailRes>(URL.roleDetail, data)
}

/** 角色列表 */
const roleList = (data: RoleListReq) => {
    return POST<RoleListRes>(URL.roleList, data)
}

/** 修改角色 */
const roleUpdate = (data: RoleUpdateReq) => {
    return POST<RoleByIdType>(URL.roleUpdate, data)
}

/** 删除角色 */
const roleDelete = (data: RoleByIdType) => {
    return POST<RoleByIdType>(URL.roleDelete, data)
}

/** 获取角色下拉列表 */
const roleSelect = (data: RoleSelectReq) => {
    return POST<RoleSelectRes>(URL.roleSelect)
}

/** 角色关联接口 */
const roleRelateApi = (data: RoleRelateApiParams) => {
    return POST<RoleByIdType>(URL.roleRelateApi, data)
}

/** 角色接口 */
export const roleApi = {
    /** 创建角色 */
    roleCreate,
    /** 角色详情 */
    roleDetail,
    /** 角色列表 */
    roleList,
    /** 修改角色 */
    roleUpdate,
    /** 删除角色 */
    roleDelete,
    /** 获取角色下拉列表 */
    roleSelect,
    /** 关联角色接口 */
    roleRelateApi
}

export default roleApi
