package repository

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// NotificationRepo 读写渠道、模板、通知组、任务和消息日志。
type NotificationRepo interface {
	CreateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error)
	UpdateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error)
	UpdateChannelStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationChannel, error)
	DeleteChannel(ctx context.Context, namespaceUID, uid string) error
	GetChannel(ctx context.Context, namespaceUID, uid string) (*bo.NotificationChannel, error)
	ListChannels(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationChannel, int, error)
	SelectChannels(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]bo.SelectItem, error)

	CreateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error)
	UpdateTemplateStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationTemplate, error)
	DeleteTemplate(ctx context.Context, namespaceUID, uid string) error
	GetTemplate(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTemplate, error)
	ListTemplates(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationTemplate, int, error)
	SelectTemplates(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]bo.SelectItem, error)
	ListPresets(ctx context.Context) ([]bo.NotificationPreset, error)

	CreateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error)
	UpdateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error)
	UpdateGroupStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationGroup, error)
	DeleteGroup(ctx context.Context, namespaceUID, uid string) error
	GetGroup(ctx context.Context, namespaceUID, uid string) (*bo.NotificationGroup, error)
	ListGroups(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationGroup, int, error)
	SelectGroups(ctx context.Context, namespaceUID string, limit int, lastUID string, activeOnly bool) ([]bo.SelectItem, error)

	CreateTestTask(ctx context.Context, in bo.NotificationTask) (*bo.NotificationTask, error)
	ListTasks(ctx context.Context, namespaceUID, status string, page, pageSize int) ([]bo.NotificationTask, int, error)
	RetryTask(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTask, error)
	CancelTask(ctx context.Context, namespaceUID, uid string) (*bo.NotificationTask, error)

	ListMessageLogs(ctx context.Context, namespaceUID, taskUID string, page, pageSize int) ([]bo.MessageLog, int, error)
	GetMessageLog(ctx context.Context, namespaceUID, uid string) (*bo.MessageLog, error)

	ClaimPending(ctx context.Context, limit int) ([]bo.NotifySnapshot, error)
	MarkSending(ctx context.Context, taskUID string) error
	ApplyNotifyReport(ctx context.Context, report bo.NotifyReport) error
	ReapStaleSending(ctx context.Context, before time.Time, limit int) (int, error)
}

// NotifySnapshot 是 merak 领取到的任务，含解密后的渠道密钥。
type NotifySnapshot = bo.NotifySnapshot
