import { POST } from '@/apis/request'
import {
    ApiAuthById,
    ApiAuthDetailRes,
    ApiAuthListReq,
    ApiAuthListRes,
    ApiAuthSelectRes,
    ApiAuthUpdate,
    CreateApiAuth
} from './types'

/** URL枚举 */
export enum URL {
    /**创建API数据 */
    authApiCreate = '/api/v1/system/api/create',
    /**删除API数据 */
    authApiDelete = '/api/v1/system/api/delete',
    /**修改API数据 */
    authApiUpdate = '/api/v1/system/api/update',
    /**获取API数据详情 */
    authApiDetail = '/api/v1/system/api/get',
    /**获取API列表数据 */
    authApiList = '/api/v1/system/api/list',
    /**获取API下拉列表 */
    authApiSelect = '/api/v1/system/api/select'
}
/**创建API数据 */
const authApiCreate = (data: CreateApiAuth) => {
    return POST<ApiAuthById>(URL.authApiCreate, data)
}
/**删除API数据 */
const authApiDelete = (data: ApiAuthById) => {
    return POST<ApiAuthById>(URL.authApiDelete, data)
}
/**修改API数据 */
const authApiUpdate = (data: ApiAuthUpdate) => {
    return POST<ApiAuthById>(URL.authApiUpdate, data)
}
/**获取API数据详情 */
const authApiDetail = (data: ApiAuthById) => {
    return POST<ApiAuthDetailRes>(URL.authApiDetail, data)
}
/**获取API列表数据 */
const authApiList = (data: ApiAuthListReq) => {
    return POST<ApiAuthListRes>(URL.authApiList, data)
}
/**获取API下拉列表 */
const authApiSelect = (data: ApiAuthListReq) => {
    return POST<ApiAuthSelectRes>(URL.authApiSelect, data)
}
/**导出接口 */
export const authApi = {
    /**创建API数据 */
    authApiCreate,
    /**删除API数据 */
    authApiDelete,
    /**修改API数据 */
    authApiUpdate,
    /**获取API数据详情 */
    authApiDetail,
    /**获取API列表数据 */
    authApiList,
    /**获取API下拉列表 */
    authApiSelect
}
export default authApi
