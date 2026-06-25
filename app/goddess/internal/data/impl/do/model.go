package do

func Models() []any {
	return []any{
		&User{},
		&OAuth2User{},
		&Namespace{},
		&Member{},
		&SysRole{},
		&SysMenu{},
		&SysFunction{},
		&SysUserRole{},
		&SysRoleMenu{},
		&SysRoleFunction{},
		&SysUserMenu{},
		&SysUserFunction{},
		&NmspacesRole{},
		&NmspacesMenu{},
		&NmspacesFunction{},
		&NmspacesMemberRole{},
		&NmspacesRoleMenu{},
		&NmspacesRoleFunction{},
		&NmspacesMemberMenu{},
		&NmspacesMemberFunction{},
	}
}
