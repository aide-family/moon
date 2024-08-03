package vobj

// DashboardChartType 仪表盘图表类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=DashboardChartType -linecomment
type DashboardChartType int

const (
	// DashboardChartTypeUnknown 未知
	DashboardChartTypeUnknown DashboardChartType = iota // 未知

	// DashboardChartTypeFullScreen 全屏
	DashboardChartTypeFullScreen // 全屏

	// DashboardChartTypeRow 行
	DashboardChartTypeRow // 行

	// DashboardChartTypeCol 列
	DashboardChartTypeCol // 列
)
