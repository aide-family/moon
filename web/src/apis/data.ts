import { Category, DomainType, ModuleType, NotifyApp } from './types'

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
    [Category.CATEGORY_NOTIFY_TYPE]: '通知类型'
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
    [ModuleType.ModelTypeApi]: '接口模块',
    [ModuleType.ModelTypeDict]: '字典模块',
    [ModuleType.ModelTypeMonitor]: '监控模块',
    [ModuleType.ModelTypeMenu]: '菜单模块',
    [ModuleType.ModelTypeUser]: '用户模块',
    [ModuleType.ModelTypeLog]: '日志模块',
    [ModuleType.ModelTypeJob]: '任务模块',
    [ModuleType.ModelTypeNotify]: '通知模块',
    [ModuleType.ModelTypeOther]: '其他模块',
    [ModuleType.ModelTypeSystem]: '系统模块',
    [ModuleType.ModelTypeBusiness]: '业务模块',
    [ModuleType.ModelTypeRole]: '角色模块',
    [ModuleType.ModelTypeConfig]: '配置模块'
}

/** NotifyApp */
const NotifyAppData: Record<NotifyApp, string> = {
    [NotifyApp.NOTIFY_APP_UNKNOWN]: '未知',
    [NotifyApp.NOTIFY_APP_DINGTALK]: '钉钉',
    [NotifyApp.NOTIFY_APP_WECHATWORK]: '企业微信',
    [NotifyApp.NOTIFY_APP_FEISHU]: '飞书'
}

export enum ActionKey {
    /** 刷新 */
    REFRESH = '__refresh__',
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
    CONFIG_DASHBOARD_CHART = '__config_dashboard_chart__'
}

export { categoryData, domainTypeData, moduleTypeData, NotifyAppData }
