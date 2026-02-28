package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

func NewStrategyMetricService(strategyMetricBiz *biz.StrategyMetricBiz) *StrategyMetricService {
	return &StrategyMetricService{
		strategyMetricBiz: strategyMetricBiz,
	}
}

type StrategyMetricService struct {
	apiv1.UnimplementedStrategyMetricServer

	strategyMetricBiz *biz.StrategyMetricBiz
}

func (s *StrategyMetricService) SaveStrategyMetric(ctx context.Context, req *apiv1.SaveStrategyMetricRequest) (*apiv1.SaveStrategyMetricReply, error) {
	if err := s.strategyMetricBiz.SaveStrategyMetric(ctx, bo.NewSaveStrategyMetricBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.SaveStrategyMetricReply{}, nil
}

func (s *StrategyMetricService) GetStrategyMetric(ctx context.Context, req *apiv1.GetStrategyMetricRequest) (*apiv1.StrategyMetricItem, error) {
	item, err := s.strategyMetricBiz.GetStrategyMetric(ctx, snowflake.ParseInt64(req.GetStrategyUID()))
	if err != nil {
		return nil, err
	}
	return item.ToAPIV1StrategyMetricItem(), nil
}

func (s *StrategyMetricService) SaveStrategyMetricLevel(ctx context.Context, req *apiv1.SaveStrategyMetricLevelRequest) (*apiv1.SaveStrategyMetricLevelReply, error) {
	if err := s.strategyMetricBiz.SaveStrategyMetricLevel(ctx, bo.NewSaveStrategyMetricLevelBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.SaveStrategyMetricLevelReply{}, nil
}

func (s *StrategyMetricService) UpdateStrategyMetricLevelStatus(ctx context.Context, req *apiv1.UpdateStrategyMetricLevelStatusRequest) (*apiv1.UpdateStrategyMetricLevelStatusReply, error) {
	if err := s.strategyMetricBiz.UpdateStrategyMetricLevelStatus(ctx, bo.NewUpdateStrategyMetricLevelStatusBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.UpdateStrategyMetricLevelStatusReply{}, nil
}

func (s *StrategyMetricService) DeleteStrategyMetricLevel(ctx context.Context, req *apiv1.DeleteStrategyMetricLevelRequest) (*apiv1.DeleteStrategyMetricLevelReply, error) {
	if err := s.strategyMetricBiz.DeleteStrategyMetricLevel(ctx, snowflake.ParseInt64(req.GetUid()), snowflake.ParseInt64(req.GetStrategyUID())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteStrategyMetricLevelReply{}, nil
}

func (s *StrategyMetricService) GetStrategyMetricLevel(ctx context.Context, req *apiv1.GetStrategyMetricLevelRequest) (*apiv1.StrategyMetricLevelItem, error) {
	item, err := s.strategyMetricBiz.GetStrategyMetricLevel(ctx, snowflake.ParseInt64(req.GetUid()), snowflake.ParseInt64(req.GetStrategyUID()))
	if err != nil {
		return nil, err
	}
	return item.ToAPIV1StrategyMetricLevelItem(), nil
}

func (s *StrategyMetricService) StrategyMetricBindReceivers(ctx context.Context, req *apiv1.StrategyMetricBindReceiversRequest) (*apiv1.StrategyMetricBindReceiversReply, error) {
	if err := s.strategyMetricBiz.StrategyMetricBindReceivers(ctx, bo.NewStrategyMetricBindReceiversBo(req)); err != nil {
		return nil, err
	}
	return &apiv1.StrategyMetricBindReceiversReply{}, nil
}
