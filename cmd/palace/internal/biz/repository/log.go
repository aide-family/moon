package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
)

type OperateLog interface {
	CreateLog(ctx context.Context, log *bo.OperateLogParams) error
	TeamCreateLog(ctx context.Context, log *bo.OperateLogParams) error
	List(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error)
	TeamList(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error)
}
