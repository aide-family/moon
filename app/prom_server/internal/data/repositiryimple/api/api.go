package api

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
)

var _ repository.ApiRepo = (*apiRepoImpl)(nil)

type apiRepoImpl struct {
	repository.UnimplementedApiRepo

	log  *log.Helper
	data *data.Data

	query.IAction[model.SysAPI]
}

func (l *apiRepoImpl) Create(ctx context.Context, apiBOList ...*bo.ApiBO) ([]*bo.ApiBO, error) {
	newModelDataList := make([]*model.SysAPI, 0, len(apiBOList))
	for _, apiBO := range apiBOList {
		newModelData := apiBO.ToModel()
		newModelDataList = append(newModelDataList, newModelData)
	}

	// 执行创建逻辑
	if err := l.WithContext(ctx).BatchCreate(newModelDataList, 100); err != nil {
		return nil, err
	}

	list := make([]*bo.ApiBO, 0, len(apiBOList))
	for _, apiItem := range newModelDataList {
		newModelData := bo.ApiModelToBO(apiItem)
		list = append(list, newModelData)
	}
	return list, nil
}

func (l *apiRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.ApiBO, error) {
	apiModelInfo, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}

	return bo.ApiModelToBO(apiModelInfo), nil
}

func (l *apiRepoImpl) Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*bo.ApiBO, error) {
	var apiModelInfoList []*model.SysAPI

	if err := l.WithContext(ctx).DB().Scopes(scopes...).Find(&apiModelInfoList).Error; err != nil {
		return nil, err
	}

	list := make([]*bo.ApiBO, 0, len(apiModelInfoList))
	for _, apiModelInfo := range apiModelInfoList {
		newModelData := bo.ApiModelToBO(apiModelInfo)
		list = append(list, newModelData)
	}

	return list, nil
}

func (l *apiRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.ApiBO, error) {
	apiModelList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	list := make([]*bo.ApiBO, 0, len(apiModelList))
	for _, apiModel := range apiModelList {
		newModelData := bo.ApiModelToBO(apiModel)
		list = append(list, newModelData)
	}
	return list, nil
}

func (l *apiRepoImpl) Delete(ctx context.Context, scopes ...query.ScopeMethod) error {
	// 不允许不带条件执行
	if len(scopes) == 0 {
		return status.Error(codes.InvalidArgument, "not allow not condition delete")
	}
	return l.WithContext(ctx).Delete(scopes...)
}

func (l *apiRepoImpl) Update(ctx context.Context, apiBO *bo.ApiBO, scopes ...query.ScopeMethod) (*bo.ApiBO, error) {
	if len(scopes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "not allow not condition update")
	}

	newModelInfo := apiBO.ToModel()
	if err := l.WithContext(ctx).Update(newModelInfo, scopes...); err != nil {
		return nil, err
	}

	return bo.ApiModelToBO(newModelInfo), nil
}

func NewApiRepo(data *data.Data, logger log.Logger) repository.ApiRepo {
	return &apiRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.repository.api")),
		data: data,

		IAction: query.NewAction[model.SysAPI](
			query.WithDB[model.SysAPI](data.DB()),
		),
	}
}
