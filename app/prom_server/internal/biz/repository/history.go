package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ HistoryRepo = (*UnimplementedHistoryRepo)(nil)

type (
	// HistoryRepo .
	HistoryRepo interface {
		mustEmbedUnimplemented()
		// GetHistoryById 通过id获取历史详情
		GetHistoryById(ctx context.Context, id uint32) (*bo.AlarmHistoryBO, error)
		// ListHistory 获取历史列表
		ListHistory(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.AlarmHistoryBO, error)
		// StorageHistory 创建历史
		StorageHistory(ctx context.Context, historyBO ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error)
		// UpdateHistoryById 通过id更新历史
		UpdateHistoryById(ctx context.Context, id uint32, historyBO *bo.AlarmHistoryBO) (*bo.AlarmHistoryBO, error)
	}

	UnimplementedHistoryRepo struct{}
)

func (UnimplementedHistoryRepo) mustEmbedUnimplemented() {}

func (UnimplementedHistoryRepo) GetHistoryById(_ context.Context, _ uint32) (*bo.AlarmHistoryBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHistoryById not implemented")
}

func (UnimplementedHistoryRepo) ListHistory(_ context.Context, _ basescopes.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.AlarmHistoryBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListHistory not implemented")
}

func (UnimplementedHistoryRepo) StorageHistory(_ context.Context, _ ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleHistory not implemented")
}

func (UnimplementedHistoryRepo) UpdateHistoryById(_ context.Context, _ uint32, _ *bo.AlarmHistoryBO) (*bo.AlarmHistoryBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateHistoryById not implemented")
}
