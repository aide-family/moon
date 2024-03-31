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
                label: '策略列表',
                key: '/home/monitor/strategy'
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
    }
]

export const breadcrumbNameMap: Record<string, string> = {
    '/home': '主页',
    '/home/system': '系统管理',
    '/home/customer': '客户管理',
    '/home/resource': '资源管理',
    '/home/flow': '流程管理',
    '/home/monitor': '监控',
    '/home/monitor/alarm-realtime': '实时告警',
    '/home/monitor/alarm-history': '告警历史',
    '/home/monitor/strategy-group': '策略组',
    '/home/monitor/strategy': '策略',
    '/home/monitor/endpoint': '数据源',
    '/home/monitor/alarm-group': '告警组',
    '/home/monitor/chat-group': '机器人组',
    '/home/system/user': '用户管理',
    '/home/system/dict': '字典管理',
    '/home/system/role': '角色管理',
    '/home/system/menu': '菜单管理',
    '/home/system/auth': '权限管理',
    '/home/customer/list': '客户列表',
    '/home/customer/supplier': '供应商',
    '/home/resource/device': '设备列表',
    '/home/resource/node': '节点列表',
    '/home/resource/account': '账号列表',
    '/home/flow/template': '流程模板',
    '/home/flow/instance': '流程实例',
    '/home/flow/task': '任务大厅',
    '/home/flow/mytask': '我的任务',
    '/home/flow/form': '流程表单'
}
