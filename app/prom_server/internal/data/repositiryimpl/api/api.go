package api

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ repository.ApiRepo = (*apiRepoImpl)(nil)

type apiRepoImpl struct {
	repository.UnimplementedApiRepo

	log  *log.Helper
	data *data.Data
}

func (l *apiRepoImpl) Create(ctx context.Context, apiBOList ...*bo.ApiBO) ([]*bo.ApiBO, error) {
	newModelDataList := slices.To(apiBOList, func(item *bo.ApiBO) *do.SysAPI {
		return item.ToModel()
	})

	// 执行创建逻辑
	if err := l.data.DB().WithContext(ctx).CreateInBatches(newModelDataList, 100).Error; err != nil {
		return nil, err
	}

	list := slices.To(newModelDataList, func(item *do.SysAPI) *bo.ApiBO {
		return bo.ApiModelToBO(item)
	})
	return list, nil
}

func (l *apiRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.ApiBO, error) {
	var apiModelInfo do.SysAPI
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&apiModelInfo).Error; err != nil {
		return nil, err
	}

	return bo.ApiModelToBO(&apiModelInfo), nil
}

func (l *apiRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.ApiBO, error) {
	var apiModelInfoList []*do.SysAPI
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Find(&apiModelInfoList).Error; err != nil {
		return nil, err
	}

	list := slices.To(apiModelInfoList, func(item *do.SysAPI) *bo.ApiBO {
		return bo.ApiModelToBO(item)
	})

	return list, nil
}

func (l *apiRepoImpl) List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.ApiBO, error) {
	var apiModelInfoList []*do.SysAPI
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, bo.Page(pgInfo))...).Find(&apiModelInfoList).Error; err != nil {
		return nil, err
	}

	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Model(&do.SysAPI{}).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}

	list := slices.To(apiModelInfoList, func(item *do.SysAPI) *bo.ApiBO {
		return bo.ApiModelToBO(item)
	})
	return list, nil
}

func (l *apiRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
	// 不允许不带条件执行
	if len(scopes) == 0 {
		return status.Error(codes.InvalidArgument, "not allow not condition delete")
	}
	var detail do.SysAPI
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&detail).Error; err != nil {
		return err
	}
	return l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := basescopes.WithTx(ctx, tx)
		// 删除关联关系
		if err := tx.Model(&detail).WithContext(txCtx).Association(do.SysAPIPreloadFieldRoles).Clear(); err != nil {
			return err
		}
		// 删除主数据
		if err := tx.Model(&detail).WithContext(txCtx).Delete(do.SysAPI{}, detail.ID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (l *apiRepoImpl) Update(ctx context.Context, apiBO *bo.ApiBO, scopes ...basescopes.ScopeMethod) (*bo.ApiBO, error) {
	if len(scopes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "not allow not condition update")
	}

	// 根据条件查询即将修改的条数, 超过1条则不允许修改
	var total int64
	if err := l.data.DB().Model(&do.SysAPI{}).WithContext(ctx).Scopes(scopes...).Count(&total).Error; err != nil {
		return nil, err
	}
	if total > 1 {
		return nil, status.Error(codes.InvalidArgument, "not allow update more than one")
	}

	newModelInfo := apiBO.ToModel()
	if err := l.data.DB().Model(&do.SysAPI{}).WithContext(ctx).Scopes(scopes...).Updates(newModelInfo).Error; err != nil {
		return nil, err
	}

	return bo.ApiModelToBO(newModelInfo), nil
}

func (l *apiRepoImpl) UpdateAll(ctx context.Context, apiBO *bo.ApiBO, scopes ...basescopes.ScopeMethod) error {
	newModelInfo := apiBO.ToModel()
	return l.data.DB().Model(&do.SysAPI{}).WithContext(ctx).Scopes(scopes...).Updates(newModelInfo).Error
}

func NewApiRepo(data *data.Data, logger log.Logger) repository.ApiRepo {
	return &apiRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.repository.api")),
		data: data,
	}
}
