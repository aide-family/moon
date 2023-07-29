package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
)

type (
	NodeRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *NodeRepo) V1(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}

var _ promBizV1.INodeRepo = (*NodeRepo)(nil)

func NewNodeRepo(data *Data, logger log.Logger) *NodeRepo {
	return &NodeRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Node"))}
}
