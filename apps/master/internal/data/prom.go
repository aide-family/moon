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

	promStrategy := p.db.PromStrategy
	inf, err := promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).Delete()
	if err != nil {
		p.logger.WithContext(ctx).Errorw("PromRepo.DeleteStrategyByID", id, "err", err)
		return err
	}

	if inf.RowsAffected != 1 {
		p.logger.WithContext(ctx).Warnw("PromRepo.DeleteStrategyByID", id, "err", "RowsAffected != 1")
	}

	return nil
}

func (p *PromRepo) UpdateStrategyByID(ctx context.Context, id int32, m map[string]any) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.UpdateStrategyByID")
	defer span.End()

	promStrategy := p.db.PromStrategy
	inf, err := promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).Updates(m)
	if err != nil {
		p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
		return err
	}

	if inf.RowsAffected != 1 {
		p.logger.WithContext(ctx).Warnw("PromRepo.UpdateStrategyByID", id, "err", "RowsAffected != 1")
	}

	return nil
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
	promGroup := p.db.PromGroup
	inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(id)).Delete()
	if err != nil {
		p.logger.WithContext(ctx).Errorw("PromRepo.DeleteGroupByID", id, "err", err)
		return err
	}

	if inf.RowsAffected != 1 {
		p.logger.WithContext(ctx).Warnw("PromRepo.DeleteGroupByID", id, "err", "RowsAffected != 1")
	}
	return nil
}

func (p *PromRepo) UpdateGroupByID(ctx context.Context, id int32, m map[string]any) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.UpdateGroupByID")
	defer span.End()
	promGroup := p.db.PromGroup
	inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(id)).Updates(m)
	if err != nil {
		p.logger.WithContext(ctx).Errorw("PromRepo.UpdateGroupByID", id, "err", err)
		return err
	}

	if inf.RowsAffected != 1 {
		p.logger.WithContext(ctx).Warnw("PromRepo.UpdateGroupByID", id, "err", "RowsAffected != 1")
	}

	return nil
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
