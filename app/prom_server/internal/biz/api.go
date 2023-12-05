package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/systemscopes"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type ApiBiz struct {
	log *log.Helper

	apiRepo  repository.ApiRepo
	dataRepo repository.DataRepo
}

func NewApiBiz(repo repository.ApiRepo, dataRepo repository.DataRepo, logger log.Logger) *ApiBiz {
	return &ApiBiz{
		apiRepo:  repo,
		dataRepo: dataRepo,
		log:      log.NewHelper(log.With(logger, "module", "biz.api")),
	}
}

// CreateApi 创建api
func (b *ApiBiz) CreateApi(ctx context.Context, apiBoList ...*bo.ApiBO) ([]*bo.ApiBO, error) {

	apiBoList, err := b.apiRepo.Create(ctx, apiBoList...)
	if err != nil {
		return nil, err
	}

	ids := slices.To[*bo.ApiBO, uint](apiBoList, func(t *bo.ApiBO) uint {
		return t.Id
	})
	b.cacheApiByIds(ids...)
	return apiBoList, nil
}

// GetApiById 获取api
func (b *ApiBiz) GetApiById(ctx context.Context, id uint) (*bo.ApiBO, error) {
	apiBO, err := b.apiRepo.Get(ctx, systemscopes.ApiInIds(id))
	if err != nil {
		return nil, err
	}

	return apiBO, nil
}

// ListApi 获取api列表
func (b *ApiBiz) ListApi(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.ApiBO, error) {
	apiBOList, err := b.apiRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return apiBOList, nil
}

// DeleteApiById 删除api
func (b *ApiBiz) DeleteApiById(ctx context.Context, id uint) error {
	if err := b.apiRepo.Delete(ctx, systemscopes.ApiInIds(id)); err != nil {
		return err
	}
	b.cacheApiByIds(id)
	return nil
}

// UpdateApiById 更新api
func (b *ApiBiz) UpdateApiById(ctx context.Context, id uint, apiBO *bo.ApiBO) (*bo.ApiBO, error) {
	apiBO, err := b.apiRepo.Update(ctx, apiBO, systemscopes.ApiInIds(id))
	if err != nil {
		return nil, err
	}
	b.cacheApiByIds(id)

	return apiBO, nil
}

// cacheApiByIds 缓存api
func (b *ApiBiz) cacheApiByIds(apiIds ...uint) {
	go func() {
		defer after.Recover(b.log)
		db, err := b.dataRepo.DB()
		if err != nil {
			return
		}
		cacheClient, err := b.dataRepo.Client()
		if err != nil {
			return
		}
		if err = model.CacheApiSimple(db, cacheClient, apiIds...); err != nil {
			b.log.Error(err)
		}
	}()
}

// UpdateApiStatusById 更新api状态
func (b *ApiBiz) UpdateApiStatusById(ctx context.Context, status valueobj.Status, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	apiBo := &bo.ApiBO{
		Status: status,
	}
	if err := b.apiRepo.UpdateAll(ctx, apiBo, systemscopes.ApiInIds(ids...)); err != nil {
		return err
	}
	b.cacheApiByIds(ids...)
	return nil
}
