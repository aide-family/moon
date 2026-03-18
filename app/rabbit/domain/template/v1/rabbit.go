package templatev1

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
	RegisterTemplateV1Factory(config.DomainConfig_DEFAULT, NewDefaultTemplate)
	RegisterTemplateV1Factory(config.DomainConfig_OUTER, NewOuterTemplate)
}

func NewDefaultTemplate(c *config.DomainConfig) (apiv1.TemplateServer, func() error, error) {
	defaultConfig := &config.DefaultConfig{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), defaultConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal default config failed: %v", err)
		}
	}

	bootstrap := &conf.Bootstrap{
		Jwt:      defaultConfig.GetJwt(),
		Database: defaultConfig.GetDatabase(),
		// Template domain does not rely on message repo, but data.New expects a valid Bootstrap.
		JobCore: &conf.JobCore{
			WorkerTotal: 1,
			Timeout:     durationpb.New(10 * time.Second),
			BufferSize:  1,
			MaxRetry:    0,
		},
		JobClusters: &config.ClusterConfig{},
	}

	helper := klog.NewHelper(klog.With(klog.GetLogger(), "domain", "rabbit.template.v1"))
	d, closeData, err := data.New(bootstrap, helper)
	if err != nil {
		return nil, nil, err
	}

	templateRepo := impl.NewTemplateRepository(d)
	templateBiz := biz.NewTemplate(templateRepo, helper)
	srv := &defaultTemplate{TemplateServer: service.NewTemplateService(templateBiz)}

	return srv, func() error {
		closeData()
		return nil
	}, nil
}

type defaultTemplate struct {
	apiv1.TemplateServer
}

func NewOuterTemplate(c *config.DomainConfig) (apiv1.TemplateServer, func() error, error) {
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
	initCfg := connect.NewDefaultConfig("rabbit.template", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewTemplateHTTPClient(httpClient)
		return &outerTemplateServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewTemplateClient(grpcConn)
		return &outerTemplateServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerTemplateServer struct {
	apiv1.UnimplementedTemplateServer

	httpClient apiv1.TemplateHTTPClient
	grpcClient apiv1.TemplateClient
}

func (o *outerTemplateServer) CreateTemplate(ctx context.Context, req *apiv1.CreateTemplateRequest) (*apiv1.CreateTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateTemplate(ctx, req)
	}
	return o.grpcClient.CreateTemplate(ctx, req)
}
func (o *outerTemplateServer) UpdateTemplate(ctx context.Context, req *apiv1.UpdateTemplateRequest) (*apiv1.UpdateTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateTemplate(ctx, req)
	}
	return o.grpcClient.UpdateTemplate(ctx, req)
}
func (o *outerTemplateServer) UpdateTemplateStatus(ctx context.Context, req *apiv1.UpdateTemplateStatusRequest) (*apiv1.UpdateTemplateStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateTemplateStatus(ctx, req)
	}
	return o.grpcClient.UpdateTemplateStatus(ctx, req)
}
func (o *outerTemplateServer) DeleteTemplate(ctx context.Context, req *apiv1.DeleteTemplateRequest) (*apiv1.DeleteTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteTemplate(ctx, req)
	}
	return o.grpcClient.DeleteTemplate(ctx, req)
}
func (o *outerTemplateServer) GetTemplate(ctx context.Context, req *apiv1.GetTemplateRequest) (*apiv1.TemplateItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetTemplate(ctx, req)
	}
	return o.grpcClient.GetTemplate(ctx, req)
}
func (o *outerTemplateServer) ListTemplate(ctx context.Context, req *apiv1.ListTemplateRequest) (*apiv1.ListTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListTemplate(ctx, req)
	}
	return o.grpcClient.ListTemplate(ctx, req)
}
func (o *outerTemplateServer) SelectTemplate(ctx context.Context, req *apiv1.SelectTemplateRequest) (*apiv1.SelectTemplateReply, error) {
	if o.httpClient != nil {
		return o.httpClient.SelectTemplate(ctx, req)
	}
	return o.grpcClient.SelectTemplate(ctx, req)
}

