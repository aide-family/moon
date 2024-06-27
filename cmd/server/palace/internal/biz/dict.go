package biz

import (
	"context"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

func NewDictBiz(dictRepo repository.Dict) *DictBiz {
	return &DictBiz{
		dictRepo: dictRepo,
	}
}

type DictBiz struct {
	dictRepo repository.Dict
}

// CreateDict 创建字典
func (b *DictBiz) CreateDict(ctx context.Context, dictParam *bo.CreateDictParams) (*model.SysDict, error) {
	dictDo, err := b.dictRepo.Create(ctx, dictParam)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return dictDo, nil
}

// UpdateDict 更新字典
func (b *DictBiz) UpdateDict(ctx context.Context, updateParam *bo.UpdateDictParams) error {
	if err := b.dictRepo.UpdateByID(ctx, updateParam); !types.IsNil(err) {
		return err
	}
	return nil
}

// ListDict 列表字典
func (b *DictBiz) ListDict(ctx context.Context, listParam *bo.QueryDictListParams) ([]*model.SysDict, error) {
	dictDos, err := b.dictRepo.FindByPage(ctx, listParam)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return dictDos, nil

}

// GetDict 获取字典
func (b *DictBiz) GetDict(ctx context.Context, dictId uint32) (*model.SysDict, error) {
	return b.dictRepo.GetByID(ctx, dictId)
}

// UpdateDictStatusByIds 更新字典状态
func (b *DictBiz) UpdateDictStatusByIds(ctx context.Context, updateParams *bo.UpdateDictStatusParams) error {
	return b.dictRepo.UpdateStatusByIds(ctx, updateParams.Status, updateParams.IDs...)
}

// DeleteDictById 删除字典
func (b *DictBiz) DeleteDictById(ctx context.Context, dictId uint32) error {
	return b.dictRepo.DeleteByID(ctx, dictId)
}
