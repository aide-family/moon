// Package biz is the business logic for the Goddess service.
package biz

import "github.com/google/wire"

var ProviderSetBiz = wire.NewSet(
	NewHealth,
	NewLoginBiz,
	NewNamespace,
	NewUser,
	NewMember,
	NewCaptcha,
	NewEmail,
	NewSysRole,
	NewSysMenu,
	NewSysFunction,
	NewSysUserRole,
	NewSysRoleMenu,
	NewSysRoleFunction,
	NewSysUserMenu,
	NewSysUserFunction,
	NewNmspacesRole,
	NewNmspacesMenu,
	NewNmspacesFunction,
	NewNmspacesMemberRole,
	NewNmspacesRoleMenu,
	NewNmspacesRoleFunction,
	NewNmspacesMemberMenu,
	NewNmspacesMemberFunction,
)
