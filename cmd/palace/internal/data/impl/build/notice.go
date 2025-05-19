package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToStrategyNotice(ctx context.Context, route do.NoticeGroup) *team.NoticeGroup {
	if validate.IsNil(route) {
		return nil
	}
	if notice, ok := route.(*team.NoticeGroup); ok {
		notice.WithContext(ctx)
		return notice
	}
	group := &team.NoticeGroup{
		Name:          route.GetName(),
		Remark:        route.GetRemark(),
		Status:        route.GetStatus(),
		Members:       ToStrategyMembers(ctx, route.GetNoticeMembers()),
		Hooks:         ToStrategyHooks(ctx, route.GetHooks()),
		EmailConfigID: route.GetEmailConfig().GetID(),
		EmailConfig:   ToStrategyEmailConfig(ctx, route.GetEmailConfig()),
		SMSConfigID:   route.GetSMSConfig().GetID(),
		SMSConfig:     ToStrategySmsConfig(ctx, route.GetSMSConfig()),
		TeamModel:     ToTeamModel(ctx, route),
	}
	group.WithContext(ctx)
	return group
}

func ToStrategyNotices(ctx context.Context, routes []do.NoticeGroup) []*team.NoticeGroup {
	if len(routes) == 0 {
		return nil
	}
	return slices.MapFilter(routes, func(route do.NoticeGroup) (*team.NoticeGroup, bool) {
		if validate.IsNil(route) {
			return nil, false
		}
		return ToStrategyNotice(ctx, route), true
	})
}

func ToStrategyHook(ctx context.Context, hook do.NoticeHook) *team.NoticeHook {
	if validate.IsNil(hook) {
		return nil
	}
	hookDo, ok := hook.(*team.NoticeHook)
	if ok {
		hookDo.WithContext(ctx)
		return hookDo
	}
	hookDo = &team.NoticeHook{
		TeamModel:    ToTeamModel(ctx, hook),
		Name:         hook.GetName(),
		Remark:       hook.GetRemark(),
		Status:       hook.GetStatus(),
		URL:          hook.GetURL(),
		Method:       hook.GetMethod(),
		Secret:       crypto.String(hook.GetSecret()),
		Headers:      crypto.NewObject(hook.GetHeaders()),
		NoticeGroups: ToStrategyNotices(ctx, hook.GetNoticeGroups()),
		APP:          hook.GetApp(),
	}
	hookDo.WithContext(ctx)
	return hookDo
}

func ToStrategyHooks(ctx context.Context, hooks []do.NoticeHook) []*team.NoticeHook {
	return slices.MapFilter(hooks, func(hook do.NoticeHook) (*team.NoticeHook, bool) {
		if validate.IsNil(hook) {
			return nil, false
		}
		return ToStrategyHook(ctx, hook), true
	})
}

func ToStrategyEmailConfig(ctx context.Context, config do.TeamEmailConfig) *team.EmailConfig {
	if validate.IsNil(config) {
		return nil
	}
	if config, ok := config.(*team.EmailConfig); ok {
		config.WithContext(ctx)
		return config
	}

	emailConfig := &team.EmailConfig{
		TeamModel: ToTeamModel(ctx, config),
		Name:      config.GetName(),
		Remark:    config.GetRemark(),
		Status:    config.GetStatus(),
		Email:     crypto.NewObject(config.GetEmailConfig()),
	}
	emailConfig.WithContext(ctx)
	return emailConfig
}

func ToStrategySmsConfig(ctx context.Context, config do.TeamSMSConfig) *team.SmsConfig {
	if validate.IsNil(config) {
		return nil
	}
	if config, ok := config.(*team.SmsConfig); ok {
		config.WithContext(ctx)
		return config
	}
	smsConfig := &team.SmsConfig{
		TeamModel: ToTeamModel(ctx, config),
		Name:      config.GetName(),
		Remark:    config.GetRemark(),
		Status:    config.GetStatus(),
		Sms:       crypto.NewObject(config.GetSMSConfig()),
		Provider:  config.GetProviderType(),
	}
	smsConfig.WithContext(ctx)
	return smsConfig
}

func ToStrategyMetricRuleLabelNotice(ctx context.Context, notice do.StrategyMetricRuleLabelNotice) *team.StrategyMetricRuleLabelNotice {
	if validate.IsNil(notice) {
		return nil
	}
	if notice, ok := notice.(*team.StrategyMetricRuleLabelNotice); ok {
		notice.WithContext(ctx)
		return notice
	}
	noticeDo := &team.StrategyMetricRuleLabelNotice{
		TeamModel:            ToTeamModel(ctx, notice),
		StrategyMetricRuleID: notice.GetStrategyMetricRuleID(),
		LabelKey:             notice.GetLabelKey(),
		LabelValue:           notice.GetLabelValue(),
		Notices:              ToStrategyNotices(ctx, notice.GetNotices()),
	}
	noticeDo.WithContext(ctx)
	return noticeDo
}

func ToStrategyMetricRuleLabelNotices(ctx context.Context, notices []do.StrategyMetricRuleLabelNotice) []*team.StrategyMetricRuleLabelNotice {
	return slices.MapFilter(notices, func(notice do.StrategyMetricRuleLabelNotice) (*team.StrategyMetricRuleLabelNotice, bool) {
		if validate.IsNil(notice) {
			return nil, false
		}
		return ToStrategyMetricRuleLabelNotice(ctx, notice), true
	})
}
