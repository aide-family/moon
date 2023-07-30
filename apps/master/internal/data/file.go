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
	FileRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *FileRepo) CreateFile(ctx context.Context, m *model.PromNodeDirFile) error {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.CreateFile")
	defer span.End()
	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFile.Create(m)
}

func (l *FileRepo) UpdateFileById(ctx context.Context, id uint32, m *model.PromNodeDirFile) error {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.UpdateFileById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFile.Where(query.PromNodeDirFile.ID.Eq(int32(id))).Updates(m); err != nil {
		return err
	}

	return nil
}

func (l *FileRepo) DeleteFileById(ctx context.Context, id uint32) error {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.DeleteFileById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFile.Where(query.PromNodeDirFile.ID.Eq(int32(id))).Delete(); err != nil {
		return err
	}

	return nil
}

func (l *FileRepo) GetFileById(ctx context.Context, id uint32) (*model.PromNodeDirFile, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.GetFileById")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFile.FindById(ctx, int32(id))
}

func (l *FileRepo) ListFile(ctx context.Context, q *promBizV1.FileListQueryParams) ([]*model.PromNodeDirFile, int64, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.ListFile")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFile.Scopes(
		func(dao gen.Dao) gen.Dao {
			if q.Keyword != "" {
				dao = dao.Where(query.PromNodeDirFile.Filename.Like("%" + q.Keyword + "%"))
			}
			return dao
		},
	).FindByPage(q.Offset, q.Limit)
}

func (l *FileRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.V1")
	defer span.End()
	return "file v1"
}

var _ promBizV1.IFileRepo = (*FileRepo)(nil)

func NewFileRepo(data *Data, logger log.Logger) *FileRepo {
	return &FileRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/File"))}
}
