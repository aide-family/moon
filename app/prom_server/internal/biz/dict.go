package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	DictBiz struct {
		log *log.Helper

		dictRepo repository.PromDictRepo
		logX     repository.SysLogRepo
	}
)

// NewDictBiz 实例化字典业务
func NewDictBiz(dictRepo repository.PromDictRepo, logX repository.SysLogRepo, logger log.Logger) *DictBiz {
	return &DictBiz{
		log:      log.NewHelper(log.With(logger, "module", "biz.dict")),
		dictRepo: dictRepo,
		logX:     logX,
	}
}

// CreateDict 创建字典
func (b *DictBiz) CreateDict(ctx context.Context, dict *bo.DictBO) (*bo.DictBO, error) {
	newDictBO, err := b.dictRepo.CreateDict(ctx, dict)
	if err != nil {
		return nil, err
	}
	b.logX.CreateSysLog(ctx, vo.ActionCreate, &bo.SysLogBo{
		ModuleName: vo.ModuleDict,
		ModuleId:   newDictBO.Id,
		Content:    newDictBO.String(),
		Title:      "创建字典",
	})
	return newDictBO, nil
}

// UpdateDict 更新字典
func (b *DictBiz) UpdateDict(ctx context.Context, dictBO *bo.DictBO) (*bo.DictBO, error) {
	// 查询
	dictDetail, err := b.dictRepo.GetDictById(ctx, dictBO.Id)
	if err != nil {
		return nil, err
	}
	newDictDO, err := b.dictRepo.UpdateDictById(ctx, dictBO.Id, dictBO)
	if err != nil {
		return nil, err
	}
	b.logX.CreateSysLog(ctx, vo.ActionUpdate, &bo.SysLogBo{
		ModuleName: vo.ModuleDict,
		ModuleId:   newDictDO.Id,
		Content:    bo.NewChangeLogBo(dictDetail, newDictDO).String(),
		Title:      "更新字典",
	})
	return newDictDO, nil
}

// BatchUpdateDictStatus 批量更新字典状态
func (b *DictBiz) BatchUpdateDictStatus(ctx context.Context, status vo.Status, ids []uint32) error {
	// 查询
	oldList, err := b.dictRepo.GetDictByIds(ctx, ids...)
	if err != nil {
		return err
	}

	if err := b.dictRepo.BatchUpdateDictStatusByIds(ctx, status, ids); err != nil {
		return err
	}

	list := slices.To(oldList, func(old *bo.DictBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleDict,
			ModuleId:   0,
			Content:    bo.NewChangeLogBo(old.Status.String(), status.String()).String(),
			Title:      "批量更新字典",
		}
	})
	b.logX.CreateSysLog(ctx, vo.ActionUpdate, list...)
	return nil
}

// DeleteDictByIds 删除字典
func (b *DictBiz) DeleteDictByIds(ctx context.Context, id ...uint32) error {
	// 查询
	dictList, err := b.dictRepo.GetDictByIds(ctx, id...)
	if err != nil {
		return err
	}

	if err = b.dictRepo.DeleteDictByIds(ctx, id...); err != nil {
		return err
	}

	list := slices.To(dictList, func(dict *bo.DictBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleDict,
			ModuleId:   dict.Id,
			Content:    dict.String(),
			Title:      "删除字典",
		}
	})

	b.logX.CreateSysLog(ctx, vo.ActionDelete, list...)
	return nil
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
func (b *DictBiz) ListDict(ctx context.Context, req *bo.ListDictRequest) ([]*bo.DictBO, error) {
	wheres := []basescopes.ScopeMethod{
		do.SysDictWhereCategory(req.Category),
		basescopes.NameLike(req.Keyword),
		basescopes.WithTrashed(req.IsDeleted),
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
		basescopes.StatusEQ(req.Status),
	}

	dictList, err := b.dictRepo.ListDict(ctx, req.Page, wheres...)
	if err != nil {
		return nil, err
	}
	return dictList, nil
}
