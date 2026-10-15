package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/alertlevel"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type AlertLevelService struct {
	v1.UnimplementedAlertLevelServiceServer
	levels *alertlevel.Usecase
}

func NewAlertLevelService(uc *alertlevel.Usecase) *AlertLevelService {
	return &AlertLevelService{levels: uc}
}

func (s *AlertLevelService) CreateAlertLevel(ctx context.Context, req *v1.CreateAlertLevelRequest) (*v1.CreateAlertLevelResponse, error) {
	item, err := s.levels.CreateAlertLevel(ctx, bo.AlertLevel{Name: req.GetName(), Color: req.GetColor(), Type: req.GetType()})
	if err != nil {
		return nil, err
	}
	return &v1.CreateAlertLevelResponse{Level: levelItem(*item)}, nil
}

func (s *AlertLevelService) UpdateAlertLevel(ctx context.Context, req *v1.UpdateAlertLevelRequest) (*v1.UpdateAlertLevelResponse, error) {
	item, err := s.levels.UpdateAlertLevel(ctx, bo.AlertLevel{UID: req.GetUid(), Name: req.GetName(), Color: req.GetColor()})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertLevelResponse{Level: levelItem(*item)}, nil
}

func (s *AlertLevelService) UpdateAlertLevelStatus(ctx context.Context, req *v1.UpdateAlertLevelStatusRequest) (*v1.UpdateAlertLevelStatusResponse, error) {
	item, err := s.levels.UpdateAlertLevelStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAlertLevelStatusResponse{Level: levelItem(*item)}, nil
}

func (s *AlertLevelService) DeleteAlertLevel(ctx context.Context, req *v1.DeleteAlertLevelRequest) (*v1.DeleteAlertLevelResponse, error) {
	if err := s.levels.DeleteAlertLevel(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteAlertLevelResponse{}, nil
}

func (s *AlertLevelService) ListAlertLevels(ctx context.Context, req *v1.ListAlertLevelsRequest) (*v1.ListAlertLevelsResponse, error) {
	items, total, page, pageSize, err := s.levels.ListAlertLevels(ctx, req.GetPage(), req.GetPageSize(), req.GetType(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListAlertLevelsResponse{Items: levelItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}
