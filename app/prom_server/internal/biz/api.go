package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/after"
	"github.com/aide-family/moon/pkg/util/slices"
)

type ApiBiz struct {
	log *log.Helper

	apiRepo  repository.ApiRepo
	dataRepo repository.DataRepo
	logX     repository.SysLogRepo
}

func NewApiBiz(
	repo repository.ApiRepo,
	dataRepo repository.DataRepo,
	logX repository.SysLogRepo,
	logger log.Logger,
) *ApiBiz {
	return &ApiBiz{
		apiRepo:  repo,
		dataRepo: dataRepo,
		logX:     logX,
		log:      log.NewHelper(log.With(logger, "module", "biz.api")),
	}
}

// CreateApi 创建api
func (b *ApiBiz) CreateApi(ctx context.Context, apiBoList ...*bo.ApiBO) ([]*bo.ApiBO, error) {
	apiBoList, err := b.apiRepo.Create(ctx, apiBoList...)
	if err != nil {
		return nil, err
	}

	ids := slices.To(apiBoList, func(t *bo.ApiBO) uint32 {
		return t.Id
	})
	b.cacheApiByIds(ids...)
	list := slices.To(apiBoList, func(item *bo.ApiBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vobj.ModuleApi,
			ModuleId:   item.Id,
			Content:    item.String(),
			Title:      "创建API",
		}
	})
	b.logX.CreateSysLog(ctx, vobj.ActionCreate, list...)
	return apiBoList, nil
}

// GetApiById 获取api
func (b *ApiBiz) GetApiById(ctx context.Context, id uint32) (*bo.ApiBO, error) {
	apiBO, err := b.apiRepo.Get(ctx, basescopes.InIds(id), do.SysAPIPreloadRoles())
	if err != nil {
		return nil, err
	}

	return apiBO, nil
}

// ListApi 获取api列表
func (b *ApiBiz) ListApi(ctx context.Context, params *bo.ApiListApiReq) ([]*bo.ApiBO, bo.Pagination, error) {
	pgInfo := bo.NewPage(params.Curr, params.Size)
	scopes := []basescopes.ScopeMethod{
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
		basescopes.StatusEQ(params.Status),
		do.SysApiLike(params.Keyword),
	}
	apiBOList, err := b.apiRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}

	return apiBOList, pgInfo, nil
}

// ListAllApi 获取api列表
func (b *ApiBiz) ListAllApi(ctx context.Context) ([]*bo.ApiBO, error) {
	apiBOList, err := b.apiRepo.Find(ctx)
	if err != nil {
		return nil, err
	}

	return apiBOList, nil
}

// DeleteApiById 删除api
func (b *ApiBiz) DeleteApiById(ctx context.Context, id uint32) error {
	// 查询
	apiBO, err := b.GetApiById(ctx, id)
	if err != nil {
		return err
	}
	if err = b.apiRepo.Delete(ctx, basescopes.InIds(id)); err != nil {
		return err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionDelete, &bo.SysLogBo{
		ModuleName: vobj.ModuleApi,
		ModuleId:   id,
		Content:    apiBO.String(),
		Title:      "删除API",
	})
	b.cacheApiByIds(id)
	return nil
}

// UpdateApiById 更新api
func (b *ApiBiz) UpdateApiById(ctx context.Context, id uint32, apiBO *bo.ApiBO) (*bo.ApiBO, error) {
	// 查询
	oldApiBO, err := b.GetApiById(ctx, id)
	if err != nil {
		return nil, err
	}
	newApiBO, err := b.apiRepo.Update(ctx, apiBO, basescopes.InIds(id))
	if err != nil {
		return nil, err
	}
	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		ModuleName: vobj.ModuleApi,
		ModuleId:   id,
		Content:    bo.NewChangeLogBo(oldApiBO, newApiBO).String(),
		Title:      "更新API",
	})
	b.cacheApiByIds(id)

	return apiBO, nil
}

// cacheApiByIds 缓存api
func (b *ApiBiz) cacheApiByIds(apiIds ...uint32) {
	go func() {
		defer after.Recover(b.log)
		db, err := b.dataRepo.DB()
		if err != nil {
			return
		}
		cacheClient, err := b.dataRepo.Cache()
		if err != nil {
			return
		}
		if err = do.CacheApiSimple(db, cacheClient, apiIds...); err != nil {
			b.log.Error(err)
		}
	}()
}

// UpdateApiStatusById 更新api状态
func (b *ApiBiz) UpdateApiStatusById(ctx context.Context, status vobj.Status, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	// 查询
	oldList, err := b.apiRepo.Find(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	apiBo := &bo.ApiBO{
		Status: status,
	}
	if err := b.apiRepo.UpdateAll(ctx, apiBo, basescopes.InIds(ids...)); err != nil {
		return err
	}
	list := slices.To(oldList, func(old *bo.ApiBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vobj.ModuleApi,
			ModuleId:   old.Id,
			Content:    bo.NewChangeLogBo(old.Status.String(), status.String()).String(),
			Title:      "更新API状态",
		}
	})
	b.cacheApiByIds(ids...)
	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, list...)

	return nil
}
