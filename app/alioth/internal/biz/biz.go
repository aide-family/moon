package biz

import (
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/biz/agent"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/agenthub"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertdictionary"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertpage"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertrule"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/auth"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/datasource"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/health"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/k8sproxy"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/member"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/message"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/namespace"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/notification"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/partition"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/role"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/sftp"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/ssh"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/user"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/worker"
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
	datasource.ProviderSet,
	alertdictionary.ProviderSet,
	alertpage.ProviderSet,
	alertevent.ProviderSet,
	alertrule.ProviderSet,
	notification.ProviderSet,
	agenthub.ProviderSet,
	agent.ProviderSet,
	ssh.ProviderSet,
	sftp.ProviderSet,
	k8sproxy.ProviderSet,
	partition.ProviderSet,
	wire.Bind(new(auth.AgentTokenLookup), new(*agent.Usecase)),
)
