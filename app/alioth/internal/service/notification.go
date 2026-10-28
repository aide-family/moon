package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	merakv1 "github.com/aide-family/moon/api/merak/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/notification"
	"github.com/aide-family/moon/app/alioth/internal/biz/worker"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

type NotificationChannelService struct {
	v1.UnimplementedNotificationChannelServiceServer
	center *notification.Center
}

func NewNotificationChannelService(center *notification.Center) *NotificationChannelService {
	return &NotificationChannelService{center: center}
}

func (s *NotificationChannelService) CreateNotificationChannel(ctx context.Context, req *v1.CreateNotificationChannelRequest) (*v1.CreateNotificationChannelResponse, error) {
	item, err := s.center.CreateChannel(ctx, channelInput("", req.GetName(), req.GetType(), req.GetApp(), req.GetEndpoint(), req.GetUsername(), req.GetPassword(), req.GetSecret(), req.GetFromAddress(), req.GetRemark()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateNotificationChannelResponse{Channel: channelItem(*item)}, nil
}

func (s *NotificationChannelService) UpdateNotificationChannel(ctx context.Context, req *v1.UpdateNotificationChannelRequest) (*v1.UpdateNotificationChannelResponse, error) {
	item, err := s.center.UpdateChannel(ctx, channelInput(req.GetUid(), req.GetName(), req.GetType(), req.GetApp(), req.GetEndpoint(), req.GetUsername(), req.GetPassword(), req.GetSecret(), req.GetFromAddress(), req.GetRemark()))
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationChannelResponse{Channel: channelItem(*item)}, nil
}

func (s *NotificationChannelService) UpdateNotificationChannelStatus(ctx context.Context, req *v1.UpdateNotificationChannelStatusRequest) (*v1.UpdateNotificationChannelStatusResponse, error) {
	item, err := s.center.UpdateChannelStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationChannelStatusResponse{Channel: channelItem(*item)}, nil
}

func (s *NotificationChannelService) DeleteNotificationChannel(ctx context.Context, req *v1.DeleteNotificationChannelRequest) (*v1.DeleteNotificationChannelResponse, error) {
	if err := s.center.DeleteChannel(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNotificationChannelResponse{}, nil
}

func (s *NotificationChannelService) GetNotificationChannel(ctx context.Context, req *v1.GetNotificationChannelRequest) (*v1.GetNotificationChannelResponse, error) {
	item, err := s.center.GetChannel(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNotificationChannelResponse{Channel: channelItem(*item)}, nil
}

func (s *NotificationChannelService) ListNotificationChannels(ctx context.Context, req *v1.ListNotificationChannelsRequest) (*v1.ListNotificationChannelsResponse, error) {
	items, total, page, pageSize, err := s.center.ListChannels(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetType(), req.GetApp())
	if err != nil {
		return nil, err
	}
	return &v1.ListNotificationChannelsResponse{Items: channelItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NotificationChannelService) SelectNotificationChannels(ctx context.Context, req *v1.SelectNotificationChannelsRequest) (*v1.SelectNotificationChannelsResponse, error) {
	items, err := s.center.SelectChannels(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNotificationChannelsResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *NotificationChannelService) TestNotificationChannel(ctx context.Context, req *v1.TestNotificationChannelRequest) (*v1.TestNotificationChannelResponse, error) {
	uid, err := s.center.TestChannel(ctx, req.GetUid(), req.GetTemplateUid(), req.GetReceivers(), req.GetSubject(), req.GetBody(), req.GetParams())
	if err != nil {
		return nil, err
	}
	return &v1.TestNotificationChannelResponse{TaskUid: uid}, nil
}

type NotificationTemplateService struct {
	v1.UnimplementedNotificationTemplateServiceServer
	center *notification.Center
}

func NewNotificationTemplateService(center *notification.Center) *NotificationTemplateService {
	return &NotificationTemplateService{center: center}
}

func (s *NotificationTemplateService) ListNotificationPresets(ctx context.Context, _ *v1.ListNotificationPresetsRequest) (*v1.ListNotificationPresetsResponse, error) {
	items, err := s.center.ListPresets(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.NotificationPresetItem, 0, len(items))
	for i := range items {
		out = append(out, &v1.NotificationPresetItem{
			Uid: items[i].UID, Name: items[i].Name, Type: items[i].Type, App: items[i].App,
			Subject: items[i].Subject, Body: items[i].Body, Description: items[i].Description,
		})
	}
	return &v1.ListNotificationPresetsResponse{Items: out}, nil
}

func (s *NotificationTemplateService) CreateNotificationTemplate(ctx context.Context, req *v1.CreateNotificationTemplateRequest) (*v1.CreateNotificationTemplateResponse, error) {
	item, err := s.center.CreateTemplate(ctx, bo.NotificationTemplate{
		Name: req.GetName(), Type: req.GetType(), App: req.GetApp(),
		Subject: req.GetSubject(), Body: req.GetBody(), Remark: req.GetRemark(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateNotificationTemplateResponse{Template: templateItem(*item)}, nil
}

func (s *NotificationTemplateService) UpdateNotificationTemplate(ctx context.Context, req *v1.UpdateNotificationTemplateRequest) (*v1.UpdateNotificationTemplateResponse, error) {
	item, err := s.center.UpdateTemplate(ctx, bo.NotificationTemplate{
		UID: req.GetUid(), Name: req.GetName(), Type: req.GetType(), App: req.GetApp(),
		Subject: req.GetSubject(), Body: req.GetBody(), Remark: req.GetRemark(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationTemplateResponse{Template: templateItem(*item)}, nil
}

func (s *NotificationTemplateService) UpdateNotificationTemplateStatus(ctx context.Context, req *v1.UpdateNotificationTemplateStatusRequest) (*v1.UpdateNotificationTemplateStatusResponse, error) {
	item, err := s.center.UpdateTemplateStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationTemplateStatusResponse{Template: templateItem(*item)}, nil
}

func (s *NotificationTemplateService) DeleteNotificationTemplate(ctx context.Context, req *v1.DeleteNotificationTemplateRequest) (*v1.DeleteNotificationTemplateResponse, error) {
	if err := s.center.DeleteTemplate(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNotificationTemplateResponse{}, nil
}

func (s *NotificationTemplateService) GetNotificationTemplate(ctx context.Context, req *v1.GetNotificationTemplateRequest) (*v1.GetNotificationTemplateResponse, error) {
	item, err := s.center.GetTemplate(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNotificationTemplateResponse{Template: templateItem(*item)}, nil
}

func (s *NotificationTemplateService) ListNotificationTemplates(ctx context.Context, req *v1.ListNotificationTemplatesRequest) (*v1.ListNotificationTemplatesResponse, error) {
	items, total, page, pageSize, err := s.center.ListTemplates(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetType(), req.GetApp())
	if err != nil {
		return nil, err
	}
	return &v1.ListNotificationTemplatesResponse{Items: templateItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NotificationTemplateService) SelectNotificationTemplates(ctx context.Context, req *v1.SelectNotificationTemplatesRequest) (*v1.SelectNotificationTemplatesResponse, error) {
	items, err := s.center.SelectTemplates(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNotificationTemplatesResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *NotificationTemplateService) TestNotificationTemplate(ctx context.Context, req *v1.TestNotificationTemplateRequest) (*v1.TestNotificationTemplateResponse, error) {
	uid, err := s.center.TestTemplate(ctx, req.GetUid(), req.GetChannelUid(), req.GetReceivers(), req.GetParams())
	if err != nil {
		return nil, err
	}
	return &v1.TestNotificationTemplateResponse{TaskUid: uid}, nil
}

func (s *NotificationTemplateService) PreviewNotificationTemplate(ctx context.Context, req *v1.PreviewNotificationTemplateRequest) (*v1.PreviewNotificationTemplateResponse, error) {
	subject, body, err := s.center.PreviewTemplate(ctx, bo.NotificationTemplate{
		Type: req.GetType(), App: req.GetApp(), Subject: req.GetSubject(), Body: req.GetBody(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.PreviewNotificationTemplateResponse{Subject: subject, Body: body}, nil
}

type NotificationGroupService struct {
	v1.UnimplementedNotificationGroupServiceServer
	center *notification.Center
}

func NewNotificationGroupService(center *notification.Center) *NotificationGroupService {
	return &NotificationGroupService{center: center}
}

func (s *NotificationGroupService) CreateNotificationGroup(ctx context.Context, req *v1.CreateNotificationGroupRequest) (*v1.CreateNotificationGroupResponse, error) {
	item, err := s.center.CreateGroup(ctx, createGroupInput(req))
	if err != nil {
		return nil, err
	}
	return &v1.CreateNotificationGroupResponse{Group: notifyGroupItem(*item)}, nil
}

func (s *NotificationGroupService) UpdateNotificationGroup(ctx context.Context, req *v1.UpdateNotificationGroupRequest) (*v1.UpdateNotificationGroupResponse, error) {
	item, err := s.center.UpdateGroup(ctx, updateGroupInput(req))
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationGroupResponse{Group: notifyGroupItem(*item)}, nil
}

func (s *NotificationGroupService) UpdateNotificationGroupStatus(ctx context.Context, req *v1.UpdateNotificationGroupStatusRequest) (*v1.UpdateNotificationGroupStatusResponse, error) {
	item, err := s.center.UpdateGroupStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationGroupStatusResponse{Group: notifyGroupItem(*item)}, nil
}

func (s *NotificationGroupService) DeleteNotificationGroup(ctx context.Context, req *v1.DeleteNotificationGroupRequest) (*v1.DeleteNotificationGroupResponse, error) {
	if err := s.center.DeleteGroup(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNotificationGroupResponse{}, nil
}

func (s *NotificationGroupService) GetNotificationGroup(ctx context.Context, req *v1.GetNotificationGroupRequest) (*v1.GetNotificationGroupResponse, error) {
	item, err := s.center.GetGroup(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNotificationGroupResponse{Group: notifyGroupItem(*item)}, nil
}

func (s *NotificationGroupService) ListNotificationGroups(ctx context.Context, req *v1.ListNotificationGroupsRequest) (*v1.ListNotificationGroupsResponse, error) {
	items, total, page, pageSize, err := s.center.ListGroups(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListNotificationGroupsResponse{Items: notifyGroupItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NotificationGroupService) SelectNotificationGroups(ctx context.Context, req *v1.SelectNotificationGroupsRequest) (*v1.SelectNotificationGroupsResponse, error) {
	items, err := s.center.SelectGroups(ctx, req.GetLimit(), req.GetLastUid(), req.GetActiveOnly())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNotificationGroupsResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

type MessageLogService struct {
	v1.UnimplementedMessageLogServiceServer
	center *notification.Center
}

func NewMessageLogService(center *notification.Center) *MessageLogService {
	return &MessageLogService{center: center}
}

func (s *MessageLogService) ListMessageLogs(ctx context.Context, req *v1.ListMessageLogsRequest) (*v1.ListMessageLogsResponse, error) {
	filter := bo.MessageLogListFilter{
		TaskUID: req.GetTaskUid(),
		Status:  req.GetStatus(),
		StartAt: unixTime(req.GetStartTime()),
		EndAt:   unixTime(req.GetEndTime()),
	}
	items, total, page, pageSize, err := s.center.ListMessageLogs(ctx, req.GetPage(), req.GetPageSize(), filter)
	if err != nil {
		return nil, err
	}
	return &v1.ListMessageLogsResponse{Items: messageLogItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *MessageLogService) GetMessageLog(ctx context.Context, req *v1.GetMessageLogRequest) (*v1.GetMessageLogResponse, error) {
	item, err := s.center.GetMessageLog(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetMessageLogResponse{Log: messageLogItem(*item)}, nil
}

func (s *MessageLogService) ListNotificationTasks(ctx context.Context, req *v1.ListNotificationTasksRequest) (*v1.ListNotificationTasksResponse, error) {
	filter := bo.NotificationTaskListFilter{
		Status:  req.GetStatus(),
		StartAt: unixTime(req.GetStartTime()),
		EndAt:   unixTime(req.GetEndTime()),
	}
	items, total, page, pageSize, err := s.center.ListTasks(ctx, req.GetPage(), req.GetPageSize(), filter)
	if err != nil {
		return nil, err
	}
	return &v1.ListNotificationTasksResponse{Items: taskItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *MessageLogService) RetryNotificationTask(ctx context.Context, req *v1.RetryNotificationTaskRequest) (*v1.RetryNotificationTaskResponse, error) {
	item, err := s.center.RetryTask(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.RetryNotificationTaskResponse{Task: taskItem(*item)}, nil
}

func (s *MessageLogService) CancelNotificationTask(ctx context.Context, req *v1.CancelNotificationTaskRequest) (*v1.CancelNotificationTaskResponse, error) {
	item, err := s.center.CancelTask(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.CancelNotificationTaskResponse{Task: taskItem(*item)}, nil
}

type NotifyService struct {
	merakv1.UnimplementedNotifyServiceServer
	center *notification.Center
	roster *worker.Roster
}

func NewNotifyService(center *notification.Center, roster *worker.Roster) *NotifyService {
	return &NotifyService{center: center, roster: roster}
}

func (s *NotifyService) ClaimTasks(ctx context.Context, _ *merakv1.ClaimTasksRequest) (*merakv1.ClaimTasksResponse, error) {
	items, err := s.center.Claim(ctx)
	if err != nil {
		return nil, err
	}
	tasks := make([]*merakv1.TaskSnapshot, 0, len(items))
	for _, item := range items {
		tasks = append(tasks, &merakv1.TaskSnapshot{
			Uid: item.UID, NamespaceUid: item.NamespaceUID, Phase: item.Phase,
			ChannelType: item.ChannelType, ChannelApp: item.ChannelApp, ChannelEndpoint: item.ChannelEndpoint,
			ChannelUsername: item.ChannelUsername, ChannelPassword: item.ChannelPassword, ChannelSecret: item.ChannelSecret,
			ChannelFrom: item.ChannelFrom, Sender: item.Sender, TemplateType: item.TemplateType, TemplateSubject: item.TemplateSubject,
			TemplateBody: item.TemplateBody, Receivers: item.Receivers, Params: item.Params,
			Subject: item.Subject, Body: item.Body, ChannelUid: item.ChannelUID, TemplateUid: item.TemplateUID,
		})
	}
	return &merakv1.ClaimTasksResponse{Tasks: tasks}, nil
}

func (s *NotifyService) MarkSending(ctx context.Context, req *merakv1.MarkSendingRequest) (*merakv1.MarkSendingResponse, error) {
	if err := s.center.MarkSending(ctx, req.GetTaskUid()); err != nil {
		return nil, err
	}
	return &merakv1.MarkSendingResponse{}, nil
}

func (s *NotifyService) ReportNotify(ctx context.Context, req *merakv1.ReportNotifyRequest) (*merakv1.ReportNotifyResponse, error) {
	if err := s.center.Report(ctx, bo.NotifyReport{
		TaskUID: req.GetTaskUid(), OK: req.GetOk(), Detail: req.GetDetail(),
		Subject: req.GetSubject(), Content: req.GetContent(),
	}); err != nil {
		return nil, err
	}
	return &merakv1.ReportNotifyResponse{}, nil
}

func (s *NotifyService) ReportWorkerStatus(ctx context.Context, req *merakv1.ReportWorkerStatusRequest) (*merakv1.ReportWorkerStatusResponse, error) {
	var startedAt *time.Time
	if req.GetBusy() {
		parsed, err := parseSlot(req.GetTaskStartedAt())
		if err != nil {
			return nil, err
		}
		startedAt = &parsed
	}
	if err := s.roster.Report(ctx, "MERAK", bo.WorkerStatus{
		InstanceUID: req.GetInstanceUid(), Hostname: req.GetHostname(), Busy: req.GetBusy(),
		TaskUID: req.GetTaskUid(), TaskName: req.GetTaskName(), TaskStartedAt: startedAt,
	}); err != nil {
		return nil, err
	}
	return &merakv1.ReportWorkerStatusResponse{}, nil
}

func createGroupInput(req *v1.CreateNotificationGroupRequest) bo.NotificationGroup {
	return bo.NotificationGroup{
		Name: req.GetName(), ChannelUID: req.GetChannelUid(), TemplateUID: req.GetTemplateUid(),
		Receivers: req.GetReceivers(), MemberUIDs: req.GetMemberUids(),
		WebhookChannelUIDs: req.GetWebhookChannelUids(), WebhookTemplateUID: req.GetWebhookTemplateUid(),
		ChannelUIDs: req.GetChannelUids(), TemplateUIDs: req.GetTemplateUids(),
		EmailReceivers: req.GetEmailReceivers(), SMSReceivers: req.GetSmsReceivers(),
		EmailMemberUIDs: req.GetEmailMemberUids(), SMSMemberUIDs: req.GetSmsMemberUids(),
		Remark: req.GetRemark(),
	}
}

func updateGroupInput(req *v1.UpdateNotificationGroupRequest) bo.NotificationGroup {
	return bo.NotificationGroup{
		UID: req.GetUid(), Name: req.GetName(), ChannelUID: req.GetChannelUid(), TemplateUID: req.GetTemplateUid(),
		Receivers: req.GetReceivers(), MemberUIDs: req.GetMemberUids(),
		WebhookChannelUIDs: req.GetWebhookChannelUids(), WebhookTemplateUID: req.GetWebhookTemplateUid(),
		ChannelUIDs: req.GetChannelUids(), TemplateUIDs: req.GetTemplateUids(),
		EmailReceivers: req.GetEmailReceivers(), SMSReceivers: req.GetSmsReceivers(),
		EmailMemberUIDs: req.GetEmailMemberUids(), SMSMemberUIDs: req.GetSmsMemberUids(),
		Remark: req.GetRemark(),
	}
}

func channelInput(uid, name, kind, app, endpoint, username, password, secret, from, remark string) bo.NotificationChannel {
	return bo.NotificationChannel{
		UID: uid, Name: name, Type: kind, App: app, Endpoint: endpoint, Username: username,
		Password: password, Secret: secret, FromAddress: from, Remark: remark,
	}
}

// unixTime 把请求里的 Unix 秒转换为 UTC 时间；0 表示不限制。
func unixTime(seconds int64) time.Time {
	if seconds <= 0 {
		return time.Time{}
	}
	return time.Unix(seconds, 0).UTC()
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
		Uid: item.UID, Name: item.Name, Type: item.Type, App: item.App, Status: item.Status,
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
		Uid: item.UID, Name: item.Name, Status: item.Status,
		ChannelUid: item.ChannelUID, TemplateUid: item.TemplateUID, Receivers: item.Receivers,
		Remark: item.Remark, MemberUids: item.MemberUIDs,
		WebhookChannelUids: item.WebhookChannelUIDs, WebhookTemplateUid: item.WebhookTemplateUID,
		ChannelUids: item.ChannelUIDs, TemplateUids: item.TemplateUIDs,
		EmailReceivers: item.EmailReceivers, SmsReceivers: item.SMSReceivers,
		EmailMemberUids: item.EmailMemberUIDs, SmsMemberUids: item.SMSMemberUIDs,
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
		TemplateUid: item.TemplateUID, Sender: item.Sender, Receivers: item.Receivers,
		Subject: item.Subject, Content: item.Content, Detail: item.Detail,
		GroupName: item.GroupName, SubscriptionName: item.SubscriptionName,
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
		RuleUid: item.RuleUID, SubscriptionUid: item.SubscriptionUID, GroupUid: item.GroupUID,
		ChannelUid: item.ChannelUID, TemplateUid: item.TemplateUID, Sender: item.Sender,
		Receivers: item.Receivers, Subject: item.Subject, Content: item.Content,
		RetriesLeft: int32(item.RetriesLeft), LastError: item.LastError,
		GroupName: item.GroupName, SubscriptionName: item.SubscriptionName,
		ChannelName: item.ChannelName, TemplateName: item.TemplateName,
	}
}

func taskItems(items []bo.NotificationTask) []*v1.NotificationTaskItem {
	out := make([]*v1.NotificationTaskItem, 0, len(items))
	for _, item := range items {
		out = append(out, taskItem(item))
	}
	return out
}
