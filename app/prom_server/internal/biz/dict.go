package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api"
	dictpb "prometheus-manager/api/dict"
	"prometheus-manager/pkg/helper/model/dictscopes"
	"prometheus-manager/pkg/helper/valueobj"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
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
func (b *DictBiz) CreateDict(ctx context.Context, dict *bo.DictBO) (*bo.DictBO, error) {
	newDictBO, err := b.dictRepo.CreateDict(ctx, dict)
	if err != nil {
		return nil, err
	}
	return newDictBO, nil
}

// UpdateDict 更新字典
func (b *DictBiz) UpdateDict(ctx context.Context, dictBO *bo.DictBO) (*bo.DictBO, error) {
	newDictDO, err := b.dictRepo.UpdateDictById(ctx, dictBO.Id, dictBO)
	if err != nil {
		return nil, err
	}
	return newDictDO, nil
}

// BatchUpdateDictStatus 批量更新字典状态
func (b *DictBiz) BatchUpdateDictStatus(ctx context.Context, status api.Status, ids []uint32) error {
	return b.dictRepo.BatchUpdateDictStatusByIds(ctx, valueobj.Status(status), ids)
}

// DeleteDictByIds 删除字典
func (b *DictBiz) DeleteDictByIds(ctx context.Context, id ...uint32) error {
	return b.dictRepo.DeleteDictByIds(ctx, id...)
}

// GetDictById 获取字典详情
func (b *DictBiz) GetDictById(ctx context.Context, id uint32) (*bo.DictBO, error) {
	dictDetail, err := b.dictRepo.GetDictById(ctx, id)
	if err != nil {
		return nil, err
	}
	return dictDetail, nil
}

// ListDict 获取字典列表
func (b *DictBiz) ListDict(ctx context.Context, req *dictpb.ListDictRequest) ([]*bo.DictBO, *query.Page, error) {
	pageReq := req.GetPage()
	pgInfo := query.NewPage(pageReq.GetCurr(), pageReq.GetSize())

	wheres := []query.ScopeMethod{
		dictscopes.WhereCategory(int32(req.GetCategory())),
		dictscopes.LikeName(req.GetKeyword()),
		dictscopes.WithTrashed(req.GetIsDeleted()),
	}

	dictList, err := b.dictRepo.ListDict(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, nil, err
	}
	return dictList, pgInfo, nil
}

// SelectDict 获取字典列表
func (b *DictBiz) SelectDict(ctx context.Context, req *dictpb.SelectDictRequest) ([]*bo.DictBO, *query.Page, error) {
	pageReq := req.GetPage()
	pgInfo := query.NewPage(pageReq.GetCurr(), pageReq.GetSize())

	wheres := []query.ScopeMethod{
		dictscopes.WhereCategory(int32(req.GetCategory())),
		dictscopes.LikeName(req.GetKeyword()),
		dictscopes.WithTrashed(req.GetIsDeleted()),
	}

	dictList, err := b.dictRepo.ListDict(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, nil, err
	}
	return dictList, pgInfo, nil
}
