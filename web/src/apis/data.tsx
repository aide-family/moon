import { Button } from 'antd'
import {
    Category,
    DomainType,
    ModuleType,
    NotifyApp,
    NotifyTemplateType,
    SysLogActionType
} from './types'
import { IconFont } from '@/components/IconFont/IconFont'
import React from 'react'

/**
 * 字典分类数据
 */
const categoryData: Record<Category, string> = {
    [Category.CATEGORY_UNKNOWN]: '全部',
    [Category.CATEGORY_PROM_LABEL]: '标签',
    [Category.CATEGORY_PROM_ANNOTATION]: '注解',
    [Category.CATEGORY_PROM_STRATEGY]: '策略',
    [Category.CATEGORY_PROM_STRATEGY_GROUP]: '策略组',
    [Category.CATEGORY_ALARM_LEVEL]: '告警级别',
    [Category.CATEGORY_ALARM_STATUS]: '告警状态',
    [Category.CATEGORY_NOTIFY_TYPE]: '通知类型',
    [Category.CATEGORY_ALARM_PAGE]: '告警页面'
}
/**  领域类型数据 */
const domainTypeData: Record<DomainType, string> = {
    [DomainType.DomainTypeMonitor]: '监控领域',
    [DomainType.DomainTypeBusiness]: '业务领域',
    [DomainType.DomainTypeOther]: '其他领域',
    [DomainType.DomainTypeSystem]: '系统领域'
}

/**  模块类型枚举 */
const moduleTypeData: Record<ModuleType, string> = {
    [ModuleType.ModelTypeOther]: '其他模块',
    [ModuleType.ModelTypeApi]: '接口模块',
    [ModuleType.ModelTypeMenu]: '菜单模块',
    [ModuleType.ModelTypeRole]: '角色模块',
    [ModuleType.ModelTypeUser]: '用户模块',
    [ModuleType.ModelTypeDict]: '字典模块',
    [ModuleType.ModuleRealtimeAlarm]: '实时告警模块',
    [ModuleType.ModuleAlarmHistory]: '告警历史模块',
    [ModuleType.ModuleStrategyGroup]: '策略组模块',
    [ModuleType.ModuleStrategy]: '策略模块',
    [ModuleType.ModuleDatasource]: '数据源模块',
    [ModuleType.ModuleAlarmPage]: '告警页面模块',
    [ModuleType.ModuleAlarmNotifyGroup]: '告警通知组模块',
    [ModuleType.ModuleAlarmNotifyHook]: '告警通知机器人模块',
    [ModuleType.ModuleDashboardChart]: '仪表盘图表模块',
    [ModuleType.ModuleDashboard]: '仪表盘模块'
}

/** NotifyApp */
const NotifyAppData: Record<NotifyApp, React.ReactNode> = {
    [NotifyApp.NOTIFY_APP_UNKNOWN]: '全部',
    [NotifyApp.NOTIFY_APP_DINGTALK]: (
        <Button type="text" icon={<IconFont type="icon-dingding" />}>
            钉钉
        </Button>
    ),
    [NotifyApp.NOTIFY_APP_WECHATWORK]: (
        <Button type="text" icon={<IconFont type="icon-qiyeweixin" />}>
            企业微信
        </Button>
    ),
    [NotifyApp.NOTIFY_APP_FEISHU]: (
        <Button type="text" icon={<IconFont type="icon-feishu" />}>
            飞书
        </Button>
    ),
    [NotifyApp.NOTIFY_APP_CUSTOM]: (
        <Button type="text" icon={<IconFont type="icon-zidingyi" />}>
            自定义hook
        </Button>
    )
}

const SysLogActionTypeData: Record<
    SysLogActionType,
    { color: string; text: string }
> = {
    [SysLogActionType.SysLogActionUnknown]: {
        color: 'gray',
        text: '未知'
    },
    [SysLogActionType.SysLogActionCreate]: {
        color: 'green',
        text: '创建'
    },
    [SysLogActionType.SysLogActionUpdate]: {
        color: 'orange',
        text: '更新'
    },
    [SysLogActionType.SysLogActionDelete]: {
        color: 'red',
        text: '删除'
    },
    [SysLogActionType.SysLogActionQuery]: {
        color: 'blue',
        text: '查询'
    },
    [SysLogActionType.SysLogActionImport]: {
        color: '#55F107',
        text: '导入'
    },
    [SysLogActionType.SysLogActionExport]: {
        color: '#9C05F1',
        text: '导出'
    }
}

export const NotifyTemplateTypeData: Record<
    NotifyTemplateType,
    React.ReactNode
> = {
    [NotifyTemplateType.NotifyTemplateTypeEmail]: (
        <Button type="text" icon={<IconFont type="icon-youjian" />}>
            邮件
        </Button>
    ),
    [NotifyTemplateType.NotifyTemplateTypeCustom]: (
        <Button type="text" icon={<IconFont type="icon-zidingyi" />}>
            自定义hook
        </Button>
    ),
    [NotifyTemplateType.NotifyTemplateTypeDingDing]: (
        <Button type="text" icon={<IconFont type="icon-dingding" />}>
            钉钉
        </Button>
    ),
    [NotifyTemplateType.NotifyTemplateTypeSms]: (
        <Button type="text" icon={<IconFont type="icon-duanxin" />}>
            短信
        </Button>
    ),
    [NotifyTemplateType.NotifyTemplateTypeFeiShu]: (
        <Button type="text" icon={<IconFont type="icon-feishu" />}>
            飞书
        </Button>
    ),
    [NotifyTemplateType.NotifyTemplateTypeWeChatWork]: (
        <Button type="text" icon={<IconFont type="icon-qiyeweixin" />}>
            企业微信
        </Button>
    )
}

export enum ActionKey {
    /** 刷新 */
    REFRESH = '__refresh__',
    /** 自动刷新 */
    AUTO_REFRESH = '__auto_refresh__',
    /** 重置 */
    RESET = '__reset__',
    /** 操作 */
    ACTION = '__action__',
    /** 列表 */
    LIST = '__list__',
    /** 详情 */
    DETAIL = '__detail__',
    /** 新增 */
    ADD = '__add__',
    /** 编辑 */
    EDIT = '__edit__',
    /** 删除 */
    DELETE = '__delete__',
    /** 启用 */
    ENABLE = '__enable__',
    /** 禁用 */
    DISABLE = '__disable__',
    /** 导入 */
    IMPORT = '__import__',
    /** 导出 */
    EXPORT = '__export__',
    /** 批量删除 */
    BATCH_DELETE = '__batch_delete__',
    /** 批量启用 */
    BATCH_ENABLE = '__batch_enable__',
    /** 批量禁用 */
    BATCH_DISABLE = '__batch_disable__',
    /** 批量导入 */
    BATCH_IMPORT = '__batch_import__',
    /** 批量导出 */
    BATCH_EXPORT = '__batch_export__',
    /** 分配权限 */
    ASSIGN_AUTH = '__assign_auth__',
    /** 分配角色 */
    ASSIGN_ROLE = '__assign_role__',
    /** 状态修改 */
    CHANGE_STATUS = '__change_status__',
    /** IKUAI */
    IKUAI = '__ikuai__',
    /** 规则组列表 */
    STRATEGY_GROUP_LIST = '__strategy_group_list__',
    /** 策略通知对象 */
    STRATEGY_NOTIFY_OBJECT = '__strategy_notify_object__',
    /** 告警介入 */
    ALARM_INTERVENTION = '__alarm_intervention__',
    /** 告警升级 */
    ALARM_UPGRADE = '__alarm_upgrade__',
    /** 告警标记 */
    ALARM_MARK = '__alarm_mark__',
    /** 跳转告警规则列表 */
    OP_KEY_STRATEGY_LIST = 'strategy-list',
    /** 绑定我的告警页面 */
    BIND_MY_ALARM_PAGES = '__bind_my_alarm_pages__',
    /** 配置大盘图表 */
    CONFIG_DASHBOARD_CHART = '__config_dashboard_chart__',
    /** 操作日志 */
    OPERATION_LOG = '__operation_log__',
    /** 个人设置 */
    SELF_SETTING = 'self-setting',
    /** 修改密码 */
    CHANGE_PASSWORD = '__change_password__',
    /** 退出登录 */
    LOGOUT = '__logout__',
    /** 切换角色 */
    SWITCH_ROLE = '__switch_role__',
    /** 告警事件图表 */
    ALARM_EVENT_CHART = '__alarm_event_chart__',
    /** 测试告警模板 */
    TEST_ALARM_TEMPLATE = '__test_alarm_template__',
    /** 绑定通知模板 */
    STRATEGY_BIND_NOTIFY_TEMPLATE = '__strategy_bind_notify_template__',
    /** 展示告警行颜色 */
    ALARM_ROW_COLOR = '__alarm_row_color__'
}

export {
    categoryData,
    domainTypeData,
    moduleTypeData,
    NotifyAppData,
    SysLogActionTypeData
}
