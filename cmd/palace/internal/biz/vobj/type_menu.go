package vobj

// MenuType menu type
//
//go:generate stringer -type=MenuType -linecomment -output=type_menu.string.go
type MenuType int8

const (
	MenuTypeUnknown            MenuType = iota // unknown
	MenuTypeMenuSystem                         // system
	MenuTypeMenuTeam                           // team
	MenuTypeMenuUser                           // user
	MenuTypeMenuTeamDatasource                 // team datasource
)
