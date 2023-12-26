import { DictSelectItem } from '@/apis/home/system/dict/types.ts'
import { PageReqType, PageResType, Status } from '@/apis/types'
import { StrategyItemType } from '../strategy/types'

type StrategyGroupItemType = {
    name?: string
    id?: string
    remark?: string
    categories?: DictSelectItem[]
    prom_strategies?: StrategyItemType[]
    strategy_total?: number
    status?: string
    created_at?: string | number
    updated_at?: string | number
}

interface StrategyGroupListRequest {
    page: PageReqType
    keyword?: string
    categoryIds?: number[]
    status?: Status
    startAt?: number
    endAt?: number
}

interface StrategyGroupListResponse {
    page: PageResType
    list: StrategyGroupItemType[]
}

export type {
    StrategyGroupItemType,
    StrategyGroupListRequest,
    StrategyGroupListResponse
}
