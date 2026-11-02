package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/alertescalation"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// AlertEscalationPolicyService 把升级策略接口转成 biz 入参；权限与校验都在 biz。
type AlertEscalationPolicyService struct {
	v1.UnimplementedAlertEscalationPolicyServiceServer
	gateway *alertescalation.Gateway
}

func NewAlertEscalationPolicyService(gateway *alertescalation.Gateway) *AlertEscalationPolicyService {
	return &AlertEscalationPolicyService{gateway: gateway}
}

func (s *AlertEscalationPolicyService) CreateAlertEscalationPolicy(ctx context.Context, req *v1.CreateAlertEscalationPolicyRequest) (*v1.CreateAlertEscalationPolicyResponse, error) {
	item, err := s.gateway.Create(ctx, escalationPolicyInput(req.GetName(), req.GetDescription(), req.GetLabels(),
		req.GetLevelUids(), req.GetDatasourceUids(), req.GetRuleUids(), req.GetSteps()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateAlertEscalationPolicyResponse{Policy: escalationPolicyItem(*item)}, nil
}

func (s *AlertEscalationPolicyService) UpdateAlertEscalationPolicy(ctx context.Context, req *v1.UpdateAlertEscalationPolicyRequest) (*v1.UpdateAlertEscalationPolicyResponse, error) {
	in := escalationPolicyInput(req.GetName(), req.GetDescription(), req.GetLabels(),
		req.GetLevelUids(), req.GetDatasourceUids(), req.GetRuleUids(), req.GetSteps())
	in.UID = req.GetUid()
	item, err := s.gateway.Update(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertEscalationPolicyResponse{Policy: escalationPolicyItem(*item)}, nil
}

func (s *AlertEscalationPolicyService) UpdateAlertEscalationPolicyStatus(ctx context.Context, req *v1.UpdateAlertEscalationPolicyStatusRequest) (*v1.UpdateAlertEscalationPolicyStatusResponse, error) {
	item, err := s.gateway.UpdateStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertEscalationPolicyStatusResponse{Policy: escalationPolicyItem(*item)}, nil
}

func (s *AlertEscalationPolicyService) DeleteAlertEscalationPolicy(ctx context.Context, req *v1.DeleteAlertEscalationPolicyRequest) (*v1.DeleteAlertEscalationPolicyResponse, error) {
	if err := s.gateway.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteAlertEscalationPolicyResponse{}, nil
}

func (s *AlertEscalationPolicyService) GetAlertEscalationPolicy(ctx context.Context, req *v1.GetAlertEscalationPolicyRequest) (*v1.GetAlertEscalationPolicyResponse, error) {
	item, err := s.gateway.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertEscalationPolicyResponse{Policy: escalationPolicyItem(*item)}, nil
}

func (s *AlertEscalationPolicyService) ListAlertEscalationPolicies(ctx context.Context, req *v1.ListAlertEscalationPoliciesRequest) (*v1.ListAlertEscalationPoliciesResponse, error) {
	items, total, pageNo, pageSize, err := s.gateway.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.AlertEscalationPolicyItem, 0, len(items))
	for i := range items {
		out = append(out, escalationPolicyItem(items[i]))
	}
	return &v1.ListAlertEscalationPoliciesResponse{
		Items: out, Total: int32(total), Page: int32(pageNo), PageSize: int32(pageSize),
	}, nil
}

func (s *AlertEscalationPolicyService) ListEventEscalations(ctx context.Context, req *v1.ListEventEscalationsRequest) (*v1.ListEventEscalationsResponse, error) {
	items, err := s.gateway.ListEventEscalations(ctx, req.GetEventUid())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.AlertEventEscalationItem, 0, len(items))
	for i := range items {
		out = append(out, eventEscalationItem(items[i]))
	}
	return &v1.ListEventEscalationsResponse{Items: out}, nil
}

func escalationPolicyInput(name, description, labels string, levelUIDs, datasourceUIDs, ruleUIDs []string, steps []*v1.AlertEscalationStepInput) bo.AlertEscalationPolicy {
	items := make([]bo.AlertEscalationStep, 0, len(steps))
	for _, step := range steps {
		items = append(items, bo.AlertEscalationStep{
			UID: step.GetUid(), DelayMinutes: int(step.GetDelayMinutes()),
			NotificationGroupUIDs: step.GetNotificationGroupUids(),
		})
	}
	return bo.AlertEscalationPolicy{
		Name: name, Description: description, Labels: labels,
		LevelUIDs: levelUIDs, DatasourceUIDs: datasourceUIDs, RuleUIDs: ruleUIDs,
		Steps: items,
	}
}

func escalationPolicyItem(in bo.AlertEscalationPolicy) *v1.AlertEscalationPolicyItem {
	steps := make([]*v1.AlertEscalationStepItem, 0, len(in.Steps))
	for _, step := range in.Steps {
		steps = append(steps, &v1.AlertEscalationStepItem{
			Uid: step.UID, StepOrder: int32(step.StepOrder), DelayMinutes: int32(step.DelayMinutes),
			NotificationGroupUids: step.NotificationGroupUIDs,
		})
	}
	return &v1.AlertEscalationPolicyItem{
		Uid: in.UID, Name: in.Name, Description: in.Description, Labels: in.Labels,
		LevelUids: in.LevelUIDs, DatasourceUids: in.DatasourceUIDs, RuleUids: in.RuleUIDs,
		Status: in.Status, CreatedBy: in.CreatedBy,
		CreatedAt: formatTime(in.CreatedAt), UpdatedAt: formatTime(in.UpdatedAt), Steps: steps,
	}
}

func eventEscalationItem(in bo.AlertEventEscalation) *v1.AlertEventEscalationItem {
	sentAt := ""
	if in.SentAt != nil {
		sentAt = formatTime(*in.SentAt)
	}
	return &v1.AlertEventEscalationItem{
		Uid: in.UID, EventUid: in.EventUID, PolicyUid: in.PolicyUID, PolicyName: in.PolicyName,
		StepOrder: int32(in.StepOrder), DelayMinutes: int32(in.DelayMinutes),
		NotificationGroupCount: int32(in.NotificationGroupCount),
		DueAt:                  formatTime(in.DueAt), Status: in.Status, SentAt: sentAt,
	}
}
