package webhookv1

import (
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	webhookdomain "github.com/aide-family/rabbit/domain/webhook"
	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl"
	"github.com/aide-family/rabbit/internal/service"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	webhookdomain.RegisterWebhookV1Factory(config.DomainConfig_DEFAULT, NewDefaultWebhook)
}

// NewDefaultWebhook creates an in-process webhook server (DEFAULT driver).
func NewDefaultWebhook(c *config.DomainConfig, driver *anypb.Any) (apiv1.WebhookServer, func() error, error) {
	defaultConfig := &config.DefaultConfig{}
	if pointer.IsNotNil(driver) {
		if err := anypb.UnmarshalTo(driver, defaultConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal default config failed: %v", err)
		}
	}
	bootstrap := &conf.Bootstrap{
		Jwt:      defaultConfig.GetJwt(),
		Database: defaultConfig.GetDatabase(),
		JobCore: &conf.JobCore{
			WorkerTotal: 1,
			Timeout:     durationpb.New(10 * time.Second),
			BufferSize:  100,
			MaxRetry:    3,
		},
		JobClusters: &config.ClusterConfig{},
	}

	helper := klog.NewHelper(klog.With(klog.GetLogger(), "domain", "rabbit.webhook.v1"))
	d, closeData, err := data.New(bootstrap, helper)
	if err != nil {
		return nil, nil, err
	}

	webhookConfigRepo := impl.NewWebhookConfigRepository(d)
	webhookConfigBiz := biz.NewWebhookConfig(webhookConfigRepo, helper)
	srv := &defaultWebhook{WebhookServer: service.NewWebhookService(webhookConfigBiz)}

	return srv, func() error {
		closeData()
		return nil
	}, nil
}

type defaultWebhook struct {
	apiv1.WebhookServer
}
