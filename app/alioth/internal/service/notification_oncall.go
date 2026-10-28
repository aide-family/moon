package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/notification"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

type NotificationOnCallService struct {
	v1.UnimplementedNotificationOnCallServiceServer
	center *notification.Center
}

func NewNotificationOnCallService(center *notification.Center) *NotificationOnCallService {
	return &NotificationOnCallService{center: center}
}

func (s *NotificationOnCallService) CreateNotificationOnCall(ctx context.Context, req *v1.CreateNotificationOnCallRequest) (*v1.CreateNotificationOnCallResponse, error) {
	in := onCallInput("", req.GetName(), req.GetLabels(), req.GetTimezone(), req.GetRemark(), req.GetLevelUids(), req.GetDatasourceUids(), req.GetDatasourceLevelUids(), req.GetRepeatIntervalSeconds(), req.GetGroups())
	item, err := s.center.CreateOnCall(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.CreateNotificationOnCallResponse{OnCall: onCallItem(*item)}, nil
}

func (s *NotificationOnCallService) UpdateNotificationOnCall(ctx context.Context, req *v1.UpdateNotificationOnCallRequest) (*v1.UpdateNotificationOnCallResponse, error) {
	in := onCallInput(req.GetUid(), req.GetName(), req.GetLabels(), req.GetTimezone(), req.GetRemark(), req.GetLevelUids(), req.GetDatasourceUids(), req.GetDatasourceLevelUids(), req.GetRepeatIntervalSeconds(), req.GetGroups())
	item, err := s.center.UpdateOnCall(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationOnCallResponse{OnCall: onCallItem(*item)}, nil
}

func (s *NotificationOnCallService) UpdateNotificationOnCallStatus(ctx context.Context, req *v1.UpdateNotificationOnCallStatusRequest) (*v1.UpdateNotificationOnCallStatusResponse, error) {
	item, err := s.center.UpdateOnCallStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationOnCallStatusResponse{OnCall: onCallItem(*item)}, nil
}

func (s *NotificationOnCallService) DeleteNotificationOnCall(ctx context.Context, req *v1.DeleteNotificationOnCallRequest) (*v1.DeleteNotificationOnCallResponse, error) {
	if err := s.center.DeleteOnCall(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNotificationOnCallResponse{}, nil
}

func (s *NotificationOnCallService) GetNotificationOnCall(ctx context.Context, req *v1.GetNotificationOnCallRequest) (*v1.GetNotificationOnCallResponse, error) {
	item, err := s.center.GetOnCall(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNotificationOnCallResponse{OnCall: onCallItem(*item)}, nil
}

func (s *NotificationOnCallService) ListNotificationOnCalls(ctx context.Context, req *v1.ListNotificationOnCallsRequest) (*v1.ListNotificationOnCallsResponse, error) {
	items, total, page, pageSize, err := s.center.ListOnCalls(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.NotificationOnCallItem, 0, len(items))
	for i := range items {
		out = append(out, onCallItem(items[i]))
	}
	return &v1.ListNotificationOnCallsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func onCallInput(uid, name, labels, timezone, remark string, levelUIDs, datasourceUIDs, datasourceLevelUIDs []string, repeatIntervalSeconds int32, groups []*v1.NotificationOnCallGroupInput) bo.NotificationOnCall {
	return bo.NotificationOnCall{
		UID: uid, Name: name, Labels: labels, Timezone: timezone, Remark: remark,
		LevelUIDs: levelUIDs, DatasourceUIDs: datasourceUIDs, DatasourceLevelUIDs: datasourceLevelUIDs,
		RepeatIntervalSeconds: int(repeatIntervalSeconds), Groups: onCallGroupInputs(groups),
	}
}

func onCallGroupInputs(groups []*v1.NotificationOnCallGroupInput) []bo.NotificationOnCallGroup {
	out := make([]bo.NotificationOnCallGroup, 0, len(groups))
	for _, group := range groups {
		out = append(out, bo.NotificationOnCallGroup{
			GroupUID: group.GetGroupUid(),
			Windows:  onCallWindowItems(group.GetWindows()),
		})
	}
	return out
}

func onCallWindowItems(windows []*v1.NotificationOnCallWindow) []bo.NotificationOnCallWindow {
	out := make([]bo.NotificationOnCallWindow, 0, len(windows))
	for _, window := range windows {
		out = append(out, bo.NotificationOnCallWindow{
			Weekdays: window.GetWeekdays(),
			Start:    window.GetStart(),
			End:      window.GetEnd(),
		})
	}
	return out
}

func onCallItem(item bo.NotificationOnCall) *v1.NotificationOnCallItem {
	groups := make([]*v1.NotificationOnCallGroupItem, 0, len(item.Groups))
	for _, group := range item.Groups {
		groups = append(groups, onCallGroupItem(group))
	}
	return &v1.NotificationOnCallItem{
		Uid: item.UID, Name: item.Name, Status: item.Status, Labels: item.Labels,
		LevelUids: item.LevelUIDs, DatasourceUids: item.DatasourceUIDs, Timezone: item.Timezone,
		DatasourceLevelUids: item.DatasourceLevelUIDs, RepeatIntervalSeconds: int32(item.RepeatIntervalSeconds),
		Groups: groups, Remark: item.Remark, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func onCallGroupItem(group bo.NotificationOnCallGroup) *v1.NotificationOnCallGroupItem {
	windows := make([]*v1.NotificationOnCallWindow, 0, len(group.Windows))
	for _, window := range group.Windows {
		windows = append(windows, &v1.NotificationOnCallWindow{
			Weekdays: window.Weekdays, Start: window.Start, End: window.End,
		})
	}
	return &v1.NotificationOnCallGroupItem{
		GroupUid: group.GroupUID, GroupName: group.GroupName, Windows: windows,
	}
}
