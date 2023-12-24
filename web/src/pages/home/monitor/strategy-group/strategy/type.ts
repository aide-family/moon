import { Map } from '@/apis/types.ts'

export type StrategyItemType = {
    id?: string
    alert?: string
    expr?: string
    duration?: string
    labels?: Map
    annotations?: Map

    group_id?: string
    alert_level_id?: string
    status?: number | string

    created_at?: string | number
    updated_at?: string | number
}
