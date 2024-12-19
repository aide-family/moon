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

// BatchUpdateTimeEngineRuleStatus 批量更新时间引擎规则状态
func (s *TimeEngineRuleService) BatchUpdateTimeEngineRuleStatus(ctx context.Context, req *pb.BatchUpdateTimeEngineRuleStatusRequest) (*pb.BatchUpdateTimeEngineRuleStatusReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineRuleModuleBuilder().WithBatchUpdateTimeEngineRuleStatusRequest(req).ToBo()
	if err := s.timeEngineRuleBiz.BatchUpdateTimeEngineRuleStatus(ctx, params); err != nil {
		return nil, err
	}
	return &pb.BatchUpdateTimeEngineRuleStatusReply{}, nil
}

// CreateTimeEngine 创建时间引擎
func (s *TimeEngineRuleService) CreateTimeEngine(ctx context.Context, req *pb.CreateTimeEngineRequest) (*pb.CreateTimeEngineReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineModuleBuilder().WithCreateTimeEngineRequest(req).ToBo()
	if err := s.timeEngineRuleBiz.CreateTimeEngine(ctx, params); err != nil {
		return nil, err
	}
	return &pb.CreateTimeEngineReply{}, nil
}

// UpdateTimeEngine 更新时间引擎
func (s *TimeEngineRuleService) UpdateTimeEngine(ctx context.Context, req *pb.UpdateTimeEngineRequest) (*pb.UpdateTimeEngineReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineModuleBuilder().WithUpdateTimeEngineRequest(req).ToBo()
	if err := s.timeEngineRuleBiz.UpdateTimeEngine(ctx, params); err != nil {
		return nil, err
	}
	return &pb.UpdateTimeEngineReply{}, nil
}

// DeleteTimeEngine 删除时间引擎
func (s *TimeEngineRuleService) DeleteTimeEngine(ctx context.Context, req *pb.DeleteTimeEngineRequest) (*pb.DeleteTimeEngineReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineModuleBuilder().WithDeleteTimeEngineRequest(req).ToBo()
	if err := s.timeEngineRuleBiz.DeleteTimeEngine(ctx, params); err != nil {
		return nil, err
	}
	return &pb.DeleteTimeEngineReply{}, nil
}

// GetTimeEngine 获取时间引擎
func (s *TimeEngineRuleService) GetTimeEngine(ctx context.Context, req *pb.GetTimeEngineRequest) (*pb.GetTimeEngineReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineModuleBuilder().WithGetTimeEngineRequest(req).ToBo()
	timeEngine, err := s.timeEngineRuleBiz.GetTimeEngine(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.GetTimeEngineReply{
		Detail: builder.NewParamsBuild(ctx).TimeEngineModuleBuilder().Do().ToAPI(timeEngine),
	}, nil
}

// ListTimeEngine 获取时间引擎列表
func (s *TimeEngineRuleService) ListTimeEngine(ctx context.Context, req *pb.ListTimeEngineRequest) (*pb.ListTimeEngineReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineModuleBuilder().WithListTimeEngineRequest(req).ToBo()
	timeEngines, err := s.timeEngineRuleBiz.ListTimeEngine(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.ListTimeEngineReply{
		List:       builder.NewParamsBuild(ctx).TimeEngineModuleBuilder().Do().ToAPIs(timeEngines),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}

// BatchUpdateTimeEngineStatus 批量更新时间引擎状态
func (s *TimeEngineRuleService) BatchUpdateTimeEngineStatus(ctx context.Context, req *pb.BatchUpdateTimeEngineStatusRequest) (*pb.BatchUpdateTimeEngineStatusReply, error) {
	params := builder.NewParamsBuild(ctx).TimeEngineModuleBuilder().WithBatchUpdateTimeEngineStatusRequest(req).ToBo()
	if err := s.timeEngineRuleBiz.BatchUpdateTimeEngineStatus(ctx, params); err != nil {
		return nil, err
	}
	return &pb.BatchUpdateTimeEngineStatusReply{}, nil
}
