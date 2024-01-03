import { DictSelectItem } from '../../system/dict/types'
import { AlarmPageSelectItem } from '../alarm-page/types'
import { PrometheusServerSelectItem } from '../endpoint/types'
import { StrategyGroupSelectItemType } from '../strategy-group/types'
import { Map, PageReqType, PageResType } from '@/apis/types'

/** 策略详情 */
interface StrategyItemType {
    id: number
    alert: string
    expr: string
    duration: string
    labels: Map
    annotations: Map
    status: number
    groupId: number
    groupInfo?: StrategyGroupSelectItemType
    alarmLevelId: number
    alarmLevelInfo?: DictSelectItem
    alarmPageIds: number[]
    alarmPageInfo?: AlarmPageSelectItem[]
    categoryIds: number[]
    categoryInfo?: DictSelectItem[]
    endpoint?: PrometheusServerSelectItem
    createdAt: number
    updatedAt: number
    deletedAt: number
    remark: string
}

interface StrategySelectItemType {
    value: number
    label: string
    category: number
    color: string
    status: number
    remark: string
    isDeleted: boolean
}

/** 策略创建请求参数 */
interface StrategyCreateRequest {
    groupId: number
    alert: string
    expr: string
    duration: string
    labels: Map
    annotations: Map
    alarmPageIds: number[]
    categoryIds: number[]
    alarmLevelId: number
    remark: string
    dataSourceId: number
}

/** 策略创建响应参数 */
interface StrategyCreateResponse {
    id: number
}

/** 策略更新请求参数 */
interface StrategyUpdateRequest {
    id: number
    groupId: number
    alert: string
    expr: string
    duration: string
    labels: Map
    annotations: Map
    alarmPageIds: number[]
    categoryIds: number[]
    alarmLevelId: Number
    remark?: string
}

/** 策略更新响应参数 */
interface StrategyUpdateResponse {
    id: number
}

/** 策略删除请求参数 */
interface StrategyDeleteRequest {
    id: number
}

/** 策略删除响应参数 */
interface StrategyDeleteResponse {
    id: number
}

/** 策略列表请求参数 */
interface StrategyListRequest {
    page: PageReqType
    keyword: string
    groupId: number
    categoryIds: number[]
    alarmLevelId: number
    status: number
    isDeleted: false
}

export const defaultStrategyListRequest: StrategyListRequest = {
    page: {
        curr: 1,
        size: 10
    },
    keyword: '',
    groupId: 0,
    categoryIds: [],
    alarmLevelId: 0,
    status: 0,
    isDeleted: false
}

/** 策略列表响应参数 */
interface StrategyListResponse {
    list: StrategyItemType[]
    page: PageResType
}

interface StrategySelectListRequest {
    page: PageReqType
    keyword: string
}

interface StrategySelectListResponse {
    list: StrategySelectItemType[]
    page: PageResType
}

interface StrategyDetailResponse {
    detail: StrategyItemType
}

export type {
    StrategyItemType,
    StrategyCreateRequest,
    StrategyCreateResponse,
    StrategyUpdateRequest,
    StrategyUpdateResponse,
    StrategyDeleteRequest,
    StrategyDeleteResponse,
    StrategyListRequest,
    StrategyListResponse,
    StrategySelectListRequest,
    StrategySelectListResponse,
    StrategyDetailResponse,
    StrategySelectItemType
}
