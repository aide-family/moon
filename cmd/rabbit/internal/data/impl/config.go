package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/do"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/rabbit/internal/data"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewConfigRepo(d *data.Data, logger log.Logger) repository.Config {
	return &configImpl{
		helper: log.NewHelper(log.With(logger, "module", "data.repo.config")),
		Data:   d,
	}
}

type configImpl struct {
	helper *log.Helper
	*data.Data
}

func (c *configImpl) GetEmailConfig(ctx context.Context, teamID uint32, name string) (bo.EmailConfig, bool) {
	key := vobj.EmailCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().HExists(ctx, key, name).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetEmailConfig", "err", err)
		return nil, false
	}
	if !exist {
		return nil, false
	}
	var emailConfig do.EmailConfig
	if err := c.GetCache().Client().HGet(ctx, key, name).Scan(&emailConfig); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetEmailConfig", "err", err)
		return nil, false
	}

	return &emailConfig, true
}

func (c *configImpl) GetEmailConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.EmailConfig, error) {
	key := vobj.EmailCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetEmailConfig", "err", err)
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	all, err := c.GetCache().Client().HMGet(ctx, key, names...).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetEmailConfig", "err", err)
		return nil, err
	}

	emailConfigs := make([]*do.EmailConfig, 0, len(all))
	if err := slices.UnmarshalBinary(all, &emailConfigs); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetEmailConfig", "err", err)
		return nil, err
	}
	return slices.Map(emailConfigs, func(v *do.EmailConfig) bo.EmailConfig { return v }), nil
}

func (c *configImpl) SetEmailConfig(ctx context.Context, teamID uint32, configs ...bo.EmailConfig) error {
	configDos := make(map[string]any, len(configs))
	for _, v := range configs {
		item := &do.EmailConfig{
			User:   v.GetUser(),
			Pass:   v.GetPass(),
			Host:   v.GetHost(),
			Port:   v.GetPort(),
			Enable: v.GetEnable(),
			Name:   v.GetName(),
		}
		configDos[item.UniqueKey()] = item
	}

	return c.GetCache().Client().HSet(ctx, vobj.EmailCacheKey.Key(teamID), configDos).Err()
}

func (c *configImpl) GetSMSConfig(ctx context.Context, teamID uint32, name string) (bo.SMSConfig, bool) {
	key := vobj.SmsCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().HExists(ctx, key, name).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetSMSConfig", "err", err)
		return nil, false
	}
	if !exist {
		return nil, false
	}
	var smsConfig do.SMSConfig
	if err := c.GetCache().Client().HGet(ctx, key, name).Scan(&smsConfig); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetSMSConfig", "err", err)
		return nil, false
	}
	return &smsConfig, true
}

func (c *configImpl) GetSMSConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.SMSConfig, error) {
	key := vobj.SmsCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetSMSConfig", "err", err)
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	all, err := c.GetCache().Client().HMGet(ctx, key, names...).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetSMSConfig", "err", err)
		return nil, err
	}

	smsConfigs := make([]*do.SMSConfig, 0, len(all))
	if err := slices.UnmarshalBinary(all, &smsConfigs); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetSMSConfig", "err", err)
		return nil, err
	}
	return slices.Map(smsConfigs, func(v *do.SMSConfig) bo.SMSConfig { return v }), nil
}

func (c *configImpl) SetSMSConfig(ctx context.Context, teamID uint32, configs ...bo.SMSConfig) error {
	configDos := make(map[string]any, len(configs))
	for _, v := range configs {
		item := &do.SMSConfig{
			AccessKeyId:     v.GetAccessKeyId(),
			AccessKeySecret: v.GetAccessKeySecret(),
			Endpoint:        v.GetEndpoint(),
			Name:            v.GetName(),
			SignName:        v.GetSignName(),
			Type:            v.GetType(),
			Enable:          v.GetEnable(),
		}
		configDos[item.UniqueKey()] = item
	}
	return c.GetCache().Client().HSet(ctx, vobj.SmsCacheKey.Key(teamID), configDos).Err()
}

func (c *configImpl) GetHookConfig(ctx context.Context, teamID uint32, name string) (bo.HookConfig, bool) {
	key := vobj.HookCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetHookConfig", "err", err)
		return nil, false
	}
	if exist == 0 {
		return nil, false
	}
	var hookConfig do.HookConfig
	if err := c.GetCache().Client().HGet(ctx, key, name).Scan(&hookConfig); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetHookConfig", "err", err)
		return nil, false
	}
	return &hookConfig, true
}

func (c *configImpl) GetHookConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.HookConfig, error) {
	key := vobj.HookCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetHookConfig", "err", err)
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	all, err := c.GetCache().Client().HMGet(ctx, key, names...).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetHookConfig", "err", err)
		return nil, err
	}

	hookConfigs := make([]*do.HookConfig, 0, len(all))
	if err := slices.UnmarshalBinary(all, &hookConfigs); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetHookConfig", "err", err)
		return nil, err
	}
	return slices.Map(hookConfigs, func(v *do.HookConfig) bo.HookConfig { return v }), nil
}

func (c *configImpl) SetHookConfig(ctx context.Context, teamID uint32, configs ...bo.HookConfig) error {
	configDos := make(map[string]any, len(configs))
	for _, v := range configs {
		item := &do.HookConfig{
			Name:     v.GetName(),
			App:      v.GetApp(),
			Url:      v.GetUrl(),
			Secret:   v.GetSecret(),
			Token:    v.GetToken(),
			Username: v.GetUsername(),
			Password: v.GetPassword(),
			Headers:  v.GetHeaders(),
			Enable:   v.GetEnable(),
		}
		configDos[item.UniqueKey()] = item
	}
	return c.GetCache().Client().HSet(ctx, vobj.HookCacheKey.Key(teamID), configDos).Err()
}

func (c *configImpl) GetNoticeGroupConfig(ctx context.Context, teamID uint32, name string) (bo.NoticeGroup, bool) {
	key := vobj.NoticeGroupCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeGroupConfig", "err", err)
		return nil, false
	}
	if exist == 0 {
		return nil, false
	}
	var noticeGroupConfig do.NoticeGroupConfig
	if err := c.GetCache().Client().HGet(ctx, key, name).Scan(&noticeGroupConfig); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeGroupConfig", "err", err)
		return nil, false
	}
	return &noticeGroupConfig, true
}

func (c *configImpl) GetNoticeGroupConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.NoticeGroup, error) {
	key := vobj.NoticeGroupCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeGroupConfig", "err", err)
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	all, err := c.GetCache().Client().HMGet(ctx, key, names...).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeGroupConfig", "err", err)
		return nil, err
	}

	noticeGroupConfigs := make([]*do.NoticeGroupConfig, 0, len(all))
	if err := slices.UnmarshalBinary(all, &noticeGroupConfigs); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeGroupConfig", "err", err)
		return nil, err
	}
	return slices.Map(noticeGroupConfigs, func(v *do.NoticeGroupConfig) bo.NoticeGroup { return v }), nil
}

func (c *configImpl) SetNoticeGroupConfig(ctx context.Context, teamID uint32, configs ...bo.NoticeGroup) error {
	configDos := make(map[string]any, len(configs))
	for _, v := range configs {
		templateMap := make(map[common.NoticeType]*do.Template, len(v.GetTemplates()))
		for _, t := range v.GetTemplates() {
			templateMap[t.GetType()] = &do.Template{
				Type:           t.GetType(),
				Template:       t.GetTemplate(),
				TemplateParams: t.GetTemplateParameters(),
			}
		}
		item := &do.NoticeGroupConfig{
			Name:            v.GetName(),
			SMSConfigName:   v.GetSmsConfigName(),
			EmailConfigName: v.GetEmailConfigName(),
			HookConfigNames: v.GetHookConfigNames(),
			SMSUserNames:    v.GetSmsUserNames(),
			EmailUserNames:  v.GetEmailUserNames(),
			Templates:       templateMap,
		}
		configDos[item.UniqueKey()] = item
	}
	return c.GetCache().Client().HSet(ctx, vobj.NoticeGroupCacheKey.Key(teamID), configDos).Err()
}

func (c *configImpl) GetNoticeUserConfig(ctx context.Context, teamID uint32, name string) (bo.NoticeUser, bool) {
	key := vobj.NoticeUserCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeUserConfig", "err", err)
		return nil, false
	}
	if exist == 0 {
		return nil, false
	}
	var noticeUserConfig do.NoticeUserConfig
	if err := c.GetCache().Client().HGet(ctx, key, name).Scan(&noticeUserConfig); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeUserConfig", "err", err)
		return nil, false
	}
	return &noticeUserConfig, true
}

func (c *configImpl) GetNoticeUserConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.NoticeUser, error) {
	key := vobj.NoticeUserCacheKey.Key(teamID)
	exist, err := c.GetCache().Client().Exists(ctx, key).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeUserConfig", "err", err)
		return nil, err
	}
	if exist == 0 {
		return nil, nil
	}
	all, err := c.GetCache().Client().HMGet(ctx, key, names...).Result()
	if err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeUserConfig", "err", err)
		return nil, err
	}

	noticeUserConfigs := make([]*do.NoticeUserConfig, 0, len(all))
	if err := slices.UnmarshalBinary(all, &noticeUserConfigs); err != nil {
		c.helper.WithContext(ctx).Errorw("method", "GetNoticeUserConfig", "err", err)
		return nil, err
	}
	return slices.Map(noticeUserConfigs, func(v *do.NoticeUserConfig) bo.NoticeUser { return v }), nil
}

func (c *configImpl) SetNoticeUserConfig(ctx context.Context, teamID uint32, configs ...bo.NoticeUser) error {
	configDos := make(map[string]any, len(configs))
	for _, v := range configs {
		item := &do.NoticeUserConfig{
			Name:  v.GetName(),
			Email: v.GetEmail(),
			Phone: v.GetPhone(),
		}
		configDos[item.UniqueKey()] = item
	}
	return c.GetCache().Client().HSet(ctx, vobj.NoticeUserCacheKey.Key(teamID), configDos).Err()
}

func (c *configImpl) RemoveConfig(ctx context.Context, params *bo.RemoveConfigParams) error {
	var key cache.K
	switch params.Type {
	case common.NoticeType_NOTICE_TYPE_EMAIL:
		key = vobj.EmailCacheKey
	case common.NoticeType_NOTICE_TYPE_SMS:
		key = vobj.SmsCacheKey
	case common.NoticeType_NOTICE_TYPE_HOOK_DINGTALK, common.NoticeType_NOTICE_TYPE_HOOK_WECHAT, common.NoticeType_NOTICE_TYPE_HOOK_FEISHU, common.NoticeType_NOTICE_TYPE_HOOK_WEBHOOK:
		key = vobj.HookCacheKey
	default:
		return nil
	}
	return c.GetCache().Client().HDel(ctx, key.Key(params.TeamID), params.Name).Err()
}
