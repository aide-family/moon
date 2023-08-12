import {M} from "@/apis/type"

// 字典类别枚举
export enum Category {
    // 未知
    CATEGORY_NONE = "CATEGORY_NONE",
    // 规则组类型
    CATEGORY_GROUP = "CATEGORY_GROUP",
    // 套餐类型
    CATEGORY_COMBO = "CATEGORY_COMBO",
    // 规则类型
    CATEGORY_STRATEGY = "CATEGORY_STRATEGY",
    // 告警等级
    CATEGORY_ALERT_LEVEL = "CATEGORY_ALERT_LEVEL",
    // 告警通知方式
    CATEGORY_ALERT_NOTIFY = "CATEGORY_ALERT_NOTIFY",
}

// 状态枚举
export enum Status {
    // 未知
    Status_NONE = "Status_NONE",
    // 启用
    Status_ENABLE = "Status_ENABLE",
    // 禁用
    Status_DISABLE = "Status_DISABLE",
}

export type StatusValue = {
    // 状态名称
    text: string,
    // 状态原始值
    value: Status,
    // number
    number: number,
    // 状态颜色
    color?: string,
    // 相反状态
    opposite?: StatusValue,
}

export type StatusMap = {
    [key in Status]: StatusValue
}

export const StatusMap: StatusMap = {
    [Status.Status_NONE]: {
        text: "未知",
        value: Status.Status_NONE,
        color: "#999999",
        number: 0,
        opposite: {
            text: "未知",
            value: Status.Status_NONE,
            color: "#999999",
            number: 0,
        },
    },
    [Status.Status_ENABLE]: {
        text: "启用",
        value: Status.Status_ENABLE,
        color: "#67C23A",
        number: 1,
        opposite: {
            text: "禁用",
            value: Status.Status_DISABLE,
            color: "#F56C6C",
            number: 2,
        }
    },
    [Status.Status_DISABLE]: {
        text: "禁用",
        value: Status.Status_DISABLE,
        color: "#F56C6C",
        number: 2,
        opposite: {
            text: "启用",
            value: Status.Status_ENABLE,
            color: "#67C23A",
            number: 1,
        }
    },
}

// 规则组
export type PromDict = {
    // 字典名称
    name: string,
    // 字典备注
    remark: string,
    // 字典创建时间, unix时间戳
    createdAt: string,
    // 字典更新时间, unix时间戳
    updatedAt: string,
    // 字典类别
    category: Category,
    // 字典颜色
    color: string,
    // 字典唯一ID
    id: number
}

// 告警等级
export type AlertLevel = {
    // 等级名称
    name: string,
    // 等级备注
    remark: string,
    // 创建时间, unix时间戳
    createdAt: string,
    // 更新时间, unix时间戳
    updatedAt: string,
    // 等级类别
    category: string,
    // 等级颜色
    color: string,
    // 等级唯一ID
    id: number
}

// 告警页面
export type AlarmPage = {
    // 页面名称
    name: string,
    // 页面备注
    remark: string,
    // 页面图标
    icon: string,
    // 页面tab颜色
    color: string,
    // 页面创建时间, unix时间戳
    createdAt: string,
    // 页面更新时间, unix时间戳
    updatedAt: string,
    // 页面唯一ID
    id: number
}

// 规则
export type PromStrategyItem = {
    // 规则所属规则组ID
    groupId: number,
    // 规则名称
    alert: string,
    // 规则PromQL表达式
    expr: string,
    // 持续时间, 单位(s|m|h|d), 表示prom ql表达式持续多久时间满足条件
    for: string,
    // 规则标签, 用于自定义告警标签信息
    labels: M,
    // 规则注释, 用于自定义告警注释信息, 例如:告警标题、告警内容等
    annotations: M,
    // 创建时间, unix时间戳
    createdAt: string,
    // 更新时间, unix时间戳
    updatedAt: string,
    // 规则tag, 用于标记规则属性, 例如告警对象领域, 业务领域等
    categories: PromDict[],
    // 规则tag id, 用于标记规则属性, 例如告警对象领域, 业务领域等, 与categories一一对应
    categorieIds: number[],
    // 告警等级ID, 用于标记告警等级, 每一个告警规则必有一个告警等级
    alertLevelId: number,
    // 告警等级, 用于标记告警等级, 每一个告警规则必有一个告警等级
    alertLevel: AlertLevel,
    // 告警页面, 用于标记告警页面, 告警发生时, 告警信息会发送到对应的告警页面
    alarmPageIds: number[],
    // 告警页面, 用于标记告警页面, 告警发生时, 告警信息会发送到对应的告警页面, 与alarmPageIds一一对应
    alarmPages: AlarmPage[],
    // 规则状态, 用于标记规则状态, 启用或禁用
    status: Status,
    id: number
}

// 规则组
export type GroupItem = {
    // 规则组唯一ID
    id: number
    // 规则组名称, 唯一
    name: string,
    // 规则组备注
    remark: string,
    // 规则组状态, 启用或禁用
    status: Status,
    // 规则组下规则数量, int64的字符串
    strategyCount: string,
    // 规则组属性, 用于标记规则组属性, 例如告警对象领域, 业务领域等
    categories: PromDict[],
    // 规则组属性ID, 用于标记规则组属性, 例如告警对象领域, 业务领域等, 与categories一一对应
    categoriesIds: number[],
    // 规则组创建时间, unix时间戳
    createdAt: string,
    // 规则组更新时间, unix时间戳
    updatedAt: string,
    // 规则组下规则列表
    promStrategies: PromStrategyItem[]
}

// 规则组
export type GroupItemRequest = {
    // 规则组状态, 启用或禁用
    status?: number,
    // 规则组下规则数量, int64的字符串
    strategyCount?: number,
    // 规则组属性ID, 用于标记规则组属性, 例如告警对象领域, 业务领域等, 与categories一一对应
    categoriesIds?: number[],
}