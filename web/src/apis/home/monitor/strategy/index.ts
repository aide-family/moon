import { POST } from '@/apis/request'
import {
    StrategyBindNotifyRequest,
    StrategyCreateRequest,
    StrategyDetailResponse,
    StrategyListRequest,
    StrategyListResponse,
    StrategySelectListRequest,
    StrategySelectListResponse,
    StrategyUpdateRequest,
    StrategyWithNOtifyItemType
} from './types'
import { IdReponse, IdsReponse, Status } from '@/apis/types'

enum URL {
    /** 获取策略列表 */
    LIST = '/api/v1/strategy/list',
    /** 获取策略详情 */
    DETAIL = '/api/v1/strategy/detail',
    /** 新增策略 */
    ADD = '/api/v1/strategy/create',
    /** 更新策略 */
    UPDATE = '/api/v1/strategy/update',
    /** 删除策略 */
    DELETE = '/api/v1/strategy/delete',
    /** 启用策略 */
    BATCH_CHANGE_STATUS = '/api/v1/strategy/status/batch/update',
    /** 批量删除策略 */
    BATCH_DELETE = '/api/v1/strategy/batch/delete',
    /** 批量导出策略 */
    BATCH_EXPORT = '/api/v1/strategy/export',
    /** 获取策略列表 */
    SELECT_LIST = '/api/v1/strategy/select',
    /** 获取通知对象明细 */
    NOTIFY_DETAIL = '/api/v1/strategy/notify/object',
    /** 绑定通知对象 */
    BIND_NOTIFY = '/api/v1/strategy/notify/object/bind'
}

/**
 * 获取策略列表
 * @param params
 */
const getStrategyList = (params: StrategyListRequest) => {
    return POST<StrategyListResponse>(URL.LIST, params)
}

/**
 * 获取策略详情
 * @param id
 */
const getStrategyDetail = (id: number) => {
    return POST<StrategyDetailResponse>(URL.DETAIL, { id }).then((res) => {
        return res.detail
    })
}

/**
 * 新增策略
 * @param params
 */
const addStrategy = (params: StrategyCreateRequest) => {
    return POST(URL.ADD, params)
}

/**
 * 更新策略
 * @param params
 */
const updateStrategy = (params: StrategyUpdateRequest) => {
    return POST(URL.UPDATE, params)
}

/**
 * 删除策略
 * @param id
 */
const deleteStrategy = (id: number) => {
    return POST<IdReponse>(URL.DELETE, { id })
}

/**
 * 批量删除策略
 * @param ids
 */
const batchDeleteStrategy = (ids: number[]) => {
    return POST<IdsReponse>(URL.BATCH_DELETE, { ids })
}

/**
 * 批量启用/禁用策略
 * @param ids
 * @param status
 */
const batchChangeStrategyStatus = (ids: number[], status: Status) => {
    return POST<IdsReponse>(URL.BATCH_CHANGE_STATUS, { ids, status })
}

/**
 * 批量导出策略
 * @param ids 策略ID
 * @returns
 */
const batchExportStrategy = (ids: number[]) => {
    // TODO 暂时没有该功能
    return POST<IdsReponse>(URL.BATCH_EXPORT, { ids })
}

/**
 * 获取策略组列表
 * @param params
 */
const getStrategySelectList = (params: StrategySelectListRequest) => {
    return POST<StrategySelectListResponse>(URL.SELECT_LIST, params)
}

/**
 * 获取通知对象详情
 * @param id 策略ID
 * @returns  通知对象详情
 */
const getNotifyDetail = (id: number) => {
    return POST<StrategyWithNOtifyItemType>(URL.NOTIFY_DETAIL, { id })
}

/**
 *  绑定通知对象
 * @param params
 * @returns
 */
const bindNotify = (params: StrategyBindNotifyRequest) => {
    return POST<{ id: number }>(URL.BIND_NOTIFY, params)
}

/** 策略模块API */
const strategyApi = {
    getStrategyList,
    getStrategyDetail,
    addStrategy,
    updateStrategy,
    deleteStrategy,
    batchDeleteStrategy,
    batchChangeStrategyStatus,
    batchExportStrategy,
    getStrategySelectList,
    getNotifyDetail,
    bindNotify
}

export default strategyApi
