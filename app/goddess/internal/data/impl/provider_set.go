// Package impl is the implementation package for the Rabbit service.
package impl

import "github.com/google/wire"

// ProviderSetImpl is a set of providers.
var ProviderSetImpl = wire.NewSet(
	NewHealthRepository,
	NewLoginRepository,
	NewNamespaceRepository,
	NewUserRepository,
	NewMemberRepository,
	NewEmailRepository,
	NewSysRoleRepository,
	NewSysMenuRepository,
	NewSysFunctionRepository,
	NewSysUserRoleRepository,
	NewSysRoleMenuRepository,
	NewSysRoleFunctionRepository,
	NewSysUserMenuRepository,
	NewSysUserFunctionRepository,
	NewNmspacesRoleRepository,
	NewNmspacesMenuRepository,
	NewNmspacesFunctionRepository,
	NewNmspacesMemberRoleRepository,
	NewNmspacesRoleMenuRepository,
	NewNmspacesRoleFunctionRepository,
	NewNmspacesMemberMenuRepository,
	NewNmspacesMemberFunctionRepository,
	NewTransaction,
)
