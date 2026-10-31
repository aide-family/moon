package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

// NotificationRepo 读写渠道、模板、通知组、任务和消息日志。
type NotificationRepo interface {
	CreateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error)
	UpdateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error)
	UpdateChannelStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationChannel, error)
	DeleteChannel(ctx context.Context, namespaceUID, uid string) error
	GetChannel(ctx context.Context, namespaceUID, uid string) (*bo.NotificationChannel, error)
	// ListChannels 的 scopes 是已授权的「渠道类型 + 应用」组合，列表只返回其中的渠道。
	ListChannels(ctx context.Context, namespaceUID string, page, pageSize int, name string, scopes []perm.ChannelScope) ([]bo.NotificationChannel, int, error)
	// ListChannelsByUIDs 按 UID 批量读取渠道，用于过滤通知组里已不再授权的渠道。
	ListChannelsByUIDs(ctx context.Context, namespaceUID string, uids []string) ([]bo.NotificationChannel, error)
	SelectChannels(ctx context.Context, namespaceUID string, limit int, lastUID string, scopes []perm.ChannelScope) ([]bo.SelectItem, error)

	CreateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error)
	UpdateTemplateStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationTemplate, error)
	DeleteTemplate(ctx context.Context, namespaceUID, uid string) error
	GetTemplate(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTemplate, error)
	// ListTemplates 的 scopes 是已授权的「模板类型 + 应用」组合。
	ListTemplates(ctx context.Context, namespaceUID string, page, pageSize int, name string, scopes []perm.TemplateScope) ([]bo.NotificationTemplate, int, error)
	// ListTemplatesByUIDs 按 UID 批量读取模板，用于过滤通知组里已不再授权的模板。
	ListTemplatesByUIDs(ctx context.Context, namespaceUID string, uids []string) ([]bo.NotificationTemplate, error)
	SelectTemplates(ctx context.Context, namespaceUID string, limit int, lastUID string, scopes []perm.TemplateScope) ([]bo.SelectItem, error)
	ListPresets(ctx context.Context) ([]bo.NotificationPreset, error)

	CreateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error)
	UpdateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error)
	UpdateGroupStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationGroup, error)
	DeleteGroup(ctx context.Context, namespaceUID, uid string) error
	GetGroup(ctx context.Context, namespaceUID, uid string) (*bo.NotificationGroup, error)
	ListGroups(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationGroup, int, error)
	SelectGroups(ctx context.Context, namespaceUID string, limit int, lastUID string, activeOnly bool) ([]bo.SelectItem, error)

	CreateSubscription(ctx context.Context, in bo.NotificationSubscription) (*bo.NotificationSubscription, error)
	UpdateSubscription(ctx context.Context, in bo.NotificationSubscription) (*bo.NotificationSubscription, error)
	UpdateSubscriptionStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationSubscription, error)
	DeleteSubscription(ctx context.Context, namespaceUID, uid string) error
	GetSubscription(ctx context.Context, namespaceUID, uid string) (*bo.NotificationSubscription, error)
	ListSubscriptions(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationSubscription, int, error)

	CreateOnCall(ctx context.Context, in bo.NotificationOnCall) (*bo.NotificationOnCall, error)
	UpdateOnCall(ctx context.Context, in bo.NotificationOnCall) (*bo.NotificationOnCall, error)
	UpdateOnCallStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationOnCall, error)
	DeleteOnCall(ctx context.Context, namespaceUID, uid string) error
	GetOnCall(ctx context.Context, namespaceUID, uid string) (*bo.NotificationOnCall, error)
	ListOnCalls(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationOnCall, int, error)

	CreateTestTask(ctx context.Context, in bo.NotificationTask) (*bo.NotificationTask, error)
	ListTasks(ctx context.Context, namespaceUID string, filter bo.NotificationTaskListFilter, page, pageSize int) ([]bo.NotificationTask, int, error)
	RetryTask(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTask, error)
	CancelTask(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTask, error)

	ListMessageLogs(ctx context.Context, namespaceUID string, filter bo.MessageLogListFilter, page, pageSize int) ([]bo.MessageLog, int, error)
	GetMessageLog(ctx context.Context, namespaceUID, uid string) (*bo.MessageLog, error)
	// BackfillMessageLogs 把主库里遗留的消息日志归档到租户分片库，兼容升级前的数据。
	BackfillMessageLogs(ctx context.Context, limit int) (int, error)

	ClaimPending(ctx context.Context, limit int) ([]bo.NotifySnapshot, error)
	MarkSending(ctx context.Context, taskUID string) error
	ApplyNotifyReport(ctx context.Context, report bo.NotifyReport) error
	ReapStaleSending(ctx context.Context, before time.Time, limit int) (int, error)
}

// NotifySnapshot 是 merak 领取到的任务，含解密后的渠道密钥。
type NotifySnapshot = bo.NotifySnapshot
