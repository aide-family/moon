package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
	"prometheus-manager/pkg/util/stringer"
)

type (
	FileRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *FileRepo) CreateFile(ctx context.Context, m *model.PromNodeDirFile) error {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.CreateFile")
	span.SetAttributes(attribute.Stringer("model", stringer.New(m)))
	defer span.End()
	return query.Use(l.data.DB()).WithContext(ctx).PromNodeDirFile.Create(m)
}

func (l *FileRepo) UpdateFileById(ctx context.Context, id uint32, m *model.PromNodeDirFile) error {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.UpdateFileById")
	span.SetAttributes(attribute.Stringer("model", stringer.New(m)))
	defer span.End()

	modelInstance := query.Use(l.data.DB()).PromNodeDirFile
	db := modelInstance.WithContext(ctx)
	if _, err := db.Where(modelInstance.ID.Eq(int32(id))).Updates(m); err != nil {
		return err
	}

	return nil
}

func (l *FileRepo) DeleteFileById(ctx context.Context, id uint32) error {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.DeleteFileById")
	span.SetAttributes(attribute.Int64("id", int64(id)))
	defer span.End()

	modelInstance := query.Use(l.data.DB()).PromNodeDirFile
	db := modelInstance.WithContext(ctx)
	if _, err := db.Where(modelInstance.ID.Eq(int32(id))).Delete(); err != nil {
		return err
	}

	return nil
}

func (l *FileRepo) GetFileById(ctx context.Context, id uint32) (*model.PromNodeDirFile, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.GetFileById")
	span.SetAttributes(attribute.Int64("id", int64(id)))
	defer span.End()

	modelInstance := query.Use(l.data.DB()).PromNodeDirFile
	db := modelInstance.WithContext(ctx)
	return db.Preload(modelInstance.Groups.Strategies).Where(modelInstance.ID.Eq(int32(id))).First()
}

func (l *FileRepo) ListFile(ctx context.Context, q *promBizV1.FileListQueryParams) ([]*model.PromNodeDirFile, int64, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "FileRepo.ListFile")
	span.SetAttributes(attribute.Stringer("query", stringer.New(q)))
	defer span.End()

	modelInstance := query.Use(l.data.DB()).PromNodeDirFile
	db := modelInstance.WithContext(ctx)
	return db.Scopes(
		func(dao gen.Dao) gen.Dao {
			if q.Keyword != "" {
				dao = dao.Where(modelInstance.Filename.Like("%" + q.Keyword + "%"))
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
