// 自定义map类型
type Map<T = any> = { [key: string]: T }

/**基础响应参数 */
type BaseResp<T = any> = {
    // 响应码, 0为成功
    code: number
    // 响应消息
    message: string
    // 响应元数据
    metadata?: Map
    // 响应数据
    data: T
}

/**基础响应格式 */
type Response<T = any> = BaseResp<T>

type Callback<T = any, R = any> = {
    OK?: (res: T) => void
    ERROR?: (err: R) => void
    setLoading?: (loading: boolean) => void
}

type PageReq = {
    current: number
    size: number
}

export const defaultPageReqInfo: PageReq = {
    current: 1,
    size: 10
}

type PageRes<T = any> = {
    current: number
    total: number
    size: number
    records?: T[]
}

interface PageResType {
    curr: number
    size: number
    total: number
}

interface PageReqType {
    curr: number
    size: number
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
    CATEGORY_NOTIFY_TYPE
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
    /** UNKNOWN 未知, 用于默认值*/
    NOTIFY_APP_UNKNOWN,

    /** DINGTALK 钉钉*/
    NOTIFY_APP_DINGTALK,

    /** WECHATWORK 企业微信*/
    NOTIFY_APP_WECHATWORK,

    /** FEISHU 飞书*/
    NOTIFY_APP_FEISHU
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
    /**  系统领域 */
    DomainTypeSystem,
    /**  监控领域 */
    DomainTypeMonitor,
    /**  业务领域 */
    DomainTypeBusiness,
    /**  其他领域 */
    DomainTypeOther
}

/**  模块类型枚举 */
export enum ModuleType {
    /**  接口模块 */
    ModelTypeApi,
    /**  菜单模块 */
    ModelTypeMenu,
    /**  角色模块 */
    ModelTypeRole,
    /**  用户模块 */
    ModelTypeUser,
    /**  字典模块 */
    ModelTypeDict,
    /**  配置模块 */
    ModelTypeConfig,
    /**  日志模块 */
    ModelTypeLog,
    /**  任务模块 */
    ModelTypeJob,
    /**  通知模块 */
    ModelTypeNotify,
    /**  系统模块 */
    ModelTypeSystem,
    /**  监控模块 */
    ModelTypeMonitor,
    /**  业务模块 */
    ModelTypeBusiness,
    /**  其他模块 */
    ModelTypeOther
}

interface IdReponse {
    id: number
}

interface IdsReponse {
    ids: number[]
}

interface Duration {
    value?: number | string | null
    unit?: string
}

export type {
    Map,
    BaseResp,
    Response,
    Callback,
    PageReq,
    PageRes,
    PageResType,
    PageReqType,
    IdReponse,
    IdsReponse,
    Duration
}
