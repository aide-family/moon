package vobj

// ChartType 系统商品类型
//
//go:generate stringer -type=ChartType -linecomment
type ChartType int

const (
	ChartTypeUnknown ChartType = iota // 未知
	ChartTypeFull                     // 全屏
	ChartTypeRow                      // 行
	ChartTypeCol                      // 列
)
