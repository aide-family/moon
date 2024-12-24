package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	alarmapi "github.com/aide-family/moon/api/admin/alarm"
	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
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
		ToAPI(*bizmodel.AlarmNoticeGroup, ...map[uint32]*adminapi.UserItem) *adminapi.AlarmNoticeGroupItem
		ToAPIs([]*bizmodel.AlarmNoticeGroup) []*adminapi.AlarmNoticeGroupItem
		ToSelect(*bizmodel.AlarmNoticeGroup) *adminapi.SelectItem
		ToSelects([]*bizmodel.AlarmNoticeGroup) []*adminapi.SelectItem
	}

	doAlarmNoticeGroupItemBuilder struct {
		ctx context.Context
	}

	// IDoLabelNoticeBuilder 告警组标签通知对象返回值构造器
	IDoLabelNoticeBuilder interface {
		ToAPI(*bizmodel.StrategyMetricsLabelNotice) *adminapi.LabelNoticeItem
		ToAPIs([]*bizmodel.StrategyMetricsLabelNotice) []*adminapi.LabelNoticeItem
	}

	doLabelNoticeBuilder struct {
		ctx context.Context
	}

	// ICreateStrategyLabelNoticeRequestBuilder 创建策略标签通知请求参数构造器
	ICreateStrategyLabelNoticeRequestBuilder interface {
		ToBo(*strategyapi.CreateStrategyLabelNoticeRequest) *bo.StrategyLabelNotice
		ToBos([]*strategyapi.CreateStrategyLabelNoticeRequest) []*bo.StrategyLabelNotice
	}

	createStrategyLabelNoticeRequestBuilder struct {
		ctx context.Context
	}

	// IMyAlarmGroupListParamsBuilder 获取我的告警组列表请求参数构造器
	IMyAlarmGroupListParamsBuilder interface {
		ToBo() *bo.MyAlarmGroupListParams
	}

	myAlarmGroupListParamsBuilder struct {
		ctx context.Context
		*alarmapi.MyAlarmGroupRequest
	}

	// IAlarmNoticeGroupModuleBuilder 告警组模块构造器
	IAlarmNoticeGroupModuleBuilder interface {
		WithCreateAlarmGroupRequest(*alarmapi.CreateAlarmGroupRequest) ICreateAlarmGroupRequestBuilder
		WithUpdateAlarmGroupRequest(*alarmapi.UpdateAlarmGroupRequest) IUpdateAlarmGroupRequestBuilder
		WithListAlarmGroupRequest(*alarmapi.ListAlarmGroupRequest) IListAlarmGroupRequestBuilder
		WithUpdateAlarmGroupStatusRequest(*alarmapi.UpdateAlarmGroupStatusRequest) IUpdateAlarmGroupStatusRequestBuilder
		APICreateStrategyLabelNoticeRequest() ICreateStrategyLabelNoticeRequestBuilder
		DoAlarmNoticeGroupItemBuilder() IDoAlarmNoticeGroupItemBuilder
		DoLabelNoticeBuilder() IDoLabelNoticeBuilder
		AlarmNoticeGroupItemBuilder() IAlarmNoticeGroupItemBuilder
		WithAPIMyAlarmGroupListRequest(*alarmapi.MyAlarmGroupRequest) IMyAlarmGroupListParamsBuilder
	}

	// IAlarmNoticeGroupItemBuilder 告警组列表返回值构造器
	IAlarmNoticeGroupItemBuilder interface {
		ToAPI(*bizmodel.StrategyMetricsLabelNotice) *api.LabelNotices
		ToAPIs([]*bizmodel.StrategyMetricsLabelNotice) []*api.LabelNotices
	}

	alarmNoticeGroupItemBuilder struct {
		ctx context.Context
	}
)

func (a *alarmNoticeGroupItemBuilder) ToAPI(group *bizmodel.StrategyMetricsLabelNotice) *api.LabelNotices {
	if types.IsNil(group) || types.IsNil(a) {
		return nil
	}

	return &api.LabelNotices{
		Key:   group.Name,
		Value: group.Value,
		ReceiverGroupIDs: types.SliceUnique(types.SliceTo(group.AlarmGroups, func(group *bizmodel.AlarmNoticeGroup) uint32 {
			return group.ID
		})),
	}
}

func (a *alarmNoticeGroupItemBuilder) ToAPIs(groups []*bizmodel.StrategyMetricsLabelNotice) []*api.LabelNotices {
	return types.SliceTo(groups, func(group *bizmodel.StrategyMetricsLabelNotice) *api.LabelNotices {
		return a.ToAPI(group)
	})
}

func (a *alarmNoticeGroupModuleBuilder) AlarmNoticeGroupItemBuilder() IAlarmNoticeGroupItemBuilder {
	return &alarmNoticeGroupItemBuilder{ctx: a.ctx}
}

// ToBo 转换为业务对象
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

func (d *doLabelNoticeBuilder) ToAPI(notice *bizmodel.StrategyMetricsLabelNotice) *adminapi.LabelNoticeItem {
	if types.IsNil(notice) || types.IsNil(d) {
		return nil
	}

	return &adminapi.LabelNoticeItem{
		Name:        notice.Name,
		Value:       notice.Value,
		AlarmGroups: NewParamsBuild(d.ctx).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPIs(notice.AlarmGroups),
	}
}

func (d *doLabelNoticeBuilder) ToAPIs(notices []*bizmodel.StrategyMetricsLabelNotice) []*adminapi.LabelNoticeItem {
	if types.IsNil(notices) || types.IsNil(d) {
		return nil
	}

	return types.SliceTo(notices, func(notice *bizmodel.StrategyMetricsLabelNotice) *adminapi.LabelNoticeItem {
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

func getUsers(ctx context.Context, userIDs ...uint32) map[uint32]*adminapi.UserItem {
	userMap := make(map[uint32]*adminapi.UserItem)
	userDoBuilder := NewParamsBuild(ctx).UserModuleBuilder().DoUserBuilder()
	if biz.RuntimeCache != nil {
		userList := biz.RuntimeCache.GetUsers(ctx, userIDs)
		for _, user := range userList {
			userMap[user.ID] = userDoBuilder.ToAPI(user)
		}
	}

	return userMap
}

func (d *doAlarmNoticeGroupItemBuilder) ToAPI(group *bizmodel.AlarmNoticeGroup, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.AlarmNoticeGroupItem {
	if types.IsNil(group) || types.IsNil(d) {
		return nil
	}

	userMap := getUsers(d.ctx, group.CreatorID)
	return &adminapi.AlarmNoticeGroupItem{
		Id:          group.ID,
		Name:        group.Name,
		Status:      api.Status(group.Status),
		CreatedAt:   group.CreatedAt.String(),
		UpdatedAt:   group.UpdatedAt.String(),
		Remark:      group.Remark,
		Creator:     userMap[group.CreatorID],
		CreatorId:   group.CreatorID,
		NoticeUsers: NewParamsBuild(d.ctx).UserModuleBuilder().DoNoticeUserBuilder().ToAPIs(group.NoticeMembers),
		Hooks:       NewParamsBuild(d.ctx).HookModuleBuilder().DoHookBuilder().ToAPIs(group.AlarmHooks),
		TimeEngines: NewParamsBuild(d.ctx).TimeEngineModuleBuilder().Do().ToAPIs(group.TimeEngines),
	}
}

func (d *doAlarmNoticeGroupItemBuilder) ToAPIs(groups []*bizmodel.AlarmNoticeGroup) []*adminapi.AlarmNoticeGroupItem {
	if types.IsNil(groups) || types.IsNil(d) {
		return nil
	}
	ids := types.SliceTo(groups, func(group *bizmodel.AlarmNoticeGroup) uint32 {
		if types.IsNil(group) {
			return 0
		}
		return group.GetCreatorID()
	})
	userMap := getUsers(d.ctx, ids...)
	return types.SliceTo(groups, func(group *bizmodel.AlarmNoticeGroup) *adminapi.AlarmNoticeGroupItem {
		return d.ToAPI(group, userMap)
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
		UpdateParam: NewParamsBuild(u.ctx).
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
			return &bo.CreateNoticeMemberParams{MemberID: member.GetMemberId(), NotifyType: vobj.NotifyType(member.GetNotifyType())}
		}),
		HookIds:       c.GetHookIds(),
		TimeEngineIds: c.GetTimeEngines(),
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

func (a *alarmNoticeGroupModuleBuilder) WithAPIMyAlarmGroupListRequest(request *alarmapi.MyAlarmGroupRequest) IMyAlarmGroupListParamsBuilder {
	return &myAlarmGroupListParamsBuilder{ctx: a.ctx, MyAlarmGroupRequest: request}
}

func (a *alarmNoticeGroupModuleBuilder) DoAlarmNoticeGroupItemBuilder() IDoAlarmNoticeGroupItemBuilder {
	return &doAlarmNoticeGroupItemBuilder{ctx: a.ctx}
}
