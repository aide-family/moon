import { DictSelectItem } from '@/apis/home/system/dict/types.ts'
import { Duration, Map, PageReqType, PageResType, Status } from '@/apis/types'
import { StrategyItemType } from '../strategy/types'

type StrategyGroupItemType = {
    name: string
    id: number
    remark?: string
    categories?: DictSelectItem[]
    enableStrategyCount: number | string
    strategyCount: number | string
    status?: Status
    createdAt?: string | number
    updatedAt?: string | number
    strategies?: StrategyItemType[]
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
    ids?: number[]
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
    status?: Status
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

interface ImportRuleItemType {
    alert: string
    expr: string
    labels: Map
    annotations: Map
    for: Duration
}

interface ImportGroupItemType {
    name: string
    rules: ImportRuleItemType[]
}

interface ImportGroupRequest {
    groups?: ImportGroupItemType[]
    datasourceId?: number
    defaultLevel?: number
    defaultAlarmPageIds?: number[]
    defaultCategoryIds?: number[]
    defaultAlarmNotifyIds?: number[]
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
    BatchChangeStrategyGroupStatusResponse,
    ImportGroupRequest,
    ImportGroupItemType,
    ImportRuleItemType
}
