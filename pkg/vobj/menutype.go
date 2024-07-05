package vobj

//go:generate go run ../../cmd/server/stringer/cmd.go -type=MenuType -linecomment
type MenuType int

const (
	MenuTypeUnknown MenuType = iota // 未知
	MenuTypeMenu                    // 菜单
	MenuTypeButton                  // 按钮
	MenuTypeDir                     // 文件夹
)
