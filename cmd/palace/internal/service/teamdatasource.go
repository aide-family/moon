package service

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/permission"
	"github.com/moon-monitor/moon/cmd/palace/internal/service/build"
	com "github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/merr"
)

type TeamDatasourceService struct {
	palace.UnimplementedTeamDatasourceServer
	teamDatasourceBiz   *biz.TeamDatasource
	teamDatasourceQuery *biz.TeamDatasourceQuery
	helper              *log.Helper
}

func NewTeamDatasourceService(
	teamDatasourceBiz *biz.TeamDatasource,
	teamDatasourceQuery *biz.TeamDatasourceQuery,
	logger log.Logger,
) *TeamDatasourceService {
	return &TeamDatasourceService{
		teamDatasourceBiz:   teamDatasourceBiz,
		teamDatasourceQuery: teamDatasourceQuery,
		helper:              log.NewHelper(log.With(logger, "module", "service.datasource")),
	}
}

func (s *TeamDatasourceService) SaveTeamMetricDatasource(ctx context.Context, req *palace.SaveTeamMetricDatasourceRequest) (*common.EmptyReply, error) {
	params := build.ToSaveTeamMetricDatasourceRequest(req)
	if err := s.teamDatasourceBiz.SaveMetricDatasource(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "保存团队数据源成功"}, nil
}

func (s *TeamDatasourceService) UpdateTeamMetricDatasourceStatus(ctx context.Context, req *palace.UpdateTeamMetricDatasourceStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateTeamMetricDatasourceStatusRequest{
		DatasourceID: req.GetDatasourceId(),
		Status:       vobj.GlobalStatus(req.GetStatus()),
	}
	if err := s.teamDatasourceBiz.UpdateMetricDatasourceStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "更新团队数据源状态成功"}, nil
}

func (s *TeamDatasourceService) DeleteTeamMetricDatasource(ctx context.Context, req *palace.DeleteTeamMetricDatasourceRequest) (*common.EmptyReply, error) {
	if err := s.teamDatasourceBiz.DeleteMetricDatasource(ctx, req.GetDatasourceId()); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "删除团队数据源成功"}, nil
}

func (s *TeamDatasourceService) GetTeamMetricDatasource(ctx context.Context, req *palace.GetTeamMetricDatasourceRequest) (*common.TeamMetricDatasourceItem, error) {
	datasource, err := s.teamDatasourceBiz.GetMetricDatasource(ctx, req.GetDatasourceId())
	if err != nil {
		return nil, err
	}

	return build.ToTeamMetricDatasourceItem(datasource), nil
}

func (s *TeamDatasourceService) ListTeamMetricDatasource(ctx context.Context, req *palace.ListTeamMetricDatasourceRequest) (*palace.ListTeamMetricDatasourceReply, error) {
	params := build.ToListTeamMetricDatasourceRequest(req)
	datasourceReply, err := s.teamDatasourceBiz.ListMetricDatasource(ctx, params)
	if err != nil {
		return nil, err
	}

	return &palace.ListTeamMetricDatasourceReply{
		Pagination: build.ToPaginationReply(datasourceReply.PaginationReply),
		Items:      build.ToTeamMetricDatasourceItems(datasourceReply.Items),
	}, nil
}

func (s *TeamDatasourceService) SyncMetricMetadata(ctx context.Context, req *palace.SyncMetricMetadataRequest) (*common.EmptyReply, error) {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorBadRequest("请选择团队")
	}
	params := &bo.SyncMetricMetadataRequest{
		DatasourceID: req.GetDatasourceId(),
		TeamID:       teamID,
	}
	if err := s.teamDatasourceBiz.SyncMetricMetadata(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "数据源元数据同步中，请稍后刷新页面查看"}, nil
}

func (s *TeamDatasourceService) MetricDatasourceQuery(ctx context.Context, req *palace.MetricDatasourceQueryRequest) (*com.MetricDatasourceQueryReply, error) {
	datasource, err := s.teamDatasourceBiz.GetMetricDatasource(ctx, req.GetDatasourceId())
	if err != nil {
		return nil, err
	}

	params := &bo.MetricDatasourceQueryRequest{
		Datasource: datasource,
		Expr:       req.GetExpr(),
		Time:       req.GetTime(),
		StartTime:  req.GetStartTime(),
		EndTime:    req.GetEndTime(),
		Step:       req.GetStep(),
	}
	reply, err := s.teamDatasourceQuery.MetricDatasourceQuery(ctx, params)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

const (
	OperationTeamDatasourceMetricDatasourceProxy = "/api.palace.TeamDatasource/MetricDatasourceProxy"
)

func (s *TeamDatasourceService) MetricDatasourceProxy(httpCtx http.Context) error {
	isContentType := false
	for k := range httpCtx.Request().Header {
		if strings.EqualFold(k, "Content-Type") {
			isContentType = true
			break
		}
	}
	if !isContentType {
		httpCtx.Header().Set("Content-Type", "application/json")
		httpCtx.Request().Header.Set("Content-Type", "application/json")
	}
	var in palace.MetricDatasourceProxyRequest
	if err := httpCtx.Bind(&in); err != nil {
		return err
	}
	if err := httpCtx.BindQuery(&in); err != nil {
		return err
	}
	if err := httpCtx.BindVars(&in); err != nil {
		return err
	}
	http.SetOperation(httpCtx, OperationTeamDatasourceMetricDatasourceProxy)

	h := httpCtx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		ctx = permission.WithTeamIDContext(ctx, in.GetTeamId())
		datasourceDo, err := s.teamDatasourceBiz.GetMetricDatasource(ctx, in.GetDatasourceId())
		if err != nil {
			return nil, err
		}

		datasource, err := build.ToMetricDatasource(datasourceDo, s.helper.Logger())
		if err != nil {
			return nil, err
		}

		return nil, datasource.Proxy(httpCtx, in.GetTarget())
	})
	_, err := h(httpCtx, &in)
	if err != nil {
		return err
	}
	return nil
}
