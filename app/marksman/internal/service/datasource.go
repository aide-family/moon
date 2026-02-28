package service

import (
	"context"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
	"github.com/bwmarrin/snowflake"
)

func NewDatasourceService(datasourceBiz *biz.DatasourceBiz) *DatasourceService {
	return &DatasourceService{
		datasourceBiz: datasourceBiz,
	}
}

type DatasourceService struct {
	apiv1.UnimplementedDatasourceServer

	datasourceBiz *biz.DatasourceBiz
}

func (s *DatasourceService) CreateDatasource(ctx context.Context, req *apiv1.CreateDatasourceRequest) (*apiv1.CreateDatasourceReply, error) {
	createBo := bo.NewCreateDatasourceBo(req)
	if err := s.datasourceBiz.CreateDatasource(ctx, createBo); err != nil {
		return nil, err
	}
	return &apiv1.CreateDatasourceReply{}, nil
}

func (s *DatasourceService) UpdateDatasource(ctx context.Context, req *apiv1.UpdateDatasourceRequest) (*apiv1.UpdateDatasourceReply, error) {
	updateBo := bo.NewUpdateDatasourceBo(req)
	if err := s.datasourceBiz.UpdateDatasource(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateDatasourceReply{}, nil
}

func (s *DatasourceService) DeleteDatasource(ctx context.Context, req *apiv1.DeleteDatasourceRequest) (*apiv1.DeleteDatasourceReply, error) {
	if err := s.datasourceBiz.DeleteDatasource(ctx, snowflake.ParseInt64(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteDatasourceReply{}, nil
}

func (s *DatasourceService) GetDatasource(ctx context.Context, req *apiv1.GetDatasourceRequest) (*apiv1.DatasourceItem, error) {
	item, err := s.datasourceBiz.GetDatasource(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return item.ToAPIV1DatasourceItem(), nil
}

func (s *DatasourceService) ListDatasource(ctx context.Context, req *apiv1.ListDatasourceRequest) (*apiv1.ListDatasourceReply, error) {
	result, err := s.datasourceBiz.ListDatasource(ctx, bo.NewListDatasourceBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListDatasourceReply(result), nil
}
