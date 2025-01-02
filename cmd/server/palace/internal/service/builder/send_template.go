package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	templateapi "github.com/aide-family/moon/api/admin/template"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ SendTemplateModuleBuild = (*sendTemplateModuleBuild)(nil)

type (
	SendTemplateModuleBuild interface {
		// WithSendTemplateCreateRequest 构建创建告警发送模板请求
		WithSendTemplateCreateRequest(*templateapi.CreateSendTemplateRequest) ICreateSendTemplateRequestBuilder
		// WithSendTemplateStatusUpdateRequest 构建更新告警发送模板状态请求
		WithSendTemplateStatusUpdateRequest(*templateapi.UpdateStatusRequest) IUpdateSendTemplateStatusRequestBuilder
		// WithSendTemplateUpdateRequest 构建更新告警发送模板请求
		WithSendTemplateUpdateRequest(*templateapi.UpdateSendTemplateRequest) IUpdateSendTemplateRequestBuilder
		// WithSendTemplateListRequest 构建查询告警发送模板列表请求
		WithSendTemplateListRequest(*templateapi.ListSendTemplateRequest) IListSendTemplateRequestBuilder
		// IDoSendTemplateBuilder 执行告警发送模板请求构建器
		IDoSendTemplateBuilder() IDoSendTemplateBuilder
	}

	sendTemplateModuleBuild struct {
		ctx context.Context
	}

	// ICreateSendTemplateRequestBuilder 创建告警发送模板请求构建器
	ICreateSendTemplateRequestBuilder interface {
		// ToBo 将请求转换为业务对象
		ToBo() *bo.CreateSendTemplate
	}

	createSendTemplateRequestBuilder struct {
		*templateapi.CreateSendTemplateRequest
		ctx context.Context
	}

	// IUpdateSendTemplateStatusRequestBuilder 更新告警发送模板状态请求构建器
	IUpdateSendTemplateStatusRequestBuilder interface {
		ToBo() *bo.UpdateSendTemplateStatusParams
	}

	updateSendTemplateStatusRequestBuilder struct {
		*templateapi.UpdateStatusRequest
		ctx context.Context
	}

	// IUpdateSendTemplateRequestBuilder 更新告警发送模板请求构建器
	IUpdateSendTemplateRequestBuilder interface {
		ToBo() *bo.UpdateSendTemplate
	}

	updateSendTemplateRequestBuilder struct {
		*templateapi.UpdateSendTemplateRequest
		ctx context.Context
	}
	// IListSendTemplateRequestBuilder 查询告警发送模板列表请求构建器
	IListSendTemplateRequestBuilder interface {
		ToBo() *bo.QuerySendTemplateListParams
	}

	listSendTemplateRequestBuilder struct {
		*templateapi.ListSendTemplateRequest
		ctx context.Context
	}

	// IDoSendTemplateBuilder 执行告警发送模板请求构建器
	IDoSendTemplateBuilder interface {
		ToAPI(imodel.ISendTemplate) *adminapi.SendTemplateItem
		ToAPIs([]imodel.ISendTemplate) []*adminapi.SendTemplateItem
	}
)

func (s *sendTemplateModuleBuild) ToAPI(template imodel.ISendTemplate) *adminapi.SendTemplateItem {
	if types.IsNil(s) || types.IsNil(template) {
		return nil
	}
	userMap := getUsers(s.ctx, template.GetCreatorID())
	return &adminapi.SendTemplateItem{
		Id:        template.GetID(),
		Name:      template.GetName(),
		Content:   template.GetContent(),
		SendType:  api.AlarmSendType(template.GetSendType()),
		Status:    api.Status(template.GetStatus()),
		CreatedAt: template.GetCreatedAt().Time.String(),
		UpdatedAt: template.GetUpdatedAt().Time.String(),
		Creator:   userMap[template.GetCreatorID()],
	}
}

func (s *sendTemplateModuleBuild) ToAPIs(templates []imodel.ISendTemplate) []*adminapi.SendTemplateItem {
	if types.IsNil(s) || types.IsNil(templates) {
		return nil
	}
	return types.SliceTo(templates, func(t imodel.ISendTemplate) *adminapi.SendTemplateItem {
		return s.ToAPI(t)
	})
}

func (l *listSendTemplateRequestBuilder) ToBo() *bo.QuerySendTemplateListParams {
	if types.IsNil(l) || types.IsNil(l.ListSendTemplateRequest) {
		return nil
	}
	return &bo.QuerySendTemplateListParams{
		Page:     types.NewPagination(l.GetPagination()),
		Keyword:  l.GetKeyword(),
		Status:   vobj.Status(l.GetStatus()),
		SendType: vobj.AlarmSendType(l.GetStatus()),
	}
}

func (u *updateSendTemplateRequestBuilder) ToBo() *bo.UpdateSendTemplate {
	if types.IsNil(u) || types.IsNil(u.UpdateSendTemplateRequest) {
		return nil
	}
	return &bo.UpdateSendTemplate{
		ID:          u.GetId(),
		UpdateParam: NewParamsBuild(u.ctx).SendTemplateModuleBuild().WithSendTemplateCreateRequest(u.UpdateSendTemplateRequest.GetData()).ToBo(),
	}
}

func (u *updateSendTemplateStatusRequestBuilder) ToBo() *bo.UpdateSendTemplateStatusParams {
	if types.IsNil(u) || types.IsNil(u.UpdateStatusRequest) {
		return nil
	}
	return &bo.UpdateSendTemplateStatusParams{
		Ids:    u.GetId(),
		Status: vobj.Status(u.GetStatus()),
	}
}

func (c *createSendTemplateRequestBuilder) ToBo() *bo.CreateSendTemplate {
	if types.IsNil(c) || types.IsNil(c.CreateSendTemplateRequest) {
		return nil
	}
	return &bo.CreateSendTemplate{
		Name:     c.GetName(),
		Content:  c.GetContent(),
		SendType: vobj.AlarmSendType(c.GetSendType()),
		Status:   vobj.Status(c.GetStatus()),
		Remark:   c.GetRemark(),
	}
}

func (s *sendTemplateModuleBuild) WithSendTemplateCreateRequest(request *templateapi.CreateSendTemplateRequest) ICreateSendTemplateRequestBuilder {
	return &createSendTemplateRequestBuilder{
		ctx:                       s.ctx,
		CreateSendTemplateRequest: request,
	}
}

func (s *sendTemplateModuleBuild) WithSendTemplateStatusUpdateRequest(request *templateapi.UpdateStatusRequest) IUpdateSendTemplateStatusRequestBuilder {
	return &updateSendTemplateStatusRequestBuilder{
		ctx:                 s.ctx,
		UpdateStatusRequest: request,
	}
}

func (s *sendTemplateModuleBuild) WithSendTemplateUpdateRequest(request *templateapi.UpdateSendTemplateRequest) IUpdateSendTemplateRequestBuilder {
	return &updateSendTemplateRequestBuilder{
		ctx:                       s.ctx,
		UpdateSendTemplateRequest: request,
	}
}

func (s *sendTemplateModuleBuild) WithSendTemplateListRequest(request *templateapi.ListSendTemplateRequest) IListSendTemplateRequestBuilder {

	return &listSendTemplateRequestBuilder{
		ctx:                     s.ctx,
		ListSendTemplateRequest: request,
	}
}

func (s *sendTemplateModuleBuild) IDoSendTemplateBuilder() IDoSendTemplateBuilder {
	return &sendTemplateModuleBuild{
		ctx: s.ctx,
	}
}
