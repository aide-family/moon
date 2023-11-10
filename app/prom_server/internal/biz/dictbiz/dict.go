package dictbiz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api"
	dictpb "prometheus-manager/api/dict"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/pkg/model/dict"
)

type (
	Biz struct {
		log *log.Helper

		dictRepo Repo
	}

	Repo interface {
		// CreateDict 创建字典
		CreateDict(ctx context.Context, dict *biz.DictDO) (*biz.DictDO, error)
		// UpdateDictById 通过id更新字典
		UpdateDictById(ctx context.Context, id uint, dict *biz.DictDO) (*biz.DictDO, error)
		// BatchUpdateDictStatusByIds 通过id批量更新字典状态
		BatchUpdateDictStatusByIds(ctx context.Context, status int32, ids []uint) error
		// DeleteDictByIds 通过id删除字典
		DeleteDictByIds(ctx context.Context, id ...uint) error
		// GetDictById 通过id获取字典详情
		GetDictById(ctx context.Context, id uint) (*biz.DictDO, error)
		// ListDict 获取字典列表
		ListDict(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*biz.DictDO, error)
	}
)

// NewBiz 实例化字典业务
func NewBiz(dictRepo Repo, logger log.Logger) *Biz {
	return &Biz{
		log:      log.NewHelper(log.With(logger, "module", "biz.dict")),
		dictRepo: dictRepo,
	}
}

// CreateDict 创建字典
func (b *Biz) CreateDict(ctx context.Context, dict *biz.DictBO) (*biz.DictBO, error) {
	newDictBO, err := b.dictRepo.CreateDict(ctx, biz.NewDictBO(dict).DO().First())
	if err != nil {
		return nil, err
	}
	return biz.NewDictDO(newDictBO).BO().First(), nil
}

// UpdateDict 更新字典
func (b *Biz) UpdateDict(ctx context.Context, dict *biz.DictBO) (*biz.DictBO, error) {
	newDictDO, err := b.dictRepo.UpdateDictById(ctx, uint(dict.Id), biz.NewDictBO(dict).DO().First())
	if err != nil {
		return nil, err
	}
	return biz.NewDictDO(newDictDO).BO().First(), nil
}

// BatchUpdateDictStatus 批量更新字典状态
func (b *Biz) BatchUpdateDictStatus(ctx context.Context, status api.Status, ids []uint) error {
	return b.dictRepo.BatchUpdateDictStatusByIds(ctx, int32(status), ids)
}

// DeleteDictByIds 删除字典
func (b *Biz) DeleteDictByIds(ctx context.Context, id ...uint) error {
	return b.dictRepo.DeleteDictByIds(ctx, id...)
}

// GetDictById 获取字典详情
func (b *Biz) GetDictById(ctx context.Context, id uint) (*biz.DictBO, error) {
	dictDetail, err := b.dictRepo.GetDictById(ctx, id)
	if err != nil {
		return nil, err
	}
	return biz.NewDictDO(dictDetail).BO().First(), nil
}

// ListDict 获取字典列表
func (b *Biz) ListDict(ctx context.Context, req *dictpb.ListDictRequest) ([]*biz.DictBO, *query.Page, error) {
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
	return biz.NewDictDO(dictList...).BO().List(), pgInfo, nil
}

// SelectDict 获取字典列表
func (b *Biz) SelectDict(ctx context.Context, req *dictpb.SelectDictRequest) ([]*biz.DictBO, *query.Page, error) {
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
	return biz.NewDictDO(dictList...).BO().List(), pgInfo, nil
}
