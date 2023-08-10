package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
)

type (
	PromRepo struct {
		logger *log.Helper
		data   *Data
		db     *query.Query
	}
)

func (p *PromRepo) StrategyDetail(ctx context.Context, id int32) (*model.PromStrategy, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.StrategyDetail")
	defer span.End()

	promStrategy := p.db.PromStrategy
	return promStrategy.WithContext(ctx).Preload(
		promStrategy.Categories,
		promStrategy.AlertLevel,
		promStrategy.AlarmPages,
	).Where(promStrategy.ID.Eq(id)).First()
}

func (p *PromRepo) DeleteStrategyByID(ctx context.Context, id int32) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.DeleteStrategyByID")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promStrategy := tx.PromStrategy

		first, err := promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).First()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.DeleteStrategyByID", id, "err", err)
			return err
		}

		promGroup := tx.PromGroup
		inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(first.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Sub(1))
		if err != nil {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteStrategyByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteStrategyByID", id, "err", "RowsAffected != 1")
		}

		inf, err = promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.DeleteStrategyByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteStrategyByID", id, "err", "RowsAffected != 1")
		}

		return nil
	})
}

func (p *PromRepo) UpdateStrategyByID(ctx context.Context, id int32, m *model.PromStrategy) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.UpdateStrategyByID")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promStrategy := tx.PromStrategy

		first, err := promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).First()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
			return err
		}

		if first.GroupID != m.GroupID {
			promGroup := tx.PromGroup
			// 源group strategy_count -1, 目标group strategy_count +1
			simple, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(first.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Sub(1))
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
				return err
			}
			if simple.RowsAffected != 1 {
				p.logger.WithContext(ctx).Warnw("PromRepo.UpdateStrategyByID", first.GroupID, "err", "RowsAffected != 1")
			}

			simple, err = promGroup.WithContext(ctx).Where(promGroup.ID.Eq(m.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Add(1))
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
				return err
			}
			if simple.RowsAffected != 1 {
				p.logger.WithContext(ctx).Warnw("PromRepo.UpdateStrategyByID", m.GroupID, "err", "RowsAffected != 1")
			}
		}

		inf, err := promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).UpdateColumnSimple(
			promStrategy.Alert.Value(m.Alert),
			promStrategy.Expr.Value(m.Expr),
			promStrategy.For.Value(m.For),
			promStrategy.Labels.Value(m.Labels),
			promStrategy.Annotations.Value(m.Annotations),
			promStrategy.AlertLevelID.Value(m.AlertLevelID),
			promStrategy.GroupID.Value(m.GroupID),
			promStrategy.GroupID.Value(m.GroupID),
		)
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.UpdateStrategyByID", id, "err", "RowsAffected != 1")
		}

		if err = promStrategy.AlarmPages.WithContext(ctx).Model(&model.PromStrategy{ID: id}).Replace(m.AlarmPages...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID AlarmPages Replace", id, "m.AlarmPages", m.AlarmPages, "err", err)
			return err
		}

		if err = promStrategy.Categories.WithContext(ctx).Model(&model.PromStrategy{ID: id}).Replace(m.Categories...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID Categories Replace", id, "m.Categories", m.Categories, "err", err)
			return err
		}

		return nil
	})
}

func (p *PromRepo) CreateStrategy(ctx context.Context, m *model.PromStrategy) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.CreateStrategy")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promStrategy := tx.PromStrategy
		if err := promStrategy.WithContext(ctx).Create(m); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", m, "err", err)
			return err
		}

		promGroup := tx.PromGroup
		rows, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(m.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Add(1))
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", m.GroupID, "err", err)
			return err
		}

		if rows.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.CreateStrategy", m.GroupID, "err", "RowsAffected != 1")
		}

		return nil
	})
}

func (p *PromRepo) GroupDetail(ctx context.Context, id int32) (*model.PromGroup, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.GroupDetail")
	defer span.End()

	promGroup := p.db.PromGroup
	return promGroup.WithContext(ctx).Preload(promGroup.PromStrategies, promGroup.Categories).Where(promGroup.ID.Eq(id)).First()
}

func (p *PromRepo) DeleteGroupByID(ctx context.Context, id int32) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.DeleteGroupByID")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promGroup := tx.PromGroup
		inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.DeleteGroupByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteGroupByID", id, "err", "RowsAffected != 1")
		}

		promStrategy := tx.PromStrategy
		inf, err = promStrategy.WithContext(ctx).Where(promStrategy.GroupID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.DeleteGroupByID PromStrategy", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteGroupByID PromStrategy", id, "err", "RowsAffected != 1")
		}

		return nil
	})
}

func (p *PromRepo) UpdateGroupByID(ctx context.Context, id int32, m *model.PromGroup) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.UpdateGroupByID")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promGroup := tx.PromGroup
		inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(id)).UpdateColumnSimple(
			promGroup.Name.Value(m.Name),
			promGroup.Remark.Value(m.Remark),
		)
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateGroupByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.UpdateGroupByID", id, "err", "RowsAffected != 1")
		}

		if err = promGroup.Categories.WithContext(ctx).Model(&model.PromGroup{ID: id}).Replace(m.Categories...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateGroupByID Categories Replace", id, "m.Categories", m.Categories, "err", err)
			return err
		}

		return nil
	})
}

func (p *PromRepo) CreateGroup(ctx context.Context, m *model.PromGroup) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.CreateGroup")
	defer span.End()
	return p.db.PromGroup.WithContext(ctx).Create(m)
}

func (p *PromRepo) V1(_ context.Context) string {
	return "/prom/v1"
}

var _ biz.IPromRepo = (*PromRepo)(nil)

func NewPromRepo(data *Data, logger log.Logger) *PromRepo {
	return &PromRepo{
		data:   data,
		logger: log.NewHelper(log.With(logger, "module", "data/Prom")),
		db:     query.Use(data.DB()),
	}
}
