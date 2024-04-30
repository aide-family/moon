package biz

import (
	"context"

	"github.com/aide-family/moon/api/perrors"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/go-kratos/kratos/v2/log"
)

func NewSysDictBiz(sysDictRepo repository.SysDictRepo, logger log.Logger) *SysDictBiz {
	return &SysDictBiz{
		sysDictRepo: sysDictRepo,
		log:         log.NewHelper(log.With(logger, "module", "biz.sys_dict")),
	}
}

type SysDictBiz struct {
	sysDictRepo repository.SysDictRepo

	log *log.Helper
}

// CreateDict 创建字典
func (b *SysDictBiz) CreateDict(ctx context.Context, dict *bo.CreateSysDictBo) (*do.SysDict, error) {
	return b.sysDictRepo.CreateDict(ctx, dict)
}

// UpdateDict 更新字典
func (b *SysDictBiz) UpdateDict(ctx context.Context, dict *bo.UpdateSysDictBo) (*do.SysDict, error) {
	if dict.ID == 0 {
		return nil, perrors.ErrorInvalidParams("不合法参数")
	}
	return b.sysDictRepo.UpdateDictById(ctx, dict.ID, dict)
}

// BatchUpdateDictStatus 批量更新字典状态
func (b *SysDictBiz) BatchUpdateDictStatus(ctx context.Context, status vobj.Status, ids []uint32) error {
	return b.sysDictRepo.BatchUpdateDictStatusByIds(ctx, status, ids)
}

// DeleteDictByIds 删除字典
func (b *SysDictBiz) DeleteDictByIds(ctx context.Context, id ...uint32) error {
	return b.sysDictRepo.DeleteDictByIds(ctx, id...)
}

// GetDictById 获取字典详情
func (b *SysDictBiz) GetDictById(ctx context.Context, id uint32) (*do.SysDict, error) {
	return b.sysDictRepo.GetDictById(ctx, id)
}

// ListDict 获取字典列表
func (b *SysDictBiz) ListDict(ctx context.Context, req *bo.ListSysDictBo) ([]*do.SysDict, error) {
	return b.sysDictRepo.ListDict(ctx, req)
}

// SelectDict 获取字典下拉列表
func (b *SysDictBiz) SelectDict(ctx context.Context, req *bo.SelectSysDictBo) ([]*do.SysDict, error) {
	return b.sysDictRepo.SelectDict(ctx, req)
}
