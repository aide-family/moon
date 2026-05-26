package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/biz/bo"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func NewAlertService(alertBiz *biz.Alert) *AlertService {
	return &AlertService{alertBiz: alertBiz}
}

type AlertService struct {
	apiv1.UnimplementedAlertServer
	alertBiz *biz.Alert
}

func (s *AlertService) ReceivePrometheusWebhook(ctx context.Context, req *apiv1.ReceivePrometheusWebhookRequest) (*apiv1.ReceivePrometheusWebhookReply, error) {
	reqBo := bo.NewReceivePrometheusWebhookBo(req)
	namespaceUID, err := reqBo.NamespaceUID()
	if err != nil {
		return nil, err
	}
	ctx = contextx.WithNamespace(ctx, namespaceUID)
	ctx = contextx.WithUserUID(ctx, namespaceUID)
	uids, err := s.alertBiz.ReceivePrometheusWebhook(ctx, reqBo)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(uids))
	for _, uid := range uids {
		out = append(out, uid.Int64())
	}
	return &apiv1.ReceivePrometheusWebhookReply{
		Total: int64(len(out)),
		Uids:  out,
	}, nil
}

func (s *AlertService) GetAlertRecord(ctx context.Context, req *apiv1.GetAlertRecordRequest) (*apiv1.AlertRecordItem, error) {
	item, err := s.alertBiz.GetAlertRecord(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return item.ToAPIV1AlertRecordItem(), nil
}

func (s *AlertService) ListAlertRecords(ctx context.Context, req *apiv1.ListAlertRecordsRequest) (*apiv1.ListAlertRecordsReply, error) {
	page, err := s.alertBiz.ListAlertRecord(ctx, bo.NewListAlertRecordBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListAlertRecordsReply(page), nil
}

func (s *AlertService) CreateAlertSubscription(ctx context.Context, req *apiv1.CreateAlertSubscriptionRequest) (*apiv1.CreateAlertSubscriptionReply, error) {
	uid, err := s.alertBiz.CreateAlertSubscription(ctx, bo.NewCreateAlertSubscriptionBo(req))
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateAlertSubscriptionReply{Uid: uid.Int64()}, nil
}

func (s *AlertService) UpdateAlertSubscription(ctx context.Context, req *apiv1.UpdateAlertSubscriptionRequest) (*apiv1.UpdateAlertSubscriptionReply, error) {
	if err := s.alertBiz.UpdateAlertSubscription(ctx, bo.NewUpdateAlertSubscriptionBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateAlertSubscriptionReply{}, nil
}

func (s *AlertService) DeleteAlertSubscription(ctx context.Context, req *apiv1.DeleteAlertSubscriptionRequest) (*apiv1.DeleteAlertSubscriptionReply, error) {
	if err := s.alertBiz.DeleteAlertSubscription(ctx, snowflake.ParseInt64(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteAlertSubscriptionReply{}, nil
}

func (s *AlertService) GetAlertSubscription(ctx context.Context, req *apiv1.GetAlertSubscriptionRequest) (*apiv1.AlertSubscriptionItem, error) {
	item, err := s.alertBiz.GetAlertSubscription(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return item.ToAPIV1(), nil
}

func (s *AlertService) ListAlertSubscriptions(ctx context.Context, req *apiv1.ListAlertSubscriptionsRequest) (*apiv1.ListAlertSubscriptionsReply, error) {
	page, err := s.alertBiz.ListAlertSubscription(ctx, bo.NewListAlertSubscriptionBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListAlertSubscriptionsReply(page), nil
}

func (s *AlertService) UpdateAlertSubscriptionStatus(ctx context.Context, req *apiv1.UpdateAlertSubscriptionStatusRequest) (*apiv1.UpdateAlertSubscriptionStatusReply, error) {
	if err := s.alertBiz.UpdateAlertSubscriptionStatus(ctx, bo.NewUpdateAlertSubscriptionStatusBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateAlertSubscriptionStatusReply{}, nil
}
