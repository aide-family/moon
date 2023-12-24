import { StrategyGroupItemType } from '@/pages/home/monitor/strategy-group/type.ts'

export const defaultData: StrategyGroupItemType[] = [
    {
        name: '规则组1',
        id: '1',
        remark: '备注',
        categories: [
            {
                name: '分类1',
                id: '1'
            },
            {
                name: '分类2',
                id: '2'
            }
        ],
        prom_strategies: [
            {
                alert: '规则1',
                id: '1',
                status: '1',
                created_at: '2021-04-01 12:00:00',
                updated_at: '2021-04-01 12:00:00'
            },
            {
                alert: '规则2',
                id: '2',
                status: '1',
                created_at: '2021-04-01 12:00:00',
                updated_at: '2021-04-01 12:00:00'
            }
        ],
        strategy_total: 2,
        status: '1',
        created_at: '2021-04-01 12:00:00',
        updated_at: '2021-04-01 12:00:00'
    }
]
