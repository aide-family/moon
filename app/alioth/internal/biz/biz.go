package biz

import (
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/biz/auth"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/health"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/member"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/namespace"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/role"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/user"
)

var ProviderSet = wire.NewSet(
	auth.ProviderSet,
	health.ProviderSet,
	namespace.ProviderSet,
	member.ProviderSet,
	user.ProviderSet,
	role.ProviderSet,
)
