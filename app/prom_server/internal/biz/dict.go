package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"

	"prometheus-manager/api"
	dictpb "prometheus-manager/api/dict"
	"prometheus-manager/pkg/model/dict"
)

type (
	DictBiz struct {
		log *log.Helper

		dictRepo repository.PromDictRepo
	}
)

// NewDictBiz 实例化字典业务
func NewDictBiz(dictRepo repository.PromDictRepo, logger log.Logger) *DictBiz {
	return &DictBiz{
		log:      log.NewHelper(log.With(logger, "module", "biz.dict")),
		dictRepo: dictRepo,
	}
}

// CreateDict 创建字典
func (b *DictBiz) CreateDict(ctx context.Context, dict *dobo.DictBO) (*dobo.DictBO, error) {
	newDictBO, err := b.dictRepo.CreateDict(ctx, dobo.NewDictBO(dict).DO().First())
	if err != nil {
		return nil, err
	}
	return dobo.NewDictDO(newDictBO).BO().First(), nil
}

// UpdateDict 更新字典
func (b *DictBiz) UpdateDict(ctx context.Context, dict *dobo.DictBO) (*dobo.DictBO, error) {
	newDictDO, err := b.dictRepo.UpdateDictById(ctx, uint(dict.Id), dobo.NewDictBO(dict).DO().First())
	if err != nil {
		return nil, err
	}
	return dobo.NewDictDO(newDictDO).BO().First(), nil
}

// BatchUpdateDictStatus 批量更新字典状态
func (b *DictBiz) BatchUpdateDictStatus(ctx context.Context, status api.Status, ids []uint) error {
	return b.dictRepo.BatchUpdateDictStatusByIds(ctx, int32(status), ids)
}

// DeleteDictByIds 删除字典
func (b *DictBiz) DeleteDictByIds(ctx context.Context, id ...uint) error {
	return b.dictRepo.DeleteDictByIds(ctx, id...)
}

// GetDictById 获取字典详情
func (b *DictBiz) GetDictById(ctx context.Context, id uint) (*dobo.DictBO, error) {
	dictDetail, err := b.dictRepo.GetDictById(ctx, id)
	if err != nil {
		return nil, err
	}
	return dobo.NewDictDO(dictDetail).BO().First(), nil
}

// ListDict 获取字典列表
func (b *DictBiz) ListDict(ctx context.Context, req *dictpb.ListDictRequest) ([]*dobo.DictBO, *query.Page, error) {
	pageReq := req.GetPage()
	pgInfo := query.NewPage(int(pageReq.GetCurr()), int(pageReq.GetSize()))

	wheres := []query.ScopeMethod{
		dict.WhereCategory(int32(req.GetCategory())),
		dict.LikeName(req.GetKeyword()),
		dict.WithTrashed(req.GetIsDeleted()),
	}

	dictList, err := b.dictRepo.ListDict(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, nil, err
	}
	return dobo.NewDictDO(dictList...).BO().List(), pgInfo, nil
}

// SelectDict 获取字典列表
func (b *DictBiz) SelectDict(ctx context.Context, req *dictpb.SelectDictRequest) ([]*dobo.DictBO, *query.Page, error) {
	pageReq := req.GetPage()
	pgInfo := query.NewPage(int(pageReq.GetCurr()), int(pageReq.GetSize()))

	wheres := []query.ScopeMethod{
		dict.WhereCategory(int32(req.GetCategory())),
		dict.LikeName(req.GetKeyword()),
		dict.WithTrashed(req.GetIsDeleted()),
	}

	dictList, err := b.dictRepo.ListDict(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, nil, err
	}
	return dobo.NewDictDO(dictList...).BO().List(), pgInfo, nil
}
