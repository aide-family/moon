package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
)

type (
	FileRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *FileRepo) V1(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}

var _ promBizV1.IFileRepo = (*FileRepo)(nil)

func NewFileRepo(data *Data, logger log.Logger) *FileRepo {
	return &FileRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/File"))}
}
