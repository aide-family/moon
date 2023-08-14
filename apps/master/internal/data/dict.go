package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/master/internal/biz"
)

type (
	DictRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IDictV1Repo = (*DictRepo)(nil)

func NewDictRepo(data *Data, logger log.Logger) *DictRepo {
	return &DictRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Dict"))}
}

func (l *DictRepo) V1(_ context.Context) string {
	return "DictRepo.V1"
}
