// Package biz is the business logic for the Goddess service.
package biz

import "github.com/google/wire"

var ProviderSetBiz = wire.NewSet(
	NewHealth,
	NewLoginBiz,
	NewNamespace,
	NewUser,
	NewMember,
)
