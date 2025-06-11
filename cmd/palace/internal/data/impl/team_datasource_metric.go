package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/data"
	"github.com/moon-monitor/moon/pkg/util/crypto"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func NewTeamMetricDatasourceRepo(data *data.Data, logger log.Logger) repository.TeamDatasourceMetric {
	return &teamMetricDatasourceImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.team.datasource.metric")),
	}
}

type teamMetricDatasourceImpl struct {
	*data.Data
	helper *log.Helper
}

func (t *teamMetricDatasourceImpl) Create(ctx context.Context, req *bo.SaveTeamMetricDatasource) error {
	metricDatasourceDo := &team.DatasourceMetric{
		Name:           req.Name,
		Status:         req.Status,
		Remark:         req.Remark,
		Driver:         req.Driver,
		Endpoint:       crypto.String(req.Endpoint),
		ScrapeInterval: req.ScrapeInterval,
		Headers:        crypto.NewObject(req.Headers),
		QueryMethod:    req.QueryMethod,
		CA:             crypto.String(req.CA),
		TLS:            crypto.NewObject(req.TLS),
		BasicAuth:      crypto.NewObject(req.BasicAuth),
		Extra:          req.Extra,
	}
	metricDatasourceDo.WithContext(ctx)
	bizMutation := getTeamBizQuery(ctx, t)
	return bizMutation.DatasourceMetric.WithContext(ctx).Create(metricDatasourceDo)
}

func (t *teamMetricDatasourceImpl) Update(ctx context.Context, req *bo.SaveTeamMetricDatasource) error {
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizMutation.DatasourceMetric
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamId),
		mutation.ID.Eq(req.ID),
	}
	mutations := []field.AssignExpr{
		mutation.Name.Value(req.Name),
		mutation.Status.Value(req.Status.GetValue()),
		mutation.Remark.Value(req.Remark),
		mutation.Driver.Value(req.Driver.GetValue()),
		mutation.Endpoint.Value(crypto.String(req.Endpoint)),
		mutation.ScrapeInterval.Value(int64(req.ScrapeInterval)),
		mutation.Headers.Value(crypto.NewObject(req.Headers)),
		mutation.QueryMethod.Value(req.QueryMethod.GetValue()),
		mutation.Extra.Value(req.Extra),
	}
	if validate.TextIsNotNull(req.CA) {
		mutations = append(mutations, mutation.CA.Value(crypto.String(req.CA)))
	}
	if validate.IsNotNil(req.TLS) {
		mutations = append(mutations, mutation.TLS.Value(crypto.NewObject(req.TLS)))
	}
	if validate.IsNotNil(req.BasicAuth) {
		mutations = append(mutations, mutation.BasicAuth.Value(crypto.NewObject(req.BasicAuth)))
	}
	_, err := mutation.WithContext(ctx).Where(wrapper...).UpdateSimple(mutations...)
	return err
}

func (t *teamMetricDatasourceImpl) UpdateStatus(ctx context.Context, req *bo.UpdateTeamMetricDatasourceStatusRequest) error {
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizMutation.DatasourceMetric
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamId),
		mutation.ID.Eq(req.DatasourceID),
	}
	_, err := mutation.WithContext(ctx).Where(wrapper...).UpdateSimple(mutation.Status.Value(req.Status.GetValue()))
	return err
}

func (t *teamMetricDatasourceImpl) Delete(ctx context.Context, datasourceID uint32) error {
	bizMutation, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizMutation.DatasourceMetric
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamId),
		mutation.ID.Eq(datasourceID),
	}
	_, err := mutation.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

func (t *teamMetricDatasourceImpl) Get(ctx context.Context, datasourceID uint32) (do.DatasourceMetric, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizQuery.DatasourceMetric
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamId),
		mutation.ID.Eq(datasourceID),
	}
	datasource, err := mutation.WithContext(ctx).Where(wrapper...).First()
	if err != nil {
		return nil, datasourceNotFound(err)
	}
	return datasource, nil
}

func (t *teamMetricDatasourceImpl) List(ctx context.Context, req *bo.ListTeamMetricDatasource) (*bo.ListTeamMetricDatasourceReply, error) {
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizQuery.DatasourceMetric
	wrapper := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamId))

	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(mutation.Status.Eq(req.Status.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		ors := []gen.Condition{
			mutation.Name.Like(req.Keyword),
			mutation.Remark.Like(req.Keyword),
			mutation.Endpoint.Eq(crypto.String(req.Keyword)),
		}
		wrapper = wrapper.Where(mutation.Or(ors...))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	datasourceDos, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	return req.ToListTeamMetricDatasourceReply(datasourceDos), nil
}

func (t *teamMetricDatasourceImpl) FindByIds(ctx context.Context, datasourceIds []uint32) ([]do.DatasourceMetric, error) {
	if len(datasourceIds) == 0 {
		return nil, nil
	}
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t)
	mutation := bizQuery.DatasourceMetric
	wrapper := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamId), mutation.ID.In(datasourceIds...))
	rows, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(rows, func(row *team.DatasourceMetric) do.DatasourceMetric { return row }), nil
}
