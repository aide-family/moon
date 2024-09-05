package build

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	hookapi "github.com/aide-family/moon/api/admin/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// HookModuleBuilder hook模块构建器
	HookModuleBuilder interface {
		WithDoAlarmHook(*bizmodel.AlarmHook) DoAlarmHookBuilder
		WithsAlarmHook([]*bizmodel.AlarmHook) DosAlarmHookBuilder
		WithAPICreateAlarmHook(*hookapi.CreateHookRequest) APICreateHookParamsBuilder
		WithAPIQueryAlarmHookListRequest(*hookapi.ListHookRequest) APIQueryAlarmHookParamsBuilder
		WithAPIUpdateAlarmHookRequest(*hookapi.UpdateHookRequest) APIUpdateAlarmHookParamsBuilder
		WithAPIUpdateAlarmHookStatusRequest(*hookapi.UpdateHookStatusRequest) APIUpdateStatusParamsBuilder

		WithContext(context.Context) HookModuleBuilder
	}

	hookModuleBuilder struct {
		ctx context.Context
	}

	// DoAlarmHookBuilder do  alarm hook builder
	DoAlarmHookBuilder interface {
		ToAPI() *adminapi.AlarmHookItem
		ToAPISelect() *adminapi.SelectItem
	}

	// DosAlarmHookBuilder do alarm hook builder
	DosAlarmHookBuilder interface {
		ToAPIs() []*adminapi.AlarmHookItem
	}

	// APICreateHookParamsBuilder api alarm hook item builder
	APICreateHookParamsBuilder interface {
		ToBo() *bo.CreateAlarmHookParams
	}

	// APIQueryAlarmHookParamsBuilder api alarm hook item builder
	APIQueryAlarmHookParamsBuilder interface {
		ToBo() *bo.QueryAlarmHookListParams
	}

	// APIUpdateAlarmHookParamsBuilder api alarm hook item builder
	APIUpdateAlarmHookParamsBuilder interface {
		ToBo() *bo.UpdateAlarmHookParams
	}

	// APIUpdateStatusParamsBuilder api alarm hook item builder
	APIUpdateStatusParamsBuilder interface {
		ToBo() *bo.UpdateAlarmHookStatusParams
	}

	doAlarmHookBuilder struct {
		hookModel *bizmodel.AlarmHook

		ctx context.Context
	}

	dosAlarmHookBuilder struct {
		hookModels []*bizmodel.AlarmHook

		ctx context.Context
	}

	apiCreateHookParamsBuilder struct {
		params *hookapi.CreateHookRequest

		ctx context.Context
	}

	apiQueryHookParamsBuilder struct {
		params *hookapi.ListHookRequest

		ctx context.Context
	}

	apiUpdateHookParamsBuilder struct {
		params *hookapi.UpdateHookRequest

		ctx context.Context
	}

	apiUpdateStatusParamsBuilder struct {
		params *hookapi.UpdateHookStatusRequest

		ctx context.Context
	}
)

func (a *doAlarmHookBuilder) ToAPISelect() *adminapi.SelectItem {
	if types.IsNil(a) || types.IsNil(a.hookModel) {
		return nil
	}
	hookModel := a.hookModel
	return &adminapi.SelectItem{
		Value:    hookModel.ID,
		Label:    hookModel.Name,
		Children: nil,
		Disabled: hookModel.DeletedAt > 0 || !hookModel.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: hookModel.Remark,
		},
	}
}

func (a *apiUpdateStatusParamsBuilder) ToBo() *bo.UpdateAlarmHookStatusParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	return &bo.UpdateAlarmHookStatusParams{
		IDs:    a.params.GetIds(),
		Status: vobj.Status(a.params.GetStatus()),
	}
}

func (a *apiUpdateHookParamsBuilder) ToBo() *bo.UpdateAlarmHookParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	return &bo.UpdateAlarmHookParams{
		ID: a.params.GetId(),
		UpdateParam: NewBuilder().
			HookModuleBuilder().
			WithContext(a.ctx).
			WithAPICreateAlarmHook(a.params.GetUpdate()).
			ToBo(),
	}
}

func (a *apiQueryHookParamsBuilder) ToBo() *bo.QueryAlarmHookListParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	return &bo.QueryAlarmHookListParams{
		Keyword: a.params.GetKeyword(),
		Status:  vobj.Status(a.params.GetStatus()),
		Page:    types.NewPagination(a.params.GetPagination()),
		Apps: types.SliceTo(a.params.HookApp, func(app api.HookApp) vobj.HookAPP {
			return vobj.HookAPP(app)
		}),
	}
}

func (a *apiCreateHookParamsBuilder) ToBo() *bo.CreateAlarmHookParams {
	if types.IsNil(a) || types.IsNil(a.params) {
		return nil
	}
	return &bo.CreateAlarmHookParams{
		Name:    a.params.GetName(),
		Remark:  a.params.GetRemark(),
		URL:     a.params.GetUrl(),
		Secret:  a.params.GetSecret(),
		HookApp: vobj.HookAPP(a.params.GetHookApp()),
		Status:  vobj.Status(a.params.GetStatus()),
	}
}

func (a *dosAlarmHookBuilder) ToAPIs() []*adminapi.AlarmHookItem {
	if types.IsNil(a) || types.IsNil(a.hookModels) {
		return nil
	}
	return types.SliceTo(a.hookModels, func(hookModel *bizmodel.AlarmHook) *adminapi.AlarmHookItem {
		return NewBuilder().
			HookModuleBuilder().
			WithContext(a.ctx).
			WithDoAlarmHook(hookModel).
			ToAPI()
	})
}

func (a *doAlarmHookBuilder) ToAPI() *adminapi.AlarmHookItem {
	if types.IsNil(a) || types.IsNil(a.hookModel) {
		return nil
	}
	return &adminapi.AlarmHookItem{
		Id:        a.hookModel.ID,
		Name:      a.hookModel.Name,
		Remark:    a.hookModel.Remark,
		Url:       a.hookModel.URL,
		Secret:    a.hookModel.Secret,
		HookApp:   api.HookApp(a.hookModel.APP),
		Status:    api.Status(a.hookModel.Status),
		CreatedAt: a.hookModel.CreatedAt.String(),
		UpdatedAt: a.hookModel.UpdatedAt.String(),
	}
}

func (h *hookModuleBuilder) WithDoAlarmHook(hook *bizmodel.AlarmHook) DoAlarmHookBuilder {
	return newDoAlarmHookBuilder(h.ctx, hook)
}

func (h *hookModuleBuilder) WithsAlarmHook(hooks []*bizmodel.AlarmHook) DosAlarmHookBuilder {
	return newDosAlarmHookBuilder(h.ctx, hooks)
}

func (h *hookModuleBuilder) WithAPICreateAlarmHook(request *hookapi.CreateHookRequest) APICreateHookParamsBuilder {
	return newAPICreateHookParamsBuilder(h.ctx, request)
}

func (h *hookModuleBuilder) WithAPIQueryAlarmHookListRequest(request *hookapi.ListHookRequest) APIQueryAlarmHookParamsBuilder {
	return newAPIQueryHookParamsBuilder(h.ctx, request)
}

func (h *hookModuleBuilder) WithAPIUpdateAlarmHookRequest(request *hookapi.UpdateHookRequest) APIUpdateAlarmHookParamsBuilder {
	return newAPIUpdateHookParamsBuilder(h.ctx, request)
}

func (h *hookModuleBuilder) WithAPIUpdateAlarmHookStatusRequest(request *hookapi.UpdateHookStatusRequest) APIUpdateStatusParamsBuilder {
	return newAPIUpdateStatusParamsBuilder(h.ctx, request)
}

func (h *hookModuleBuilder) WithContext(ctx context.Context) HookModuleBuilder {
	if types.IsNil(h) {
		return newHookModuleBuilder(ctx)
	}
	h.ctx = ctx
	return h
}

func newHookModuleBuilder(ctx context.Context) HookModuleBuilder {
	return &hookModuleBuilder{ctx: ctx}
}

func newDoAlarmHookBuilder(ctx context.Context, hook *bizmodel.AlarmHook) DoAlarmHookBuilder {
	return &doAlarmHookBuilder{ctx: ctx, hookModel: hook}
}

func newDosAlarmHookBuilder(ctx context.Context, hooks []*bizmodel.AlarmHook) DosAlarmHookBuilder {
	return &dosAlarmHookBuilder{
		ctx:        ctx,
		hookModels: hooks,
	}
}

func newAPICreateHookParamsBuilder(ctx context.Context, request *hookapi.CreateHookRequest) APICreateHookParamsBuilder {
	return &apiCreateHookParamsBuilder{ctx: ctx, params: request}
}

func newAPIQueryHookParamsBuilder(ctx context.Context, request *hookapi.ListHookRequest) APIQueryAlarmHookParamsBuilder {
	return &apiQueryHookParamsBuilder{ctx: ctx, params: request}
}

func newAPIUpdateHookParamsBuilder(ctx context.Context, request *hookapi.UpdateHookRequest) APIUpdateAlarmHookParamsBuilder {
	return &apiUpdateHookParamsBuilder{ctx: ctx, params: request}
}

func newAPIUpdateStatusParamsBuilder(ctx context.Context, request *hookapi.UpdateHookStatusRequest) APIUpdateStatusParamsBuilder {
	return &apiUpdateStatusParamsBuilder{ctx: ctx, params: request}
}
