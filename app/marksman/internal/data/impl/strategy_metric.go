package impl

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
	"github.com/aide-family/marksman/internal/data/impl/do"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewStrategyMetricRepository(d *data.Data) (repository.StrategyMetric, error) {
	query.SetDefault(d.DB())
	return &strategyMetricRepository{db: d.DB()}, nil
}

type strategyMetricRepository struct {
	db *gorm.DB
}

// GetEvaluateMetricStrategies implements [repository.StrategyMetric].
func (r *strategyMetricRepository) GetEvaluateMetricStrategies(ctx context.Context) ([]*bo.EvaluateMetricStrategyBo, error) {
	ns := contextx.GetNamespace(ctx)
	strategyGroupQuery := query.StrategyGroup
	strategyGroupQueryWrapper := strategyGroupQuery.WithContext(ctx)
	strategyGroupQueryWrapper = strategyGroupQueryWrapper.Where(
		strategyGroupQuery.NamespaceUID.Eq(ns.Int64()),
		strategyGroupQuery.Status.Eq(int32(enum.GlobalStatus_ENABLED)),
	)
	strategyGroups, err := strategyGroupQueryWrapper.Select(strategyGroupQuery.ID).Find()
	if err != nil {
		return nil, err
	}
	if len(strategyGroups) == 0 {
		return nil, nil
	}
	strategyGroupIds := make([]int64, 0, len(strategyGroups))
	for _, strategyGroup := range strategyGroups {
		strategyGroupIds = append(strategyGroupIds, strategyGroup.ID.Int64())
	}
	strategyQuery := query.Strategy
	strategyQueryWrapper := strategyQuery.WithContext(ctx)
	strategyQueryWrapper = strategyQueryWrapper.Where(
		strategyQuery.NamespaceUID.Eq(ns.Int64()),
		strategyQuery.Status.Eq(int32(enum.GlobalStatus_ENABLED)),
		strategyQuery.StrategyGroupUID.In(strategyGroupIds...),
	)
	var strategyIds []int64
	strategies, err := strategyQueryWrapper.Select(strategyQuery.ID).Find()
	if err != nil {
		return nil, err
	}
	if len(strategies) == 0 {
		return nil, nil
	}
	for _, strategy := range strategies {
		strategyIds = append(strategyIds, strategy.ID.Int64())
	}
	strategyMetricQuery := query.StrategyMetric
	strategyMetricLevelQuery := query.StrategyMetricLevel
	levelQuery := query.Level
	strategyMetrics, err := strategyMetricQuery.WithContext(ctx).Where(
		strategyMetricQuery.NamespaceUID.Eq(ns.Int64()),
		strategyMetricQuery.StrategyUID.In(strategyIds...),
	).Preload(
		strategyMetricQuery.StrategyLevels.On(strategyMetricLevelQuery.Status.Eq(int32(enum.GlobalStatus_ENABLED))),
		strategyMetricQuery.StrategyLevels.Level.On(levelQuery.Status.Eq(int32(enum.GlobalStatus_ENABLED))),
	).Find()
	if err != nil {
		return nil, err
	}

	datasourceQuery := query.Datasource
	datasourceQueryWrapper := datasourceQuery.WithContext(ctx)
	datasourceQueryWrapper = datasourceQueryWrapper.Where(
		datasourceQuery.NamespaceUID.Eq(ns.Int64()),
		datasourceQuery.Status.Eq(int32(enum.GlobalStatus_ENABLED)),
	)
	datasourceList, err := datasourceQueryWrapper.Find()
	if err != nil {
		return nil, err
	}
	datasourceMap := make(map[int64]*do.Datasource)
	datasourceIds := make([]int64, 0, len(datasourceList))
	for _, datasource := range datasourceList {
		datasourceMap[datasource.ID.Int64()] = datasource
		datasourceIds = append(datasourceIds, datasource.ID.Int64())
	}

	evaluateStrategies := make([]*bo.EvaluateMetricStrategyBo, 0, len(strategies))
	for _, strategyDetail := range strategyMetrics {
		for _, strategyLevel := range strategyDetail.StrategyLevels {
			datasources := strategyDetail.DatasourceUIDs.List()
			if len(datasources) == 0 {
				datasources = datasourceIds
			}

			for _, datasourceUID := range datasources {
				datasourceDetail, ok := datasourceMap[datasourceUID]
				if !ok {
					continue
				}
				evaluateStrategies = append(evaluateStrategies, convert.ToEvaluateMetricStrategyBo(strategyDetail, strategyLevel, datasourceDetail))
			}
		}
	}
	return evaluateStrategies, nil
}

func (r *strategyMetricRepository) CreateStrategyMetric(ctx context.Context, req *bo.SaveStrategyMetricBo) error {
	m := convert.ToStrategyMetricDo(ctx, req)
	return query.StrategyMetric.WithContext(ctx).Create(m)
}

func (r *strategyMetricRepository) UpdateStrategyMetric(ctx context.Context, req *bo.SaveStrategyMetricBo) error {
	strategyMetricMutation := query.StrategyMetric
	columns := []field.AssignExpr{
		strategyMetricMutation.Expr.Value(req.Expr),
		strategyMetricMutation.Labels.Value(safety.NewMap(req.Labels)),
		strategyMetricMutation.Summary.Value(req.Summary),
		strategyMetricMutation.Description.Value(req.Description),
		strategyMetricMutation.DatasourceUIDs.Value(safety.NewSlice(req.DatasourceUIDs)),
	}
	_, err := strategyMetricMutation.WithContext(ctx).Where(strategyMetricMutation.StrategyUID.Eq(req.StrategyUID.Int64())).UpdateColumnSimple(columns...)
	return err
}

func (r *strategyMetricRepository) GetStrategyMetric(ctx context.Context, strategyUID snowflake.ID) (*bo.StrategyMetricItemBo, error) {
	strategyMetricQuery := query.StrategyMetric
	wrapper := strategyMetricQuery.WithContext(ctx).Where(strategyMetricQuery.StrategyUID.Eq(strategyUID.Int64()))
	wrapper = wrapper.Preload(
		field.Associations,
		strategyMetricQuery.StrategyLevels.Level,
		strategyMetricQuery.Strategy.StrategyGroup,
	)
	m, err := wrapper.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("strategy metric not found")
		}
		return nil, err
	}

	return convert.ToStrategyMetricItemBo(m), nil
}

func (r *strategyMetricRepository) CreateStrategyMetricLevel(ctx context.Context, req *bo.SaveStrategyMetricLevelBo) error {
	m := convert.ToStrategyMetricLevelDo(ctx, req)
	return query.StrategyMetricLevel.WithContext(ctx).Create(m)
}

func (r *strategyMetricRepository) UpdateStrategyMetricLevel(ctx context.Context, req *bo.SaveStrategyMetricLevelBo) error {
	columns := []field.AssignExpr{
		query.StrategyMetricLevel.Mode.Value(int32(req.Mode)),
		query.StrategyMetricLevel.Condition.Value(int32(req.Condition)),
		query.StrategyMetricLevel.Values.Value(safety.NewSlice(req.Values)),
		query.StrategyMetricLevel.DurationSec.Value(req.DurationSec),
		query.StrategyMetricLevel.Status.Value(int32(req.Status)),
	}
	conditions := []gen.Condition{
		query.StrategyMetricLevel.StrategyUID.Eq(req.StrategyUID.Int64()),
		query.StrategyMetricLevel.LevelUID.Eq(req.LevelUID.Int64()),
	}
	_, err := query.StrategyMetricLevel.WithContext(ctx).Where(conditions...).UpdateColumnSimple(columns...)
	return err
}

func (r *strategyMetricRepository) UpdateStrategyMetricLevelStatus(ctx context.Context, req *bo.UpdateStrategyMetricLevelStatusBo) error {
	sml := query.StrategyMetricLevel
	info, err := sml.WithContext(ctx).Where(
		sml.StrategyUID.Eq(req.StrategyUID.Int64()),
		sml.LevelUID.Eq(req.LevelUID.Int64()),
	).Update(sml.Status, req.Status)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("strategy metric level not found")
	}
	return nil
}

func (r *strategyMetricRepository) DeleteStrategyMetricLevel(ctx context.Context, levelUID snowflake.ID, strategyUID snowflake.ID) error {
	sml := query.StrategyMetricLevel
	info, err := sml.WithContext(ctx).Where(
		sml.StrategyUID.Eq(strategyUID.Int64()),
		sml.LevelUID.Eq(levelUID.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("strategy metric level not found")
	}
	return nil
}

func (r *strategyMetricRepository) GetStrategyMetricLevelByStrategyAndLevel(ctx context.Context, strategyUID snowflake.ID, levelUID snowflake.ID) (*bo.StrategyMetricLevelItemBo, error) {
	sml := query.StrategyMetricLevel
	row, err := sml.WithContext(ctx).Where(
		sml.StrategyUID.Eq(strategyUID.Int64()),
		sml.LevelUID.Eq(levelUID.Int64()),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("strategy metric level not found")
		}
		return nil, err
	}
	var levelDo *do.Level
	if row.LevelUID != 0 {
		levelDo, err = query.Level.WithContext(ctx).Where(query.Level.ID.Eq(row.LevelUID.Int64())).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, merr.ErrorNotFound("strategy metric level level not found")
			}
			return nil, err
		}
	}
	return convert.ToStrategyMetricLevelItemBo(row, levelDo), nil
}

func (r *strategyMetricRepository) StrategyMetricBindReceivers(ctx context.Context, req *bo.StrategyMetricBindReceiversBo) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		smr := q.StrategyMetricReceiver
		w := smr.WithContext(ctx).Where(smr.StrategyUID.Eq(req.StrategyUID.Int64()))
		w = w.Where(smr.LevelUID.Eq(req.LevelUID.Int64()))
		_, err := w.Delete()
		if err != nil {
			return err
		}
		if len(req.ReceiverUIDs) == 0 {
			return nil
		}
		rows := make([]*do.StrategyMetricReceiver, 0, len(req.ReceiverUIDs))
		for _, recUID := range req.ReceiverUIDs {
			rows = append(rows, &do.StrategyMetricReceiver{
				StrategyUID: req.StrategyUID,
				ReceiverUID: recUID,
				LevelUID:    req.LevelUID,
			})
		}
		return tx.CreateInBatches(rows, 100).Error
	})
}
