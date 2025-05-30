package service

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/api/palace"
	palacecommon "github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/merr"
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

func (s *TeamDatasourceService) SaveTeamMetricDatasource(ctx context.Context, req *palace.SaveTeamMetricDatasourceRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToSaveTeamMetricDatasourceRequest(req)
	if err := s.teamDatasourceBiz.SaveMetricDatasource(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamDatasourceService) UpdateTeamMetricDatasourceStatus(ctx context.Context, req *palace.UpdateTeamMetricDatasourceStatusRequest) (*palacecommon.EmptyReply, error) {
	params := &bo.UpdateTeamMetricDatasourceStatusRequest{
		DatasourceID: req.GetDatasourceId(),
		Status:       vobj.GlobalStatus(req.GetStatus()),
	}
	if err := s.teamDatasourceBiz.UpdateMetricDatasourceStatus(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamDatasourceService) DeleteTeamMetricDatasource(ctx context.Context, req *palace.DeleteTeamMetricDatasourceRequest) (*palacecommon.EmptyReply, error) {
	if err := s.teamDatasourceBiz.DeleteMetricDatasource(ctx, req.GetDatasourceId()); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamDatasourceService) GetTeamMetricDatasource(ctx context.Context, req *palace.GetTeamMetricDatasourceRequest) (*palacecommon.TeamMetricDatasourceItem, error) {
	datasource, err := s.teamDatasourceBiz.GetMetricDatasource(ctx, req.GetDatasourceId())
	if err != nil {
		return nil, err
	}

	return build.ToTeamMetricDatasourceItem(datasource), nil
}

func (s *TeamDatasourceService) DatasourceSelect(ctx context.Context, req *palace.DatasourceSelectRequest) (*palace.DatasourceSelectReply, error) {
	params := build.ToDatasourceSelectRequest(req)
	datasourceReply, err := s.teamDatasourceBiz.DatasourceSelect(ctx, params)
	if err != nil {
		return nil, err
	}

	return &palace.DatasourceSelectReply{
		Pagination: build.ToPaginationReply(datasourceReply.PaginationReply),
		Items:      build.ToSelectItems(datasourceReply.Items),
	}, nil
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

func (s *TeamDatasourceService) SyncMetricMetadata(ctx context.Context, req *palace.SyncMetricMetadataRequest) (*palacecommon.EmptyReply, error) {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorBadRequest("please select a team")
	}
	params := &bo.SyncMetricMetadataRequest{
		DatasourceID: req.GetDatasourceId(),
		TeamID:       teamID,
	}
	if err := s.teamDatasourceBiz.SyncMetricMetadata(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamDatasourceService) MetricDatasourceQuery(ctx context.Context, req *palace.MetricDatasourceQueryRequest) (*common.MetricDatasourceQueryReply, error) {
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

func (s *TeamDatasourceService) GetMetricDatasourceMetadata(ctx context.Context, req *palace.GetMetricDatasourceMetadataRequest) (*palacecommon.TeamMetricDatasourceMetadataItem, error) {
	params := &bo.GetMetricDatasourceMetadataRequest{
		DatasourceID: req.GetDatasourceId(),
		ID:           req.GetMetadataId(),
	}
	metadata, err := s.teamDatasourceBiz.GetMetricDatasourceMetadata(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToTeamMetricDatasourceMetadataItem(metadata), nil
}

func (s *TeamDatasourceService) ListMetricDatasourceMetadata(ctx context.Context, req *palace.ListMetricDatasourceMetadataRequest) (*palace.ListMetricDatasourceMetadataReply, error) {
	params := build.ToListMetricDatasourceMetadataRequest(req)
	metadata, err := s.teamDatasourceBiz.ListMetricDatasourceMetadata(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.ListMetricDatasourceMetadataReply{
		Pagination: build.ToPaginationReply(metadata.PaginationReply),
		Items:      build.ToTeamMetricDatasourceMetadataItems(metadata.Items),
	}, nil
}

func (s *TeamDatasourceService) UpdateMetricDatasourceMetadata(ctx context.Context, req *palace.UpdateMetricDatasourceMetadataRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToUpdateMetricDatasourceMetadataRequest(req)
	if err := s.teamDatasourceBiz.UpdateMetricDatasourceMetadata(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamDatasourceService) MetricDatasourceProxyHandler(httpCtx http.Context) error {
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
	http.SetOperation(httpCtx, palace.OperationTeamDatasourceMetricDatasourceProxy)

	h := httpCtx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		datasourceDo, err := s.teamDatasourceBiz.GetMetricDatasource(ctx, in.GetDatasourceId())
		if err != nil {
			return nil, err
		}

		datasource, err := bo.ToMetricDatasource(datasourceDo, s.helper.Logger())
		if err != nil {
			return nil, err
		}

		return nil, datasource.Proxy(httpCtx, in.GetTarget())
	})

	if _, err := h(httpCtx, &in); err != nil {
		return err
	}
	return nil
}
