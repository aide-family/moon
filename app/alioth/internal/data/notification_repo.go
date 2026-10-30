package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/messagelog"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/messagelogshard"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationoncallgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationpreset"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationsubscription"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationsubscriptiongroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtemplate"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/alerttmpl"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *notificationRepo) CreateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error) {
	kind, app, status, err := channelEnums(in.Type, in.App, "ACTIVE")
	if err != nil {
		return nil, err
	}
	password, err := r.box.encrypt(in.Password)
	if err != nil {
		return nil, err
	}
	secret, err := r.box.encrypt(in.Secret)
	if err != nil {
		return nil, err
	}
	row, err := r.client.NotificationChannel.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetType(kind).
		SetApp(app).
		SetStatus(status).
		SetEndpoint(in.Endpoint).
		SetUsername(in.Username).
		SetPasswordCipher(password).
		SetSecretCipher(secret).
		SetFromAddress(in.FromAddress).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationChannel(row)
	return &item, nil
}

func (r *notificationRepo) UpdateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error) {
	row, err := r.client.NotificationChannel.Query().Where(
		notificationchannel.NamespaceUIDEQ(in.NamespaceUID),
		notificationchannel.UIDEQ(in.UID),
		notificationchannel.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	kind, app, _, err := channelEnums(in.Type, in.App, string(row.Status))
	if err != nil {
		return nil, err
	}
	password := row.PasswordCipher
	if in.Password != "" {
		password, err = r.box.encrypt(in.Password)
		if err != nil {
			return nil, err
		}
	}
	secret := row.SecretCipher
	if in.Secret != "" {
		secret, err = r.box.encrypt(in.Secret)
		if err != nil {
			return nil, err
		}
	}
	updated, err := r.client.NotificationChannel.UpdateOne(row).
		SetName(in.Name).
		SetType(kind).
		SetApp(app).
		SetEndpoint(in.Endpoint).
		SetUsername(in.Username).
		SetPasswordCipher(password).
		SetSecretCipher(secret).
		SetFromAddress(in.FromAddress).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationChannel(updated)
	return &item, nil
}

func (r *notificationRepo) UpdateChannelStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationChannel, error) {
	value := notificationchannel.StatusACTIVE
	if status == "DISABLED" {
		value = notificationchannel.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.NotificationChannel.Update().Where(
		notificationchannel.NamespaceUIDEQ(namespaceUID),
		notificationchannel.UIDEQ(uid),
		notificationchannel.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetChannel(ctx, namespaceUID, uid)
}

func (r *notificationRepo) DeleteChannel(ctx context.Context, namespaceUID, uid string) error {
	groups, err := r.client.NotificationGroup.Query().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	for _, group := range groups {
		if group.ChannelUID == uid || containsString(group.ChannelUids, uid) || containsString(group.WebhookChannelUids, uid) {
			return merr.Conflict("channel is used by a notification group")
		}
	}
	n, err := r.client.NotificationChannel.Update().Where(
		notificationchannel.NamespaceUIDEQ(namespaceUID),
		notificationchannel.UIDEQ(uid),
		notificationchannel.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *notificationRepo) GetChannel(ctx context.Context, namespaceUID, uid string) (*bo.NotificationChannel, error) {
	row, err := r.client.NotificationChannel.Query().Where(
		notificationchannel.NamespaceUIDEQ(namespaceUID),
		notificationchannel.UIDEQ(uid),
		notificationchannel.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationChannel(row)
	return &item, nil
}

func (r *notificationRepo) ListChannels(ctx context.Context, namespaceUID string, page, pageSize int, name, kind, app string, types []string) ([]bo.NotificationChannel, int, error) {
	query := r.client.NotificationChannel.Query().Where(notificationchannel.NamespaceUIDEQ(namespaceUID), notificationchannel.DeletedAtIsNil())
	if name != "" {
		query = query.Where(notificationchannel.NameContainsFold(name))
	}
	if kind = strings.ToUpper(strings.TrimSpace(kind)); kind != "" {
		query = query.Where(notificationchannel.TypeEQ(notificationchannel.Type(kind)))
	}
	if app = strings.ToUpper(strings.TrimSpace(app)); app != "" {
		query = query.Where(notificationchannel.AppEQ(notificationchannel.App(app)))
	}
	if len(types) > 0 {
		values := make([]notificationchannel.Type, 0, len(types))
		for _, item := range types {
			values = append(values, notificationchannel.Type(item))
		}
		query = query.Where(notificationchannel.TypeIn(values...))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(notificationchannel.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.NotificationChannel, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNotificationChannel(row))
	}
	return items, total, nil
}

func (r *notificationRepo) ListChannelsByUIDs(ctx context.Context, namespaceUID string, uids []string) ([]bo.NotificationChannel, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	rows, err := r.client.NotificationChannel.Query().Where(
		notificationchannel.NamespaceUIDEQ(namespaceUID),
		notificationchannel.UIDIn(uids...),
		notificationchannel.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.NotificationChannel, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNotificationChannel(row))
	}
	return items, nil
}

func (r *notificationRepo) SelectChannels(ctx context.Context, namespaceUID string, limit int, lastUID string, types []string) ([]bo.SelectItem, error) {
	query := r.client.NotificationChannel.Query().Where(
		notificationchannel.NamespaceUIDEQ(namespaceUID),
		notificationchannel.DeletedAtIsNil(),
		notificationchannel.StatusEQ(notificationchannel.StatusACTIVE),
	)
	if lastUID != "" {
		query = query.Where(notificationchannel.UIDGT(lastUID))
	}
	if len(types) > 0 {
		values := make([]notificationchannel.Type, 0, len(types))
		for _, item := range types {
			values = append(values, notificationchannel.Type(item))
		}
		query = query.Where(notificationchannel.TypeIn(values...))
	}
	rows, err := query.Order(ent.Asc(notificationchannel.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SelectItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.SelectItem{UID: row.UID, Name: row.Name})
	}
	return items, nil
}

func (r *notificationRepo) CreateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error) {
	kind, status, err := templateEnums(in.Type, "ACTIVE")
	if err != nil {
		return nil, err
	}
	app, err := templateApp(in.App)
	if err != nil {
		return nil, err
	}
	row, err := r.client.NotificationTemplate.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetType(kind).
		SetApp(app).
		SetStatus(status).
		SetSubject(in.Subject).
		SetBody(in.Body).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationTemplate(row)
	return &item, nil
}

func (r *notificationRepo) UpdateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error) {
	row, err := r.client.NotificationTemplate.Query().Where(
		notificationtemplate.NamespaceUIDEQ(in.NamespaceUID),
		notificationtemplate.UIDEQ(in.UID),
		notificationtemplate.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	kind, _, err := templateEnums(in.Type, string(row.Status))
	if err != nil {
		return nil, err
	}
	app, err := templateApp(in.App)
	if err != nil {
		return nil, err
	}
	updated, err := r.client.NotificationTemplate.UpdateOne(row).
		SetName(in.Name).
		SetType(kind).
		SetApp(app).
		SetSubject(in.Subject).
		SetBody(in.Body).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationTemplate(updated)
	return &item, nil
}

func (r *notificationRepo) UpdateTemplateStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationTemplate, error) {
	value := notificationtemplate.StatusACTIVE
	if status == "DISABLED" {
		value = notificationtemplate.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.NotificationTemplate.Update().Where(
		notificationtemplate.NamespaceUIDEQ(namespaceUID),
		notificationtemplate.UIDEQ(uid),
		notificationtemplate.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetTemplate(ctx, namespaceUID, uid)
}

func (r *notificationRepo) DeleteTemplate(ctx context.Context, namespaceUID, uid string) error {
	groups, err := r.client.NotificationGroup.Query().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	for _, group := range groups {
		if group.TemplateUID == uid || group.WebhookTemplateUID == uid || containsString(group.TemplateUids, uid) {
			return merr.Conflict("template is used by a notification group")
		}
	}
	n, err := r.client.NotificationTemplate.Update().Where(
		notificationtemplate.NamespaceUIDEQ(namespaceUID),
		notificationtemplate.UIDEQ(uid),
		notificationtemplate.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *notificationRepo) GetTemplate(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTemplate, error) {
	row, err := r.client.NotificationTemplate.Query().Where(
		notificationtemplate.NamespaceUIDEQ(namespaceUID),
		notificationtemplate.UIDEQ(uid),
		notificationtemplate.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationTemplate(row)
	return &item, nil
}

func (r *notificationRepo) ListTemplates(ctx context.Context, namespaceUID string, page, pageSize int, name, kind, app string, types []string) ([]bo.NotificationTemplate, int, error) {
	query := r.client.NotificationTemplate.Query().Where(notificationtemplate.NamespaceUIDEQ(namespaceUID), notificationtemplate.DeletedAtIsNil())
	if name != "" {
		query = query.Where(notificationtemplate.NameContainsFold(name))
	}
	if kind = strings.ToUpper(strings.TrimSpace(kind)); kind != "" {
		query = query.Where(notificationtemplate.TypeEQ(notificationtemplate.Type(kind)))
	}
	if app = strings.ToUpper(strings.TrimSpace(app)); app != "" {
		query = query.Where(notificationtemplate.AppEQ(notificationtemplate.App(app)))
	}
	if len(types) > 0 {
		values := make([]notificationtemplate.Type, 0, len(types))
		for _, item := range types {
			values = append(values, notificationtemplate.Type(item))
		}
		query = query.Where(notificationtemplate.TypeIn(values...))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(notificationtemplate.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.NotificationTemplate, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNotificationTemplate(row))
	}
	return items, total, nil
}

func (r *notificationRepo) ListTemplatesByUIDs(ctx context.Context, namespaceUID string, uids []string) ([]bo.NotificationTemplate, error) {
	if len(uids) == 0 {
		return nil, nil
	}
	rows, err := r.client.NotificationTemplate.Query().Where(
		notificationtemplate.NamespaceUIDEQ(namespaceUID),
		notificationtemplate.UIDIn(uids...),
		notificationtemplate.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.NotificationTemplate, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNotificationTemplate(row))
	}
	return items, nil
}

func (r *notificationRepo) SelectTemplates(ctx context.Context, namespaceUID string, limit int, lastUID string, types []string) ([]bo.SelectItem, error) {
	query := r.client.NotificationTemplate.Query().Where(
		notificationtemplate.NamespaceUIDEQ(namespaceUID),
		notificationtemplate.DeletedAtIsNil(),
		notificationtemplate.StatusEQ(notificationtemplate.StatusACTIVE),
	)
	if lastUID != "" {
		query = query.Where(notificationtemplate.UIDGT(lastUID))
	}
	if len(types) > 0 {
		values := make([]notificationtemplate.Type, 0, len(types))
		for _, item := range types {
			values = append(values, notificationtemplate.Type(item))
		}
		query = query.Where(notificationtemplate.TypeIn(values...))
	}
	rows, err := query.Order(ent.Asc(notificationtemplate.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SelectItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.SelectItem{UID: row.UID, Name: row.Name})
	}
	return items, nil
}

// ListPresets 返回系统预置模板，按 sort 排序。
func (r *notificationRepo) ListPresets(ctx context.Context) ([]bo.NotificationPreset, error) {
	rows, err := r.client.NotificationPreset.Query().Where(
		notificationpreset.StatusEQ(notificationpreset.StatusACTIVE),
	).Order(ent.Asc(notificationpreset.FieldSort), ent.Asc(notificationpreset.FieldName)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.NotificationPreset, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.NotificationPreset{
			UID: row.UID, Name: row.Name, Type: string(row.Type), App: string(row.App),
			Subject: row.Subject, Body: row.Body, Description: row.Description, Sort: row.Sort,
		})
	}
	return items, nil
}

func (r *notificationRepo) CreateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error) {
	row, err := r.client.NotificationGroup.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetStatus(notificationgroup.StatusACTIVE).
		SetChannelUids(in.ChannelUIDs).
		SetTemplateUids(in.TemplateUIDs).
		SetEmailReceivers(in.EmailReceivers).
		SetSmsReceivers(in.SMSReceivers).
		SetEmailMemberUids(in.EmailMemberUIDs).
		SetSmsMemberUids(in.SMSMemberUIDs).
		// 旧字段同时写入，保证回滚到旧版本时仍有可用数据。
		SetChannelUID(in.ChannelUID).
		SetTemplateUID(in.TemplateUID).
		SetReceivers(in.Receivers).
		SetMemberUids(in.MemberUIDs).
		SetWebhookChannelUids(in.WebhookChannelUIDs).
		SetWebhookTemplateUID(in.WebhookTemplateUID).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationGroup(row)
	return &item, nil
}

func (r *notificationRepo) UpdateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error) {
	row, err := r.client.NotificationGroup.Query().Where(
		notificationgroup.NamespaceUIDEQ(in.NamespaceUID),
		notificationgroup.UIDEQ(in.UID),
		notificationgroup.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	updated, err := r.client.NotificationGroup.UpdateOne(row).
		SetName(in.Name).
		SetChannelUids(in.ChannelUIDs).
		SetTemplateUids(in.TemplateUIDs).
		SetEmailReceivers(in.EmailReceivers).
		SetSmsReceivers(in.SMSReceivers).
		SetEmailMemberUids(in.EmailMemberUIDs).
		SetSmsMemberUids(in.SMSMemberUIDs).
		SetChannelUID(in.ChannelUID).
		SetTemplateUID(in.TemplateUID).
		SetReceivers(in.Receivers).
		SetMemberUids(in.MemberUIDs).
		SetWebhookChannelUids(in.WebhookChannelUIDs).
		SetWebhookTemplateUID(in.WebhookTemplateUID).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationGroup(updated)
	return &item, nil
}

func (r *notificationRepo) UpdateGroupStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationGroup, error) {
	value := notificationgroup.StatusACTIVE
	if status == "DISABLED" {
		value = notificationgroup.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.NotificationGroup.Update().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.UIDEQ(uid),
		notificationgroup.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetGroup(ctx, namespaceUID, uid)
}

func (r *notificationRepo) DeleteGroup(ctx context.Context, namespaceUID, uid string) error {
	used, err := r.client.NotificationSubscriptionGroup.Query().Where(
		notificationsubscriptiongroup.NamespaceUIDEQ(namespaceUID),
		notificationsubscriptiongroup.GroupUIDEQ(uid),
		notificationsubscriptiongroup.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if used {
		return merr.Conflict("group is bound to a notification subscription")
	}
	usedByOnCall, err := r.client.NotificationOnCallGroup.Query().Where(
		notificationoncallgroup.NamespaceUIDEQ(namespaceUID),
		notificationoncallgroup.GroupUIDEQ(uid),
		notificationoncallgroup.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if usedByOnCall {
		return merr.Conflict("group is bound to a notification on-call")
	}
	n, err := r.client.NotificationGroup.Update().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.UIDEQ(uid),
		notificationgroup.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *notificationRepo) GetGroup(ctx context.Context, namespaceUID, uid string) (*bo.NotificationGroup, error) {
	row, err := r.client.NotificationGroup.Query().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.UIDEQ(uid),
		notificationgroup.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationGroup(row)
	return &item, nil
}

func (r *notificationRepo) ListGroups(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationGroup, int, error) {
	query := r.client.NotificationGroup.Query().Where(notificationgroup.NamespaceUIDEQ(namespaceUID), notificationgroup.DeletedAtIsNil())
	if name != "" {
		query = query.Where(notificationgroup.NameContainsFold(name))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(notificationgroup.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.NotificationGroup, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNotificationGroup(row))
	}
	return items, total, nil
}

func (r *notificationRepo) SelectGroups(ctx context.Context, namespaceUID string, limit int, lastUID string, activeOnly bool) ([]bo.SelectItem, error) {
	query := r.client.NotificationGroup.Query().Where(notificationgroup.NamespaceUIDEQ(namespaceUID), notificationgroup.DeletedAtIsNil())
	if activeOnly {
		query = query.Where(notificationgroup.StatusEQ(notificationgroup.StatusACTIVE))
	}
	if lastUID != "" {
		query = query.Where(notificationgroup.UIDGT(lastUID))
	}
	rows, err := query.Order(ent.Asc(notificationgroup.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SelectItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.SelectItem{UID: row.UID, Name: row.Name})
	}
	return items, nil
}

func (r *notificationRepo) CreateTestTask(ctx context.Context, in bo.NotificationTask) (*bo.NotificationTask, error) {
	channel, err := r.client.NotificationChannel.Query().Where(
		notificationchannel.NamespaceUIDEQ(in.NamespaceUID),
		notificationchannel.UIDEQ(in.ChannelUID),
		notificationchannel.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	receivers := in.Receivers
	if receivers == "" {
		receivers = notificationReceiver(channel)
	}
	row, err := r.client.NotificationTask.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetEventUID("").
		SetPhase(notificationtask.PhaseTEST).
		SetStatus(notificationtask.StatusPENDING).
		SetChannelUID(in.ChannelUID).
		SetTemplateUID(in.TemplateUID).
		SetSender(notificationSender(channel)).
		SetReceivers(receivers).
		SetSubject(in.Subject).
		SetBody(in.Body).
		SetContent(in.Body).
		SetParams(in.Params).
		SetRetriesLeft(3).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationTask(row)
	return &item, nil
}

func (r *notificationRepo) ListTasks(ctx context.Context, namespaceUID string, filter bo.NotificationTaskListFilter, page, pageSize int) ([]bo.NotificationTask, int, error) {
	query := r.client.NotificationTask.Query().Where(notificationtask.NamespaceUIDEQ(namespaceUID))
	if filter.Status != "" {
		query = query.Where(notificationtask.StatusEQ(notificationtask.Status(filter.Status)))
	}
	if !filter.StartAt.IsZero() {
		query = query.Where(notificationtask.CreatedAtGTE(filter.StartAt))
	}
	if !filter.EndAt.IsZero() {
		query = query.Where(notificationtask.CreatedAtLT(filter.EndAt))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(notificationtask.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.NotificationTask, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNotificationTask(row))
	}
	if err := r.fillTaskNames(ctx, namespaceUID, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// fillTaskNames 给任务列表补出通知组、订阅、渠道和模板名称。
// 这些记录可能已经软删除，名称仍要显示，方便定位历史投递来源。
func (r *notificationRepo) fillTaskNames(ctx context.Context, namespaceUID string, items []bo.NotificationTask) error {
	if len(items) == 0 {
		return nil
	}
	groupUIDs := make([]string, 0, len(items))
	subscriptionUIDs := make([]string, 0, len(items))
	channelUIDs := make([]string, 0, len(items))
	templateUIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item.GroupUID != "" {
			groupUIDs = append(groupUIDs, item.GroupUID)
		}
		if item.SubscriptionUID != "" {
			subscriptionUIDs = append(subscriptionUIDs, item.SubscriptionUID)
		}
		if item.ChannelUID != "" {
			channelUIDs = append(channelUIDs, item.ChannelUID)
		}
		if item.TemplateUID != "" {
			templateUIDs = append(templateUIDs, item.TemplateUID)
		}
	}
	groupNames, err := func() (map[string]string, error) {
		rows, err := r.client.NotificationGroup.Query().Where(
			notificationgroup.NamespaceUIDEQ(namespaceUID),
			notificationgroup.UIDIn(uniqueStrings(groupUIDs)...),
		).All(ctx)
		if err != nil {
			return nil, err
		}
		out := make(map[string]string, len(rows))
		for _, row := range rows {
			out[row.UID] = row.Name
		}
		return out, nil
	}()
	if err != nil {
		return r.fail(err)
	}
	subscriptionNames, err := func() (map[string]string, error) {
		rows, err := r.client.NotificationSubscription.Query().Where(
			notificationsubscription.NamespaceUIDEQ(namespaceUID),
			notificationsubscription.UIDIn(uniqueStrings(subscriptionUIDs)...),
		).All(ctx)
		if err != nil {
			return nil, err
		}
		out := make(map[string]string, len(rows))
		for _, row := range rows {
			out[row.UID] = row.Name
		}
		return out, nil
	}()
	if err != nil {
		return r.fail(err)
	}
	channelNames, err := func() (map[string]string, error) {
		rows, err := r.client.NotificationChannel.Query().Where(
			notificationchannel.NamespaceUIDEQ(namespaceUID),
			notificationchannel.UIDIn(uniqueStrings(channelUIDs)...),
		).All(ctx)
		if err != nil {
			return nil, err
		}
		out := make(map[string]string, len(rows))
		for _, row := range rows {
			out[row.UID] = row.Name
		}
		return out, nil
	}()
	if err != nil {
		return r.fail(err)
	}
	templateNames, err := func() (map[string]string, error) {
		rows, err := r.client.NotificationTemplate.Query().Where(
			notificationtemplate.NamespaceUIDEQ(namespaceUID),
			notificationtemplate.UIDIn(uniqueStrings(templateUIDs)...),
		).All(ctx)
		if err != nil {
			return nil, err
		}
		out := make(map[string]string, len(rows))
		for _, row := range rows {
			out[row.UID] = row.Name
		}
		return out, nil
	}()
	if err != nil {
		return r.fail(err)
	}
	for i := range items {
		items[i].GroupName = groupNames[items[i].GroupUID]
		items[i].SubscriptionName = subscriptionNames[items[i].SubscriptionUID]
		items[i].ChannelName = channelNames[items[i].ChannelUID]
		items[i].TemplateName = templateNames[items[i].TemplateUID]
	}
	return nil
}

func (r *notificationRepo) RetryTask(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTask, error) {
	// 条件更新代替“先读后写”：merak 可能在同一时刻把任务改成 SENDING，
	// 只有仍然是 FAILED 的一行才允许重置。
	updated, err := r.client.NotificationTask.Update().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.UIDEQ(uid),
		notificationtask.StatusEQ(notificationtask.StatusFAILED),
	).
		SetStatus(notificationtask.StatusPENDING).
		SetRetriesLeft(3).
		SetLastError("").
		ClearNextAttemptAt().
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if updated == 0 {
		return nil, r.notifyTaskConflict(ctx, namespaceUID, uid, "only failed tasks can be retried")
	}
	row, err := r.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.UIDEQ(uid),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationTask(row)
	return &item, nil
}

func (r *notificationRepo) CancelTask(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTask, error) {
	// 只允许取消未开始投递的任务；状态条件放在 WHERE 里，避免与 ClaimPending 竞态。
	updated, err := r.client.NotificationTask.Update().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.UIDEQ(uid),
		notificationtask.StatusIn(notificationtask.StatusPENDING, notificationtask.StatusFAILED),
	).
		SetStatus(notificationtask.StatusCANCELLED).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if updated == 0 {
		return nil, r.notifyTaskConflict(ctx, namespaceUID, uid, "task cannot be cancelled")
	}
	row, err := r.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.UIDEQ(uid),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNotificationTask(row)
	return &item, nil
}

// notifyTaskConflict 在条件更新没有命中时给出稳定语义：不存在返回 404，状态不允许返回 400。
func (r *notificationRepo) notifyTaskConflict(ctx context.Context, namespaceUID, uid, reason string) error {
	exists, err := r.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.UIDEQ(uid),
	).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if !exists {
		return merr.NotFound("not found")
	}
	return merr.Invalid(reason)
}

func (r *notificationRepo) ListMessageLogs(ctx context.Context, namespaceUID string, filter bo.MessageLogListFilter, page, pageSize int) ([]bo.MessageLog, int, error) {
	// 多取到当前页末尾即可：分片库和主库兜底分别倒序返回，合并后截取当前页。
	need := page * pageSize
	sharded, shardedTotal, err := r.messageLogs.list(ctx, namespaceUID, filter, need)
	if err != nil {
		return nil, 0, err
	}
	fallback, fallbackTotal, err := r.listMainMessageLogs(ctx, namespaceUID, filter, need)
	if err != nil {
		return nil, 0, err
	}
	items := mergeMessageLogs(sharded, fallback, need)
	total := shardedTotal + fallbackTotal
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []bo.MessageLog{}, total, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	items = items[start:end]
	if err := r.fillLogSources(ctx, namespaceUID, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// listMainMessageLogs 读取尚未归档到租户库的主库行。
// 新日志提交后会立即归档并删除，这里只兜底归档失败和升级前的历史数据。
func (r *notificationRepo) listMainMessageLogs(ctx context.Context, namespaceUID string, filter bo.MessageLogListFilter, limit int) ([]bo.MessageLog, int, error) {
	query := r.client.MessageLog.Query().Where(messagelog.NamespaceUIDEQ(namespaceUID))
	if filter.TaskUID != "" {
		query = query.Where(messagelog.TaskUIDEQ(filter.TaskUID))
	}
	if filter.Status != "" {
		query = query.Where(messagelog.StatusEQ(messagelog.Status(filter.Status)))
	}
	if !filter.StartAt.IsZero() {
		query = query.Where(messagelog.CreatedAtGTE(filter.StartAt))
	}
	if !filter.EndAt.IsZero() {
		query = query.Where(messagelog.CreatedAtLT(filter.EndAt))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	if total == 0 {
		return nil, 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	rows, err := query.
		Order(ent.Desc(messagelog.FieldCreatedAt), ent.Desc(messagelog.FieldUID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.MessageLog, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMessageLog(row))
	}
	return items, total, nil
}

// fillLogSources 沿着消息日志的任务标识补出通知组和订阅名称。
// 用户看到一封邮件时可以直接判断它来自哪个订阅，而不是只看渠道和接收方。
func (r *notificationRepo) fillLogSources(ctx context.Context, namespaceUID string, items []bo.MessageLog) error {
	if len(items) == 0 {
		return nil
	}
	taskUIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item.TaskUID != "" {
			taskUIDs = append(taskUIDs, item.TaskUID)
		}
	}
	taskUIDs = uniqueStrings(taskUIDs)
	if len(taskUIDs) == 0 {
		return nil
	}
	tasks, err := r.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.UIDIn(taskUIDs...),
	).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	byTask := make(map[string]*ent.NotificationTask, len(tasks))
	groupUIDs := make([]string, 0, len(tasks))
	subscriptionUIDs := make([]string, 0, len(tasks))
	for _, task := range tasks {
		byTask[task.UID] = task
		if task.GroupUID != "" {
			groupUIDs = append(groupUIDs, task.GroupUID)
		}
		if task.SubscriptionUID != "" {
			subscriptionUIDs = append(subscriptionUIDs, task.SubscriptionUID)
		}
	}
	groupNames := map[string]string{}
	if values := uniqueStrings(groupUIDs); len(values) > 0 {
		rows, err := r.client.NotificationGroup.Query().Where(
			notificationgroup.NamespaceUIDEQ(namespaceUID),
			notificationgroup.UIDIn(values...),
		).All(ctx)
		if err != nil {
			return r.fail(err)
		}
		for _, row := range rows {
			groupNames[row.UID] = row.Name
		}
	}
	subscriptionNames := map[string]string{}
	if values := uniqueStrings(subscriptionUIDs); len(values) > 0 {
		rows, err := r.client.NotificationSubscription.Query().Where(
			notificationsubscription.NamespaceUIDEQ(namespaceUID),
			notificationsubscription.UIDIn(values...),
		).All(ctx)
		if err != nil {
			return r.fail(err)
		}
		for _, row := range rows {
			subscriptionNames[row.UID] = row.Name
		}
	}
	for i := range items {
		task := byTask[items[i].TaskUID]
		if task == nil {
			continue
		}
		items[i].GroupName = groupNames[task.GroupUID]
		items[i].SubscriptionName = subscriptionNames[task.SubscriptionUID]
	}
	return nil
}

func (r *notificationRepo) GetMessageLog(ctx context.Context, namespaceUID, uid string) (*bo.MessageLog, error) {
	// 归档后的日志按索引直接定位周表；索引缺失时回落到主库兜底行。
	if r.messageLogs != nil {
		index, err := r.client.MessageLogShard.Query().Where(
			messagelogshard.NamespaceUIDEQ(namespaceUID),
			messagelogshard.MessageUIDEQ(uid),
		).Only(ctx)
		if err == nil {
			item, err := r.messageLogs.get(ctx, namespaceUID, index.DatabaseName, index.TableName, uid)
			if err != nil {
				return nil, err
			}
			items := []bo.MessageLog{*item}
			if err := r.fillLogSources(ctx, namespaceUID, items); err != nil {
				return nil, err
			}
			return &items[0], nil
		}
		if !ent.IsNotFound(err) {
			return nil, r.fail(err)
		}
	}
	row, err := r.client.MessageLog.Query().Where(messagelog.NamespaceUIDEQ(namespaceUID), messagelog.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toMessageLog(row)
	return &item, nil
}

// archiveMessageLog 把主库里的消息日志搬到租户库周表。
// 顺序是写入周表、记录分片索引、删除主库行：任一步失败都保留主库行，
// 列表继续从主库兜底，调度器的回填任务会重试。
func (r *notificationRepo) archiveMessageLog(ctx context.Context, item bo.MessageLog) {
	if r.messageLogs == nil {
		return
	}
	database, table, week, err := r.messageLogs.archive(ctx, item)
	if err != nil {
		r.log.Error("archive message log " + item.UID + ": " + err.Error())
		return
	}
	if err := recordMessageLogShard(ctx, r.client, item, database, table, week); err != nil {
		r.log.Error("record message log shard " + item.UID + ": " + err.Error())
		return
	}
	if _, err := r.client.MessageLog.Delete().
		Where(messagelog.UIDEQ(item.UID), messagelog.NamespaceUIDEQ(item.NamespaceUID)).
		Exec(ctx); err != nil {
		r.log.Error("delete archived message log " + item.UID + ": " + err.Error())
	}
}

// recordMessageLogShard 记录日志所在租户库和周表，详情按 UID 直接定位，不用扫描所有周表。
func recordMessageLogShard(ctx context.Context, client *ent.Client, item bo.MessageLog, database, table string, week time.Time) error {
	existing, err := client.MessageLogShard.Query().Where(messagelogshard.MessageUIDEQ(item.UID)).Only(ctx)
	if err == nil {
		_, err = client.MessageLogShard.Update().Where(messagelogshard.IDEQ(existing.ID)).
			SetNamespaceUID(item.NamespaceUID).
			SetDatabaseName(database).
			SetTableName(table).
			SetWeekStart(week).
			Save(ctx)
		return err
	}
	if !ent.IsNotFound(err) {
		return err
	}
	_, err = client.MessageLogShard.Create().
		SetUID(id.New()).
		SetNamespaceUID(item.NamespaceUID).
		SetMessageUID(item.UID).
		SetDatabaseName(database).
		SetTableName(table).
		SetWeekStart(week).
		Save(ctx)
	return err
}

// BackfillMessageLogs 把主库里遗留的消息日志归档到租户分片库。
// 新日志在投递回写后立即归档，这里是升级旧数据和归档失败重试的兜底。
func (r *notificationRepo) BackfillMessageLogs(ctx context.Context, limit int) (int, error) {
	if r.messageLogs == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.client.MessageLog.Query().
		Order(ent.Asc(messagelog.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	items := make([]bo.MessageLog, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMessageLog(row))
	}
	for _, item := range items {
		r.archiveMessageLog(ctx, item)
	}
	return len(items), nil
}

func (r *notificationRepo) ClaimPending(ctx context.Context, limit int) ([]bo.NotifySnapshot, error) {
	now := time.Now().UTC()
	rows, err := r.client.NotificationTask.Query().Where(
		notificationtask.StatusEQ(notificationtask.StatusPENDING),
		notificationtask.Or(
			notificationtask.NextAttemptAtIsNil(),
			notificationtask.NextAttemptAtLTE(now),
		),
	).Order(ent.Asc(notificationtask.FieldCreatedAt)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	out := make([]bo.NotifySnapshot, 0, len(rows))
	for _, row := range rows {
		snap, err := r.snapshot(ctx, row)
		if err != nil {
			// 渠道或模板被删除时，坏任务不能卡住整个队列：标记 FAILED 后继续领取后面的任务。
			if _, updateErr := r.client.NotificationTask.Update().Where(
				notificationtask.UIDEQ(row.UID),
				notificationtask.StatusEQ(notificationtask.StatusPENDING),
			).SetStatus(notificationtask.StatusFAILED).SetLastError(err.Error()).Save(ctx); updateErr != nil {
				r.log.Error(fmt.Sprintf("drop invalid task %s: %v", row.UID, updateErr))
			}
			continue
		}
		if row.Sender == "" && snap.Sender != "" {
			// 历史任务没有发送方快照：首次领取时保存当前渠道身份，后续日志仍有可靠值。
			if _, err := r.client.NotificationTask.Update().Where(
				notificationtask.UIDEQ(row.UID), notificationtask.SenderEQ(""),
			).SetSender(snap.Sender).Save(ctx); err != nil {
				return nil, r.fail(err)
			}
		}
		out = append(out, snap)
	}
	return out, nil
}

// retryDelay 按剩余重试次数给出指数退避，避免失败任务每 2 秒被重新领取。
func retryDelay(retries int) time.Duration {
	switch retries {
	case 2:
		return 10 * time.Second
	case 1:
		return 30 * time.Second
	default:
		return time.Minute
	}
}

func (r *notificationRepo) MarkSending(ctx context.Context, taskUID string) error {
	n, err := r.client.NotificationTask.Update().Where(
		notificationtask.UIDEQ(taskUID),
		notificationtask.StatusEQ(notificationtask.StatusPENDING),
	).SetStatus(notificationtask.StatusSENDING).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.Invalid("task is not pending")
	}
	return nil
}

func (r *notificationRepo) ApplyNotifyReport(ctx context.Context, report bo.NotifyReport) error {
	// 分区由 partition.Maintainer 定时预创建，写路径只做条件更新与日志插入。
	row, err := r.client.NotificationTask.Query().Where(notificationtask.UIDEQ(report.TaskUID)).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	if row.Status != notificationtask.StatusSENDING {
		// 只有 SENDING 接受回写。重复或迟到的报告不能把 SENT 退回 FAILED，也不能重复写消息日志。
		return nil
	}
	logStatus := messagelog.StatusFAILED
	taskStatus := notificationtask.StatusFAILED
	retries := row.RetriesLeft
	lastError := report.Detail
	if report.OK {
		logStatus = messagelog.StatusSENT
		taskStatus = notificationtask.StatusSENT
		lastError = ""
	} else if retries > 0 {
		retries--
		if retries > 0 {
			taskStatus = notificationtask.StatusPENDING
		}
	}
	// 新版 merak 回传已渲染内容；旧版/历史报告不带这些字段时保留任务快照。
	subject, content := row.Subject, row.Content
	if content == "" {
		content = row.Body
	}
	if report.Subject != "" {
		subject = report.Subject
	}
	if report.Content != "" {
		content = report.Content
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	update := tx.NotificationTask.Update().
		Where(
			notificationtask.UIDEQ(row.UID),
			notificationtask.StatusEQ(notificationtask.StatusSENDING),
			// retries_left 一起做乐观校验：reaper 与迟到回写并发时只有一方能成功。
			notificationtask.RetriesLeftEQ(row.RetriesLeft),
		).
		SetStatus(taskStatus).
		SetRetriesLeft(retries).
		SetSubject(subject).
		SetContent(content).
		SetLastError(lastError)
	if taskStatus == notificationtask.StatusPENDING {
		update.SetNextAttemptAt(time.Now().UTC().Add(retryDelay(retries)))
	} else {
		update.ClearNextAttemptAt()
	}
	affected, err := update.Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if affected == 0 {
		// 状态已被其他写入者推进（重复回写或 reaper 先到）：不写消息日志，整体放弃本次回写。
		return nil
	}
	logUID := id.New()
	logAt := time.Now().UTC()
	if err := tx.MessageLog.Create().
		SetUID(logUID).
		SetNamespaceUID(row.NamespaceUID).
		SetTaskUID(row.UID).
		SetStatus(logStatus).
		SetChannelUID(row.ChannelUID).
		SetTemplateUID(row.TemplateUID).
		SetSender(row.Sender).
		SetReceivers(row.Receivers).
		SetSubject(subject).
		SetContent(content).
		SetDetail(report.Detail).
		SetCreatedAt(logAt).
		Exec(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	// 事务提交后再归档，主库保留兜底行；归档失败由调度器回填重试。
	r.archiveMessageLog(ctx, bo.MessageLog{
		UID: logUID, NamespaceUID: row.NamespaceUID, TaskUID: row.UID, Status: string(logStatus),
		ChannelUID: row.ChannelUID, TemplateUID: row.TemplateUID, Sender: row.Sender,
		Receivers: row.Receivers, Subject: subject, Content: content, Detail: report.Detail,
		CreatedAt: logAt,
	})
	return nil
}

// ReapStaleSending 把长时间停留在 SENDING 的任务标记为失败。
// merak 在改为 SENDING 后崩溃时不会回写，锁过期也不会自动释放任务。
func (r *notificationRepo) ReapStaleSending(ctx context.Context, before time.Time, limit int) (int, error) {
	rows, err := r.client.NotificationTask.Query().Where(
		notificationtask.StatusEQ(notificationtask.StatusSENDING),
		notificationtask.UpdatedAtLTE(before),
	).Order(ent.Asc(notificationtask.FieldCreatedAt)).Limit(limit).All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	reaped := 0
	for _, row := range rows {
		tx, err := r.client.Tx(ctx)
		if err != nil {
			return reaped, r.fail(err)
		}
		affected, err := tx.NotificationTask.Update().
			Where(
				notificationtask.UIDEQ(row.UID),
				notificationtask.StatusEQ(notificationtask.StatusSENDING),
				notificationtask.UpdatedAtLTE(before),
			).
			SetStatus(notificationtask.StatusFAILED).
			SetLastError("delivery timed out").
			ClearNextAttemptAt().
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return reaped, r.fail(err)
		}
		if affected == 0 {
			// 回写已经先到，任务不再停留在 SENDING：跳过日志，避免重复记录。
			_ = tx.Rollback()
			continue
		}
		content := row.Content
		if content == "" {
			content = row.Body
		}
		logUID := id.New()
		logAt := time.Now().UTC()
		if err := tx.MessageLog.Create().
			SetUID(logUID).
			SetNamespaceUID(row.NamespaceUID).
			SetTaskUID(row.UID).
			SetStatus(messagelog.StatusFAILED).
			SetChannelUID(row.ChannelUID).
			SetTemplateUID(row.TemplateUID).
			SetSender(row.Sender).
			SetReceivers(row.Receivers).
			SetSubject(row.Subject).
			SetContent(content).
			SetDetail("delivery timed out").
			SetCreatedAt(logAt).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return reaped, r.fail(err)
		}
		if err := tx.Commit(); err != nil {
			return reaped, r.fail(err)
		}
		r.archiveMessageLog(ctx, bo.MessageLog{
			UID: logUID, NamespaceUID: row.NamespaceUID, TaskUID: row.UID, Status: string(messagelog.StatusFAILED),
			ChannelUID: row.ChannelUID, TemplateUID: row.TemplateUID, Sender: row.Sender,
			Receivers: row.Receivers, Subject: row.Subject, Content: content,
			Detail: "delivery timed out", CreatedAt: logAt,
		})
		reaped++
	}
	return reaped, nil
}

func (r *notificationRepo) snapshot(ctx context.Context, row *ent.NotificationTask) (bo.NotifySnapshot, error) {
	channel, err := r.client.NotificationChannel.Query().Where(
		notificationchannel.UIDEQ(row.ChannelUID),
		notificationchannel.NamespaceUIDEQ(row.NamespaceUID),
		notificationchannel.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return bo.NotifySnapshot{}, r.fail(err)
	}
	if channel.Status != notificationchannel.StatusACTIVE {
		return bo.NotifySnapshot{}, merr.Invalid("channel is disabled")
	}
	sender := row.Sender
	if sender == "" {
		sender = notificationSender(channel)
	}
	password, err := r.box.decrypt(channel.PasswordCipher)
	if err != nil {
		return bo.NotifySnapshot{}, err
	}
	secret, err := r.box.decrypt(channel.SecretCipher)
	if err != nil {
		return bo.NotifySnapshot{}, err
	}
	subject := row.Subject
	body := row.Body
	templateType := ""
	templateUID := row.TemplateUID
	if templateUID != "" {
		template, err := r.client.NotificationTemplate.Query().Where(
			notificationtemplate.UIDEQ(templateUID),
			notificationtemplate.NamespaceUIDEQ(row.NamespaceUID),
			notificationtemplate.DeletedAtIsNil(),
		).Only(ctx)
		if err != nil {
			return bo.NotifySnapshot{}, r.fail(err)
		}
		if template.Status != notificationtemplate.StatusACTIVE {
			return bo.NotifySnapshot{}, merr.Invalid("template is disabled")
		}
		templateType = string(template.Type)
		if subject == "" {
			subject = template.Subject
		}
		if body == "" {
			body = template.Body
		}
		return bo.NotifySnapshot{
			UID: row.UID, NamespaceUID: row.NamespaceUID, Phase: string(row.Phase),
			ChannelUID: channel.UID, ChannelType: string(channel.Type), ChannelApp: string(channel.App),
			ChannelEndpoint: channel.Endpoint, ChannelUsername: channel.Username,
			ChannelPassword: password, ChannelSecret: secret, ChannelFrom: channel.FromAddress, Sender: sender,
			TemplateUID: template.UID, TemplateType: templateType, TemplateSubject: template.Subject, TemplateBody: template.Body,
			Receivers: row.Receivers, Params: row.Params, Subject: subject, Body: body,
		}, nil
	}
	return bo.NotifySnapshot{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Phase: string(row.Phase),
		ChannelUID: channel.UID, ChannelType: string(channel.Type), ChannelApp: string(channel.App),
		ChannelEndpoint: channel.Endpoint, ChannelUsername: channel.Username,
		ChannelPassword: password, ChannelSecret: secret, ChannelFrom: channel.FromAddress, Sender: sender,
		TemplateUID: templateUID, TemplateType: templateType, TemplateSubject: subject, TemplateBody: body,
		Receivers: row.Receivers, Params: row.Params, Subject: subject, Body: body,
	}, nil
}

func toNotificationChannel(row *ent.NotificationChannel) bo.NotificationChannel {
	return bo.NotificationChannel{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name,
		Type: string(row.Type), App: string(row.App), Status: string(row.Status),
		Endpoint: row.Endpoint, Username: row.Username, FromAddress: row.FromAddress, Remark: row.Remark,
		HasPassword: row.PasswordCipher != "", HasSecret: row.SecretCipher != "",
	}
}

func toNotificationTemplate(row *ent.NotificationTemplate) bo.NotificationTemplate {
	return bo.NotificationTemplate{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name,
		Type: string(row.Type), App: string(row.App), Status: string(row.Status), Subject: row.Subject, Body: row.Body, Remark: row.Remark,
	}
}

func toNotificationGroup(row *ent.NotificationGroup) bo.NotificationGroup {
	channelUIDs := row.ChannelUids
	templateUIDs := row.TemplateUids
	emailReceivers := row.EmailReceivers
	emailMembers := row.EmailMemberUids
	// 历史数据只有旧版单渠道字段时，读出来直接映射成新结构，避免升级后通知组不可编辑。
	if len(channelUIDs) == 0 && row.ChannelUID != "" {
		channelUIDs = append([]string{row.ChannelUID}, row.WebhookChannelUids...)
		templateUIDs = append([]string{row.TemplateUID}, repeatedString(row.WebhookTemplateUID, len(row.WebhookChannelUids))...)
		emailReceivers = row.Receivers
		emailMembers = row.MemberUids
	}
	return bo.NotificationGroup{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Status: string(row.Status),
		ChannelUIDs: channelUIDs, TemplateUIDs: templateUIDs,
		EmailReceivers: emailReceivers, SMSReceivers: row.SmsReceivers,
		EmailMemberUIDs: emailMembers, SMSMemberUIDs: row.SmsMemberUids,
		Remark:     row.Remark,
		ChannelUID: row.ChannelUID, TemplateUID: row.TemplateUID, Receivers: row.Receivers,
		MemberUIDs: row.MemberUids, WebhookChannelUIDs: row.WebhookChannelUids,
		WebhookTemplateUID: row.WebhookTemplateUID,
	}
}

func repeatedString(value string, count int) []string {
	if value == "" || count <= 0 {
		return nil
	}
	out := make([]string, count)
	for i := range out {
		out[i] = value
	}
	return out
}

func toNotificationTask(row *ent.NotificationTask) bo.NotificationTask {
	return bo.NotificationTask{
		UID: row.UID, NamespaceUID: row.NamespaceUID, EventUID: row.EventUID,
		Phase: string(row.Phase), Status: string(row.Status), RuleUID: row.RuleUID,
		Fingerprint:     row.Fingerprint,
		SubscriptionUID: row.SubscriptionUID, GroupUID: row.GroupUID,
		ChannelUID: row.ChannelUID, TemplateUID: row.TemplateUID, Sender: row.Sender, Receivers: row.Receivers,
		Subject: row.Subject, Body: row.Body, Content: row.Content, Params: row.Params, RetriesLeft: row.RetriesLeft, LastError: row.LastError,
	}
}

func toMessageLog(row *ent.MessageLog) bo.MessageLog {
	return bo.MessageLog{
		UID: row.UID, NamespaceUID: row.NamespaceUID, TaskUID: row.TaskUID, Status: string(row.Status),
		ChannelUID: row.ChannelUID, TemplateUID: row.TemplateUID, Sender: row.Sender,
		Receivers: row.Receivers, Subject: row.Subject, Content: row.Content,
		Detail: row.Detail, CreatedAt: row.CreatedAt,
	}
}

func channelEnums(kind, app, status string) (notificationchannel.Type, notificationchannel.App, notificationchannel.Status, error) {
	if kind != "SMTP" && kind != "WEBHOOK" && kind != "SMS" {
		return "", "", "", merr.Invalid("invalid request")
	}
	switch app {
	case "OTHER", "DINGTALK", "WECHAT", "FEISHU", "SMS_ALICLOUD":
	default:
		return "", "", "", merr.Invalid("invalid request")
	}
	st := notificationchannel.StatusACTIVE
	if status == "DISABLED" {
		st = notificationchannel.StatusDISABLED
	} else if status != "ACTIVE" {
		return "", "", "", merr.Invalid("invalid request")
	}
	return notificationchannel.Type(kind), notificationchannel.App(app), st, nil
}

func templateApp(app string) (notificationtemplate.App, error) {
	switch app {
	case "OTHER", "DINGTALK", "WECHAT", "FEISHU", "SMS_ALICLOUD":
		return notificationtemplate.App(app), nil
	default:
		return "", merr.Invalid("invalid request")
	}
}

func templateEnums(kind, status string) (notificationtemplate.Type, notificationtemplate.Status, error) {
	switch kind {
	case "EMAIL", "WEBHOOK", "SMS_ALICLOUD":
	default:
		return "", "", merr.Invalid("invalid request")
	}
	st := notificationtemplate.StatusACTIVE
	if status == "DISABLED" {
		st = notificationtemplate.StatusDISABLED
	} else if status != "ACTIVE" {
		return "", "", merr.Invalid("invalid request")
	}
	return notificationtemplate.Type(kind), st, nil
}

// paramsJSON 把告警字段打成模板参数。字段与告警摘要、描述模板共用 pkg/alerttmpl，
// 所以 $labels、$value、.Labels、.Value、{{ . }} 在通知模板里同样可用。
func paramsJSON(event notifyEvent) string {
	labels := map[string]string{}
	if event.Labels != "" {
		_ = json.Unmarshal([]byte(event.Labels), &labels)
	}
	value, _ := strconv.ParseFloat(strings.TrimSpace(event.Value), 64)
	data := alerttmpl.NewData(labels, event.Labels, value, event.ExternalURL, map[string]any{
		"eventUid": event.EventUID, "phase": event.Phase, "summary": event.Summary,
		"description": event.Description, "fingerprint": event.Fingerprint,
	})
	raw, _ := json.Marshal(data)
	return string(raw)
}
