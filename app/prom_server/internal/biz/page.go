package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/alarm/page"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/helper/model/page"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type (
	PageBiz struct {
		log *log.Helper

		pageRepo repository.PageRepo
	}
)

// NewPageBiz 实例化页面业务
func NewPageBiz(pageRepo repository.PageRepo, logger log.Logger) *PageBiz {
	return &PageBiz{
		log: log.NewHelper(log.With(logger, "module", "biz.alarm.page")),

		pageRepo: pageRepo,
	}
}

// CreatePage 创建页面
func (p *PageBiz) CreatePage(ctx context.Context, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	pageBO, err := p.pageRepo.CreatePage(ctx, pageBO)
	if err != nil {
		return nil, err
	}

	return pageBO, nil
}

// UpdatePage 通过id更新页面
func (p *PageBiz) UpdatePage(ctx context.Context, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	pageBO, err := p.pageRepo.UpdatePageById(ctx, uint(pageBO.Id), pageBO)
	if err != nil {
		return nil, err
	}

	return pageBO, nil
}

// BatchUpdatePageStatusByIds 通过id批量更新页面状态
func (p *PageBiz) BatchUpdatePageStatusByIds(ctx context.Context, status api.Status, ids []uint32) error {
	alarmPageIds := slices.To(ids, func(t uint32) uint {
		return uint(t)
	})
	return p.pageRepo.BatchUpdatePageStatusByIds(ctx, valueobj.Status(status), alarmPageIds)
}

// DeletePageByIds 通过id删除页面
func (p *PageBiz) DeletePageByIds(ctx context.Context, ids ...uint32) error {
	alarmPageIds := slices.To(ids, func(t uint32) uint {
		return uint(t)
	})
	return p.pageRepo.DeletePageByIds(ctx, alarmPageIds...)
}

// GetPageById 通过id获取页面详情
func (p *PageBiz) GetPageById(ctx context.Context, id uint32) (*bo.AlarmPageBO, error) {
	pageBO, err := p.pageRepo.GetPageById(ctx, uint(id))
	if err != nil {
		return nil, err
	}

	return pageBO, nil
}

// ListPage 获取页面列表
func (p *PageBiz) ListPage(ctx context.Context, req *pb.ListAlarmPageRequest) ([]*bo.AlarmPageBO, query.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		page.LikePageName(req.GetKeyword()),
		page.StatusEQ(int32(req.GetStatus())),
	}

	pageBos, err := p.pageRepo.ListPage(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}

	return pageBos, pgInfo, nil
}

// SelectPageList 获取页面列表
func (p *PageBiz) SelectPageList(ctx context.Context, req *pb.SelectAlarmPageRequest) ([]*bo.AlarmPageBO, query.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		page.LikePageName(req.GetKeyword()),
		page.StatusEQ(int32(req.GetStatus())),
	}

	pageBos, err := p.pageRepo.ListPage(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}
	return pageBos, pgInfo, nil
}
