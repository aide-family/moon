package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/pkg/helper/model/system"
)

type ApiBiz struct {
	log *log.Helper

	apiRepo repository.ApiRepo
}

func NewApiBiz(repo repository.ApiRepo, logger log.Logger) *ApiBiz {
	return &ApiBiz{
		apiRepo: repo,
		log:     log.NewHelper(log.With(logger, "module", "biz.api")),
	}
}

// CreateApi 创建api
func (b *ApiBiz) CreateApi(ctx context.Context, apiDoList ...*dobo.ApiBO) ([]*dobo.ApiBO, error) {
	apiDOList := dobo.NewApiBO(apiDoList...).DO().List()
	apiDOList, err := b.apiRepo.Create(ctx, apiDOList...)
	if err != nil {
		return nil, err
	}

	return dobo.NewApiDO(apiDOList...).BO().List(), nil
}

// GetApiById 获取api
func (b *ApiBiz) GetApiById(ctx context.Context, id uint) (*dobo.ApiBO, error) {
	apiDO, err := b.apiRepo.Get(ctx, system.ApiInIds(id))
	if err != nil {
		return nil, err
	}

	return dobo.NewApiDO(apiDO).BO().First(), nil
}

// ListApi 获取api列表
func (b *ApiBiz) ListApi(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.ApiBO, error) {
	apiDOList, err := b.apiRepo.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return dobo.NewApiDO(apiDOList...).BO().List(), nil
}

// DeleteApiById 删除api
func (b *ApiBiz) DeleteApiById(ctx context.Context, id uint) error {
	return b.apiRepo.Delete(ctx, system.ApiInIds(id))
}

// UpdateApiById 更新api
func (b *ApiBiz) UpdateApiById(ctx context.Context, id uint, apiBo *dobo.ApiBO) (*dobo.ApiBO, error) {
	apiDO := dobo.NewApiBO(apiBo).DO().First()
	apiDO, err := b.apiRepo.Update(ctx, apiDO, system.ApiInIds(id))
	if err != nil {
		return nil, err
	}

	return dobo.NewApiDO(apiDO).BO().First(), nil
}
