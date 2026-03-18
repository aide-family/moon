package senderv1

import (
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl"
	"github.com/aide-family/rabbit/internal/service"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	RegisterSenderV1Factory(config.DomainConfig_DEFAULT, NewDefaultSender)
}

// NewDefaultSender creates an in-process sender server (DEFAULT driver).
func NewDefaultSender(c *config.DomainConfig) (apiv1.SenderServer, func() error, error) {
	defaultConfig := &config.DefaultConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), defaultConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
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

	helper := klog.NewHelper(klog.With(klog.GetLogger(), "domain", "rabbit.sender.v1"))
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
	messageBiz := biz.NewMessage(messageLogRepo, messageRepo, helper)

	srv := &defaultSender{SenderServer: service.NewSenderService(emailBiz, webhookBiz, messageBiz)}
	return srv, func() error {
		closeData()
		return nil
	}, nil
}

type defaultSender struct {
	apiv1.SenderServer
}
