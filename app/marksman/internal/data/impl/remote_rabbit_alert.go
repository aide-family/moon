package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"

	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func newRabbitAlert(c *config.ExternalDomainConfig) (apiv1.AlertServer, func() error, error) {
	outer := c
	if outer == nil {
		return nil, nil, merr.ErrorInternalServer("external domain config is required")
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.alert", outer.GetEndpoints(), timeout, externalNetwork(outer))

	switch externalNetwork(outer) {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewAlertHTTPClient(httpClient)
		return &outerAlertServer{cfg: outer, httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewAlertClient(grpcConn)
		return &outerAlertServer{cfg: outer, grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", externalNetwork(outer))
	}
}

type outerAlertServer struct {
	apiv1.UnimplementedAlertServer

	cfg        *config.ExternalDomainConfig
	httpClient apiv1.AlertHTTPClient
	grpcClient apiv1.AlertClient
}

func (o *outerAlertServer) ReceivePrometheusWebhook(ctx context.Context, req *apiv1.ReceivePrometheusWebhookRequest) (*apiv1.ReceivePrometheusWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ReceivePrometheusWebhook(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ReceivePrometheusWebhook(externalContext(ctx, o.cfg), req)
}

func (o *outerAlertServer) CreateAlertSubscription(ctx context.Context, req *apiv1.CreateAlertSubscriptionRequest) (*apiv1.CreateAlertSubscriptionReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateAlertSubscription(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.CreateAlertSubscription(externalContext(ctx, o.cfg), req)
}

func (o *outerAlertServer) UpdateAlertSubscription(ctx context.Context, req *apiv1.UpdateAlertSubscriptionRequest) (*apiv1.UpdateAlertSubscriptionReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateAlertSubscription(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateAlertSubscription(externalContext(ctx, o.cfg), req)
}

func (o *outerAlertServer) DeleteAlertSubscription(ctx context.Context, req *apiv1.DeleteAlertSubscriptionRequest) (*apiv1.DeleteAlertSubscriptionReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteAlertSubscription(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.DeleteAlertSubscription(externalContext(ctx, o.cfg), req)
}

func (o *outerAlertServer) GetAlertSubscription(ctx context.Context, req *apiv1.GetAlertSubscriptionRequest) (*apiv1.AlertSubscriptionItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetAlertSubscription(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.GetAlertSubscription(externalContext(ctx, o.cfg), req)
}

func (o *outerAlertServer) ListAlertSubscriptions(ctx context.Context, req *apiv1.ListAlertSubscriptionsRequest) (*apiv1.ListAlertSubscriptionsReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListAlertSubscriptions(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.ListAlertSubscriptions(externalContext(ctx, o.cfg), req)
}

func (o *outerAlertServer) UpdateAlertSubscriptionStatus(ctx context.Context, req *apiv1.UpdateAlertSubscriptionStatusRequest) (*apiv1.UpdateAlertSubscriptionStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateAlertSubscriptionStatus(externalContext(ctx, o.cfg), req)
	}
	return o.grpcClient.UpdateAlertSubscriptionStatus(externalContext(ctx, o.cfg), req)
}
