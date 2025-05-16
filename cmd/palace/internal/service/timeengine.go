package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	api "github.com/aide-family/moon/pkg/api/palace"
	palacecommon "github.com/aide-family/moon/pkg/api/palace/common"
)

type TimeEngineService struct {
	api.UnimplementedTimeEngineServer

	timeEngineBiz *biz.TimeEngine
}

func NewTimeEngineService(timeEngineBiz *biz.TimeEngine) *TimeEngineService {
	return &TimeEngineService{
		timeEngineBiz: timeEngineBiz,
	}
}

func (s *TimeEngineService) SaveTimeEngine(ctx context.Context, req *api.SaveTimeEngineRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToSaveTimeEngineRequest(req)
	if err := s.timeEngineBiz.SaveTimeEngine(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TimeEngineService) UpdateTimeEngineStatus(ctx context.Context, req *api.UpdateTimeEngineStatusRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToUpdateTimeEngineStatusRequest(req)
	if err := s.timeEngineBiz.UpdateTimeEngineStatus(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TimeEngineService) DeleteTimeEngine(ctx context.Context, req *api.DeleteTimeEngineRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToDeleteTimeEngineRequest(req)
	if err := s.timeEngineBiz.DeleteTimeEngine(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TimeEngineService) GetTimeEngine(ctx context.Context, req *api.GetTimeEngineRequest) (*palacecommon.TimeEngineItem, error) {
	params := build.ToGetTimeEngineRequest(req)
	timeEngineDo, err := s.timeEngineBiz.GetTimeEngine(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToTimeEngineItem(timeEngineDo), nil
}

func (s *TimeEngineService) ListTimeEngine(ctx context.Context, req *api.ListTimeEngineRequest) (*api.ListTimeEngineReply, error) {
	params := build.ToListTimeEngineRequest(req)
	reply, err := s.timeEngineBiz.ListTimeEngine(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToListTimeEngineReply(reply), nil
}

func (s *TimeEngineService) SaveTimeEngineRule(ctx context.Context, req *api.SaveTimeEngineRuleRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToSaveTimeEngineRuleRequest(req)
	if err := s.timeEngineBiz.SaveTimeEngineRule(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TimeEngineService) UpdateTimeEngineRuleStatus(ctx context.Context, req *api.UpdateTimeEngineRuleStatusRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToUpdateTimeEngineRuleStatusRequest(req)
	if err := s.timeEngineBiz.UpdateTimeEngineRuleStatus(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TimeEngineService) DeleteTimeEngineRule(ctx context.Context, req *api.DeleteTimeEngineRuleRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToDeleteTimeEngineRuleRequest(req)
	if err := s.timeEngineBiz.DeleteTimeEngineRule(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TimeEngineService) GetTimeEngineRule(ctx context.Context, req *api.GetTimeEngineRuleRequest) (*palacecommon.TimeEngineItemRule, error) {
	params := build.ToGetTimeEngineRuleRequest(req)
	timeEngineRuleDo, err := s.timeEngineBiz.GetTimeEngineRule(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToTimeEngineItemRule(timeEngineRuleDo), nil
}

func (s *TimeEngineService) ListTimeEngineRule(ctx context.Context, req *api.ListTimeEngineRuleRequest) (*api.ListTimeEngineRuleReply, error) {
	params := build.ToListTimeEngineRuleRequest(req)
	reply, err := s.timeEngineBiz.ListTimeEngineRule(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToListTimeEngineRuleReply(reply), nil
}
