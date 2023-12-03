package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ AlarmRealtimeRepo = (*UnimplementedAlarmRealtimeRepo)(nil)

type (
	AlarmRealtimeRepo interface {
		unimplementedAlarmRealtimeRepo()
		GetRealtimeDetailById(ctx context.Context, id uint32, scopes ...query.ScopeMethod) (*dobo.AlarmRealtimeBO, error)
		GetRealtimeList(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.AlarmRealtimeBO, error)
		AlarmIntervene(ctx context.Context, realtimeAlarmID uint32, req *dobo.AlarmInterveneBO) error
		AlarmUpgrade(ctx context.Context, realtimeAlarmID uint32, req *dobo.AlarmUpgradeBO) error
		AlarmSuppress(ctx context.Context, realtimeAlarmID uint32, req *dobo.AlarmSuppressBO) error
		// AppendAlarmBeenNotifyMembers 添加发送记录-通知成员
		AppendAlarmBeenNotifyMembers(ctx context.Context, realtimeAlarmID uint32, req *dobo.AlarmBeenNotifyMemberBO) error
		// AppendAlarmBeenNotifyChatGroups 添加发送记录-通知群组
		AppendAlarmBeenNotifyChatGroups(ctx context.Context, realtimeAlarmID uint32, req *dobo.PromAlarmBeenNotifyChatGroupBO) error
		// GetRealtimeCount 统计相关
		GetRealtimeCount(ctx context.Context, scopes ...query.ScopeMethod) (int64, error)
	}

	UnimplementedAlarmRealtimeRepo struct{}
)

func (UnimplementedAlarmRealtimeRepo) AppendAlarmBeenNotifyChatGroups(_ context.Context, _ uint32, _ *dobo.PromAlarmBeenNotifyChatGroupBO) error {
	return status.Error(codes.Unimplemented, "method AppendAlarmBeenNotifyChatGroups not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AppendAlarmBeenNotifyMembers(_ context.Context, _ uint32, _ *dobo.AlarmBeenNotifyMemberBO) error {
	return status.Error(codes.Unimplemented, "method AlarmBeenNotifyMembers not implemented")
}

func (UnimplementedAlarmRealtimeRepo) GetRealtimeCount(_ context.Context, _ ...query.ScopeMethod) (int64, error) {
	return 0, status.Error(codes.Unimplemented, "method GetRealtimeCount not implemented")
}

func (UnimplementedAlarmRealtimeRepo) GetRealtimeDetailById(_ context.Context, _ uint32, _ ...query.ScopeMethod) (*dobo.AlarmRealtimeBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetRealtimeDetail not implemented")
}

func (UnimplementedAlarmRealtimeRepo) GetRealtimeList(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*dobo.AlarmRealtimeBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetRealtimeList not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AlarmIntervene(_ context.Context, _ uint32, _ *dobo.AlarmInterveneBO) error {
	return status.Error(codes.Unimplemented, "method AlarmIntervene not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AlarmUpgrade(_ context.Context, _ uint32, _ *dobo.AlarmUpgradeBO) error {
	return status.Error(codes.Unimplemented, "method AlarmUpgrade not implemented")
}

func (UnimplementedAlarmRealtimeRepo) AlarmSuppress(_ context.Context, _ uint32, _ *dobo.AlarmSuppressBO) error {
	return status.Error(codes.Unimplemented, "method AlarmSuppress not implemented")
}

func (UnimplementedAlarmRealtimeRepo) unimplementedAlarmRealtimeRepo() {}
