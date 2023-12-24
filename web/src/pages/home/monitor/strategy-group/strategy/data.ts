import { StrategyItemType } from '@/pages/home/monitor/strategy-group/strategy/type.ts'

export const defaultData: StrategyItemType[] = [
    {
        alert: 'cpu',
        id: '1',
        duration: '3m',
        expr: "container_processes{beta_kubernetes_io_arch=~\"amd64\"}",
        labels: {},
        annotations: {},
        status:1,
        group_id: '1',
        alert_level_id: '1',
        created_at: '2023-11-01 18:00:00',
        updated_at: '2023-11-01 18:00:00'
    },
]
