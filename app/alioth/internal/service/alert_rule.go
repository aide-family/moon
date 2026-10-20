package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertrule"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type AlertRuleService struct {
	v1.UnimplementedAlertRuleServiceServer
	rules *alertrule.Usecase
}

func NewAlertRuleService(uc *alertrule.Usecase) *AlertRuleService {
	return &AlertRuleService{rules: uc}
}

func (s *AlertRuleService) CreateAlertRuleGroup(ctx context.Context, req *v1.CreateAlertRuleGroupRequest) (*v1.CreateAlertRuleGroupResponse, error) {
	item, err := s.rules.CreateGroup(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.CreateAlertRuleGroupResponse{Group: groupItem(*item)}, nil
}

func (s *AlertRuleService) UpdateAlertRuleGroup(ctx context.Context, req *v1.UpdateAlertRuleGroupRequest) (*v1.UpdateAlertRuleGroupResponse, error) {
	item, err := s.rules.UpdateGroup(ctx, req.GetUid(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertRuleGroupResponse{Group: groupItem(*item)}, nil
}

func (s *AlertRuleService) UpdateAlertRuleGroupStatus(ctx context.Context, req *v1.UpdateAlertRuleGroupStatusRequest) (*v1.UpdateAlertRuleGroupStatusResponse, error) {
	item, err := s.rules.UpdateGroupStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertRuleGroupStatusResponse{Group: groupItem(*item)}, nil
}

func (s *AlertRuleService) DeleteAlertRuleGroup(ctx context.Context, req *v1.DeleteAlertRuleGroupRequest) (*v1.DeleteAlertRuleGroupResponse, error) {
	if err := s.rules.DeleteGroup(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteAlertRuleGroupResponse{}, nil
}

func (s *AlertRuleService) ListAlertRuleGroups(ctx context.Context, req *v1.ListAlertRuleGroupsRequest) (*v1.ListAlertRuleGroupsResponse, error) {
	items, total, page, pageSize, err := s.rules.ListGroups(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListAlertRuleGroupsResponse{Items: groupItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *AlertRuleService) CreateAlertRule(ctx context.Context, req *v1.CreateAlertRuleRequest) (*v1.CreateAlertRuleResponse, error) {
	item, err := s.rules.CreateRule(ctx, ruleInput("", req))
	if err != nil {
		return nil, err
	}
	return &v1.CreateAlertRuleResponse{Rule: ruleItem(*item)}, nil
}

func (s *AlertRuleService) UpdateAlertRule(ctx context.Context, req *v1.UpdateAlertRuleRequest) (*v1.UpdateAlertRuleResponse, error) {
	in := bo.AlertRule{
		UID: req.GetUid(), GroupUID: req.GetGroupUid(), Name: req.GetName(), Type: req.GetType(), Driver: req.GetDriver(),
		Expr: req.GetExpr(), Labels: req.GetLabels(), DatasourceLabels: req.GetDatasourceLabels(),
		Summary: req.GetSummary(), Description: req.GetDescription(), Levels: levelInputs(req.GetLevels()),
	}
	item, err := s.rules.UpdateRule(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertRuleResponse{Rule: ruleItem(*item)}, nil
}

func (s *AlertRuleService) UpdateAlertRuleStatus(ctx context.Context, req *v1.UpdateAlertRuleStatusRequest) (*v1.UpdateAlertRuleStatusResponse, error) {
	item, err := s.rules.UpdateRuleStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertRuleStatusResponse{Rule: ruleItem(*item)}, nil
}

func (s *AlertRuleService) DeleteAlertRule(ctx context.Context, req *v1.DeleteAlertRuleRequest) (*v1.DeleteAlertRuleResponse, error) {
	if err := s.rules.DeleteRule(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteAlertRuleResponse{}, nil
}

func (s *AlertRuleService) GetAlertRule(ctx context.Context, req *v1.GetAlertRuleRequest) (*v1.GetAlertRuleResponse, error) {
	item, err := s.rules.GetRule(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertRuleResponse{Rule: ruleItem(*item)}, nil
}

func (s *AlertRuleService) ListAlertRules(ctx context.Context, req *v1.ListAlertRulesRequest) (*v1.ListAlertRulesResponse, error) {
	items, total, page, pageSize, err := s.rules.ListRules(ctx, req.GetPage(), req.GetPageSize(), req.GetGroupUid(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListAlertRulesResponse{Items: ruleItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func ruleInput(uid string, req *v1.CreateAlertRuleRequest) bo.AlertRule {
	return bo.AlertRule{
		UID: uid, GroupUID: req.GetGroupUid(), Name: req.GetName(), Type: req.GetType(), Driver: req.GetDriver(),
		Expr: req.GetExpr(), Labels: req.GetLabels(), DatasourceLabels: req.GetDatasourceLabels(),
		Summary: req.GetSummary(), Description: req.GetDescription(), Levels: levelInputs(req.GetLevels()),
	}
}
