package alarm

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/alarm"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
)

// TimeEngineRuleService 时间引擎规则模块 service
type TimeEngineRuleService struct {
	pb.UnimplementedTimeEngineRuleServer

	timeEngineRuleBiz *biz.TimeEngineRuleBiz
}

// NewTimeEngineRuleService 创建时间引擎规则模块 service
func NewTimeEngineRuleService(timeEngineRuleBiz *biz.TimeEngineRuleBiz) *TimeEngineRuleService {
	return &TimeEngineRuleService{
		timeEngineRuleBiz: timeEngineRuleBiz,
	}
}

// CreateTimeEngineRule 创建时间引擎规则
func (s *TimeEngineRuleService) CreateTimeEngineRule(ctx context.Context, req *pb.CreateTimeEngineRuleRequest) (*pb.CreateTimeEngineRuleReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineRuleModuleBuilder().WithCreateTimeEngineRuleRequest(req).ToBo()
	if err := s.timeEngineRuleBiz.CreateTimeEngineRule(ctx, params); err != nil {
		return nil, err
	}
	return &pb.CreateTimeEngineRuleReply{}, nil
}

// UpdateTimeEngineRule 更新时间引擎规则
func (s *TimeEngineRuleService) UpdateTimeEngineRule(ctx context.Context, req *pb.UpdateTimeEngineRuleRequest) (*pb.UpdateTimeEngineRuleReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineRuleModuleBuilder().WithUpdateTimeEngineRuleRequest(req).ToBo()
	if err := s.timeEngineRuleBiz.UpdateTimeEngineRule(ctx, params); err != nil {
		return nil, err
	}
	return &pb.UpdateTimeEngineRuleReply{}, nil
}

// DeleteTimeEngineRule 删除时间引擎规则
func (s *TimeEngineRuleService) DeleteTimeEngineRule(ctx context.Context, req *pb.DeleteTimeEngineRuleRequest) (*pb.DeleteTimeEngineRuleReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineRuleModuleBuilder().WithDeleteTimeEngineRuleRequest(req).ToBo()
	if err := s.timeEngineRuleBiz.DeleteTimeEngineRule(ctx, params); err != nil {
		return nil, err
	}
	return &pb.DeleteTimeEngineRuleReply{}, nil
}

// GetTimeEngineRule 获取时间引擎规则
func (s *TimeEngineRuleService) GetTimeEngineRule(ctx context.Context, req *pb.GetTimeEngineRuleRequest) (*pb.GetTimeEngineRuleReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineRuleModuleBuilder().WithGetTimeEngineRuleRequest(req).ToBo()
	timeEngineRule, err := s.timeEngineRuleBiz.GetTimeEngineRule(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.GetTimeEngineRuleReply{
		Detail: builder.NewParamsBuild(ctx).TimeEngineRuleModuleBuilder().Do().ToAPI(timeEngineRule),
	}, nil
}

// ListTimeEngineRule 获取时间引擎规则列表
func (s *TimeEngineRuleService) ListTimeEngineRule(ctx context.Context, req *pb.ListTimeEngineRuleRequest) (*pb.ListTimeEngineRuleReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineRuleModuleBuilder().WithListTimeEngineRuleRequest(req).ToBo()
	timeEngineRules, err := s.timeEngineRuleBiz.ListTimeEngineRule(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.ListTimeEngineRuleReply{
		List:       builder.NewParamsBuild(ctx).TimeEngineRuleModuleBuilder().Do().ToAPIs(timeEngineRules),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}
