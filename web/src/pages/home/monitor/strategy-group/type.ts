import { StrategyItemType } from '@/pages/home/monitor/strategy-group/strategy/type.ts'
import { DictSelectItem } from '@/apis/home/system/dict/types.ts'

export type StrategyGroupItemType = {
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
