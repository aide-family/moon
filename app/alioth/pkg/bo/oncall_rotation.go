package bo

import "time"

// OnCallRotation 是一条值班轮换：按固定周期与时区在成员之间轮转。
type OnCallRotation struct {
	UID          string
	NamespaceUID string
	Name         string
	Remark       string
	// Timezone 是 IANA 时区名，保存规范名称。
	Timezone string
	// Period 取 DAILY 或 WEEKLY，决定一轮班次的长度。
	Period string
	// StartsAt 是轮换起点，第一班从该时刻开始（UTC）。
	StartsAt time.Time
	// HandoffHour 是本地时区的交班时刻，0–23。
	HandoffHour int
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Members 按 Position 升序。
	Members []OnCallRotationMember
}

// OnCallRotationMember 是轮换里的一个参与成员，附带展示用的名称与联系方式。
type OnCallRotationMember struct {
	UID          string
	NamespaceUID string
	RotationUID  string
	MemberUID    string
	Name         string
	Email        string
	Phone        string
	Position     int
}

// OnCallRotationShift 是一个排班班次。
type OnCallRotationShift struct {
	StartAt time.Time
	EndAt   time.Time
	Member  OnCallRotationMember
	// Current 表示该班次包含查询时刻。
	Current bool
}

// MyOnCallShift 是当前登录账号在一次轮换里的某个班次。
// 个人视图只关心「我的班」，所以不带其他成员的展示信息。
type MyOnCallShift struct {
	RotationUID  string
	RotationName string
	// Timezone 是轮换时区，前端按它换算并标注展示时间。
	Timezone string
	StartAt  time.Time
	EndAt    time.Time
	// Current 表示该班次包含查询时刻。
	Current bool
}
