package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/gen"
)

type (
	teamSendTemplateRepoImpl struct {
		data *data.Data
	}
)

// NewTeamSendTemplateRepository 创建团队发送模板仓库
func NewTeamSendTemplateRepository(data *data.Data) repository.TeamSendTemplate {
	return &teamSendTemplateRepoImpl{
		data: data,
	}
}

func (t *teamSendTemplateRepoImpl) Create(ctx context.Context, params *bo.CreateSendTemplate) error {
	templateModel := createTeamSendTemplateParamToModel(ctx, params)
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	return bizQuery.WithContext(ctx).SysSendTemplate.Create(templateModel)
}

func (t *teamSendTemplateRepoImpl) UpdateByID(ctx context.Context, params *bo.UpdateSendTemplate) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	id := params.ID
	param := params.UpdateParam
	sendTemplateModel := createTeamSendTemplateParamToModel(ctx, param)
	if _, err := bizQuery.WithContext(ctx).SysSendTemplate.Where(bizQuery.SysSendTemplate.ID.Eq(id)).Updates(sendTemplateModel); err != nil {
		return err
	}
	return nil
}

func (t *teamSendTemplateRepoImpl) DeleteByID(ctx context.Context, ID uint32) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	if _, err := bizQuery.SysSendTemplate.WithContext(ctx).Where(bizQuery.SysSendTemplate.ID.Eq(ID)).Delete(); err != nil {
		return err
	}
	return nil
}

func (t *teamSendTemplateRepoImpl) FindByPage(ctx context.Context, params *bo.QuerySendTemplateListParams) ([]imodel.ISendTemplate, error) {
	return t.listSendTemplateModels(ctx, params)
}

func (t *teamSendTemplateRepoImpl) UpdateStatusByIds(ctx context.Context, params *bo.UpdateSendTemplateStatusParams) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	status := params.Status
	ids := params.Ids
	_, err = bizQuery.SysSendTemplate.WithContext(ctx).Where(bizQuery.SysSendTemplate.ID.In(ids...)).UpdateSimple(bizQuery.SysSendTemplate.Status.Value(status.GetValue()))
	if err != nil {
		return err
	}
	return nil
}

func (t *teamSendTemplateRepoImpl) GetByID(ctx context.Context, ID uint32) (imodel.ISendTemplate, error) {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return nil, err
	}
	return bizQuery.SysSendTemplate.Where(bizQuery.SysSendTemplate.ID.Eq(ID)).First()
}

func createTeamSendTemplateParamToModel(ctx context.Context, param *bo.CreateSendTemplate) *bizmodel.SysSendTemplate {
	if types.IsNil(param) {
		return nil
	}
	templateModel := &bizmodel.SysSendTemplate{
		Name:     param.Name,
		Content:  param.Content,
		SendType: param.SendType,
		Status:   param.Status,
		Remark:   param.Remark,
	}
	templateModel.WithContext(ctx)
	return templateModel
}

func (t *teamSendTemplateRepoImpl) listSendTemplateModels(ctx context.Context, params *bo.QuerySendTemplateListParams) ([]imodel.ISendTemplate, error) {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return nil, err
	}
	queryWrapper := bizQuery.SysSendTemplate.WithContext(ctx)
	var wheres []gen.Condition
	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.SysSendTemplate.Status.Eq(params.Status.GetValue()))
	}

	if params.SendType.IsUnknown() {
		wheres = append(wheres, bizQuery.SysSendTemplate.SendType.Eq(params.SendType.GetValue()))
	}

	if types.TextIsNull(params.Keyword) {
		queryWrapper = queryWrapper.Or(bizQuery.SysSendTemplate.Name.Like(params.Keyword))
		queryWrapper = queryWrapper.Or(bizQuery.SysSendTemplate.Remark.Like(params.Keyword))
	}
	queryWrapper = queryWrapper.Where(wheres...)
	if queryWrapper, err = types.WithPageQuery(queryWrapper, params.Page); err != nil {
		return nil, err
	}

	dbTemplate, err := queryWrapper.Order(bizQuery.SysSendTemplate.ID.Desc()).Find()
	if !types.IsNil(err) {
		return nil, err
	}

	templateList := types.SliceTo(dbTemplate, func(item *bizmodel.SysSendTemplate) imodel.ISendTemplate {
		return item
	})

	return templateList, nil
}
