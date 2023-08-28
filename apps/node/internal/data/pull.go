package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/api"
	"prometheus-manager/api/strategy/v1/pull"

	"prometheus-manager/apps/node/internal/biz"
	"prometheus-manager/apps/node/internal/conf"
)

type (
	PullRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IPullRepo = (*PullRepo)(nil)

func NewPullRepo(data *Data, logger log.Logger) *PullRepo {
	return &PullRepo{data: data, logger: log.NewHelper(log.With(logger, "module", pullModuleName))}
}

func (l *PullRepo) Datasources(ctx context.Context) (*pull.DatasourcesReply, error) {
	ctx, span := otel.Tracer(pullModuleName).Start(ctx, "PullRepo.Datasources")
	defer span.End()

	datasource := conf.Get().GetStrategy().GetPromDatasources()
	promDatasource := make([]*api.Datasource, 0, len(datasource))
	for _, v := range datasource {
		promDatasource = append(promDatasource, &api.Datasource{
			Name:   v.Name,
			Type:   v.Type,
			Url:    v.Url,
			Access: v.Access,
		})
	}

	return &pull.DatasourcesReply{Datasource: promDatasource, Response: &api.Response{Message: "获取node配置成功"}}, nil
}

func (l *PullRepo) PullStrategies(ctx context.Context) (*biz.StrategyLoad, error) {
	_, span := otel.Tracer(pullModuleName).Start(ctx, "PullRepo.PullStrategies")
	defer span.End()

	return &biz.StrategyLoad{
		StrategyDirs: strategies,
		LoadTime:     loadTime,
	}, nil
}

func (l *PullRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer(pullModuleName).Start(ctx, "PullRepo.V1")
	defer span.End()
	return "version is v1"
}
