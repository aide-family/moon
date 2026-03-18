package senderv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
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
	RegisterSenderV1Factory(config.DomainConfig_OUTER, NewOuterSender)
}

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
		// Sender depends on message repository, which requires jobCore/jobClusters.
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

func NewOuterSender(c *config.DomainConfig) (apiv1.SenderServer, func() error, error) {
	outer := &config.OuterServerConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), outer, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal outer server config failed: %v", err)
		}
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.sender", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewSenderHTTPClient(httpClient)
		return &outerSenderServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewSenderClient(grpcConn)
		return &outerSenderServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerSenderServer struct {
	apiv1.UnimplementedSenderServer

	httpClient apiv1.SenderHTTPClient
	grpcClient apiv1.SenderClient
}

func (o *outerSenderServer) SendMessage(ctx context.Context, req *apiv1.SendMessageRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendMessage(ctx, req)
	}
	return o.grpcClient.SendMessage(ctx, req)
}
func (o *outerSenderServer) SendEmail(ctx context.Context, req *apiv1.SendEmailRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendEmail(ctx, req)
	}
	return o.grpcClient.SendEmail(ctx, req)
}
func (o *outerSenderServer) SendEmailWithTemplate(ctx context.Context, req *apiv1.SendEmailWithTemplateRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendEmailWithTemplate(ctx, req)
	}
	return o.grpcClient.SendEmailWithTemplate(ctx, req)
}
func (o *outerSenderServer) SendWebhook(ctx context.Context, req *apiv1.SendWebhookRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendWebhook(ctx, req)
	}
	return o.grpcClient.SendWebhook(ctx, req)
}
func (o *outerSenderServer) SendWebhookWithTemplate(ctx context.Context, req *apiv1.SendWebhookWithTemplateRequest) (*apiv1.SendReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SendWebhookWithTemplate(ctx, req)
	}
	return o.grpcClient.SendWebhookWithTemplate(ctx, req)
}

