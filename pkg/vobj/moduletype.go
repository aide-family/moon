package vobj

// ModuleType 数据状态
//
//go:generate stringer -type=ModuleType -linecomment
type ModuleType int

const (
	// ModuleTypeUnknown 未知
	ModuleTypeUnknown ModuleType = iota // 未知

	// ModuleTypeMenu 菜单模块
	ModuleTypeMenu // 菜单模块
)

var a ModuleType

func c() {
	a.GetValue()
}
