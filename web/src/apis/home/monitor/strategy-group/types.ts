import { DictSelectItem } from '@/apis/home/system/dict/types.ts'
import { PageReqType, PageResType, Status } from '@/apis/types'
import { StrategyItemType } from '../strategy/types'

type StrategyGroupItemType = {
    name: string
    id: number
    remark?: string
    categories?: DictSelectItem[]
    prom_strategies?: StrategyItemType[]
    strategy_total?: number
    status?: Status
    createdAt?: string | number
    updatedAt?: string | number
}

type StrategyGroupSelectItemType = {
    label: string
    value: string
    category: number
    color: string
    status: Status
    remark: string
    isDeleted: boolean
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

interface CreateStrategGrouRequest {
    name: string
    remark?: string
    categoryIds?: number[]
}

interface CreategypGrouResponse {
    id: number
}

interface UpdateStrategyGroupRequest {
    id: number
    name?: string
    remark?: string
    categoryIds?: number[]
}

interface UpdateStrategyGroupResponse {
    id: number
}

interface getStrategyGroupDetailRequest {
    id: number
}

interface getStrategyGroupDetailResponse {
    detail: StrategyGroupItemType
}

interface DeleteStrategyGroupRequest {
    id: number
}

interface DeleteStrategyGroupResponse {
    id: number
}

interface SelectStrategyGroupRequest {
    page: PageReqType
    keyword?: string
}

interface SelectStrategyGroupResponse {
    page: PageResType
    list: StrategyGroupSelectItemType[]
}

interface BatchChangeStrategyGroupStatusRequest {
    ids: number[]
    status: Status
}

interface BatchChangeStrategyGroupStatusResponse {
    ids: number[]
}

export type {
    StrategyGroupItemType,
    StrategyGroupListRequest,
    StrategyGroupListResponse,
    CreateStrategGrouRequest,
    CreategypGrouResponse,
    UpdateStrategyGroupRequest,
    UpdateStrategyGroupResponse,
    getStrategyGroupDetailRequest,
    getStrategyGroupDetailResponse,
    DeleteStrategyGroupRequest,
    DeleteStrategyGroupResponse,
    StrategyGroupSelectItemType,
    SelectStrategyGroupRequest,
    SelectStrategyGroupResponse,
    BatchChangeStrategyGroupStatusRequest,
    BatchChangeStrategyGroupStatusResponse
}
