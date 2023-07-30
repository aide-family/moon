package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"gorm.io/gen"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
)

type (
	RuleRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *RuleRepo) CreateRule(ctx context.Context, m *model.PromNodeDirFileGroupStrategy) error {
	ctx, span := otel.Tracer("data").Start(ctx, "RuleRepo.CreateRule")
	defer span.End()
	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroupStrategy.Create(m)
}

func (l *RuleRepo) UpdateRuleById(ctx context.Context, id uint32, m *model.PromNodeDirFileGroupStrategy) error {
	ctx, span := otel.Tracer("data").Start(ctx, "RuleRepo.UpdateRuleById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroupStrategy.Where(query.PromNodeDirFileGroupStrategy.ID.Eq(int32(id))).Updates(m); err != nil {
		return err
	}

	return nil
}

func (l *RuleRepo) DeleteRuleById(ctx context.Context, id uint32) error {
	ctx, span := otel.Tracer("data").Start(ctx, "RuleRepo.DeleteRuleById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroupStrategy.Where(query.PromNodeDirFileGroupStrategy.ID.Eq(int32(id))).Delete(); err != nil {
		return err
	}

	return nil
}

func (l *RuleRepo) GetRuleById(ctx context.Context, id uint32) (*model.PromNodeDirFileGroupStrategy, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "RuleRepo.GetRuleById")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroupStrategy.FindById(ctx, int32(id))
}

func (l *RuleRepo) ListRule(ctx context.Context, q *promBizV1.RuleListQueryParams) ([]*model.PromNodeDirFileGroupStrategy, int64, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "RuleRepo.ListRule")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroupStrategy.Scopes(
		func(dao gen.Dao) gen.Dao {
			if q.Keyword != "" {
				dao = dao.Where(query.PromNodeDirFileGroupStrategy.Alert.Like(q.Keyword))
			}
			return dao
		},
	).FindByPage(q.Offset, q.Limit)
}

func (l *RuleRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer("data").Start(ctx, "RuleRepo.V1")
	defer span.End()
	return "rule v1"
}

var _ promBizV1.IRuleRepo = (*RuleRepo)(nil)

func NewRuleRepo(data *Data, logger log.Logger) *RuleRepo {
	return &RuleRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Rule"))}
}
