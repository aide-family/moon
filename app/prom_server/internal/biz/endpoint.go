package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	EndpointBiz struct {
		log *log.Helper

		endpointRepo repository.EndpointRepo
		logX         repository.SysLogRepo
	}
)

func NewEndpointBiz(endpointRepo repository.EndpointRepo, logX repository.SysLogRepo, logger log.Logger) *EndpointBiz {
	return &EndpointBiz{
		log:          log.NewHelper(log.With(logger, "module", "biz.Endpoint")),
		endpointRepo: endpointRepo,
		logX:         logX,
	}
}

// AppendEndpoint 新增
func (b *EndpointBiz) AppendEndpoint(ctx context.Context, endpoint *bo.EndpointBO) (*bo.EndpointBO, error) {
	newData, err := b.endpointRepo.Append(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	b.logX.CreateSysLog(ctx, vo.ActionCreate, &bo.SysLogBo{
		Content:    newData.String(),
		Title:      "新增数据源",
		ModuleId:   newData.Id,
		ModuleName: vo.ModuleDatasource,
	})
	return newData, nil
}

// UpdateEndpointById 更新
func (b *EndpointBiz) UpdateEndpointById(ctx context.Context, id uint32, endpoint *bo.EndpointBO) (*bo.EndpointBO, error) {
	// 查询
	oldData, err := b.endpointRepo.Get(ctx, basescopes.InIds(id))
	if err != nil {
		return nil, err
	}
	updateInfo := endpoint
	updateInfo.Id = id
	newData, err := b.endpointRepo.Update(ctx, updateInfo)
	if err != nil {
		return nil, err
	}

	b.logX.CreateSysLog(ctx, vo.ActionUpdate, &bo.SysLogBo{
		Content:    bo.NewChangeLogBo(oldData, newData).String(),
		Title:      "更新数据源",
		ModuleId:   newData.Id,
		ModuleName: vo.ModuleDatasource,
	})

	return newData, nil
}

// UpdateStatusByIds 批量更新状态
func (b *EndpointBiz) UpdateStatusByIds(ctx context.Context, ids []uint32, status vo.Status) error {
	if len(ids) == 0 {
		return nil
	}
	// 查询
	oldDataList, err := b.endpointRepo.GetByParams(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	if err := b.endpointRepo.UpdateStatus(ctx, ids, status); err != nil {
		return err
	}
	list := slices.To(oldDataList, func(item *bo.EndpointBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			Content:    bo.NewChangeLogBo(item.Status.String(), status.String()).String(),
			Title:      "更新数据源状态",
			ModuleId:   item.Id,
			ModuleName: vo.ModuleDatasource,
		}
	})
	b.logX.CreateSysLog(ctx, vo.ActionUpdate, list...)
	return nil
}

// DetailById 查询详情
func (b *EndpointBiz) DetailById(ctx context.Context, id uint32) (*bo.EndpointBO, error) {
	return b.endpointRepo.Get(ctx, basescopes.InIds(id))
}

// DeleteEndpointById 删除
func (b *EndpointBiz) DeleteEndpointById(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}
	// 查询
	oldDataList, err := b.endpointRepo.GetByParams(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	if err = b.endpointRepo.Delete(ctx, ids); err != nil {
		return err
	}
	list := slices.To(oldDataList, func(item *bo.EndpointBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			Content:    item.String(),
			Title:      "删除数据源",
			ModuleId:   item.Id,
			ModuleName: vo.ModuleDatasource,
		}
	})
	b.logX.CreateSysLog(ctx, vo.ActionDelete, list...)
	return nil
}

// ListEndpoint 查询
func (b *EndpointBiz) ListEndpoint(ctx context.Context, params *bo.ListEndpointReq) ([]*bo.EndpointBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.NameLike(params.Keyword),
		basescopes.StatusEQ(params.Status),
		basescopes.UpdateAtDesc(),
	}

	list, err := b.endpointRepo.List(ctx, params.Page, wheres...)
	if err != nil {
		return nil, err
	}
	return list, nil
}
