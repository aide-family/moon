package alertv1

import (
	"time"

	_ "github.com/aide-family/goddess/domain/member/v1"
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	alertdomain "github.com/aide-family/rabbit/domain/alert"
	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl"
	"github.com/aide-family/rabbit/internal/service"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	alertdomain.RegisterAlertV1Factory(config.DomainConfig_DEFAULT, func(c *config.DomainConfig, driver *anypb.Any) (apiv1.AlertServer, func() error, error) {
		return NewDefaultAlert(c, driver, nil, nil)
	})
}

// NewDefaultAlert creates an in-process alert server (DEFAULT driver).
// goddessDriver and memberDomain are required when embedding from marksman; memberDomain may be nil to use DEFAULT/v1.
func NewDefaultAlert(
	c *config.DomainConfig,
	rabbitDriver *anypb.Any,
	goddessDriver *anypb.Any,
	memberDomain *config.DomainConfig,
) (apiv1.AlertServer, func() error, error) {
	if pointer.IsNil(goddessDriver) {
		return nil, nil, merr.ErrorInternalServer("goddess domain driver is required for alert DEFAULT driver")
	}
	defaultConfig := &config.DefaultConfig{}
	if pointer.IsNotNil(rabbitDriver) {
		if err := anypb.UnmarshalTo(rabbitDriver, defaultConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal default config failed: %v", err)
		}
	}
	if pointer.IsNil(memberDomain) {
		memberDomain = &config.DomainConfig{
			Driver:  config.DomainConfig_DEFAULT,
			Version: "v1",
		}
	}

	bootstrap := &conf.Bootstrap{
		Jwt:                 defaultConfig.GetJwt(),
		Database:            defaultConfig.GetDatabase(),
		GoddessDomainDriver: goddessDriver,
		MemberDomain:        memberDomain,
		JobCore: &conf.JobCore{
			WorkerTotal: 1,
			Timeout:     durationpb.New(10 * time.Second),
			BufferSize:  100,
			MaxRetry:    3,
		},
		JobClusters: &config.ClusterConfig{},
	}

	helper := klog.NewHelper(klog.With(klog.GetLogger(), "domain", "rabbit.alert.v1"))
	d, closeData, err := data.New(bootstrap, helper)
	if err != nil {
		return nil, nil, err
	}

	emailConfigRepo := impl.NewEmailConfigRepository(d)
	webhookConfigRepo := impl.NewWebhookConfigRepository(d)
	messageLogRepo := impl.NewMessageLogRepository(d)
	messageRetryLogRepo := impl.NewMessageRetryLogRepository(d)
	templateRepo := impl.NewTemplateRepository(d)

	messageRepo, err := impl.NewMessageRepository(bootstrap, d, messageLogRepo)
	if err != nil {
		closeData()
		return nil, nil, err
	}

	jobBiz := biz.NewJob(messageRepo, helper)
	messageLogBiz := biz.NewMessageLog(messageLogRepo, messageRetryLogRepo, jobBiz, helper)
	templateBiz := biz.NewTemplate(templateRepo, helper)
	emailConfigBiz := biz.NewEmailConfig(emailConfigRepo, helper)
	webhookConfigBiz := biz.NewWebhookConfig(webhookConfigRepo, helper)
	emailBiz := biz.NewEmail(emailConfigBiz, templateBiz, messageLogBiz, helper)
	webhookBiz := biz.NewWebhook(webhookConfigBiz, messageLogBiz, templateBiz, helper)

	memberRepo, err := impl.NewMemberRepository(bootstrap, d)
	if err != nil {
		closeData()
		return nil, nil, err
	}

	recipientGroupRepo := impl.NewRecipientGroupRepository(d)
	recipientMemberRepo := impl.NewRecipientMemberRepository(d)
	recipientGroupBiz := biz.NewRecipientGroup(memberRepo, recipientGroupRepo, recipientMemberRepo, helper)

	alertSubscriptionRepo := impl.NewAlertSubscriptionRepository(d)
	alertBiz := biz.NewAlert(alertSubscriptionRepo, memberRepo, recipientGroupBiz, emailBiz, webhookBiz, helper)
	srv := &defaultAlert{AlertServer: service.NewAlertService(alertBiz)}

	return srv, func() error {
		closeData()
		return nil
	}, nil
}

type defaultAlert struct {
	apiv1.AlertServer
}
