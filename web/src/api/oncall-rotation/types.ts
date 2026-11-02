/** 轮换周期：每天 / 每周。 */
export type OnCallRotationPeriod = "DAILY" | "WEEKLY"

/** 值班轮换状态：启用 / 停用。 */
export type OnCallRotationStatus = "ACTIVE" | "DISABLED"

/** 详情里的轮换成员，位置由服务端按提交顺序重写。 */
export interface OnCallRotationMemberItem {
  memberUid: string
  memberName: string
  /** 成员邮箱；没有邮箱时为空。 */
  email: string
  /** 成员手机号；没有手机号时为空。 */
  phone: string
  /** 轮转顺序，从 0 开始。 */
  position: number
}

/** 保存时提交的成员。服务端按数组顺序重写 position。 */
export interface OnCallRotationMemberInput {
  memberUid: string
}

/** 值班轮换详情。 */
export interface OnCallRotationItem {
  uid: string
  name: string
  remark: string
  /** IANA 时区名。 */
  timezone: string
  period: OnCallRotationPeriod | string
  /** 轮换起点，RFC3339；第一班从该时刻开始。 */
  startsAt: string
  /** 本地时区的交班时刻，0–23。 */
  handoffHour: number
  status: OnCallRotationStatus | string
  createdAt: string
  updatedAt: string
  members: OnCallRotationMemberItem[]
}

export interface ListOnCallRotationsParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListOnCallRotationsResult {
  items: OnCallRotationItem[]
  total: number
  page: number
  pageSize: number
}

/** 创建与更新共用的载荷；更新时由接口函数补上 uid。 */
export interface SaveOnCallRotationParams {
  name: string
  remark?: string
  timezone: string
  period: OnCallRotationPeriod | string
  startsAt: string
  handoffHour: number
  /** 至少一个，顺序即轮转顺序。 */
  members: OnCallRotationMemberInput[]
}

/** 一个班次。 */
export interface OnCallRotationShiftItem {
  startAt: string
  endAt: string
  memberUid: string
  memberName: string
  email: string
  /** 是否当前值班班次。 */
  current: boolean
}

export interface OnCallRotationScheduleResult {
  shifts: OnCallRotationShiftItem[]
}

export interface OnCallRotationScheduleParams {
  uid: string
  /** 展示未来多少天，1–30；不传按后端默认 14 天。 */
  days?: number
}
