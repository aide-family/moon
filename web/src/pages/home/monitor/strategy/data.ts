import { StrategyItemType } from '@/apis/home/monitor/strategy/types'

export const defaultData: StrategyItemType[] = [
    {
        alert: 'cpu',
        id: 1,
        duration: '3m',
        expr: 'container_processes{beta_kubernetes_io_arch=~"amd64"}',
        labels: {},
        annotations: {},
        status: 1,
        groupId: 1,
        alarmLevelId: 1,
        createdAt: 1635876800,
        updatedAt: 1635876800,
        alarmPageIds: [1],
        categoryIds: [1, 2, 3],
        deletedAt: 0,
        remark: '备注'
    }
]
