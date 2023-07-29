package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	promV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
)

type (
	DirRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *DirRepo) V1(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}

var _ promV1.IDirRepo = (*DirRepo)(nil)

func NewDirRepo(data *Data, logger log.Logger) *DirRepo {
	return &DirRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Dir"))}
}
