import { DictSelectItem } from '../../system/dict/types'
import { NotifyItem } from '../alarm-notify/types'
import { PrometheusServerSelectItem } from '../endpoint/types'
import { StrategyGroupSelectItemType } from '../strategy-group/types'
import { Duration, Map, PageReqType, PageResType } from '@/apis/types'

/** 策略详情 */
interface StrategyItemType {
    id: number
    alert: string
    expr: string
    duration?: Duration
    labels?: Map
    annotations?: Map
    status: number
    groupId: number
    groupInfo?: StrategyGroupSelectItemType
    alarmLevelId: number
    alarmLevelInfo?: DictSelectItem
    alarmPageIds: number[]
    alarmPageInfo?: DictSelectItem[]
    categoryIds: number[]
    categoryInfo?: DictSelectItem[]
    createdAt: number
    updatedAt: number
    deletedAt: number
    remark: string
    dataSourceId: number
    dataSource: PrometheusServerSelectItem
    maxSuppress?: Duration
    // 告警通知间隔
    sendInterval?: Duration
    // 是否发送告警通知
    sendRecover?: boolean
}

interface StrategyWithNOtifyItemType {
    detail: StrategyItemType
    notifyObjectList: NotifyItem[]
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
    duration?: Duration
    labels?: Map
    annotations?: Map
    alarmPageIds: number[]
    categoryIds: number[]
    alarmLevelId: number
    remark?: string
    dataSourceId: number
    // 最大抑制时常
    maxSuppress?: Duration
    // 告警通知间隔
    sendInterval?: Duration
    // 是否发送告警通知
    sendRecover?: boolean
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
    duration?: Duration
    labels?: Map
    annotations?: Map
    alarmPageIds: number[]
    categoryIds: number[]
    alarmLevelId: Number
    remark?: string
    dataSourceId?: number
    // 最大抑制时常
    maxSuppress?: Duration
    // 告警通知间隔
    sendInterval?: Duration
    // 是否发送告警通知
    sendRecover?: boolean
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
    strategyId?: number
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

interface StrategyBindNotifyRequest {
    id: number
    notifyObjectIds: number[]
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
    StrategySelectItemType,
    StrategyWithNOtifyItemType,
    StrategyBindNotifyRequest
}
