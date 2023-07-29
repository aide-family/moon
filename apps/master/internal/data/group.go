package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
)

type (
	GroupRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *GroupRepo) V1(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}

var _ promBizV1.IGroupRepo = (*GroupRepo)(nil)

func NewGroupRepo(data *Data, logger log.Logger) *GroupRepo {
	return &GroupRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Group"))}
}
