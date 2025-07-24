// Package system is a system package for kratos.
package system

func Models() []any {
	return []any{
		&OperateLog{},
		&Menu{},
		&Role{},
		&SendMessageLog{},
		&Team{},
		&TeamAudit{},
		&TeamInviteLink{},
		&TeamInviteUser{},
		&TeamRole{},
		&TeamMember{},
		&User{},
		&UserConfigTable{},
		&UserConfigTheme{},
		&UserOAuth{},
	}
}
