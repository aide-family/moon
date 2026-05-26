package alertv1

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	alertdomain "github.com/aide-family/rabbit/domain/alert"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func init() {
	alertdomain.RegisterAlertV1Factory(config.DomainConfig_OUTER, NewOuterAlert)
}

func NewOuterAlert(c *config.DomainConfig, driver *anypb.Any) (apiv1.AlertServer, func() error, error) {
	outer := &config.OuterServerConfig{}
	if pointer.IsNotNil(driver) {
		if err := anypb.UnmarshalTo(driver, outer, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal outer server config failed: %v", err)
		}
	}

	timeout := 10 * time.Second
	if outer.GetTimeout() != nil && outer.GetTimeout().AsDuration() > 0 {
		timeout = outer.GetTimeout().AsDuration()
	}
	initCfg := connect.NewDefaultConfig("rabbit.alert", outer.GetAddress(), timeout, outer.GetProtocol().String())

	switch outer.GetProtocol().String() {
	case connect.ProtocolHTTP:
		httpClient, err := connect.InitHTTPClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewAlertHTTPClient(httpClient)
		return &outerAlertServer{httpClient: client}, httpClient.Close, nil
	case connect.ProtocolGRPC:
		grpcConn, err := connect.InitGRPCClient(initCfg)
		if err != nil {
			return nil, nil, err
		}
		client := apiv1.NewAlertClient(grpcConn)
		return &outerAlertServer{grpcClient: client}, grpcConn.Close, nil
	default:
		return nil, nil, merr.ErrorInternalServer("unknown protocol: %s", outer.GetProtocol().String())
	}
}

type outerAlertServer struct {
	apiv1.UnimplementedAlertServer

	httpClient apiv1.AlertHTTPClient
	grpcClient apiv1.AlertClient
}

func (o *outerAlertServer) ReceivePrometheusWebhook(ctx context.Context, req *apiv1.ReceivePrometheusWebhookRequest) (*apiv1.ReceivePrometheusWebhookReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ReceivePrometheusWebhook(ctx, req)
	}
	return o.grpcClient.ReceivePrometheusWebhook(ctx, req)
}

func (o *outerAlertServer) GetAlertRecord(ctx context.Context, req *apiv1.GetAlertRecordRequest) (*apiv1.AlertRecordItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetAlertRecord(ctx, req)
	}
	return o.grpcClient.GetAlertRecord(ctx, req)
}

func (o *outerAlertServer) ListAlertRecords(ctx context.Context, req *apiv1.ListAlertRecordsRequest) (*apiv1.ListAlertRecordsReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListAlertRecords(ctx, req)
	}
	return o.grpcClient.ListAlertRecords(ctx, req)
}

func (o *outerAlertServer) CreateAlertSubscription(ctx context.Context, req *apiv1.CreateAlertSubscriptionRequest) (*apiv1.CreateAlertSubscriptionReply, error) {
	if o.httpClient != nil {
		return o.httpClient.CreateAlertSubscription(ctx, req)
	}
	return o.grpcClient.CreateAlertSubscription(ctx, req)
}

func (o *outerAlertServer) UpdateAlertSubscription(ctx context.Context, req *apiv1.UpdateAlertSubscriptionRequest) (*apiv1.UpdateAlertSubscriptionReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateAlertSubscription(ctx, req)
	}
	return o.grpcClient.UpdateAlertSubscription(ctx, req)
}

func (o *outerAlertServer) DeleteAlertSubscription(ctx context.Context, req *apiv1.DeleteAlertSubscriptionRequest) (*apiv1.DeleteAlertSubscriptionReply, error) {
	if o.httpClient != nil {
		return o.httpClient.DeleteAlertSubscription(ctx, req)
	}
	return o.grpcClient.DeleteAlertSubscription(ctx, req)
}

func (o *outerAlertServer) GetAlertSubscription(ctx context.Context, req *apiv1.GetAlertSubscriptionRequest) (*apiv1.AlertSubscriptionItem, error) {
	if o.httpClient != nil {
		return o.httpClient.GetAlertSubscription(ctx, req)
	}
	return o.grpcClient.GetAlertSubscription(ctx, req)
}

func (o *outerAlertServer) ListAlertSubscriptions(ctx context.Context, req *apiv1.ListAlertSubscriptionsRequest) (*apiv1.ListAlertSubscriptionsReply, error) {
	if o.httpClient != nil {
		return o.httpClient.ListAlertSubscriptions(ctx, req)
	}
	return o.grpcClient.ListAlertSubscriptions(ctx, req)
}

func (o *outerAlertServer) UpdateAlertSubscriptionStatus(ctx context.Context, req *apiv1.UpdateAlertSubscriptionStatusRequest) (*apiv1.UpdateAlertSubscriptionStatusReply, error) {
	if o.httpClient != nil {
		return o.httpClient.UpdateAlertSubscriptionStatus(ctx, req)
	}
	return o.grpcClient.UpdateAlertSubscriptionStatus(ctx, req)
}
