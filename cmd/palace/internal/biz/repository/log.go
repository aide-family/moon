package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
)

type OperateLog interface {
	OperateLog(ctx context.Context, log *bo.AddOperateLog) error
	List(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error)
	TeamOperateLog(ctx context.Context, log *bo.AddOperateLog) error
	TeamList(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error)
}
