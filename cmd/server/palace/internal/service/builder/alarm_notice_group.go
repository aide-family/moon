package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	alarmapi "github.com/aide-family/moon/api/admin/alarm"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IAlarmNoticeGroupModuleBuilder = (*alarmNoticeGroupModuleBuilder)(nil)

type (
	alarmNoticeGroupModuleBuilder struct {
		ctx context.Context
	}

	// ICreateAlarmGroupRequestBuilder 创建告警组请求参数构造器
	ICreateAlarmGroupRequestBuilder interface {
		ToBo() *bo.CreateAlarmNoticeGroupParams
	}

	createAlarmGroupRequestBuilder struct {
		ctx context.Context
		*alarmapi.CreateAlarmGroupRequest
	}

	// IUpdateAlarmGroupRequestBuilder 更新告警组请求参数构造器
	IUpdateAlarmGroupRequestBuilder interface {
		ToBo() *bo.UpdateAlarmNoticeGroupParams
	}

	updateAlarmGroupRequestBuilder struct {
		ctx context.Context
		*alarmapi.UpdateAlarmGroupRequest
	}

	// IListAlarmGroupRequestBuilder 获取告警组列表请求参数构造器
	IListAlarmGroupRequestBuilder interface {
		ToBo() *bo.QueryAlarmNoticeGroupListParams
	}

	listAlarmGroupRequestBuilder struct {
		ctx context.Context
		*alarmapi.ListAlarmGroupRequest
	}

	// IUpdateAlarmGroupStatusRequestBuilder 更新告警组状态请求参数构造器
	IUpdateAlarmGroupStatusRequestBuilder interface {
		ToBo() *bo.UpdateAlarmNoticeGroupStatusParams
	}

	updateAlarmGroupStatusRequestBuilder struct {
		ctx context.Context
		*alarmapi.UpdateAlarmGroupStatusRequest
	}

	// IDoAlarmNoticeGroupItemBuilder 告警组列表返回值构造器
	IDoAlarmNoticeGroupItemBuilder interface {
		ToAPI(*bizmodel.AlarmNoticeGroup) *adminapi.AlarmNoticeGroupItem
		ToAPIs([]*bizmodel.AlarmNoticeGroup) []*adminapi.AlarmNoticeGroupItem
		ToSelect(*bizmodel.AlarmNoticeGroup) *adminapi.SelectItem
		ToSelects([]*bizmodel.AlarmNoticeGroup) []*adminapi.SelectItem
	}

	doAlarmNoticeGroupItemBuilder struct {
		ctx context.Context
	}

	// IDoLabelNoticeBuilder 告警组标签通知对象返回值构造器
	IDoLabelNoticeBuilder interface {
		ToAPI(*bizmodel.StrategyLabelNotice) *adminapi.LabelNoticeItem
		ToAPIs([]*bizmodel.StrategyLabelNotice) []*adminapi.LabelNoticeItem
	}

	doLabelNoticeBuilder struct {
		ctx context.Context
	}

	ICreateStrategyLabelNoticeRequestBuilder interface {
		ToBo(*strategyapi.CreateStrategyLabelNoticeRequest) *bo.StrategyLabelNotice
		ToBos([]*strategyapi.CreateStrategyLabelNoticeRequest) []*bo.StrategyLabelNotice
	}

	createStrategyLabelNoticeRequestBuilder struct {
		ctx context.Context
	}

	IMyAlarmGroupListParamsBuilder interface {
		ToBo() *bo.MyAlarmGroupListParams
	}

	myAlarmGroupListParamsBuilder struct {
		ctx context.Context
		*alarmapi.MyAlarmGroupRequest
	}

	IAlarmNoticeGroupModuleBuilder interface {
		WithCreateAlarmGroupRequest(*alarmapi.CreateAlarmGroupRequest) ICreateAlarmGroupRequestBuilder
		WithUpdateAlarmGroupRequest(*alarmapi.UpdateAlarmGroupRequest) IUpdateAlarmGroupRequestBuilder
		WithListAlarmGroupRequest(*alarmapi.ListAlarmGroupRequest) IListAlarmGroupRequestBuilder
		WithUpdateAlarmGroupStatusRequest(*alarmapi.UpdateAlarmGroupStatusRequest) IUpdateAlarmGroupStatusRequestBuilder
		APICreateStrategyLabelNoticeRequest() ICreateStrategyLabelNoticeRequestBuilder
		DoAlarmNoticeGroupItemBuilder() IDoAlarmNoticeGroupItemBuilder
		DoLabelNoticeBuilder() IDoLabelNoticeBuilder
		WithMyAlarmGroupListRequest(*alarmapi.MyAlarmGroupRequest) IMyAlarmGroupListParamsBuilder
	}
)

func (a *myAlarmGroupListParamsBuilder) ToBo() *bo.MyAlarmGroupListParams {
	if types.IsNil(a) || types.IsNil(a.MyAlarmGroupRequest) {
		return nil
	}
	return &bo.MyAlarmGroupListParams{
		Keyword: a.GetKeyword(),
		Status:  vobj.Status(a.GetStatus()),
		Page:    types.NewPagination(a.GetPagination()),
	}
}

func (c *createStrategyLabelNoticeRequestBuilder) ToBo(request *strategyapi.CreateStrategyLabelNoticeRequest) *bo.StrategyLabelNotice {
	if types.IsNil(request) || types.IsNil(c) {
		return nil
	}

	return &bo.StrategyLabelNotice{
		Name:          request.Name,
		Value:         request.Value,
		AlarmGroupIds: request.GetAlarmGroupIds(),
	}
}

func (c *createStrategyLabelNoticeRequestBuilder) ToBos(requests []*strategyapi.CreateStrategyLabelNoticeRequest) []*bo.StrategyLabelNotice {
	if types.IsNil(requests) || types.IsNil(c) {
		return nil
	}

	return types.SliceTo(requests, func(request *strategyapi.CreateStrategyLabelNoticeRequest) *bo.StrategyLabelNotice {
		return c.ToBo(request)
	})
}

func (a *alarmNoticeGroupModuleBuilder) APICreateStrategyLabelNoticeRequest() ICreateStrategyLabelNoticeRequestBuilder {
	return &createStrategyLabelNoticeRequestBuilder{ctx: a.ctx}
}

func (d *doLabelNoticeBuilder) ToAPI(notice *bizmodel.StrategyLabelNotice) *adminapi.LabelNoticeItem {
	if types.IsNil(notice) || types.IsNil(d) {
		return nil
	}

	return &adminapi.LabelNoticeItem{
		Name:        notice.Name,
		Value:       notice.Value,
		AlarmGroups: NewParamsBuild().WithContext(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(notice.AlarmGroups),
	}
}

func (d *doLabelNoticeBuilder) ToAPIs(notices []*bizmodel.StrategyLabelNotice) []*adminapi.LabelNoticeItem {
	if types.IsNil(notices) || types.IsNil(d) {
		return nil
	}

	return types.SliceTo(notices, func(notice *bizmodel.StrategyLabelNotice) *adminapi.LabelNoticeItem {
		return d.ToAPI(notice)
	})
}

func (a *alarmNoticeGroupModuleBuilder) DoLabelNoticeBuilder() IDoLabelNoticeBuilder {
	return &doLabelNoticeBuilder{ctx: a.ctx}
}

func (d *doAlarmNoticeGroupItemBuilder) ToSelect(group *bizmodel.AlarmNoticeGroup) *adminapi.SelectItem {
	if types.IsNil(group) || types.IsNil(d) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    group.ID,
		Label:    group.Name,
		Disabled: group.DeletedAt > 0 || !group.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: group.Remark,
		},
	}
}

func (d *doAlarmNoticeGroupItemBuilder) ToSelects(groups []*bizmodel.AlarmNoticeGroup) []*adminapi.SelectItem {
	if types.IsNil(groups) || types.IsNil(d) {
		return nil
	}

	return types.SliceTo(groups, func(group *bizmodel.AlarmNoticeGroup) *adminapi.SelectItem {
		return d.ToSelect(group)
	})
}

func (d *doAlarmNoticeGroupItemBuilder) ToAPI(group *bizmodel.AlarmNoticeGroup) *adminapi.AlarmNoticeGroupItem {
	if types.IsNil(group) || types.IsNil(d) {
		return nil
	}

	return &adminapi.AlarmNoticeGroupItem{
		Id:          group.ID,
		Name:        group.Name,
		Status:      api.Status(group.Status),
		CreatedAt:   group.CreatedAt.String(),
		UpdatedAt:   group.UpdatedAt.String(),
		Remark:      group.Remark,
		Creator:     "", // TODO impl
		CreatorId:   group.CreatorID,
		NoticeUsers: NewParamsBuild().WithContext(d.ctx).UserModuleBuilder().DoNoticeUserBuilder().ToAPIs(group.NoticeMembers),
		Hooks:       NewParamsBuild().WithContext(d.ctx).HookModuleBuilder().DoHookBuilder().ToAPIs(group.AlarmHooks),
	}
}

func (d *doAlarmNoticeGroupItemBuilder) ToAPIs(groups []*bizmodel.AlarmNoticeGroup) []*adminapi.AlarmNoticeGroupItem {
	if types.IsNil(groups) || types.IsNil(d) {
		return nil
	}

	return types.SliceTo(groups, func(group *bizmodel.AlarmNoticeGroup) *adminapi.AlarmNoticeGroupItem {
		return d.ToAPI(group)
	})
}

func (u *updateAlarmGroupStatusRequestBuilder) ToBo() *bo.UpdateAlarmNoticeGroupStatusParams {
	if types.IsNil(u) || types.IsNil(u.UpdateAlarmGroupStatusRequest) {
		return nil
	}
	return &bo.UpdateAlarmNoticeGroupStatusParams{
		IDs:    u.GetIds(),
		Status: vobj.Status(u.GetStatus()),
	}
}

func (l *listAlarmGroupRequestBuilder) ToBo() *bo.QueryAlarmNoticeGroupListParams {
	if types.IsNil(l) || types.IsNil(l.ListAlarmGroupRequest) {
		return nil
	}
	return &bo.QueryAlarmNoticeGroupListParams{
		Keyword: l.GetKeyword(),
		Status:  vobj.Status(l.GetStatus()),
		Page:    types.NewPagination(l.GetPagination()),
	}
}

func (u *updateAlarmGroupRequestBuilder) ToBo() *bo.UpdateAlarmNoticeGroupParams {
	if types.IsNil(u) || types.IsNil(u.UpdateAlarmGroupRequest) {
		return nil
	}
	return &bo.UpdateAlarmNoticeGroupParams{
		ID: u.GetId(),
		UpdateParam: NewParamsBuild().WithContext(u.ctx).
			AlarmNoticeGroupModuleBuilder().
			WithCreateAlarmGroupRequest(u.GetUpdate()).
			ToBo(),
	}
}

func (c *createAlarmGroupRequestBuilder) ToBo() *bo.CreateAlarmNoticeGroupParams {
	if types.IsNil(c) || types.IsNil(c.CreateAlarmGroupRequest) {
		return nil
	}

	return &bo.CreateAlarmNoticeGroupParams{
		Name:   c.GetName(),
		Remark: c.GetRemark(),
		Status: vobj.Status(c.GetStatus()),
		NoticeMembers: types.SliceTo(c.NoticeMember, func(member *alarmapi.CreateNoticeMemberRequest) *bo.CreateNoticeMemberParams {
			return &bo.CreateNoticeMemberParams{
				MemberID:   member.GetMemberId(),
				NotifyType: vobj.NotifyType(member.GetNotifyType()),
			}
		}),
		HookIds: c.GetHookIds(),
	}
}

func (a *alarmNoticeGroupModuleBuilder) WithCreateAlarmGroupRequest(request *alarmapi.CreateAlarmGroupRequest) ICreateAlarmGroupRequestBuilder {
	return &createAlarmGroupRequestBuilder{ctx: a.ctx, CreateAlarmGroupRequest: request}
}

func (a *alarmNoticeGroupModuleBuilder) WithUpdateAlarmGroupRequest(request *alarmapi.UpdateAlarmGroupRequest) IUpdateAlarmGroupRequestBuilder {
	return &updateAlarmGroupRequestBuilder{ctx: a.ctx, UpdateAlarmGroupRequest: request}
}

func (a *alarmNoticeGroupModuleBuilder) WithListAlarmGroupRequest(request *alarmapi.ListAlarmGroupRequest) IListAlarmGroupRequestBuilder {
	return &listAlarmGroupRequestBuilder{ctx: a.ctx, ListAlarmGroupRequest: request}
}

func (a *alarmNoticeGroupModuleBuilder) WithUpdateAlarmGroupStatusRequest(request *alarmapi.UpdateAlarmGroupStatusRequest) IUpdateAlarmGroupStatusRequestBuilder {
	return &updateAlarmGroupStatusRequestBuilder{ctx: a.ctx, UpdateAlarmGroupStatusRequest: request}
}

func (a *alarmNoticeGroupModuleBuilder) WithMyAlarmGroupListRequest(request *alarmapi.MyAlarmGroupRequest) IMyAlarmGroupListParamsBuilder {
	return &myAlarmGroupListParamsBuilder{ctx: a.ctx, MyAlarmGroupRequest: request}
}

func (a *alarmNoticeGroupModuleBuilder) DoAlarmNoticeGroupItemBuilder() IDoAlarmNoticeGroupItemBuilder {
	return &doAlarmNoticeGroupItemBuilder{ctx: a.ctx}
}
