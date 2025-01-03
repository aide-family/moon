package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/gen"
)

// NewSendTemplateRepository new send template repository
func NewSendTemplateRepository(data *data.Data) repository.SendTemplateRepo {
	return &sendTemplateRepositoryImpl{data: data}
}

type sendTemplateRepositoryImpl struct {
	data *data.Data
}

func (s *sendTemplateRepositoryImpl) GetTemplateInfoByName(ctx context.Context, name string) (imodel.ISendTemplate, error) {
	mainQuery := query.Use(s.data.GetMainDB(ctx))
	return mainQuery.SysSendTemplate.WithContext(ctx).Where(mainQuery.SysSendTemplate.Name.Eq(name)).First()
}

func (s *sendTemplateRepositoryImpl) Create(ctx context.Context, params *bo.CreateSendTemplate) error {
	templateModel := createSendTemplateParamToModel(ctx, params)
	mainQuery := query.Use(s.data.GetMainDB(ctx))
	return mainQuery.WithContext(ctx).SysSendTemplate.Create(templateModel)
}

func (s *sendTemplateRepositoryImpl) UpdateByID(ctx context.Context, params *bo.UpdateSendTemplate) error {
	mainQuery := query.Use(s.data.GetMainDB(ctx))
	templateModel := createSendTemplateParamToModel(ctx, params.UpdateParam)
	if _, err := mainQuery.SysSendTemplate.WithContext(ctx).
		Update(mainQuery.SysSendTemplate.ID.Eq(params.ID), templateModel); err != nil {
		return err
	}
	return nil
}

func (s *sendTemplateRepositoryImpl) DeleteByID(ctx context.Context, ID uint32) error {
	mainQuery := query.Use(s.data.GetMainDB(ctx))
	if _, err := mainQuery.SysSendTemplate.WithContext(ctx).Where(mainQuery.SysSendTemplate.ID.Eq(ID)).Delete(); err != nil {
		return err
	}
	return nil
}

func (s *sendTemplateRepositoryImpl) FindByPage(ctx context.Context, params *bo.QuerySendTemplateListParams) ([]imodel.ISendTemplate, error) {
	return s.listSendTemplateModels(ctx, params)
}

func (s *sendTemplateRepositoryImpl) UpdateStatusByIds(ctx context.Context, params *bo.UpdateSendTemplateStatusParams) error {
	status := params.Status
	ids := params.Ids
	mainQuery := query.Use(s.data.GetMainDB(ctx))
	if _, err := mainQuery.StrategyTemplate.WithContext(ctx).Where(mainQuery.SysSendTemplate.ID.In(ids...)).UpdateSimple(mainQuery.SysSendTemplate.Status.Value(status.GetValue())); err != nil {
		return err
	}
	return nil
}

func (s *sendTemplateRepositoryImpl) GetByID(ctx context.Context, ID uint32) (imodel.ISendTemplate, error) {
	mainQuery := query.Use(s.data.GetMainDB(ctx))
	return mainQuery.SysSendTemplate.WithContext(ctx).Where(mainQuery.SysSendTemplate.ID.Eq(ID)).First()
}

func createSendTemplateParamToModel(ctx context.Context, param *bo.CreateSendTemplate) *model.SysSendTemplate {
	if types.IsNil(param) {
		return nil
	}
	templateModel := &model.SysSendTemplate{
		Name:     param.Name,
		Content:  param.Content,
		SendType: param.SendType,
		Status:   param.Status,
		Remark:   param.Remark,
	}
	templateModel.WithContext(ctx)
	return templateModel
}

func (s *sendTemplateRepositoryImpl) listSendTemplateModels(ctx context.Context, params *bo.QuerySendTemplateListParams) ([]imodel.ISendTemplate, error) {
	sendQuery := query.Use(s.data.GetMainDB(ctx)).SysSendTemplate
	queryWrapper := sendQuery.WithContext(ctx)
	var wheres []gen.Condition
	if !params.Status.IsUnknown() {
		wheres = append(wheres, sendQuery.Status.Eq(params.Status.GetValue()))
	}

	if params.SendType.IsUnknown() {
		wheres = append(wheres, sendQuery.SendType.Eq(params.SendType.GetValue()))
	}

	if !types.TextIsNull(params.Keyword) {
		queryWrapper = queryWrapper.Or(sendQuery.Name.Like(params.Keyword))
		queryWrapper = queryWrapper.Or(sendQuery.Remark.Like(params.Keyword))
	}
	var err error
	queryWrapper = queryWrapper.Where(wheres...)
	if queryWrapper, err = types.WithPageQuery(queryWrapper, params.Page); err != nil {
		return nil, err
	}

	dbTemplate, err := queryWrapper.Order(sendQuery.ID.Desc()).Find()
	if !types.IsNil(err) {
		return nil, err
	}

	templateList := types.SliceTo(dbTemplate, func(item *model.SysSendTemplate) imodel.ISendTemplate {
		return item
	})

	return templateList, nil
}
