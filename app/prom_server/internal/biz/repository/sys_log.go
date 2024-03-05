package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

var _ SysLogRepo = (*UnimplementedSysLogRepo)(nil)

type (
	SysLogRepo interface {
		mustEmbedUnimplemented()
		// CreateSysLog 创建日志
		CreateSysLog(ctx context.Context, action vo.Action, logInfo ...*bo.SysLogBo)
		// ListSysLog 获取日志列表
		ListSysLog(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.SysLogBo, error)
	}

	UnimplementedSysLogRepo struct{}
)

func (UnimplementedSysLogRepo) mustEmbedUnimplemented() {}

func (UnimplementedSysLogRepo) CreateSysLog(_ context.Context, _ vo.Action, _ ...*bo.SysLogBo) {}

func (UnimplementedSysLogRepo) ListSysLog(_ context.Context, _ bo.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.SysLogBo, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}
