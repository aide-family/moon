package impl

import (
	"context"
	"errors"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
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

func (r *strategyMetricRepository) CreateStrategyMetric(ctx context.Context, req *bo.SaveStrategyMetricBo) error {
	m := convert.ToStrategyMetricDo(req)
	return query.StrategyMetric.WithContext(ctx).Create(m)
}

func (r *strategyMetricRepository) UpdateStrategyMetric(ctx context.Context, req *bo.SaveStrategyMetricBo) error {
	strategyMetricMutation := query.StrategyMetric
	columns := []field.AssignExpr{
		strategyMetricMutation.Expr.Value(req.Expr),
		strategyMetricMutation.Labels.Value(safety.NewMap(req.Labels)),
		strategyMetricMutation.Summary.Value(req.Summary),
		strategyMetricMutation.Description.Value(req.Description),
		strategyMetricMutation.Status.Value(int32(req.Status)),
		strategyMetricMutation.DatasourceUIDs.Value(safety.NewSlice(req.DatasourceUIDs)),
	}
	_, err := strategyMetricMutation.WithContext(ctx).Where(strategyMetricMutation.StrategyUID.Eq(req.StrategyUID.Int64())).UpdateColumnSimple(columns...)
	return err
}

func (r *strategyMetricRepository) GetStrategyMetric(ctx context.Context, strategyUID snowflake.ID) (*bo.StrategyMetricItemBo, error) {
	strategyMetricQuery := query.StrategyMetric
	m, err := strategyMetricQuery.WithContext(ctx).Where(strategyMetricQuery.StrategyUID.Eq(strategyUID.Int64())).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, merr.ErrorNotFound("strategy metric not found")
		}
		return nil, err
	}
	// load levels
	strategyMetricLevelQuery := query.StrategyMetricLevel
	strategyMetricLevelDos, err := strategyMetricLevelQuery.WithContext(ctx).Where(strategyMetricLevelQuery.StrategyUID.Eq(strategyUID.Int64())).Find()
	if err != nil {
		return nil, err
	}

	levelUIDs := make([]snowflake.ID, 0, len(strategyMetricLevelDos))
	for _, row := range strategyMetricLevelDos {
		levelUIDs = append(levelUIDs, row.LevelUID)
	}
	levelUIDsSlice := safety.NewSlice(levelUIDs).Sort(func(a snowflake.ID, b snowflake.ID) bool {
		return a < b
	}).Uniq(func(a snowflake.ID, b snowflake.ID) bool {
		return a == b
	})

	levelUIDsInt64s := safety.ConvertSlice(levelUIDsSlice.List(), func(v snowflake.ID) int64 {
		return v.Int64()
	})
	levelQuery := query.Level
	levelDos, err := levelQuery.WithContext(ctx).Where(levelQuery.ID.In(levelUIDsInt64s...)).Find()
	if err != nil {
		return nil, err
	}
	levelMap := make(map[snowflake.ID]*do.Level)
	for _, levelDo := range levelDos {
		levelMap[levelDo.ID] = levelDo
	}

	levels := make([]*bo.StrategyMetricLevelItemBo, 0, len(strategyMetricLevelDos))
	for _, row := range strategyMetricLevelDos {
		levelDo := levelMap[row.LevelUID]
		levels = append(levels, convert.ToStrategyMetricLevelItemBo(row, levelDo))
	}
	return convert.ToStrategyMetricItemBo(m, levels), nil
}

func (r *strategyMetricRepository) CreateStrategyMetricLevel(ctx context.Context, req *bo.SaveStrategyMetricLevelBo) error {
	m := convert.ToStrategyMetricLevelDo(req)
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
		sml.ID.Eq(req.UID.Int64()),
	).Update(sml.Status, req.Status)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("strategy metric level not found")
	}
	return nil
}

func (r *strategyMetricRepository) DeleteStrategyMetricLevel(ctx context.Context, uid snowflake.ID, strategyUID snowflake.ID) error {
	sml := query.StrategyMetricLevel
	info, err := sml.WithContext(ctx).Where(
		sml.StrategyUID.Eq(strategyUID.Int64()),
		sml.ID.Eq(uid.Int64()),
	).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("strategy metric level not found")
	}
	return nil
}

func (r *strategyMetricRepository) GetStrategyMetricLevel(ctx context.Context, uid snowflake.ID, strategyUID snowflake.ID) (*bo.StrategyMetricLevelItemBo, error) {
	sml := query.StrategyMetricLevel
	row, err := sml.WithContext(ctx).Where(
		sml.StrategyUID.Eq(strategyUID.Int64()),
		sml.ID.Eq(uid.Int64()),
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
