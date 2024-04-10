package biz

import (
	"context"

	"github.com/aide-cloud/universal/base/slices"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

type AlarmRealtimeBiz struct {
	log *log.Helper

	dataRepo      repository.DataRepo
	realtimeRepo  repository.AlarmRealtimeRepo
	alarmPageRepo repository.PageRepo
}

func NewAlarmRealtime(
	dataRepo repository.DataRepo,
	realtimeRepo repository.AlarmRealtimeRepo,
	alarmPageRepo repository.PageRepo,
	logger log.Logger,
) *AlarmRealtimeBiz {
	return &AlarmRealtimeBiz{
		log: log.NewHelper(log.With(logger, "module", "biz.AlarmRealtimeBiz")),

		dataRepo:      dataRepo,
		realtimeRepo:  realtimeRepo,
		alarmPageRepo: alarmPageRepo,
	}
}

// GetRealtimeDetailById 通过id获取实时告警详情
func (l *AlarmRealtimeBiz) GetRealtimeDetailById(ctx context.Context, id uint32) (*bo.AlarmRealtimeBO, error) {
	return l.realtimeRepo.GetRealtimeDetailById(ctx, id)
}

// GetRealtimeList 获取实时告警列表
func (l *AlarmRealtimeBiz) GetRealtimeList(ctx context.Context, req *bo.ListRealtimeReq) ([]*bo.AlarmRealtimeBO, error) {
	strategyIds, err := l.alarmPageRepo.GetStrategyIds(ctx, do.StrategyInAlarmPageIds(req.AlarmPageId))
	if err != nil {
		return nil, err
	}

	if len(strategyIds) == 0 {
		return []*bo.AlarmRealtimeBO{}, nil
	}

	// 去重
	strategyIds = slices.MergeUnique(append(strategyIds, req.StrategyIds...))

	wheres := []basescopes.ScopeMethod{
		do.PromAlarmRealtimeLike(req.Keyword),
		do.PromAlarmRealtimeInLevelIds(req.LevelIds...),
		do.PromAlarmRealtimeInStrategyIds(strategyIds...),
		do.PromAlarmRealtimeBetweenEventAt(req.StartAt, req.EndAt),
		// 还在告警的数据
		basescopes.StatusEQ(vobj.StatusEnabled),
		do.PromAlarmRealtimeEventAtDesc(),
		//预加载告警等级
		do.PromAlarmRealtimePreloadLevel(),
		do.PromAlarmRealtimePreloadStrategy(),
	}
	return l.realtimeRepo.GetRealtimeList(ctx, req.Page, wheres...)
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
