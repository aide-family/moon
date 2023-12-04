package biz

import (
	"context"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
)

type AlarmRealtimeBiz struct {
	log *log.Helper

	dataRepo     repository.DataRepo
	realtimeRepo repository.AlarmRealtimeRepo
}

func NewAlarmRealtime(
	dataRepo repository.DataRepo,
	realtimeRepo repository.AlarmRealtimeRepo,
	logger log.Logger,
) *AlarmRealtimeBiz {
	return &AlarmRealtimeBiz{
		log: log.NewHelper(log.With(logger, "module", "biz.AlarmRealtimeBiz")),

		dataRepo:     dataRepo,
		realtimeRepo: realtimeRepo,
	}
}

// GetRealtimeDetailById 通过id获取实时告警详情
func (l *AlarmRealtimeBiz) GetRealtimeDetailById(ctx context.Context, id uint32) (*bo.AlarmRealtimeBO, error) {
	return l.realtimeRepo.GetRealtimeDetailById(ctx, id)
}

// GetRealtimeList 获取实时告警列表
func (l *AlarmRealtimeBiz) GetRealtimeList(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.AlarmRealtimeBO, error) {
	return l.realtimeRepo.GetRealtimeList(ctx, pgInfo, scopes...)
}

// AlarmIntervene 告警干预/介入
func (l *AlarmRealtimeBiz) AlarmIntervene(ctx context.Context, id uint32, req *bo.AlarmInterveneBO) error {
	return l.realtimeRepo.AlarmIntervene(ctx, id, req)
}

// AlarmUpgrade 告警升级
func (l *AlarmRealtimeBiz) AlarmUpgrade(ctx context.Context, id uint32, req *bo.AlarmUpgradeBO) error {
	return l.realtimeRepo.AlarmUpgrade(ctx, id, req)
}

// AlarmSuppress 告警抑制
func (l *AlarmRealtimeBiz) AlarmSuppress(ctx context.Context, id uint32, req *bo.AlarmSuppressBO) error {
	return l.realtimeRepo.AlarmSuppress(ctx, id, req)
}

// HandleRealtime 创建实时告警
func (l *AlarmRealtimeBiz) HandleRealtime(ctx context.Context, req ...*bo.AlarmRealtimeBO) ([]*bo.AlarmRealtimeBO, error) {
	if len(req) == 0 {
		return nil, nil
	}
	realtimeAlarmBOs, err := l.realtimeRepo.Create(ctx, req...)
	if err != nil {
		return nil, err
	}

	if err = l.realtimeRepo.CacheByHistoryId(ctx, realtimeAlarmBOs...); err != nil {
		// TODO 需要告警, 一般不会失败, 失败时可能缓存组件异常了
		return nil, err
	}

	return realtimeAlarmBOs, nil
}
