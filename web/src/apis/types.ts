// 自定义map类型
type Map<T = any> = { [key: string]: T }

interface PageReqType {
    curr: number
    size: number
}

interface PageResType extends PageReqType {
    total: number
}

export const defaultPageReqInfo: PageReqType = {
    curr: 1,
    size: 10
}

/**系统状态, 用于表达数据是否可用 */
export enum Status {
    /**UNKNOWN 未知, 用于默认值 */
    STATUS_UNKNOWN,

    /** ENABLED 启用*/
    STATUS_ENABLED,

    /** DISABLED 禁用*/
    STATUS_DISABLED
}

/** 系统状态映射 */
export const StatusMap: Record<Status, { color: string; text: string }> = {
    [Status.STATUS_UNKNOWN]: {
        color: '#888',
        text: '未知'
    },
    [Status.STATUS_ENABLED]: {
        color: 'green',
        text: '启用'
    },
    [Status.STATUS_DISABLED]: {
        color: 'red',
        text: '禁用'
    }
}

export enum AlarmApp {}

/** 告警状态 */
export enum AlarmStatus {
    /**UNKNOWN 未知, 用于默认值 */
    ALARM_STATUS_UNKNOWN,

    /** ALARM 告警*/
    ALARM_STATUS_ALARM,

    /**RESOLVE 已解决 */
    ALARM_STATUS_RESOLVE
}

/** 分类, 区分字典中的各个模块数据 */
export enum Category {
    // UNKNOWN 未知, 用于默认值
    CATEGORY_UNKNOWN,
    /** PromLabel 标签*/
    CATEGORY_PROM_LABEL,

    /** PromAnnotation 注解 */
    CATEGORY_PROM_ANNOTATION,

    /** PromStrategy 策略*/
    CATEGORY_PROM_STRATEGY,

    /** PromStrategyGroup 策略组*/
    CATEGORY_PROM_STRATEGY_GROUP,

    /** AlarmLevel 告警级别*/
    CATEGORY_ALARM_LEVEL,

    /** AlarmStatus 告警状态*/
    CATEGORY_ALARM_STATUS,

    /** NotifyType 通知类型*/
    CATEGORY_NOTIFY_TYPE,

    /** CATEGORY_ALARM_PAGE 报警页面 */
    CATEGORY_ALARM_PAGE
}

/** 通知类型, 用于区分通知方式*/
export enum NotifyType {
    /** UNKNOWN 未知, 用于默认值*/
    NOTIFY_TYPE_UNKNOWN,

    /** EMAIL 邮件*/
    NOTIFY_TYPE_EMAIL,

    /** SMS 短信*/
    NOTIFY_TYPE_SMS,

    /** phone 电话*/
    NOTIFY_TYPE_PHONE
}

/** 通知应用, 用于区分通知方式*/
export enum NotifyApp {
    /** NotifyAppUnknown */
    NOTIFY_APP_UNKNOWN,

    /** DINGTALK 钉钉*/
    NOTIFY_APP_DINGTALK,

    /** WECHATWORK 企业微信*/
    NOTIFY_APP_WECHATWORK,

    /** FEISHU 飞书*/
    NOTIFY_APP_FEISHU,

    /** 自定义 */
    NOTIFY_APP_CUSTOM
}
/** 验证码类型枚举*/
export enum CaptchaType {
    /**UNKNOWN 未知, 用于默认值 */
    CaptchaTypeUnknown,
    /**  audio captcha 音频形式的验证码 */
    CaptchaTypeAudio,
    /** string captcha 字符串形式的验证码 */
    CaptchaTypeString,
    /** math captcha 数学公式形式的验证码 */
    CaptchaTypeMath,
    /** chinese captcha 中文形式的验证码 */
    CaptchaTypeChinese,
    /** digit captcha 数字形式的验证码 */
    CaptchaTypeDigit
}
/**  性别, 用于区分用户性别 */
export enum Gender {
    /**UNKNOWN 未知, 用于默认值 */
    Gender_UNKNOWN,

    /**MALE 男 */
    Gender_MALE,

    /**FEMALE 女 */
    Gender_FEMALE
}

/**  领域类型枚举 */
export enum DomainType {
    /**  其他领域 */
    DomainTypeOther,
    /**  系统领域 */
    DomainTypeSystem,
    /**  监控领域 */
    DomainTypeMonitor,
    /**  业务领域 */
    DomainTypeBusiness
}

/**  模块类型枚举 */
export enum ModuleType {
    /** 其他模块 */
    ModelTypeOther,
    /** 接口模块 */
    ModelTypeApi,
    /** 菜单模块 */
    ModelTypeMenu,
    /** 角色模块 */
    ModelTypeRole,
    /** 用户模块 */
    ModelTypeUser,
    /** 字典模块 */
    ModelTypeDict,
    /** 实时告警模块 */
    ModuleRealtimeAlarm,
    /** 告警历史模块 */
    ModuleAlarmHistory,
    /** 策略组模块 */
    ModuleStrategyGroup,
    /** 策略模块 */
    ModuleStrategy,
    /** 数据源模块 */
    ModuleDatasource,
    /** 告警页面模块 */
    ModuleAlarmPage,
    /** 告警通知组模块 */
    ModuleAlarmNotifyGroup,
    /** 告警通知机器人模块 */
    ModuleAlarmNotifyHook,
    /** 仪表盘图表模块 */
    ModuleDashboardChart,
    /** 仪表盘模块 */
    ModuleDashboard
}

// 系统日志操作类型
export enum SysLogActionType {
    // 未知操作类型
    SysLogActionUnknown,
    // 创建
    SysLogActionCreate,
    // 更新\
    SysLogActionUpdate,
    // 删除
    SysLogActionDelete,
    // 查询
    SysLogActionQuery,
    // 导入
    SysLogActionImport,
    // 导出
    SysLogActionExport
}

interface IdReponse {
    id: number
}

interface IdsReponse {
    ids: number[]
}

interface Duration {
    value?: number | string
    unit?: string
}

export enum MessageType {
    /** 未知 */
    MessageTypeUnknown,

    /** 信息 */
    /** 告警 */
    MessageTypeAlarm,

    /** 通知 */
    MessageTypeNotify,

    /** 系统通知 */
    MessageTypeSystemNotify
}

export type Message = {
    msgType: MessageType
    content?: string
    title?: string
    biz: string
}

/**
 * const (
	// NotifyTemplateTypeCustom  自定义通知模板
	NotifyTemplateTypeCustom NotifyTemplateType = iota
	// NotifyTemplateTypeEmail 邮件通知模板
	NotifyTemplateTypeEmail
	// NotifyTemplateTypeSms 短信通知模板
	NotifyTemplateTypeSms
	// NotifyTemplateTypeWeChatWork 企业微信通知模板
	NotifyTemplateTypeWeChatWork
	// NotifyTemplateTypeFeiShu 飞书通知模板
	NotifyTemplateTypeFeiShu
	// NotifyTemplateTypeDingDing 钉钉通知模板
	NotifyTemplateTypeDingDing
)
 */
export enum NotifyTemplateType {
    NotifyTemplateTypeCustom,
    NotifyTemplateTypeEmail,
    NotifyTemplateTypeSms,
    NotifyTemplateTypeWeChatWork,
    NotifyTemplateTypeFeiShu,
    NotifyTemplateTypeDingDing
}

/**
 * 
enum DatasourceType {
  // UNKNOWN 未知, 用于默认值
  DATASOURCE_TYPE_UNKNOWN = 0;

  // Prometheus
  DATASOURCE_TYPE_PROMETHEUS = 1;

  // VictoriaMetrics
  DATASOURCE_TYPE_VICTORIAMETRICS = 2;

  // Elasticsearch
  DATASOURCE_TYPE_ELASTICSEARCH = 3;

  // Influxdb
  DATASOURCE_TYPE_INFLUXDB = 4;

  // Clickhouse
  DATASOURCE_TYPE_CLICKHOUSE = 5;

  // loki
  DATASOURCE_TYPE_LOKI = 6;
}
 */
export enum DatasourceType {
    DATASOURCE_TYPE_UNKNOWN,
    DATASOURCE_TYPE_PROMETHEUS,
    DATASOURCE_TYPE_VICTORIAMETRICS,
    DATASOURCE_TYPE_ELASTICSEARCH,
    DATASOURCE_TYPE_INFLUXDB,
    DATASOURCE_TYPE_CLICKHOUSE,
    DATASOURCE_TYPE_LOKI
}

export enum ChartType {
    // 全部
    ChartTypeAll,
    // 全屏
    ChartTypeFull,
    // 整行
    ChartTypeRow,
    // 整列
    ChartTypeCol,
}

export type { Map, PageResType, PageReqType, IdReponse, IdsReponse, Duration }
