package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

func NewLevelService(levelBiz *biz.LevelBiz) *LevelService {
	return &LevelService{
		levelBiz: levelBiz,
	}
}

type LevelService struct {
	apiv1.UnimplementedLevelServer

	levelBiz *biz.LevelBiz
}

func (s *LevelService) CreateLevel(ctx context.Context, req *apiv1.CreateLevelRequest) (*apiv1.CreateLevelReply, error) {
	createBo := bo.NewCreateLevelBo(req)
	if err := s.levelBiz.CreateLevel(ctx, createBo); err != nil {
		return nil, err
	}
	return &apiv1.CreateLevelReply{}, nil
}

func (s *LevelService) UpdateLevel(ctx context.Context, req *apiv1.UpdateLevelRequest) (*apiv1.UpdateLevelReply, error) {
	updateBo := bo.NewUpdateLevelBo(req)
	if err := s.levelBiz.UpdateLevel(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateLevelReply{}, nil
}

func (s *LevelService) UpdateLevelStatus(ctx context.Context, req *apiv1.UpdateLevelStatusRequest) (*apiv1.UpdateLevelStatusReply, error) {
	statusBo := bo.NewUpdateLevelStatusBo(req)
	if err := s.levelBiz.UpdateLevelStatus(ctx, statusBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateLevelStatusReply{}, nil
}

func (s *LevelService) DeleteLevel(ctx context.Context, req *apiv1.DeleteLevelRequest) (*apiv1.DeleteLevelReply, error) {
	if err := s.levelBiz.DeleteLevel(ctx, snowflake.ParseInt64(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteLevelReply{}, nil
}

func (s *LevelService) GetLevel(ctx context.Context, req *apiv1.GetLevelRequest) (*apiv1.LevelItem, error) {
	item, err := s.levelBiz.GetLevel(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return item.ToAPIV1LevelItem(), nil
}

func (s *LevelService) ListLevel(ctx context.Context, req *apiv1.ListLevelRequest) (*apiv1.ListLevelReply, error) {
	result, err := s.levelBiz.ListLevel(ctx, bo.NewListLevelBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListLevelReply(result), nil
}

func (s *LevelService) SelectLevel(ctx context.Context, req *apiv1.SelectLevelRequest) (*apiv1.SelectLevelReply, error) {
	result, err := s.levelBiz.SelectLevel(ctx, bo.NewSelectLevelBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectLevelReply(result), nil
}
