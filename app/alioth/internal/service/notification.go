package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	merakv1 "github.com/aide-family/Alioth/api/merak/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/notification"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type NotificationChannelService struct {
	v1.UnimplementedNotificationChannelServiceServer
	notify *notification.Usecase
}

func NewNotificationChannelService(uc *notification.Usecase) *NotificationChannelService {
	return &NotificationChannelService{notify: uc}
}

func (s *NotificationChannelService) CreateNotificationChannel(ctx context.Context, req *v1.CreateNotificationChannelRequest) (*v1.CreateNotificationChannelResponse, error) {
	item, err := s.notify.CreateChannel(ctx, channelInput("", req.GetName(), req.GetType(), req.GetApp(), req.GetEndpoint(), req.GetUsername(), req.GetPassword(), req.GetSecret(), req.GetFromAddress(), req.GetRemark()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateNotificationChannelResponse{Channel: channelItem(*item)}, nil
}

func (s *NotificationChannelService) UpdateNotificationChannel(ctx context.Context, req *v1.UpdateNotificationChannelRequest) (*v1.UpdateNotificationChannelResponse, error) {
	item, err := s.notify.UpdateChannel(ctx, channelInput(req.GetUid(), req.GetName(), req.GetType(), req.GetApp(), req.GetEndpoint(), req.GetUsername(), req.GetPassword(), req.GetSecret(), req.GetFromAddress(), req.GetRemark()))
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationChannelResponse{Channel: channelItem(*item)}, nil
}

func (s *NotificationChannelService) UpdateNotificationChannelStatus(ctx context.Context, req *v1.UpdateNotificationChannelStatusRequest) (*v1.UpdateNotificationChannelStatusResponse, error) {
	item, err := s.notify.UpdateChannelStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationChannelStatusResponse{Channel: channelItem(*item)}, nil
}

func (s *NotificationChannelService) DeleteNotificationChannel(ctx context.Context, req *v1.DeleteNotificationChannelRequest) (*v1.DeleteNotificationChannelResponse, error) {
	if err := s.notify.DeleteChannel(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNotificationChannelResponse{}, nil
}

func (s *NotificationChannelService) GetNotificationChannel(ctx context.Context, req *v1.GetNotificationChannelRequest) (*v1.GetNotificationChannelResponse, error) {
	item, err := s.notify.GetChannel(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNotificationChannelResponse{Channel: channelItem(*item)}, nil
}

func (s *NotificationChannelService) ListNotificationChannels(ctx context.Context, req *v1.ListNotificationChannelsRequest) (*v1.ListNotificationChannelsResponse, error) {
	items, total, page, pageSize, err := s.notify.ListChannels(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListNotificationChannelsResponse{Items: channelItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NotificationChannelService) SelectNotificationChannels(ctx context.Context, req *v1.SelectNotificationChannelsRequest) (*v1.SelectNotificationChannelsResponse, error) {
	items, err := s.notify.SelectChannels(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNotificationChannelsResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *NotificationChannelService) TestNotificationChannel(ctx context.Context, req *v1.TestNotificationChannelRequest) (*v1.TestNotificationChannelResponse, error) {
	uid, err := s.notify.TestChannel(ctx, req.GetUid(), req.GetTemplateUid(), req.GetReceivers(), req.GetSubject(), req.GetBody(), req.GetParams())
	if err != nil {
		return nil, err
	}
	return &v1.TestNotificationChannelResponse{TaskUid: uid}, nil
}

type NotificationTemplateService struct {
	v1.UnimplementedNotificationTemplateServiceServer
	notify *notification.Usecase
}

func NewNotificationTemplateService(uc *notification.Usecase) *NotificationTemplateService {
	return &NotificationTemplateService{notify: uc}
}

func (s *NotificationTemplateService) CreateNotificationTemplate(ctx context.Context, req *v1.CreateNotificationTemplateRequest) (*v1.CreateNotificationTemplateResponse, error) {
	item, err := s.notify.CreateTemplate(ctx, bo.NotificationTemplate{
		Name: req.GetName(), Type: req.GetType(), Subject: req.GetSubject(), Body: req.GetBody(), Remark: req.GetRemark(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateNotificationTemplateResponse{Template: templateItem(*item)}, nil
}

func (s *NotificationTemplateService) UpdateNotificationTemplate(ctx context.Context, req *v1.UpdateNotificationTemplateRequest) (*v1.UpdateNotificationTemplateResponse, error) {
	item, err := s.notify.UpdateTemplate(ctx, bo.NotificationTemplate{
		UID: req.GetUid(), Name: req.GetName(), Type: req.GetType(), Subject: req.GetSubject(), Body: req.GetBody(), Remark: req.GetRemark(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationTemplateResponse{Template: templateItem(*item)}, nil
}

func (s *NotificationTemplateService) UpdateNotificationTemplateStatus(ctx context.Context, req *v1.UpdateNotificationTemplateStatusRequest) (*v1.UpdateNotificationTemplateStatusResponse, error) {
	item, err := s.notify.UpdateTemplateStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationTemplateStatusResponse{Template: templateItem(*item)}, nil
}

func (s *NotificationTemplateService) DeleteNotificationTemplate(ctx context.Context, req *v1.DeleteNotificationTemplateRequest) (*v1.DeleteNotificationTemplateResponse, error) {
	if err := s.notify.DeleteTemplate(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNotificationTemplateResponse{}, nil
}

func (s *NotificationTemplateService) GetNotificationTemplate(ctx context.Context, req *v1.GetNotificationTemplateRequest) (*v1.GetNotificationTemplateResponse, error) {
	item, err := s.notify.GetTemplate(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNotificationTemplateResponse{Template: templateItem(*item)}, nil
}

func (s *NotificationTemplateService) ListNotificationTemplates(ctx context.Context, req *v1.ListNotificationTemplatesRequest) (*v1.ListNotificationTemplatesResponse, error) {
	items, total, page, pageSize, err := s.notify.ListTemplates(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListNotificationTemplatesResponse{Items: templateItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NotificationTemplateService) SelectNotificationTemplates(ctx context.Context, req *v1.SelectNotificationTemplatesRequest) (*v1.SelectNotificationTemplatesResponse, error) {
	items, err := s.notify.SelectTemplates(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNotificationTemplatesResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *NotificationTemplateService) TestNotificationTemplate(ctx context.Context, req *v1.TestNotificationTemplateRequest) (*v1.TestNotificationTemplateResponse, error) {
	uid, err := s.notify.TestTemplate(ctx, req.GetUid(), req.GetChannelUid(), req.GetReceivers(), req.GetParams())
	if err != nil {
		return nil, err
	}
	return &v1.TestNotificationTemplateResponse{TaskUid: uid}, nil
}

type NotificationGroupService struct {
	v1.UnimplementedNotificationGroupServiceServer
	notify *notification.Usecase
}

func NewNotificationGroupService(uc *notification.Usecase) *NotificationGroupService {
	return &NotificationGroupService{notify: uc}
}

func (s *NotificationGroupService) CreateNotificationGroup(ctx context.Context, req *v1.CreateNotificationGroupRequest) (*v1.CreateNotificationGroupResponse, error) {
	item, err := s.notify.CreateGroup(ctx, bo.NotificationGroup{
		Name: req.GetName(), ChannelUID: req.GetChannelUid(), TemplateUID: req.GetTemplateUid(),
		Receivers: req.GetReceivers(), Remark: req.GetRemark(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateNotificationGroupResponse{Group: notifyGroupItem(*item)}, nil
}

func (s *NotificationGroupService) UpdateNotificationGroup(ctx context.Context, req *v1.UpdateNotificationGroupRequest) (*v1.UpdateNotificationGroupResponse, error) {
	item, err := s.notify.UpdateGroup(ctx, bo.NotificationGroup{
		UID: req.GetUid(), Name: req.GetName(), ChannelUID: req.GetChannelUid(), TemplateUID: req.GetTemplateUid(),
		Receivers: req.GetReceivers(), Remark: req.GetRemark(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationGroupResponse{Group: notifyGroupItem(*item)}, nil
}

func (s *NotificationGroupService) UpdateNotificationGroupStatus(ctx context.Context, req *v1.UpdateNotificationGroupStatusRequest) (*v1.UpdateNotificationGroupStatusResponse, error) {
	item, err := s.notify.UpdateGroupStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationGroupStatusResponse{Group: notifyGroupItem(*item)}, nil
}

func (s *NotificationGroupService) DeleteNotificationGroup(ctx context.Context, req *v1.DeleteNotificationGroupRequest) (*v1.DeleteNotificationGroupResponse, error) {
	if err := s.notify.DeleteGroup(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNotificationGroupResponse{}, nil
}

func (s *NotificationGroupService) GetNotificationGroup(ctx context.Context, req *v1.GetNotificationGroupRequest) (*v1.GetNotificationGroupResponse, error) {
	item, err := s.notify.GetGroup(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNotificationGroupResponse{Group: notifyGroupItem(*item)}, nil
}

func (s *NotificationGroupService) ListNotificationGroups(ctx context.Context, req *v1.ListNotificationGroupsRequest) (*v1.ListNotificationGroupsResponse, error) {
	items, total, page, pageSize, err := s.notify.ListGroups(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListNotificationGroupsResponse{Items: notifyGroupItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NotificationGroupService) SelectNotificationGroups(ctx context.Context, req *v1.SelectNotificationGroupsRequest) (*v1.SelectNotificationGroupsResponse, error) {
	items, err := s.notify.SelectGroups(ctx, req.GetLimit(), req.GetLastUid(), req.GetActiveOnly())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNotificationGroupsResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

type MessageLogService struct {
	v1.UnimplementedMessageLogServiceServer
	notify *notification.Usecase
}

func NewMessageLogService(uc *notification.Usecase) *MessageLogService {
	return &MessageLogService{notify: uc}
}

func (s *MessageLogService) ListMessageLogs(ctx context.Context, req *v1.ListMessageLogsRequest) (*v1.ListMessageLogsResponse, error) {
	items, total, page, pageSize, err := s.notify.ListMessageLogs(ctx, req.GetPage(), req.GetPageSize(), req.GetTaskUid())
	if err != nil {
		return nil, err
	}
	return &v1.ListMessageLogsResponse{Items: messageLogItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *MessageLogService) GetMessageLog(ctx context.Context, req *v1.GetMessageLogRequest) (*v1.GetMessageLogResponse, error) {
	item, err := s.notify.GetMessageLog(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetMessageLogResponse{Log: messageLogItem(*item)}, nil
}

func (s *MessageLogService) ListNotificationTasks(ctx context.Context, req *v1.ListNotificationTasksRequest) (*v1.ListNotificationTasksResponse, error) {
	items, total, page, pageSize, err := s.notify.ListTasks(ctx, req.GetPage(), req.GetPageSize(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.ListNotificationTasksResponse{Items: taskItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *MessageLogService) RetryNotificationTask(ctx context.Context, req *v1.RetryNotificationTaskRequest) (*v1.RetryNotificationTaskResponse, error) {
	item, err := s.notify.RetryTask(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.RetryNotificationTaskResponse{Task: taskItem(*item)}, nil
}

func (s *MessageLogService) CancelNotificationTask(ctx context.Context, req *v1.CancelNotificationTaskRequest) (*v1.CancelNotificationTaskResponse, error) {
	item, err := s.notify.CancelTask(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.CancelNotificationTaskResponse{Task: taskItem(*item)}, nil
}

type NotifyService struct {
	merakv1.UnimplementedNotifyServiceServer
	notify *notification.Usecase
}

func NewNotifyService(uc *notification.Usecase) *NotifyService {
	return &NotifyService{notify: uc}
}

func (s *NotifyService) ClaimTasks(ctx context.Context, _ *merakv1.ClaimTasksRequest) (*merakv1.ClaimTasksResponse, error) {
	items, err := s.notify.Claim(ctx)
	if err != nil {
		return nil, err
	}
	tasks := make([]*merakv1.TaskSnapshot, 0, len(items))
	for _, item := range items {
		tasks = append(tasks, &merakv1.TaskSnapshot{
			Uid: item.UID, NamespaceUid: item.NamespaceUID, Phase: item.Phase,
			ChannelType: item.ChannelType, ChannelApp: item.ChannelApp, ChannelEndpoint: item.ChannelEndpoint,
			ChannelUsername: item.ChannelUsername, ChannelPassword: item.ChannelPassword, ChannelSecret: item.ChannelSecret,
			ChannelFrom: item.ChannelFrom, TemplateType: item.TemplateType, TemplateSubject: item.TemplateSubject,
			TemplateBody: item.TemplateBody, Receivers: item.Receivers, Params: item.Params,
			Subject: item.Subject, Body: item.Body, ChannelUid: item.ChannelUID, TemplateUid: item.TemplateUID,
		})
	}
	return &merakv1.ClaimTasksResponse{Tasks: tasks}, nil
}

func (s *NotifyService) MarkSending(ctx context.Context, req *merakv1.MarkSendingRequest) (*merakv1.MarkSendingResponse, error) {
	if err := s.notify.MarkSending(ctx, req.GetTaskUid()); err != nil {
		return nil, err
	}
	return &merakv1.MarkSendingResponse{}, nil
}

func (s *NotifyService) ReportNotify(ctx context.Context, req *merakv1.ReportNotifyRequest) (*merakv1.ReportNotifyResponse, error) {
	if err := s.notify.Report(ctx, bo.NotifyReport{TaskUID: req.GetTaskUid(), OK: req.GetOk(), Detail: req.GetDetail()}); err != nil {
		return nil, err
	}
	return &merakv1.ReportNotifyResponse{}, nil
}

func channelInput(uid, name, kind, app, endpoint, username, password, secret, from, remark string) bo.NotificationChannel {
	return bo.NotificationChannel{
		UID: uid, Name: name, Type: kind, App: app, Endpoint: endpoint, Username: username,
		Password: password, Secret: secret, FromAddress: from, Remark: remark,
	}
}

func channelItem(item bo.NotificationChannel) *v1.NotificationChannelItem {
	return &v1.NotificationChannelItem{
		Uid: item.UID, Name: item.Name, Type: item.Type, App: item.App, Status: item.Status,
		Endpoint: item.Endpoint, Username: item.Username, FromAddress: item.FromAddress, Remark: item.Remark,
		HasPassword: item.HasPassword, HasSecret: item.HasSecret,
	}
}

func channelItems(items []bo.NotificationChannel) []*v1.NotificationChannelItem {
	out := make([]*v1.NotificationChannelItem, 0, len(items))
	for _, item := range items {
		out = append(out, channelItem(item))
	}
	return out
}

func templateItem(item bo.NotificationTemplate) *v1.NotificationTemplateItem {
	return &v1.NotificationTemplateItem{
		Uid: item.UID, Name: item.Name, Type: item.Type, Status: item.Status,
		Subject: item.Subject, Body: item.Body, Remark: item.Remark,
	}
}

func templateItems(items []bo.NotificationTemplate) []*v1.NotificationTemplateItem {
	out := make([]*v1.NotificationTemplateItem, 0, len(items))
	for _, item := range items {
		out = append(out, templateItem(item))
	}
	return out
}

func notifyGroupItem(item bo.NotificationGroup) *v1.NotificationGroupItem {
	return &v1.NotificationGroupItem{
		Uid: item.UID, Name: item.Name, Status: item.Status, ChannelUid: item.ChannelUID,
		TemplateUid: item.TemplateUID, Receivers: item.Receivers, Remark: item.Remark,
	}
}

func notifyGroupItems(items []bo.NotificationGroup) []*v1.NotificationGroupItem {
	out := make([]*v1.NotificationGroupItem, 0, len(items))
	for _, item := range items {
		out = append(out, notifyGroupItem(item))
	}
	return out
}

func messageLogItem(item bo.MessageLog) *v1.MessageLogItem {
	return &v1.MessageLogItem{
		Uid: item.UID, TaskUid: item.TaskUID, Status: item.Status, ChannelUid: item.ChannelUID,
		TemplateUid: item.TemplateUID, Receivers: item.Receivers, Detail: item.Detail,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func messageLogItems(items []bo.MessageLog) []*v1.MessageLogItem {
	out := make([]*v1.MessageLogItem, 0, len(items))
	for _, item := range items {
		out = append(out, messageLogItem(item))
	}
	return out
}

func taskItem(item bo.NotificationTask) *v1.NotificationTaskItem {
	return &v1.NotificationTaskItem{
		Uid: item.UID, EventUid: item.EventUID, Phase: item.Phase, Status: item.Status,
		RuleUid: item.RuleUID, GroupUid: item.GroupUID, ChannelUid: item.ChannelUID,
		TemplateUid: item.TemplateUID, Receivers: item.Receivers, Subject: item.Subject,
		RetriesLeft: int32(item.RetriesLeft), LastError: item.LastError,
	}
}

func taskItems(items []bo.NotificationTask) []*v1.NotificationTaskItem {
	out := make([]*v1.NotificationTaskItem, 0, len(items))
	for _, item := range items {
		out = append(out, taskItem(item))
	}
	return out
}
