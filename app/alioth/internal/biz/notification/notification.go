package notification

import (
	"context"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct {
	repo   repository.NotificationRepo
	access access.Checker
}

func NewUsecase(repo repository.NotificationRepo, checker access.Checker) *Usecase {
	return &Usecase{repo: repo, access: checker}
}

func (uc *Usecase) CreateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	if err := validateChannel(in); err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.repo.CreateChannel(ctx, in)
	return redactChannel(item), uc.access.Wrap(err)
}

func (uc *Usecase) UpdateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	if err := validateChannel(in); err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.repo.UpdateChannel(ctx, in)
	return redactChannel(item), uc.access.Wrap(err)
}

func (uc *Usecase) UpdateChannelStatus(ctx context.Context, uid, status string) (*bo.NotificationChannel, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateChannelStatus(ctx, ns, uid, status)
	return redactChannel(item), uc.access.Wrap(err)
}

func (uc *Usecase) DeleteChannel(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.repo.DeleteChannel(ctx, ns, uid))
}

func (uc *Usecase) GetChannel(ctx context.Context, uid string) (*bo.NotificationChannel, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.GetChannel(ctx, ns, uid)
	return redactChannel(item), uc.access.Wrap(err)
}

func (uc *Usecase) ListChannels(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.NotificationChannel, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.repo.ListChannels(ctx, ns, p, s, strings.TrimSpace(name))
	out := make([]bo.NotificationChannel, 0, len(items))
	for i := range items {
		if item := redactChannel(&items[i]); item != nil {
			out = append(out, *item)
		}
	}
	return out, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) SelectChannels(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	items, err := uc.repo.SelectChannels(ctx, ns, int(limit), lastUID)
	return items, uc.access.Wrap(err)
}

// TestChannel 只写入 PENDING 任务。短信类型直接拒绝，避免被报成 SENT。
func (uc *Usecase) TestChannel(ctx context.Context, channelUID, templateUID, receivers, subject, body, params string) (string, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return "", err
	}
	channel, err := uc.repo.GetChannel(ctx, ns, channelUID)
	if err != nil {
		return "", uc.access.Wrap(err)
	}
	if channel.App == "SMS_ALICLOUD" {
		return "", merr.Invalid("sms delivery is not implemented")
	}
	if templateUID != "" {
		template, err := uc.repo.GetTemplate(ctx, ns, templateUID)
		if err != nil {
			return "", uc.access.Wrap(err)
		}
		if template.Type == "SMS_ALICLOUD" {
			return "", merr.Invalid("sms delivery is not implemented")
		}
		if subject == "" {
			subject = template.Subject
		}
		if body == "" {
			body = template.Body
		}
	}
	task, err := uc.repo.CreateTestTask(ctx, bo.NotificationTask{
		NamespaceUID: ns,
		ChannelUID:   channelUID,
		TemplateUID:  templateUID,
		Receivers:    strings.TrimSpace(receivers),
		Subject:      subject,
		Body:         body,
		Params:       params,
	})
	if err != nil {
		return "", uc.access.Wrap(err)
	}
	return task.UID, nil
}

func (uc *Usecase) CreateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.repo.CreateTemplate(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.repo.UpdateTemplate(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateTemplateStatus(ctx context.Context, uid, status string) (*bo.NotificationTemplate, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateTemplateStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) DeleteTemplate(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.repo.DeleteTemplate(ctx, ns, uid))
}

func (uc *Usecase) GetTemplate(ctx context.Context, uid string) (*bo.NotificationTemplate, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.GetTemplate(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListTemplates(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.NotificationTemplate, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.repo.ListTemplates(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) SelectTemplates(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	items, err := uc.repo.SelectTemplates(ctx, ns, int(limit), lastUID)
	return items, uc.access.Wrap(err)
}

func (uc *Usecase) TestTemplate(ctx context.Context, templateUID, channelUID, receivers, params string) (string, error) {
	return uc.TestChannel(ctx, channelUID, templateUID, receivers, "", "", params)
}

func (uc *Usecase) CreateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	if _, err := uc.repo.GetChannel(ctx, ns, in.ChannelUID); err != nil {
		return nil, uc.access.Wrap(err)
	}
	if _, err := uc.repo.GetTemplate(ctx, ns, in.TemplateUID); err != nil {
		return nil, uc.access.Wrap(err)
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.repo.CreateGroup(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	if _, err := uc.repo.GetChannel(ctx, ns, in.ChannelUID); err != nil {
		return nil, uc.access.Wrap(err)
	}
	if _, err := uc.repo.GetTemplate(ctx, ns, in.TemplateUID); err != nil {
		return nil, uc.access.Wrap(err)
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := uc.repo.UpdateGroup(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateGroupStatus(ctx context.Context, uid, status string) (*bo.NotificationGroup, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateGroupStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) DeleteGroup(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.repo.DeleteGroup(ctx, ns, uid))
}

func (uc *Usecase) GetGroup(ctx context.Context, uid string) (*bo.NotificationGroup, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.GetGroup(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListGroups(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.NotificationGroup, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.repo.ListGroups(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) SelectGroups(ctx context.Context, limit int32, lastUID string, activeOnly bool) ([]bo.SelectItem, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	items, err := uc.repo.SelectGroups(ctx, ns, int(limit), lastUID, activeOnly)
	return items, uc.access.Wrap(err)
}

func (uc *Usecase) ListMessageLogs(ctx context.Context, pageNo, pageSize int32, taskUID string) ([]bo.MessageLog, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.repo.ListMessageLogs(ctx, ns, taskUID, p, s)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) GetMessageLog(ctx context.Context, uid string) (*bo.MessageLog, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.GetMessageLog(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListTasks(ctx context.Context, pageNo, pageSize int32, status string) ([]bo.NotificationTask, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.repo.ListTasks(ctx, ns, status, p, s)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) RetryTask(ctx context.Context, uid string) (*bo.NotificationTask, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.RetryTask(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) CancelTask(ctx context.Context, uid string) (*bo.NotificationTask, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.CancelTask(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

// Claim 返回 PENDING 任务。只接受名为 merak 的 service key。
func (uc *Usecase) Claim(ctx context.Context) ([]bo.NotifySnapshot, error) {
	if err := requireMerak(ctx); err != nil {
		return nil, err
	}
	return uc.repo.ClaimPending(ctx, 20)
}

func (uc *Usecase) MarkSending(ctx context.Context, taskUID string) error {
	if err := requireMerak(ctx); err != nil {
		return err
	}
	return uc.repo.MarkSending(ctx, taskUID)
}

func (uc *Usecase) Report(ctx context.Context, report bo.NotifyReport) error {
	if err := requireMerak(ctx); err != nil {
		return err
	}
	return uc.repo.ApplyNotifyReport(ctx, report)
}

func requireMerak(ctx context.Context) error {
	name, ok := actor.ServiceName(ctx)
	if !ok || name != "merak" {
		return merr.Forbidden("service key cannot call this api")
	}
	return nil
}

func validateChannel(in bo.NotificationChannel) error {
	if strings.TrimSpace(in.Endpoint) == "" {
		return merr.Invalid("invalid request")
	}
	if in.Type == "SMTP" && in.App == "SMS_ALICLOUD" {
		return merr.Invalid("invalid request")
	}
	return nil
}

func redactChannel(item *bo.NotificationChannel) *bo.NotificationChannel {
	if item == nil {
		return nil
	}
	out := *item
	out.Password = ""
	out.Secret = ""
	return &out
}
