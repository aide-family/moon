package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type NotificationGroup interface {
	CreateNotificationGroup(ctx context.Context, req *bo.CreateNotificationGroupBo) (snowflake.ID, error)
	NotificationGroupNameTaken(ctx context.Context, name string, excludeUID snowflake.ID) (bool, error)
	UpdateNotificationGroup(ctx context.Context, req *bo.UpdateNotificationGroupBo) error
	UpdateNotificationGroupStatus(ctx context.Context, req *bo.UpdateNotificationGroupStatusBo) error
	DeleteNotificationGroup(ctx context.Context, uid snowflake.ID) error
	GetNotificationGroup(ctx context.Context, uid snowflake.ID) (*bo.NotificationGroupItemBo, error)
	ListNotificationGroup(ctx context.Context, req *bo.ListNotificationGroupBo) (*bo.PageResponseBo[*bo.NotificationGroupItemBo], error)
}

type NotificationResourceResolver interface {
	GetWebhookName(ctx context.Context, uid int64) (string, error)
	GetTemplateName(ctx context.Context, uid int64) (string, error)
	GetEmailConfigName(ctx context.Context, uid int64) (string, error)
}
