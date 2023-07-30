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
	GroupRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *GroupRepo) CreateGroup(ctx context.Context, m *model.PromNodeDirFileGroup) error {
	ctx, span := otel.Tracer("data").Start(ctx, "GroupRepo.CreateGroup")
	defer span.End()
	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroup.Create(m)
}

func (l *GroupRepo) UpdateGroupById(ctx context.Context, id uint32, m *model.PromNodeDirFileGroup) error {
	ctx, span := otel.Tracer("data").Start(ctx, "GroupRepo.UpdateGroupById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroup.Where(query.PromNodeDirFileGroup.ID.Eq(int32(id))).Updates(m); err != nil {
		return err
	}

	return nil
}

func (l *GroupRepo) DeleteGroupById(ctx context.Context, id uint32) error {
	ctx, span := otel.Tracer("data").Start(ctx, "GroupRepo.DeleteGroupById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroup.Where(query.PromNodeDirFileGroup.ID.Eq(int32(id))).Delete(); err != nil {
		return err
	}

	return nil
}

func (l *GroupRepo) GetGroupById(ctx context.Context, id uint32) (*model.PromNodeDirFileGroup, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "GroupRepo.GetGroupById")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroup.FindById(ctx, int32(id))
}

func (l *GroupRepo) ListGroup(ctx context.Context, q *promBizV1.GroupListQueryParams) ([]*model.PromNodeDirFileGroup, int64, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "GroupRepo.ListGroup")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFileGroup.Scopes(
		func(dao gen.Dao) gen.Dao {
			if q.Keyword != "" {
				dao = dao.Where(query.PromNodeDirFileGroup.Name.Like("%" + q.Keyword + "%"))
			}
			return dao
		},
	).FindByPage(q.Offset, q.Limit)
}

func (l *GroupRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer("data").Start(ctx, "GroupRepo.V1")
	defer span.End()
	return "group v1"
}

var _ promBizV1.IGroupRepo = (*GroupRepo)(nil)

func NewGroupRepo(data *Data, logger log.Logger) *GroupRepo {
	return &GroupRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Group"))}
}
