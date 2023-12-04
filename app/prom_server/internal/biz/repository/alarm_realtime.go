package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ AlarmRealtimeRepo = (*UnimplementedAlarmRealtimeRepo)(nil)

type (
	AlarmRealtimeRepo interface {
		unimplementedAlarmRealtimeRepo()
		GetRealtimeDetailById(ctx context.Context, id uint32, scopes ...query.ScopeMethod) (*bo.AlarmRealtimeBO, error)
		GetRealtimeList(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.AlarmRealtimeBO, error)
		AlarmIntervene(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmInterveneBO) error
		AlarmUpgrade(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmUpgradeBO) error
		AlarmSuppress(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmSuppressBO) error
		// AppendAlarmBeenNotifyMembers 添加发送记录-通知成员
		AppendAlarmBeenNotifyMembers(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmBeenNotifyMemberBO) error
		// AppendAlarmBeenNotifyChatGroups 添加发送记录-通知群组
		AppendAlarmBeenNotifyChatGroups(ctx context.Context, realtimeAlarmID uint32, req *bo.PromAlarmBeenNotifyChatGroupBO) error
		// GetRealtimeCount 统计相关
		GetRealtimeCount(ctx context.Context, scopes ...query.ScopeMethod) (int64, error)
		// Create 创建实时告警信息, 并缓存
		Create(ctx context.Context, req ...*bo.AlarmRealtimeBO) ([]*bo.AlarmRealtimeBO, error)
		// CacheByHistoryId 缓存
		CacheByHistoryId(ctx context.Context, req ...*bo.AlarmRealtimeBO) error
		// DeleteCacheByHistoryId 删除
		DeleteCacheByHistoryId(ctx context.Context, historyId ...uint32) error
	}

	UnimplementedAlarmRealtimeRepo struct{}
)

func (UnimplementedAlarmRealtimeRepo) CacheByHistoryId(_ context.Context, _ ...*bo.AlarmRealtimeBO) error {
	return status.Error(codes.Unimplemented, "method CacheByHistoryId not implemented")
}

func (UnimplementedAlarmRealtimeRepo) Create(_ context.Context, _ ...*bo.AlarmRealtimeBO) ([]*bo.AlarmRealtimeBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedAlarmRealtimeRepo) DeleteCacheByHistoryId(_ context.Context, _ ...uint32) error {
	return status.Error(codes.Unimplemented, "method DeleteCacheByHistoryId not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AppendAlarmBeenNotifyChatGroups(_ context.Context, _ uint32, _ *bo.PromAlarmBeenNotifyChatGroupBO) error {
	return status.Error(codes.Unimplemented, "method AppendAlarmBeenNotifyChatGroups not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AppendAlarmBeenNotifyMembers(_ context.Context, _ uint32, _ *bo.AlarmBeenNotifyMemberBO) error {
	return status.Error(codes.Unimplemented, "method AlarmBeenNotifyMembers not implemented")
}

func (UnimplementedAlarmRealtimeRepo) GetRealtimeCount(_ context.Context, _ ...query.ScopeMethod) (int64, error) {
	return 0, status.Error(codes.Unimplemented, "method GetRealtimeCount not implemented")
}

func (UnimplementedAlarmRealtimeRepo) GetRealtimeDetailById(_ context.Context, _ uint32, _ ...query.ScopeMethod) (*bo.AlarmRealtimeBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetRealtimeDetail not implemented")
}

func (UnimplementedAlarmRealtimeRepo) GetRealtimeList(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*bo.AlarmRealtimeBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetRealtimeList not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AlarmIntervene(_ context.Context, _ uint32, _ *bo.AlarmInterveneBO) error {
	return status.Error(codes.Unimplemented, "method AlarmIntervene not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AlarmUpgrade(_ context.Context, _ uint32, _ *bo.AlarmUpgradeBO) error {
	return status.Error(codes.Unimplemented, "method AlarmUpgrade not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AlarmSuppress(_ context.Context, _ uint32, _ *bo.AlarmSuppressBO) error {
	return status.Error(codes.Unimplemented, "method AlarmSuppress not implemented")
}

func (UnimplementedAlarmRealtimeRepo) unimplementedAlarmRealtimeRepo() {}
