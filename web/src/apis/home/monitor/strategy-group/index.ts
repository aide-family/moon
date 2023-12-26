import { POST } from '@/apis/request'
import { StrategyGroupListRequest, StrategyGroupListResponse } from './types'

enum URL {
    LIST = '/api/v1/strategy/group/list'
}

/**
 * 获取策略组列表
 * @param params
 */
const getStrategyGroupList = (params: StrategyGroupListRequest) => {
    return POST<StrategyGroupListResponse>(URL.LIST, params)
}

const strategyGroupApi = {
    getStrategyGroupList
}

export default strategyGroupApi
