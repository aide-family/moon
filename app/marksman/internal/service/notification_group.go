package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

func NewNotificationGroupService(notificationGroupBiz *biz.NotificationGroupBiz) *NotificationGroupService {
	return &NotificationGroupService{
		notificationGroupBiz: notificationGroupBiz,
	}
}

type NotificationGroupService struct {
	apiv1.UnimplementedNotificationGroupServer

	notificationGroupBiz *biz.NotificationGroupBiz
}

func (s *NotificationGroupService) CreateNotificationGroup(ctx context.Context, req *apiv1.CreateNotificationGroupRequest) (*apiv1.CreateNotificationGroupReply, error) {
	createBo := bo.NewCreateNotificationGroupBo(req)
	uid, err := s.notificationGroupBiz.CreateNotificationGroup(ctx, createBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateNotificationGroupReply{Uid: uid.Int64()}, nil
}

func (s *NotificationGroupService) UpdateNotificationGroup(ctx context.Context, req *apiv1.UpdateNotificationGroupRequest) (*apiv1.UpdateNotificationGroupReply, error) {
	updateBo := bo.NewUpdateNotificationGroupBo(req)
	if err := s.notificationGroupBiz.UpdateNotificationGroup(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateNotificationGroupReply{}, nil
}

func (s *NotificationGroupService) UpdateNotificationGroupStatus(ctx context.Context, req *apiv1.UpdateNotificationGroupStatusRequest) (*apiv1.UpdateNotificationGroupStatusReply, error) {
	statusBo := bo.NewUpdateNotificationGroupStatusBo(req)
	if err := s.notificationGroupBiz.UpdateNotificationGroupStatus(ctx, statusBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateNotificationGroupStatusReply{}, nil
}

func (s *NotificationGroupService) DeleteNotificationGroup(ctx context.Context, req *apiv1.DeleteNotificationGroupRequest) (*apiv1.DeleteNotificationGroupReply, error) {
	if err := s.notificationGroupBiz.DeleteNotificationGroup(ctx, snowflake.ParseInt64(req.GetUid())); err != nil {
		return nil, err
	}
	return &apiv1.DeleteNotificationGroupReply{}, nil
}

func (s *NotificationGroupService) GetNotificationGroup(ctx context.Context, req *apiv1.GetNotificationGroupRequest) (*apiv1.NotificationGroupItem, error) {
	item, err := s.notificationGroupBiz.GetNotificationGroup(ctx, snowflake.ParseInt64(req.GetUid()))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1NotificationGroupItem(item), nil
}

func (s *NotificationGroupService) ListNotificationGroup(ctx context.Context, req *apiv1.ListNotificationGroupRequest) (*apiv1.ListNotificationGroupReply, error) {
	result, err := s.notificationGroupBiz.ListNotificationGroup(ctx, bo.NewListNotificationGroupBo(req))
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListNotificationGroupReply(result), nil
}
