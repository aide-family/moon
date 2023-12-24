import { POST } from '@/apis/request'
import type {
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
    NotifySelectReply
} from './types'

enum URL {
    /** 创建通知对象 */
    CREATE = '/api/v1/prom/notify/create',
    /** 更新通知对象 */
    UPDATE = '/api/v1/prom/notify/update',
    /** 删除通知对象 */
    DELETE = '/api/v1/prom/notify/delete',
    /** 查询通知对象详情 */
    DETAIL = '/api/v1/prom/notify/get',
    /** 查询通知对象列表 */
    LIST = '/api/v1/prom/notify/list',
    /** 获取通知对象列表(用于下拉选择) */
    SELECT = '/api/v1/prom/notify/select'
}

/** 创建通知对象 */
const createNotify = (data: NotifyCreateRequest) => {
    return POST<NotifyCreateReply>(URL.CREATE, data)
}

/** 更新通知对象 */
const updateNotify = (data: NotifyUpdateRequest) => {
    return POST<NotifyUpdateReply>(URL.UPDATE, data)
}

/** 删除通知对象 */
const deleteNotify = (data: NotifyDeleteRequest) => {
    return POST<NotifyDeleteReply>(URL.DELETE, data)
}

/** 通知对象列表 */
const listNotify = (data: NotifyListRequest) => {
    return POST<NotifyListReply>(URL.LIST, data)
}

/** 通知对象详情 */
const detailNotify = (data: NotifyDetailRequest) => {
    return POST<NotifyDetailReply>(URL.DETAIL, data)
}

/** 获取通知对象列表(用于下拉选择) */
const selectNotify = (data: NotifySelectRequest) => {
    return POST<NotifySelectReply>(URL.SELECT, data)
}

/** 通知对象接口 */
const notifyApi = {
    listNotify,
    detailNotify,
    selectNotify,
    updateNotify,
    createNotify,
    deleteNotify
}

export default notifyApi
