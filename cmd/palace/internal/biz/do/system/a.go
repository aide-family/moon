package system

func Models() []any {
	return []any{
		&OperateLog{},
		&Menu{},
		&Resource{},
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
