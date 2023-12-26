import { DictSelectItem } from '../../system/dict/types'
import { AlarmPageItem } from '../alarm-page/types'
import { StrategyGroupItemType } from '../strategy-group/types'
import { Map } from '@/apis/types'

interface StrategyItemType {
    id: number
    alert: string
    expr: string
    duration: string
    labels: Map
    annotations: Map
    status: number
    groupId: number
    groupInfo?: StrategyGroupItemType
    alarmLevelId: number
    alarmLevelInfo?: DictSelectItem
    alarmPageIds: number[]
    alarmPageInfo?: AlarmPageItem
    categoryIds: number[]
    categoryInfo?: DictSelectItem[]
    createdAt: number
    updatedAt: number
    deletedAt: number
    remark: string
}

export type { StrategyItemType }
