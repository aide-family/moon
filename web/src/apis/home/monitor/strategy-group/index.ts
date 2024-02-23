import { POST } from '@/apis/request'
import {
    BatchChangeStrategyGroupStatusRequest,
    BatchChangeStrategyGroupStatusResponse,
    CreateStrategGrouRequest,
    CreategypGrouResponse,
    DeleteStrategyGroupRequest,
    DeleteStrategyGroupResponse,
    ImportGroupRequest,
    SelectStrategyGroupRequest,
    SelectStrategyGroupResponse,
    StrategyGroupListRequest,
    StrategyGroupListResponse,
    UpdateStrategyGroupRequest,
    UpdateStrategyGroupResponse,
    getStrategyGroupDetailRequest,
    getStrategyGroupDetailResponse
} from './types'

enum URL {
    LIST = '/api/v1/strategy/group/list',
    CREATE = '/api/v1/strategy/group/create',
    UPDATE = '/api/v1/strategy/group/update',
    DELETE = '/api/v1/strategy/group/delete',
    DETAIL = '/api/v1/strategy/group/get',
    SELECT = '/api/v1/strategy/group/select',
    BATCH_CHANGE_STATUS = '/api/v1/strategy/group/status/batch/update',
    BATCH_IMPORT = '/api/v1/strategy/group/import'
}

/**
 * 获取策略组列表
 * @param params
 */
const getStrategyGroupList = (params: StrategyGroupListRequest) => {
    return POST<StrategyGroupListResponse>(URL.LIST, params)
}

const createSteategyGroup = (params: CreateStrategGrouRequest) => {
    return POST<CreategypGrouResponse>(URL.CREATE, params)
}

const updateSteategyGroup = (params: UpdateStrategyGroupRequest) => {
    return POST<UpdateStrategyGroupResponse>(URL.UPDATE, params)
}

const deleteSteategyGroup = (params: DeleteStrategyGroupRequest) => {
    return POST<DeleteStrategyGroupResponse>(URL.DELETE, params)
}

const getStrategyGroupDetail = (params: getStrategyGroupDetailRequest) => {
    return POST<getStrategyGroupDetailResponse>(URL.DETAIL, params)
}

const getStrategyGroupSelect = (params: SelectStrategyGroupRequest) => {
    return POST<SelectStrategyGroupResponse>(URL.SELECT, params)
}

const batchChangeStatus = (params: BatchChangeStrategyGroupStatusRequest) => {
    return POST<BatchChangeStrategyGroupStatusResponse>(
        URL.BATCH_CHANGE_STATUS,
        params
    )
}

const batchImport = (params: ImportGroupRequest) => {
    return POST<{ ids: number[] }>(URL.BATCH_IMPORT, params)
}

const strategyGroupApi = {
    getStrategyGroupList,
    createSteategyGroup,
    updateSteategyGroup,
    deleteSteategyGroup,
    getStrategyGroupDetail,
    getStrategyGroupSelect,
    batchChangeStatus,
    batchImport
}

export default strategyGroupApi
