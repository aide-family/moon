package biz

import (
	"context"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/go-kratos/kratos/v2/log"
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
func (b *EndpointBiz) AppendEndpoint(ctx context.Context, endpoint *bo.CreateEndpointReq) (*bo.EndpointBO, error) {
	newData, err := b.endpointRepo.Append(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionCreate, &bo.SysLogBo{
		Content:    newData.String(),
		Title:      "新增数据源",
		ModuleId:   newData.Id,
		ModuleName: vobj.ModuleDatasource,
	})
	return newData, nil
}

// UpdateEndpointById 更新
func (b *EndpointBiz) UpdateEndpointById(ctx context.Context, endpoint *bo.UpdateEndpointReq) (*bo.EndpointBO, error) {
	// 查询
	oldData, err := b.endpointRepo.Get(ctx, basescopes.InIds(endpoint.Id))
	if err != nil {
		return nil, err
	}
	updateInfo := oldData
	updateInfo.Name = endpoint.Name
	updateInfo.Endpoint = endpoint.Endpoint
	updateInfo.Remark = endpoint.Remark
	updateInfo.DatasourceCategory = endpoint.DatasourceCategory
	if endpoint.Password != "" && endpoint.Username != "" {
		updateInfo.BasicAuth = strategy.NewBasicAuth(endpoint.Username, endpoint.Password)
	}

	newData, err := b.endpointRepo.Update(ctx, updateInfo)
	if err != nil {
		return nil, err
	}

	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		Content:    bo.NewChangeLogBo(oldData, newData).String(),
		Title:      "更新数据源",
		ModuleId:   newData.Id,
		ModuleName: vobj.ModuleDatasource,
	})

	return newData, nil
}

// UpdateStatusByIds 批量更新状态
func (b *EndpointBiz) UpdateStatusByIds(ctx context.Context, ids []uint32, status vobj.Status) error {
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
			ModuleName: vobj.ModuleDatasource,
		}
	})
	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, list...)
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
			ModuleName: vobj.ModuleDatasource,
		}
	})
	b.logX.CreateSysLog(ctx, vobj.ActionDelete, list...)
	return nil
}

// ListEndpoint 查询
func (b *EndpointBiz) ListEndpoint(ctx context.Context, params *bo.ListEndpointReq) ([]*bo.EndpointBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.NameLike(params.Keyword),
		basescopes.StatusEQ(params.Status),
		basescopes.UpdateAtDesc(),
		do.EndpointInDatasourceType(params.DatasourceCategoryList...),
	}

	list, err := b.endpointRepo.List(ctx, params.Page, wheres...)
	if err != nil {
		return nil, err
	}
	return list, nil
}
