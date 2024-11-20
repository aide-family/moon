package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	alarmapi "github.com/aide-family/moon/api/admin/alarm"
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IAlarmSendModuleBuilder = (*alarmSendModuleBuilder)(nil)

type (
	alarmSendModuleBuilder struct {
		ctx context.Context
	}

	// ICreateAlarmSendRequestBuilder 创建告警发送记录
	ICreateAlarmSendRequestBuilder interface {
		ToBo() *bo.CreateAlarmSendParams
	}

	createAlarmSendRequestBuilder struct {
		ctx context.Context
		*hookapi.SendMsgRequest
	}

	// IUpdateAlarmSendRequestBuilder 更新告警发送记录请求参数构造器
	IUpdateAlarmSendRequestBuilder interface {
		ToBo() *bo.UpdateAlarmSendParams
	}

	updateAlarmSendRequestBuilder struct {
		ctx context.Context
	}

	// IListAlarmSendRequestBuilder 获取告警发送历史列表请求参数构造器
	IListAlarmSendRequestBuilder interface {
		ToBo() *bo.QueryAlarmSendHistoryListParams
	}

	listAlarmSendRequestBuilder struct {
		ctx context.Context
		*alarmapi.ListAlarmSendRequest
	}

	// IDoAlarmSendItemBuilder 告警发送记录条目构造器
	IDoAlarmSendItemBuilder interface {
		ToAPI(history *alarmmodel.AlarmSendHistory, group *bizmodel.AlarmNoticeGroup) *adminapi.AlarmSendItem
		ToAPIs(histories []*alarmmodel.AlarmSendHistory) []*adminapi.AlarmSendItem
	}

	doAlarmSendItemBuilder struct {
		ctx context.Context
	}

	// IAlarmSendModuleBuilder 告警发送记录模块
	IAlarmSendModuleBuilder interface {
		WithListAlarmSendRequest(ctx context.Context, req *alarmapi.ListAlarmSendRequest) IListAlarmSendRequestBuilder
		WithDoAlarmSendItem(ctx context.Context) IDoAlarmSendItemBuilder

		WithCreateAlarmSendRequest(ctx context.Context, req *hookapi.SendMsgRequest) ICreateAlarmSendRequestBuilder
	}
)

func (d *createAlarmSendRequestBuilder) ToBo() *bo.CreateAlarmSendParams {
	if types.IsNil(d) || types.IsNil(d.SendMsgRequest) {
		return nil
	}
	return &bo.CreateAlarmSendParams{
		RequestID: d.RequestID,
		SendData:  d.Json,
		Route:     d.Route,
	}
}

func (d *doAlarmSendItemBuilder) ToAPI(history *alarmmodel.AlarmSendHistory, group *bizmodel.AlarmNoticeGroup) *adminapi.AlarmSendItem {
	if types.IsNil(history) || types.IsNil(d) {
		return nil
	}
	return &adminapi.AlarmSendItem{
		Id:           history.ID,
		AlarmGroupId: history.AlarmGroupID,
		SendData:     history.SendData,
		RetryNumber:  int32(history.RetryNumber),
		RequestId:    history.RequestID,
		SendStatus:   api.SendStatus(history.SendStatus),
		SendTime:     history.SendTime,
		AlarmGroup:   NewParamsBuild(context.TODO()).AlarmNoticeGroupModuleBuilder().DoAlarmNoticeGroupItemBuilder().ToAPI(group),
	}
}

func (d *doAlarmSendItemBuilder) ToAPIs(histories []*alarmmodel.AlarmSendHistory) []*adminapi.AlarmSendItem {
	if types.IsNil(histories) || types.IsNil(d) {
		return nil
	}
	return types.SliceTo(histories, func(history *alarmmodel.AlarmSendHistory) *adminapi.AlarmSendItem {
		return d.ToAPI(history, nil)
	})
}

func (a *listAlarmSendRequestBuilder) ToBo() *bo.QueryAlarmSendHistoryListParams {
	if types.IsNil(a) || types.IsNil(a.ListAlarmSendRequest) {
		return nil
	}
	return &bo.QueryAlarmSendHistoryListParams{
		Keyword: a.GetKeyword(),
		Page:    types.NewPagination(a.GetPagination()),
		SendStatus: types.To(a.GetSendStatus(), func(status api.SendStatus) vobj.SendStatus {
			return vobj.SendStatus(status)
		}),
		StartSendTime: a.GetSendStartTime(),
		EndSendTime:   a.GetSendEndTime(),
	}
}

func (a *alarmSendModuleBuilder) WithListAlarmSendRequest(ctx context.Context, req *alarmapi.ListAlarmSendRequest) IListAlarmSendRequestBuilder {
	return &listAlarmSendRequestBuilder{
		ctx:                  ctx,
		ListAlarmSendRequest: req,
	}
}

func (a *alarmSendModuleBuilder) WithDoAlarmSendItem(ctx context.Context) IDoAlarmSendItemBuilder {
	return &doAlarmSendItemBuilder{
		ctx: ctx,
	}
}

func (a *alarmSendModuleBuilder) WithCreateAlarmSendRequest(ctx context.Context, req *hookapi.SendMsgRequest) ICreateAlarmSendRequestBuilder {
	return &createAlarmSendRequestBuilder{
		ctx:            ctx,
		SendMsgRequest: req,
	}
}
