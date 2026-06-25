// Package service is a service package for kratos.
package service

import "github.com/google/wire"

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewHealthService,
	NewAuthService,
	NewCaptchaService,
	NewNamespaceService,
	NewUserService,
	NewMemberService,
	NewSelfService,
	NewSysRoleService,
	NewSysMenuService,
	NewSysFunctionService,
	NewSysUserRoleService,
	NewSysRoleMenuService,
	NewSysRoleFunctionService,
	NewSysUserMenuService,
	NewSysUserFunctionService,
	NewNmspacesRoleService,
	NewNmspacesMenuService,
	NewNmspacesFunctionService,
	NewNmspacesMemberRoleService,
	NewNmspacesRoleMenuService,
	NewNmspacesRoleFunctionService,
	NewNmspacesMemberMenuService,
	NewNmspacesMemberFunctionService,
)
