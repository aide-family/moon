package service

import (
	"context"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/service/build"
	"github.com/aide-family/moon/pkg/api/rabbit/common"
	apiv1 "github.com/aide-family/moon/pkg/api/rabbit/v1"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/go-kratos/kratos/v2/log"
)

type SyncService struct {
	apiv1.UnimplementedSyncServer
	configBiz *biz.Config
	helper    *log.Helper
}

func NewSyncService(configBiz *biz.Config, logger log.Logger) *SyncService {
	return &SyncService{
		configBiz: configBiz,
		helper:    log.NewHelper(log.With(logger, "module", "service.sync")),
	}
}
func (s *SyncService) Sms(ctx context.Context, req *apiv1.SyncSmsRequest) (*common.EmptyReply, error) {
	smss := build.ToSMSConfigs(req.GetSmss())
	params := &bo.SetSMSConfigParams{
		TeamID:  req.GetTeamId(),
		Configs: smss,
	}
	if err := s.configBiz.SetSMSConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) Email(ctx context.Context, req *apiv1.SyncEmailRequest) (*common.EmptyReply, error) {
	emails := slices.Map(req.GetEmails(), func(emailItem *common.EmailConfig) bo.EmailConfig {
		return emailItem
	})
	params := &bo.SetEmailConfigParams{
		TeamID:  req.GetTeamId(),
		Configs: emails,
	}
	if err := s.configBiz.SetEmailConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) Hook(ctx context.Context, req *apiv1.SyncHookRequest) (*common.EmptyReply, error) {
	hooks := slices.Map(req.GetHooks(), func(hookItem *common.HookConfig) bo.HookConfig {
		return hookItem
	})
	params := &bo.SetHookConfigParams{
		TeamID:  req.GetTeamId(),
		Configs: hooks,
	}
	if err := s.configBiz.SetHookConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) NoticeGroup(ctx context.Context, req *apiv1.SyncNoticeGroupRequest) (*common.EmptyReply, error) {
	noticeGroups := slices.Map(req.GetNoticeGroups(), func(noticeGroupItem *common.NoticeGroup) bo.NoticeGroup {
		templates := slices.Map(noticeGroupItem.GetTemplates(), func(templateItem *common.NoticeGroup_Template) bo.Template {
			return templateItem
		})
		return bo.NewNoticeGroup(
			bo.WithNoticeGroupOptionName(noticeGroupItem.GetName()),
			bo.WithNoticeGroupOptionSmsConfigName(noticeGroupItem.GetSmsConfigName()),
			bo.WithNoticeGroupOptionEmailConfigName(noticeGroupItem.GetEmailConfigName()),
			bo.WithNoticeGroupOptionHookConfigNames(noticeGroupItem.GetHookConfigNames()),
			bo.WithNoticeGroupOptionSmsUserNames(noticeGroupItem.GetSmsUserNames()),
			bo.WithNoticeGroupOptionEmailUserNames(noticeGroupItem.GetEmailUserNames()),
			bo.WithNoticeGroupOptionTemplates(templates),
		)
	})
	params := &bo.SetNoticeGroupConfigParams{
		TeamID:  req.GetTeamId(),
		Configs: noticeGroups,
	}
	if err := s.configBiz.SetNoticeGroupConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) NoticeUser(ctx context.Context, req *apiv1.SyncNoticeUserRequest) (*common.EmptyReply, error) {
	noticeUsers := slices.Map(req.GetNoticeUsers(), func(noticeUserItem *common.NoticeUser) bo.NoticeUser {
		return noticeUserItem
	})
	params := &bo.SetNoticeUserConfigParams{
		TeamID:  req.GetTeamId(),
		Configs: noticeUsers,
	}
	if err := s.configBiz.SetNoticeUserConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) Remove(ctx context.Context, req *apiv1.RemoveRequest) (*common.EmptyReply, error) {
	params := &bo.RemoveConfigParams{
		TeamID: req.GetTeamId(),
		Name:   req.GetName(),
		Type:   req.GetType(),
	}
	if err := s.configBiz.RemoveConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}
