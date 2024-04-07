import type { ItemType } from 'antd/es/menu/hooks/useItems'
import { IconFont } from '@/components/IconFont/IconFont'

export const defaultMenuItems: ItemType[] = [
    {
        label: '主页',
        key: '/home',
        icon: <IconFont type="icon-monitor3" />
    },
    {
        label: '系统管理',
        key: '/home/system',
        icon: <IconFont type="icon-xitongguanli2" />,
        children: [
            {
                label: '用户管理',
                key: '/home/system/user'
            },
            {
                label: '角色管理',
                key: '/home/system/role'
            },
            {
                label: '字典管理',
                key: '/home/system/dict'
            },
            {
                label: '权限管理',
                key: '/home/system/auth'
            }
        ]
    },
    {
        label: '监控',
        key: '/home/monitor',
        icon: <IconFont type="icon-Prometheus" />,
        children: [
            {
                label: '实时告警',
                key: '/home/monitor/alarm-realtime'
            },
            {
                label: '告警历史',
                key: '/home/monitor/alarm-history'
            },
            {
                label: '策略组',
                key: '/home/monitor/strategy-group'
            },
            {
                label: '数据源',
                key: '/home/monitor/endpoint'
            },
            {
                label: '告警组',
                key: '/home/monitor/alarm-group'
            },
            {
                label: '机器人',
                key: '/home/monitor/chat-hook'
            }
        ]
    },
    {
        label: 'Kubernetes',
        key: '/home/k8s',
        icon: <IconFont type="icon-kubernetes11" />,
        children: [
            {
                label: '集群',
                key: '/home/k8s/cluster',
                icon: <IconFont type="icon-org-gear" />,
                children: [
                    {
                        label: '角色管理',
                        key: '/home/k8s/cluster/role'
                    },
                    {
                        label: '命名空间',
                        key: '/home/k8s/cluster/namespace'
                    },
                    {
                        label: '节点',
                        key: '/home/k8s/cluster/node'
                    },
                    {
                        label: '配置',
                        key: '/home/k8s/cluster/config'
                    },
                    {
                        label: '存储',
                        key: '/home/k8s/cluster/storage'
                    }
                ]
            },
            {
                label: '工作负载',
                key: '/home/k8s/workload',
                icon: <IconFont type="icon-dept" />,
                children: [
                    {
                        label: '定时任务',
                        key: '/home/k8s/workload/cronjob'
                    },
                    {
                        label: '守护进程集',
                        key: '/home/k8s/workload/daemon-set'
                    },
                    {
                        label: '副本集',
                        key: '/home/k8s/workload/replica-sets'
                    },
                    {
                        label: '控制器',
                        key: '/home/k8s/workload/deployment'
                    },
                    {
                        label: '容器组',
                        key: '/home/k8s/workload/pods'
                    },
                    {
                        label: '有状态集',
                        key: '/home/k8s/workload/stateful-set'
                    }
                ]
            },
            {
                label: '服务',
                key: '/home/k8s/service',
                icon: <IconFont type="icon-rongqifuwu1" />,
                children: [
                    {
                        label: 'Ingresses',
                        key: '/home/k8s/service/ingress'
                    },
                    {
                        label: 'Services',
                        key: '/home/k8s/service/service'
                    }
                ]
            }
        ]
    }
]

export type breadcrumbNameType = {
    name: string
    disabled?: boolean
}

export const breadcrumbNameMap: Record<string, breadcrumbNameType> = {
    '/home': {
        name: '主页'
    },
    '/home/monitor': {
        name: '监控',
        disabled: true
    },
    '/home/monitor/alarm-realtime': {
        name: '实时告警'
    },
    '/home/monitor/alarm-history': {
        name: '告警历史'
    },
    '/home/monitor/strategy-group': {
        name: '策略组'
    },
    '/home/monitor/strategy-group/strategy': {
        name: '策略列表'
    },
    '/home/monitor/endpoint': {
        name: '数据源'
    },
    '/home/monitor/alarm-group': {
        name: '告警组'
    },
    '/home/monitor/chat-group': {
        name: '机器人组'
    },
    '/home/system': {
        name: '系统管理',
        disabled: true
    },
    '/home/system/user': {
        name: '用户管理'
    },
    '/home/system/dict': {
        name: '字典管理'
    },
    '/home/system/role': {
        name: '角色管理'
    },
    '/home/system/menu': {
        name: '菜单管理'
    },
    '/home/system/auth': {
        name: '权限管理'
    }
}
