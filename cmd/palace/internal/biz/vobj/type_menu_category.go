package vobj

// MenuCategory represents the category of menu.
//
//go:generate stringer -type=MenuCategory -linecomment -output=type_menu_category.string.go
type MenuCategory int8

const (
	MenuCategoryUnknown MenuCategory = iota
	MenuCategoryMenu
	MenuCategoryButton
)
