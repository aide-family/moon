package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewDictBiz 创建字典业务
func NewDictBiz(dictRepo repository.Dict, teamDictRepo repository.TeamDict) *DictBiz {
	return &DictBiz{
		dictRepo:     dictRepo,
		teamDictRepo: teamDictRepo,
	}
}

// DictBiz 字典业务
type DictBiz struct {
	dictRepo     repository.Dict
	teamDictRepo repository.TeamDict
}

func (b *DictBiz) getDictRepo(ctx context.Context) repository.Dict {
	if middleware.GetSourceType(ctx).IsSystem() {
		return b.dictRepo
	}
	return b.teamDictRepo
}

// CreateDict 创建字典
func (b *DictBiz) CreateDict(ctx context.Context, dictParam *bo.CreateDictParams) (imodel.IDict, error) {
	dictDo, err := b.getDictRepo(ctx).Create(ctx, dictParam)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return dictDo, nil
}

// UpdateDict 更新字典
func (b *DictBiz) UpdateDict(ctx context.Context, updateParam *bo.UpdateDictParams) error {
	if err := b.getDictRepo(ctx).UpdateByID(ctx, updateParam); !types.IsNil(err) {
		return err
	}
	return nil
}

// ListDict 列表字典
func (b *DictBiz) ListDict(ctx context.Context, listParam *bo.QueryDictListParams) ([]imodel.IDict, error) {
	dictDos, err := b.getDictRepo(ctx).FindByPage(ctx, listParam)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return dictDos, nil

}

// GetDict 获取字典
func (b *DictBiz) GetDict(ctx context.Context, id uint32) (imodel.IDict, error) {
	dictDetail, err := b.getDictRepo(ctx).GetByID(ctx, id)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastDictNotFound(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return dictDetail, nil
}

// UpdateDictStatusByIds 更新字典状态
func (b *DictBiz) UpdateDictStatusByIds(ctx context.Context, updateParams *bo.UpdateDictStatusParams) error {
	return b.getDictRepo(ctx).UpdateStatusByIds(ctx, updateParams)
}

// DeleteDictByID 删除字典
func (b *DictBiz) DeleteDictByID(ctx context.Context, id uint32) error {
	return b.getDictRepo(ctx).DeleteByID(ctx, id)
}
