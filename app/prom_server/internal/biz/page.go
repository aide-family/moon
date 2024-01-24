package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api"
	pb "prometheus-manager/api/alarm/page"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type (
	AlarmPageBiz struct {
		log *log.Helper

		pageRepo     repository.PageRepo
		realtimeRepo repository.AlarmRealtimeRepo
	}
)

// NewPageBiz 实例化页面业务
func NewPageBiz(pageRepo repository.PageRepo, realtimeRepo repository.AlarmRealtimeRepo, logger log.Logger) *AlarmPageBiz {
	return &AlarmPageBiz{
		log: log.NewHelper(log.With(logger, "module", "biz.alarm.page")),

		pageRepo:     pageRepo,
		realtimeRepo: realtimeRepo,
	}
}

// GetStrategyIds 获取策略id列表
func (p *AlarmPageBiz) GetStrategyIds(ctx context.Context, ids ...uint32) ([]uint32, error) {
	return p.pageRepo.GetStrategyIds(ctx, basescopes.InTableNamePromStrategyAlarmPageFieldPromAlarmPageIds(ids...))
}

// CreatePage 创建页面
func (p *AlarmPageBiz) CreatePage(ctx context.Context, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	pageBO, err := p.pageRepo.CreatePage(ctx, pageBO)
	if err != nil {
		return nil, err
	}

	return pageBO, nil
}

// UpdatePage 通过id更新页面
func (p *AlarmPageBiz) UpdatePage(ctx context.Context, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	pageBO, err := p.pageRepo.UpdatePageById(ctx, pageBO.Id, pageBO)
	if err != nil {
		return nil, err
	}

	return pageBO, nil
}

// BatchUpdatePageStatusByIds 通过id批量更新页面状态
func (p *AlarmPageBiz) BatchUpdatePageStatusByIds(ctx context.Context, status api.Status, ids []uint32) error {
	return p.pageRepo.BatchUpdatePageStatusByIds(ctx, vo.Status(status), ids)
}

// DeletePageByIds 通过id删除页面
func (p *AlarmPageBiz) DeletePageByIds(ctx context.Context, ids ...uint32) error {
	return p.pageRepo.DeletePageByIds(ctx, ids...)
}

// GetPageById 通过id获取页面详情
func (p *AlarmPageBiz) GetPageById(ctx context.Context, id uint32) (*bo.AlarmPageBO, error) {
	pageBO, err := p.pageRepo.GetPageById(ctx, id)
	if err != nil {
		return nil, err
	}

	return pageBO, nil
}

// GetPageRealtimeById 通过id获取页面详情
func (p *AlarmPageBiz) GetPageRealtimeById(ctx context.Context, id uint32, wheres ...basescopes.ScopeMethod) (*bo.AlarmPageBO, error) {
	pageBO, err := p.pageRepo.Get(ctx, append(wheres, basescopes.InIds(id))...)
	if err != nil {
		return nil, err
	}

	return pageBO, nil
}

// ListPage 获取页面列表
func (p *AlarmPageBiz) ListPage(ctx context.Context, req *pb.ListAlarmPageRequest) ([]*bo.AlarmPageBO, basescopes.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []basescopes.ScopeMethod{
		basescopes.NameLike(req.GetKeyword()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
	}

	pageBos, err := p.pageRepo.ListPage(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}

	return pageBos, pgInfo, nil
}

// SelectPageList 获取页面列表
func (p *AlarmPageBiz) SelectPageList(ctx context.Context, req *pb.SelectAlarmPageRequest) ([]*bo.AlarmPageBO, basescopes.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []basescopes.ScopeMethod{
		basescopes.NameLike(req.GetKeyword()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
	}

	pageBos, err := p.pageRepo.ListPage(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}
	return pageBos, pgInfo, nil
}

// CountAlarmPageByIds 通过id列表获取各页面报警数量
func (p *AlarmPageBiz) CountAlarmPageByIds(ctx context.Context, ids ...uint32) (map[uint32]int64, error) {
	strategyAlarmPages, err := p.pageRepo.GetPromStrategyAlarmPage(ctx, basescopes.InTableNamePromStrategyAlarmPageFieldPromAlarmPageIds(ids...))
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
