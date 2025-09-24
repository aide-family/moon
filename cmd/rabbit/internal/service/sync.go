package service

import (
	"context"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/do"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/rabbit/internal/service/build"
	"github.com/aide-family/moon/pkg/api/rabbit/common"
	apiv1 "github.com/aide-family/moon/pkg/api/rabbit/v1"
	"github.com/aide-family/moon/pkg/middler/permission"
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
	teamID := permission.GetTeamIDByContextWithZeroValue(ctx)
	smss := build.ToSMSConfigs(req.GetSmss())
	params := &bo.SetSMSConfigParams{
		TeamID:  teamID,
		Configs: smss,
	}
	if err := s.configBiz.SetSMSConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) Email(ctx context.Context, req *apiv1.SyncEmailRequest) (*common.EmptyReply, error) {
	teamID := permission.GetTeamIDByContextWithZeroValue(ctx)
	emails := slices.Map(req.GetEmails(), func(emailItem *common.EmailConfig) bo.EmailConfig {
		return emailItem
	})
	params := &bo.SetEmailConfigParams{
		TeamID:  teamID,
		Configs: emails,
	}
	if err := s.configBiz.SetEmailConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) Hook(ctx context.Context, req *apiv1.SyncHookRequest) (*common.EmptyReply, error) {
	teamID := permission.GetTeamIDByContextWithZeroValue(ctx)
	hooks := slices.Map(req.GetHooks(), func(hookItem *common.HookConfig) bo.HookConfig {
		return &do.HookConfig{
			Name:     hookItem.GetName(),
			App:      vobj.APP(hookItem.GetApp()),
			URL:      hookItem.GetUrl(),
			Secret:   hookItem.GetSecret(),
			Token:    hookItem.GetToken(),
			Username: hookItem.GetUsername(),
			Password: hookItem.GetPassword(),
			Headers:  hookItem.GetHeaders(),
			Enable:   hookItem.GetEnable(),
		}
	})
	params := &bo.SetHookConfigParams{
		TeamID:  teamID,
		Configs: hooks,
	}
	if err := s.configBiz.SetHookConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) NoticeGroup(ctx context.Context, req *apiv1.SyncNoticeGroupRequest) (*common.EmptyReply, error) {
	teamID := permission.GetTeamIDByContextWithZeroValue(ctx)
	noticeGroups := slices.Map(req.GetNoticeGroups(), func(noticeGroupItem *common.NoticeGroupConfig) bo.NoticeGroup {
		templates := slices.Map(noticeGroupItem.GetTemplates(), func(templateItem *common.NoticeGroupConfig_Template) bo.Template {
			return &do.Template{
				Type:           vobj.APP(templateItem.GetType()),
				Template:       templateItem.GetTemplate(),
				TemplateParams: templateItem.GetTemplateParameters(),
				Subject:        templateItem.GetSubject(),
			}
		})
		return bo.NewNoticeGroup(
			bo.WithNoticeGroupOptionName(noticeGroupItem.GetName()),
			bo.WithNoticeGroupOptionSmsConfigName(noticeGroupItem.GetSmsConfigName()),
			bo.WithNoticeGroupOptionEmailConfigName(noticeGroupItem.GetEmailConfigName()),
			bo.WithNoticeGroupOptionHookReceivers(noticeGroupItem.GetHookReceivers()),
			bo.WithNoticeGroupOptionSmsReceivers(noticeGroupItem.GetSmsReceivers()),
			bo.WithNoticeGroupOptionEmailReceivers(noticeGroupItem.GetEmailReceivers()),
			bo.WithNoticeGroupOptionTemplates(templates),
		)
	})
	params := &bo.SetNoticeGroupConfigParams{
		TeamID:  teamID,
		Configs: noticeGroups,
	}
	if err := s.configBiz.SetNoticeGroupConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SyncService) Remove(ctx context.Context, req *apiv1.RemoveRequest) (*common.EmptyReply, error) {
	teamID := permission.GetTeamIDByContextWithZeroValue(ctx)
	params := &bo.RemoveConfigParams{
		TeamID: teamID,
		Name:   req.GetName(),
		Type:   req.GetType(),
	}
	if err := s.configBiz.RemoveConfig(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}
