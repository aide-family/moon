package api

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
)

var _ repository.ApiRepo = (*apiRepoImpl)(nil)

type apiRepoImpl struct {
	repository.UnimplementedApiRepo

	log  *log.Helper
	data *data.Data

	query.IAction[model.SysApi]
}

func (l *apiRepoImpl) Create(ctx context.Context, apiDoList ...*dobo.ApiDO) ([]*dobo.ApiDO, error) {
	newModelDataList := make([]*model.SysApi, 0, len(apiDoList))
	for _, apiDo := range apiDoList {
		newModelData := apiDo.ToModel()
		newModelDataList = append(newModelDataList, newModelData)
	}

	// 执行创建逻辑
	if err := l.WithContext(ctx).BatchCreate(newModelDataList, 100); err != nil {
		return nil, err
	}

	list := make([]*dobo.ApiDO, 0, len(apiDoList))
	for _, apiItem := range newModelDataList {
		newModelData := dobo.ApiModelToDO(apiItem)
		list = append(list, newModelData)
	}
	return list, nil
}

func (l *apiRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.ApiDO, error) {
	apiModelInfo, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}

	return dobo.ApiModelToDO(apiModelInfo), nil
}

func (l *apiRepoImpl) Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*dobo.ApiDO, error) {
	var apiModelInfoList []*model.SysApi

	if err := l.WithContext(ctx).DB().Scopes(scopes...).Find(&apiModelInfoList).Error; err != nil {
		return nil, err
	}

	list := make([]*dobo.ApiDO, 0, len(apiModelInfoList))
	for _, apiModelInfo := range apiModelInfoList {
		newModelData := dobo.ApiModelToDO(apiModelInfo)
		list = append(list, newModelData)
	}

	return list, nil
}

func (l *apiRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.ApiDO, error) {
	apiModelList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	list := make([]*dobo.ApiDO, 0, len(apiModelList))
	for _, apiModel := range apiModelList {
		newModelData := dobo.ApiModelToDO(apiModel)
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

func (l *apiRepoImpl) Update(ctx context.Context, apiDo *dobo.ApiDO, scopes ...query.ScopeMethod) (*dobo.ApiDO, error) {
	if len(scopes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "not allow not condition update")
	}

	newModelInfo := apiDo.ToModel()
	if err := l.WithContext(ctx).Update(newModelInfo, scopes...); err != nil {
		return nil, err
	}

	return dobo.ApiModelToDO(newModelInfo), nil
}

func NewApiRepo(data *data.Data, logger log.Logger) repository.ApiRepo {
	return &apiRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.repository.api")),
		data: data,

		IAction: query.NewAction[model.SysApi](
			query.WithDB[model.SysApi](data.DB()),
		),
	}
}
