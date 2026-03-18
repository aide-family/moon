package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewNotificationGroup(repo repository.NotificationGroup, helper *klog.Helper) *NotificationGroupBiz {
	return &NotificationGroupBiz{
		repo:   repo,
		helper: klog.NewHelper(klog.With(helper.Logger(), "biz", "notification_group")),
	}
}

type NotificationGroupBiz struct {
	repo   repository.NotificationGroup
	helper *klog.Helper
}

func (b *NotificationGroupBiz) CreateNotificationGroup(ctx context.Context, req *bo.CreateNotificationGroupBo) (snowflake.ID, error) {
	taken, err := b.repo.NotificationGroupNameTaken(ctx, req.Name, 0)
	if err != nil {
		b.helper.Errorw("msg", "check notification group name taken failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("check name failed").WithCause(err)
	}
	if taken {
		return 0, merr.ErrorParams("notification group name already exists, please use another name")
	}
	uid, err := b.repo.CreateNotificationGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "create notification group failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("create notification group failed").WithCause(err)
	}
	return uid, nil
}

func (b *NotificationGroupBiz) UpdateNotificationGroup(ctx context.Context, req *bo.UpdateNotificationGroupBo) error {
	taken, err := b.repo.NotificationGroupNameTaken(ctx, req.Name, req.UID)
	if err != nil {
		b.helper.Errorw("msg", "check notification group name taken failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("check name failed").WithCause(err)
	}
	if taken {
		return merr.ErrorParams("notification group name already exists, please use another name")
	}
	if err := b.repo.UpdateNotificationGroup(ctx, req); err != nil {
		b.helper.Errorw("msg", "update notification group failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update notification group failed").WithCause(err)
	}
	return nil
}

func (b *NotificationGroupBiz) UpdateNotificationGroupStatus(ctx context.Context, req *bo.UpdateNotificationGroupStatusBo) error {
	if err := b.repo.UpdateNotificationGroupStatus(ctx, req); err != nil {
		b.helper.Errorw("msg", "update notification group status failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update notification group status failed").WithCause(err)
	}
	return nil
}

func (b *NotificationGroupBiz) DeleteNotificationGroup(ctx context.Context, uid snowflake.ID) error {
	if err := b.repo.DeleteNotificationGroup(ctx, uid); err != nil {
		b.helper.Errorw("msg", "delete notification group failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete notification group failed").WithCause(err)
	}
	return nil
}

func (b *NotificationGroupBiz) GetNotificationGroup(ctx context.Context, uid snowflake.ID) (*bo.NotificationGroupItemBo, error) {
	item, err := b.repo.GetNotificationGroup(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("notification group %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "get notification group failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get notification group failed").WithCause(err)
	}
	return item, nil
}

func (b *NotificationGroupBiz) ListNotificationGroup(ctx context.Context, req *bo.ListNotificationGroupBo) (*bo.PageResponseBo[*bo.NotificationGroupItemBo], error) {
	result, err := b.repo.ListNotificationGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "list notification group failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list notification group failed").WithCause(err)
	}
	return result, nil
}
