package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ HistoryRepo = (*UnimplementedHistoryRepo)(nil)

type (
	// HistoryRepo .
	HistoryRepo interface {
		mustEmbedUnimplemented()
		// GetHistoryById 通过id获取历史详情
		GetHistoryById(ctx context.Context, id uint) (*bo.AlarmHistoryBO, error)
		// ListHistory 获取历史列表
		ListHistory(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.AlarmHistoryBO, error)
		// StorageHistory 创建历史
		StorageHistory(ctx context.Context, historyBO ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error)
		// UpdateHistoryById 通过id更新历史
		UpdateHistoryById(ctx context.Context, id uint, historyBO *bo.AlarmHistoryBO) (*bo.AlarmHistoryBO, error)
	}

	UnimplementedHistoryRepo struct{}
)

func (UnimplementedHistoryRepo) mustEmbedUnimplemented() {}

func (UnimplementedHistoryRepo) GetHistoryById(_ context.Context, _ uint) (*bo.AlarmHistoryBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHistoryById not implemented")
}

func (UnimplementedHistoryRepo) ListHistory(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*bo.AlarmHistoryBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListHistory not implemented")
}

func (UnimplementedHistoryRepo) StorageHistory(_ context.Context, _ ...*bo.AlarmHistoryBO) ([]*bo.AlarmHistoryBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HandleHistory not implemented")
}

func (UnimplementedHistoryRepo) UpdateHistoryById(_ context.Context, _ uint, _ *bo.AlarmHistoryBO) (*bo.AlarmHistoryBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateHistoryById not implemented")
}
