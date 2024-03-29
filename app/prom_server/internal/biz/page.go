package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type (
	AlarmPageBiz struct {
		log *log.Helper

		pageRepo     repository.PageRepo
		realtimeRepo repository.AlarmRealtimeRepo
		logX         repository.SysLogRepo
	}
)

// NewPageBiz 实例化页面业务
func NewPageBiz(pageRepo repository.PageRepo, realtimeRepo repository.AlarmRealtimeRepo, logX repository.SysLogRepo, logger log.Logger) *AlarmPageBiz {
	return &AlarmPageBiz{
		log: log.NewHelper(log.With(logger, "module", "biz.alarm.page")),

		pageRepo:     pageRepo,
		realtimeRepo: realtimeRepo,
		logX:         logX,
	}
}

// GetStrategyIds 获取策略id列表
func (p *AlarmPageBiz) GetStrategyIds(ctx context.Context, ids ...uint32) ([]uint32, error) {
	return p.pageRepo.GetStrategyIds(ctx, do.StrategyInAlarmPageIds(ids...))
}

// CountAlarmPageByIds 通过id列表获取各页面报警数量
func (p *AlarmPageBiz) CountAlarmPageByIds(ctx context.Context, ids ...uint32) (map[uint32]int64, error) {
	strategyAlarmPages, err := p.pageRepo.GetPromStrategyAlarmPage(ctx, do.StrategyInAlarmPageIds(ids...))
	if err != nil {
		return nil, err
	}
	if len(strategyAlarmPages) == 0 {
		return nil, nil
	}
	alarmPageIdMap := make(map[uint32]map[uint32]struct{})
	strategyIds := make([]uint32, 0, len(strategyAlarmPages))
	for _, strategyAlarmPage := range strategyAlarmPages {
		if _, ok := alarmPageIdMap[strategyAlarmPage.PromAlarmPageID]; !ok {
			alarmPageIdMap[strategyAlarmPage.PromAlarmPageID] = make(map[uint32]struct{})
		}
		alarmPageIdMap[strategyAlarmPage.PromAlarmPageID][strategyAlarmPage.PromStrategyID] = struct{}{}
		strategyIds = append(strategyIds, strategyAlarmPage.PromStrategyID)
	}
	// 按策略id统计实时告警数量
	realtimeAlarmCount, err := p.realtimeRepo.CountRealtimeAlarmByStrategyIds(ctx, strategyIds...)
	if err != nil {
		return nil, err
	}

	alarmNumCount := make(map[uint32]int64)
	for alarmId, m := range alarmPageIdMap {
		for strategyId := range m {
			alarmNumCount[alarmId] += realtimeAlarmCount[strategyId]
		}
	}

	return alarmNumCount, nil
}

// GetUserAlarmPages 获取用户告警页面列表
func (p *AlarmPageBiz) GetUserAlarmPages(ctx context.Context, userId uint32) ([]*bo.DictBO, error) {
	pageBos, err := p.pageRepo.UserPageList(ctx, userId)
	if err != nil {
		return nil, err
	}

	return pageBos, nil
}

// BindUserPages 绑定用户页面
func (p *AlarmPageBiz) BindUserPages(ctx context.Context, userId uint32, pageIds []uint32) error {
	return p.pageRepo.BindUserPages(ctx, userId, pageIds)
}
