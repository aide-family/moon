package vobj

// MenuProcessType represents the type of menu process.
//
//go:generate stringer -type=MenuProcessType -linecomment -output=type_menu_process.string.go
type MenuProcessType int8

const (
	MenuProcessTypeUnknown MenuProcessType = 0
	MenuProcessTypeLogin   MenuProcessType = 1 << (iota - 1)
	MenuProcessTypeTeam
	MenuProcessTypeLog
	MenuProcessTypeDataPermission
	MenuProcessTypeAdmin
)

func (m MenuProcessType) IsContainsLogin() bool {
	return m&MenuProcessTypeLogin == MenuProcessTypeLogin
}

func (m MenuProcessType) IsContainsTeam() bool {
	return m&MenuProcessTypeTeam == MenuProcessTypeTeam
}

func (m MenuProcessType) IsContainsLog() bool {
	return m&MenuProcessTypeLog == MenuProcessTypeLog
}

func (m MenuProcessType) IsContainsDataPermission() bool {
	return m&MenuProcessTypeDataPermission == MenuProcessTypeDataPermission
}

func (m MenuProcessType) IsContainsAdmin() bool {
	return m&MenuProcessTypeAdmin == MenuProcessTypeAdmin
}

func (m MenuProcessType) IsContainsAll() bool {
	return m.IsContainsLogin() && m.IsContainsTeam() && m.IsContainsLog() && m.IsContainsDataPermission() && m.IsContainsAdmin()
}
