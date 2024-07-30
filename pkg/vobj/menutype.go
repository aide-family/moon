package vobj

// MenuType 菜单类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=MenuType -linecomment
type MenuType int

const (
	// MenuTypeUnknown 未知
	MenuTypeUnknown MenuType = iota // 未知

	// MenuTypeMenu 菜单
	MenuTypeMenu // 菜单

	// MenuTypeButton 按钮
	MenuTypeButton // 按钮

	// MenuTypeDir 文件夹
	MenuTypeDir // 文件夹
)
