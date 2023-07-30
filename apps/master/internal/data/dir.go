package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"gorm.io/gen"
	promV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
)

type (
	DirRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *DirRepo) CreateDir(ctx context.Context, m *model.PromNodeDir) error {
	ctx, span := otel.Tracer("data").Start(ctx, "DirRepo.CreateDir")
	defer span.End()
	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDir.Create(m)
}

func (l *DirRepo) UpdateDirById(ctx context.Context, id uint32, m *model.PromNodeDir) error {
	ctx, span := otel.Tracer("data").Start(ctx, "DirRepo.UpdateDirById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNodeDir.Where(query.PromNodeDir.ID.Eq(int32(id))).Updates(m); err != nil {
		return err
	}

	return nil
}

func (l *DirRepo) DeleteDirById(ctx context.Context, id uint32) error {
	ctx, span := otel.Tracer("data").Start(ctx, "DirRepo.DeleteDirById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNodeDir.Where(query.PromNodeDir.ID.Eq(int32(id))).Delete(); err != nil {
		return err
	}

	return nil
}

func (l *DirRepo) GetDirById(ctx context.Context, id uint32) (*model.PromNodeDir, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "DirRepo.GetDirById")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDir.FindById(ctx, int32(id))
}

func (l *DirRepo) ListDir(ctx context.Context, q *promV1.DirListQueryParams) ([]*model.PromNodeDir, int64, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "DirRepo.ListDir")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDir.Scopes(
		func(dao gen.Dao) gen.Dao {
			if q.Keyword != "" {
				dao = dao.Where(query.PromNodeDir.Path.Like("%" + q.Keyword + "%"))
			}
			return dao
		},
	).FindByPage(q.Offset, q.Limit)
}

func (l *DirRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer("data").Start(ctx, "DirRepo.V1")
	defer span.End()
	return "dir v1"
}

var _ promV1.IDirRepo = (*DirRepo)(nil)

func NewDirRepo(data *Data, logger log.Logger) *DirRepo {
	return &DirRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Dir"))}
}
