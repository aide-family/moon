package biz

import (
	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/biz/agent"
	"github.com/aide-family/moon/app/alioth/internal/biz/agenthub"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertdictionary"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertpage"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertrule"
	"github.com/aide-family/moon/app/alioth/internal/biz/auth"
	"github.com/aide-family/moon/app/alioth/internal/biz/cluster"
	"github.com/aide-family/moon/app/alioth/internal/biz/datasource"
	"github.com/aide-family/moon/app/alioth/internal/biz/feature"
	"github.com/aide-family/moon/app/alioth/internal/biz/health"
	"github.com/aide-family/moon/app/alioth/internal/biz/member"
	"github.com/aide-family/moon/app/alioth/internal/biz/message"
	"github.com/aide-family/moon/app/alioth/internal/biz/migration"
	"github.com/aide-family/moon/app/alioth/internal/biz/namespace"
	"github.com/aide-family/moon/app/alioth/internal/biz/notification"
	"github.com/aide-family/moon/app/alioth/internal/biz/order"
	"github.com/aide-family/moon/app/alioth/internal/biz/partition"
	"github.com/aide-family/moon/app/alioth/internal/biz/probe"
	"github.com/aide-family/moon/app/alioth/internal/biz/role"
	"github.com/aide-family/moon/app/alioth/internal/biz/ssh"
	"github.com/aide-family/moon/app/alioth/internal/biz/user"
	"github.com/aide-family/moon/app/alioth/internal/biz/worker"
)

var ProviderSet = wire.NewSet(
	auth.ProviderSet,
	health.ProviderSet,
	namespace.ProviderSet,
	member.ProviderSet,
	message.ProviderSet,
	user.ProviderSet,
	worker.ProviderSet,
	role.ProviderSet,
	feature.ProviderSet,
	order.ProviderSet,
	datasource.ProviderSet,
	probe.ProviderSet,
	alertdictionary.ProviderSet,
	alertpage.ProviderSet,
	alertevent.ProviderSet,
	alertrule.ProviderSet,
	notification.ProviderSet,
	agenthub.ProviderSet,
	agent.ProviderSet,
	ssh.ProviderSet,
	partition.ProviderSet,
	migration.ProviderSet,
	cluster.ProviderSet,
	wire.Bind(new(auth.AgentTokenLookup), new(*agent.Registry)),
)
