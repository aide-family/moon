package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

func NewStrategyService(strategyBiz *biz.StrategyBiz) *StrategyService {
	return &StrategyService{
		strategyBiz: strategyBiz,
	}
}

type StrategyService struct {
	apiv1.UnimplementedStrategyServer

	strategyBiz *biz.StrategyBiz
}

func (s *StrategyService) CreateStrategyGroup(ctx context.Context, req *apiv1.CreateStrategyGroupRequest) (*apiv1.CreateStrategyGroupReply, error) {
	if err := s.strategyBiz.CreateStrategyGroup(ctx, bo.NewCreateStrategyGroupBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.CreateStrategyGroupReply{}, nil
}

func (s *StrategyService) UpdateStrategyGroup(ctx context.Context, req *apiv1.UpdateStrategyGroupRequest) (*apiv1.UpdateStrategyGroupReply, error) {
	if err := s.strategyBiz.UpdateStrategyGroup(ctx, bo.NewUpdateStrategyGroupBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateStrategyGroupReply{}, nil
}

func (s *StrategyService) UpdateStrategyGroupStatus(ctx context.Context, req *apiv1.UpdateStrategyGroupStatusRequest) (*apiv1.UpdateStrategyGroupStatusReply, error) {
	if err := s.strategyBiz.UpdateStrategyGroupStatus(ctx, bo.NewUpdateStrategyGroupStatusBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateStrategyGroupStatusReply{}, nil
}

func (s *StrategyService) DeleteStrategyGroup(ctx context.Context, req *apiv1.DeleteStrategyGroupRequest) (*apiv1.DeleteStrategyGroupReply, error) {
	if err := s.strategyBiz.DeleteStrategyGroup(ctx, snowflake.ParseInt64(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteStrategyGroupReply{}, nil
}

func (s *StrategyService) GetStrategyGroup(ctx context.Context, req *apiv1.GetStrategyGroupRequest) (*apiv1.StrategyGroupItem, error) {
	item, err := s.strategyBiz.GetStrategyGroup(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return item.ToAPIV1StrategyGroupItem(), nil
}

func (s *StrategyService) ListStrategyGroup(ctx context.Context, req *apiv1.ListStrategyGroupRequest) (*apiv1.ListStrategyGroupReply, error) {
	result, err := s.strategyBiz.ListStrategyGroup(ctx, bo.NewListStrategyGroupBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListStrategyGroupReply(result), nil
}

func (s *StrategyService) SelectStrategyGroup(ctx context.Context, req *apiv1.SelectStrategyGroupRequest) (*apiv1.SelectStrategyGroupReply, error) {
	result, err := s.strategyBiz.SelectStrategyGroup(ctx, bo.NewSelectStrategyGroupBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectStrategyGroupReply(result), nil
}

func (s *StrategyService) StrategyGroupBindReceivers(ctx context.Context, req *apiv1.StrategyGroupBindReceiversRequest) (*apiv1.StrategyGroupBindReceiversReply, error) {
	if err := s.strategyBiz.StrategyGroupBindReceivers(ctx, bo.NewStrategyGroupBindReceiversBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.StrategyGroupBindReceiversReply{}, nil
}

func (s *StrategyService) CreateStrategy(ctx context.Context, req *apiv1.CreateStrategyRequest) (*apiv1.CreateStrategyReply, error) {
	if err := s.strategyBiz.CreateStrategy(ctx, bo.NewCreateStrategyBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.CreateStrategyReply{}, nil
}

func (s *StrategyService) UpdateStrategy(ctx context.Context, req *apiv1.UpdateStrategyRequest) (*apiv1.UpdateStrategyReply, error) {
	if err := s.strategyBiz.UpdateStrategy(ctx, bo.NewUpdateStrategyBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateStrategyReply{}, nil
}

func (s *StrategyService) UpdateStrategyStatus(ctx context.Context, req *apiv1.UpdateStrategyStatusRequest) (*apiv1.UpdateStrategyStatusReply, error) {
	if err := s.strategyBiz.UpdateStrategyStatus(ctx, bo.NewUpdateStrategyStatusBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateStrategyStatusReply{}, nil
}

func (s *StrategyService) DeleteStrategy(ctx context.Context, req *apiv1.DeleteStrategyRequest) (*apiv1.DeleteStrategyReply, error) {
	if err := s.strategyBiz.DeleteStrategy(ctx, snowflake.ParseInt64(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteStrategyReply{}, nil
}

func (s *StrategyService) GetStrategy(ctx context.Context, req *apiv1.GetStrategyRequest) (*apiv1.StrategyItem, error) {
	item, err := s.strategyBiz.GetStrategy(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return item.ToAPIV1StrategyItem(), nil
}

func (s *StrategyService) ListStrategy(ctx context.Context, req *apiv1.ListStrategyRequest) (*apiv1.ListStrategyReply, error) {
	result, err := s.strategyBiz.ListStrategy(ctx, bo.NewListStrategyBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListStrategyReply(result), nil
}
