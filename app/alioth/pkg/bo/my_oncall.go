package bo

// MyOnCallGroupWindow 是当前登录账号所在通知组在一条值班路由上的接收时段。
// Windows 为空表示全天候接收；Current 表示查询时刻是否落在任一窗口内。
type MyOnCallGroupWindow struct {
	OnCallUID  string
	OnCallName string
	GroupUID   string
	GroupName  string
	// Timezone 是值班路由的 IANA 时区，前端按它展示窗口。
	Timezone string
	Current  bool
	Windows  []NotificationOnCallWindow
}
