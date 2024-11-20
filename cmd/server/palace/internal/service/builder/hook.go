package builder

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

var _ IHookModuleBuilder = (*hookModuleBuilder)(nil)

type (
	hookModuleBuilder struct {
		ctx context.Context
	}

	// IHookModuleBuilder 钩子模块构造器
	IHookModuleBuilder interface {
		// WithCreateHookRequest 创建钩子请求参数构造器
		WithCreateHookRequest(*hookapi.CreateHookRequest) ICreateHookRequestBuilder
		// WithUpdateHookRequest 更新钩子请求参数构造器
		WithUpdateHookRequest(*hookapi.UpdateHookRequest) IUpdateHookRequestBuilder
		// WithListHookRequest 获取钩子列表请求参数构造器
		WithListHookRequest(*hookapi.ListHookRequest) IListHookRequestBuilder
		// WithUpdateHookStatusRequest 更新钩子状态请求参数构造器
		WithUpdateHookStatusRequest(*hookapi.UpdateHookStatusRequest) IUpdateHookStatusRequestBuilder
		// DoHookBuilder 钩子条目构造器
		DoHookBuilder() IDoHookBuilder
	}

	// ICreateHookRequestBuilder 创建钩子请求参数构造器
	ICreateHookRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.CreateAlarmHookParams
	}

	createHookRequestBuilder struct {
		ctx context.Context
		*hookapi.CreateHookRequest
	}

	// IUpdateHookRequestBuilder 更新钩子请求参数构造器
	IUpdateHookRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateAlarmHookParams
	}

	updateHookRequestBuilder struct {
		ctx context.Context
		*hookapi.UpdateHookRequest
	}

	// IListHookRequestBuilder 获取钩子列表请求参数构造器
	IListHookRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryAlarmHookListParams
	}

	listHookRequestBuilder struct {
		ctx context.Context
		*hookapi.ListHookRequest
	}

	// IUpdateHookStatusRequestBuilder 更新钩子状态请求参数构造器
	IUpdateHookStatusRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateAlarmHookStatusParams
	}

	updateHookStatusRequestBuilder struct {
		ctx context.Context
		*hookapi.UpdateHookStatusRequest
	}

	// IDoHookBuilder 钩子条目构造器
	IDoHookBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*bizmodel.AlarmHook, ...map[uint32]*adminapi.UserItem) *adminapi.AlarmHookItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*bizmodel.AlarmHook) []*adminapi.AlarmHookItem
		// ToSelect 转换为选择对象
		ToSelect(*bizmodel.AlarmHook) *adminapi.SelectItem
		// ToSelects 转换为选择对象列表
		ToSelects([]*bizmodel.AlarmHook) []*adminapi.SelectItem
	}

	doHookBuilder struct {
		ctx context.Context
	}
)

func (d *doHookBuilder) ToAPI(hook *bizmodel.AlarmHook, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.AlarmHookItem {
	if types.IsNil(d) || types.IsNil(hook) {
		return nil
	}

	userMap := getUsers(d.ctx, userMaps, hook.CreatorID)

	return &adminapi.AlarmHookItem{
		Id:        hook.ID,
		Name:      hook.Name,
		Status:    api.Status(hook.Status),
		CreatedAt: hook.CreatedAt.String(),
		UpdatedAt: hook.UpdatedAt.String(),
		Remark:    hook.Remark,
		Creator:   userMap[hook.CreatorID],
		HookApp:   api.HookApp(hook.APP),
		Secret:    hook.Secret,
		Url:       hook.URL,
	}
}

func (d *doHookBuilder) ToAPIs(hooks []*bizmodel.AlarmHook) []*adminapi.AlarmHookItem {
	if types.IsNil(d) || types.IsNil(hooks) {
		return nil
	}
	ids := types.SliceTo(hooks, func(hook *bizmodel.AlarmHook) uint32 {
		return hook.CreatorID
	})
	userMap := getUsers(d.ctx, nil, ids...)

	return types.SliceTo(hooks, func(hook *bizmodel.AlarmHook) *adminapi.AlarmHookItem {
		return d.ToAPI(hook, userMap)
	})
}

func (d *doHookBuilder) ToSelect(hook *bizmodel.AlarmHook) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(hook) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    hook.ID,
		Label:    hook.Name,
		Children: nil,
		Disabled: hook.DeletedAt > 0 || !hook.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: hook.Remark,
		},
	}
}

func (d *doHookBuilder) ToSelects(hooks []*bizmodel.AlarmHook) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(hooks) {
		return nil
	}

	return types.SliceTo(hooks, func(hook *bizmodel.AlarmHook) *adminapi.SelectItem {
		return d.ToSelect(hook)
	})
}

func (u *updateHookStatusRequestBuilder) ToBo() *bo.UpdateAlarmHookStatusParams {
	if types.IsNil(u) || types.IsNil(u.UpdateHookStatusRequest) {
		return nil
	}

	return &bo.UpdateAlarmHookStatusParams{
		IDs:    u.GetIds(),
		Status: vobj.Status(u.GetStatus()),
	}
}

func (l *listHookRequestBuilder) ToBo() *bo.QueryAlarmHookListParams {
	if types.IsNil(l) || types.IsNil(l.ListHookRequest) {
		return nil
	}

	return &bo.QueryAlarmHookListParams{
		Keyword: l.GetKeyword(),
		Page:    types.NewPagination(l.GetPagination()),
		Name:    l.GetName(),
		Status:  vobj.Status(l.GetStatus()),
		Apps:    types.SliceTo(l.GetHookApp(), func(app api.HookApp) vobj.HookAPP { return vobj.HookAPP(app) }),
	}
}

func (u *updateHookRequestBuilder) ToBo() *bo.UpdateAlarmHookParams {
	if types.IsNil(u) || types.IsNil(u.UpdateHookRequest) {
		return nil
	}

	return &bo.UpdateAlarmHookParams{
		ID:          u.GetId(),
		UpdateParam: NewParamsBuild(u.ctx).HookModuleBuilder().WithCreateHookRequest(u.GetUpdate()).ToBo(),
	}
}

func (c *createHookRequestBuilder) ToBo() *bo.CreateAlarmHookParams {
	if types.IsNil(c) || types.IsNil(c.CreateHookRequest) {
		return nil
	}

	return &bo.CreateAlarmHookParams{
		Name:    c.GetName(),
		Remark:  c.GetRemark(),
		URL:     c.GetUrl(),
		Secret:  c.GetSecret(),
		HookApp: vobj.HookAPP(c.GetHookApp()),
		Status:  vobj.StatusEnable,
	}
}

func (h *hookModuleBuilder) WithCreateHookRequest(request *hookapi.CreateHookRequest) ICreateHookRequestBuilder {
	return &createHookRequestBuilder{ctx: h.ctx, CreateHookRequest: request}
}

func (h *hookModuleBuilder) WithUpdateHookRequest(request *hookapi.UpdateHookRequest) IUpdateHookRequestBuilder {
	return &updateHookRequestBuilder{ctx: h.ctx, UpdateHookRequest: request}
}

func (h *hookModuleBuilder) WithListHookRequest(request *hookapi.ListHookRequest) IListHookRequestBuilder {
	return &listHookRequestBuilder{ctx: h.ctx, ListHookRequest: request}
}

func (h *hookModuleBuilder) WithUpdateHookStatusRequest(request *hookapi.UpdateHookStatusRequest) IUpdateHookStatusRequestBuilder {
	return &updateHookStatusRequestBuilder{ctx: h.ctx, UpdateHookStatusRequest: request}
}

func (h *hookModuleBuilder) DoHookBuilder() IDoHookBuilder {
	return &doHookBuilder{ctx: h.ctx}
}
