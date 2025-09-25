package impl

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/do"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/rabbit/internal/conf"
	"github.com/aide-family/moon/cmd/rabbit/internal/data"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewConfigRepo(d *data.Data, bc *conf.Bootstrap, logger log.Logger) (repository.Config, error) {
	configRepo := &configImpl{
		helper: log.NewHelper(log.With(logger, "module", "data.repo.config")),
		bc:     bc,
		Data:   d,
	}
	if err := configRepo.initConfig(); err != nil {
		return nil, err
	}
	return configRepo, nil
}

type configImpl struct {
	helper *log.Helper
	bc     *conf.Bootstrap
	*data.Data
}

func (c *configImpl) initConfig() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.initEmailConfig(ctx, c.bc.GetEmailConfigs()); err != nil {
		return err
	}
	if err := c.initSMSConfig(ctx, c.bc.GetSmsConfigs()); err != nil {
		return err
	}
	if err := c.initHookConfig(ctx, c.bc.GetHookConfigs()); err != nil {
		return err
	}
	if err := c.initNoticeGroupConfig(ctx, c.bc.GetNoticeGroupConfigs()); err != nil {
		return err
	}
	return nil
}

func (c *configImpl) initEmailConfig(ctx context.Context, configs []*conf.EmailConfigs) error {
	emailGroup := make(map[uint32][]*do.EmailConfig)
	for _, v := range configs {
		teamID := v.GetTeamId()
		emailConfigs := v.GetConfigs()
		if _, ok := emailGroup[teamID]; !ok {
			emailGroup[teamID] = make([]*do.EmailConfig, 0, len(emailConfigs))
		}
		for _, v := range emailConfigs {
			emailGroup[teamID] = append(emailGroup[teamID], &do.EmailConfig{
				User:   v.GetUser(),
				Pass:   v.GetPass(),
				Host:   v.GetHost(),
				Port:   v.GetPort(),
				Enable: v.GetEnable(),
				Name:   v.GetName(),
			})
		}
	}
	for teamID, v := range emailGroup {
		if err := c.setEmailConfig(ctx, teamID, v...); err != nil {
			return err
		}
	}
	return nil
}

func (c *configImpl) initSMSConfig(ctx context.Context, configs []*conf.SMSConfigs) error {
	smsGroup := make(map[uint32][]*do.SMSConfig)
	for _, v := range configs {
		teamID := v.GetTeamId()
		smsConfigs := v.GetConfigs()
		if _, ok := smsGroup[teamID]; !ok {
			smsGroup[teamID] = make([]*do.SMSConfig, 0, len(smsConfigs))
		}
		for _, v := range smsConfigs {
			provider := v.GetType()
			var smsConfig *do.SMSConfig
			switch provider {
			case conf.SMSConfig_ALIYUN:
				aliyunConfig := v.GetAliyun()
				smsConfig = &do.SMSConfig{
					AccessKeyID:     aliyunConfig.GetAccessKeyId(),
					AccessKeySecret: aliyunConfig.GetAccessKeySecret(),
					Endpoint:        aliyunConfig.GetEndpoint(),
					Name:            aliyunConfig.GetName(),
					SignName:        aliyunConfig.GetSignName(),
					Type:            vobj.SMSProviderTypeAliyun,
					Enable:          v.GetEnable(),
				}
			default:
				c.helper.WithContext(ctx).Warnw("method", "initSMSConfig", "err", merr.ErrorParams("No SMS configuration is available"))
				continue
			}
			smsGroup[teamID] = append(smsGroup[teamID], smsConfig)
		}
	}
	for teamID, v := range smsGroup {
		if err := c.setSMSConfig(ctx, teamID, v...); err != nil {
			return err
		}
	}
	return nil
}

func (c *configImpl) initHookConfig(ctx context.Context, configs []*conf.HookConfigs) error {
	hookGroup := make(map[uint32][]*do.HookConfig)
	for _, v := range configs {
		teamID := v.GetTeamId()
		hookConfigs := v.GetConfigs()
		if _, ok := hookGroup[teamID]; !ok {
			hookGroup[teamID] = make([]*do.HookConfig, 0, len(hookConfigs))
		}
		for _, v := range hookConfigs {
			hookGroup[teamID] = append(hookGroup[teamID], &do.HookConfig{
				Name:     v.GetName(),
				App:      vobj.APP(v.GetApp()),
				URL:      v.GetUrl(),
				Secret:   v.GetSecret(),
				Token:    v.GetToken(),
				Username: v.GetUsername(),
				Password: v.GetPassword(),
				Headers:  v.GetHeaders(),
				Enable:   v.GetEnable(),
			})
		}
	}
	for teamID, v := range hookGroup {
		if err := c.setHookConfig(ctx, teamID, v...); err != nil {
			return err
		}
	}
	return nil
}

func (c *configImpl) initNoticeGroupConfig(ctx context.Context, configs []*conf.NoticeGroupConfigs) error {
	noticeGroupGroup := make(map[uint32][]*do.NoticeGroupConfig)
	for _, v := range configs {
		teamID := v.GetTeamId()
		noticeGroupConfigs := v.GetConfigs()
		if _, ok := noticeGroupGroup[teamID]; !ok {
			noticeGroupGroup[teamID] = make([]*do.NoticeGroupConfig, 0, len(noticeGroupConfigs))
		}
		for _, v := range noticeGroupConfigs {
			templates := slices.ToMapWithValue(v.GetTemplates(), func(v *conf.NoticeGroupConfig_Template) (vobj.APP, *do.Template) {
				return vobj.APP(v.GetType()), &do.Template{
					Type:           vobj.APP(v.GetType()),
					Template:       v.GetTemplate(),
					TemplateParams: v.GetTemplateParameters(),
					Subject:        v.GetSubject(),
				}
			})
			noticeGroupGroup[teamID] = append(noticeGroupGroup[teamID], &do.NoticeGroupConfig{
				Name:            v.GetName(),
				SMSConfigName:   v.GetSmsConfigName(),
				EmailConfigName: v.GetEmailConfigName(),
				HookReceivers:   v.GetHookReceivers(),
				SMSReceivers:    v.GetSmsReceivers(),
				EmailReceivers:  v.GetEmailReceivers(),
				Templates:       templates,
			})
		}
	}
	for teamID, v := range noticeGroupGroup {
		if err := c.setNoticeGroupConfig(ctx, teamID, v...); err != nil {
			return err
		}
	}
	return nil
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
	configDos := slices.Map(configs, func(v bo.EmailConfig) *do.EmailConfig {
		return &do.EmailConfig{
			User:   v.GetUser(),
			Pass:   v.GetPass(),
			Host:   v.GetHost(),
			Port:   v.GetPort(),
			Enable: v.GetEnable(),
			Name:   v.GetName(),
		}
	})

	return c.setEmailConfig(ctx, teamID, configDos...)
}

func (c *configImpl) setEmailConfig(ctx context.Context, teamID uint32, configs ...*do.EmailConfig) error {
	configDos := slices.ToMapWithValue(configs, func(v *do.EmailConfig) (string, any) { return v.UniqueKey(), v })
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
	configDos := slices.Map(configs, func(v bo.SMSConfig) *do.SMSConfig {
		return &do.SMSConfig{
			AccessKeyID:     v.GetAccessKeyID(),
			AccessKeySecret: v.GetAccessKeySecret(),
			Endpoint:        v.GetEndpoint(),
			Name:            v.GetName(),
			SignName:        v.GetSignName(),
			Type:            v.GetType(),
			Enable:          v.GetEnable(),
		}
	})
	return c.setSMSConfig(ctx, teamID, configDos...)
}

func (c *configImpl) setSMSConfig(ctx context.Context, teamID uint32, configs ...*do.SMSConfig) error {
	configDos := slices.ToMapWithValue(configs, func(v *do.SMSConfig) (string, any) { return v.UniqueKey(), v })
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
	configDos := slices.Map(configs, func(v bo.HookConfig) *do.HookConfig {
		return &do.HookConfig{
			Name:     v.GetName(),
			App:      v.GetApp(),
			URL:      v.GetURL(),
			Secret:   v.GetSecret(),
			Token:    v.GetToken(),
			Username: v.GetUsername(),
			Password: v.GetPassword(),
			Headers:  v.GetHeaders(),
			Enable:   v.GetEnable(),
		}
	})
	return c.setHookConfig(ctx, teamID, configDos...)
}

func (c *configImpl) setHookConfig(ctx context.Context, teamID uint32, configs ...*do.HookConfig) error {
	configDos := slices.ToMapWithValue(configs, func(v *do.HookConfig) (string, any) { return v.UniqueKey(), v })
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
	configDos := slices.Map(configs, func(v bo.NoticeGroup) *do.NoticeGroupConfig {
		templates := make(map[vobj.APP]*do.Template, len(v.GetTemplates()))
		for _, t := range v.GetTemplates() {
			templates[t.GetType()] = &do.Template{
				Type:           t.GetType(),
				Template:       t.GetTemplate(),
				TemplateParams: t.GetTemplateParameters(),
				Subject:        t.GetSubject(),
			}
		}
		return &do.NoticeGroupConfig{
			Name:            v.GetName(),
			SMSConfigName:   v.GetSmsConfigName(),
			EmailConfigName: v.GetEmailConfigName(),
			HookReceivers:   v.GetHookReceivers(),
			SMSReceivers:    v.GetSmsReceivers(),
			EmailReceivers:  v.GetEmailReceivers(),
			Templates:       templates,
		}
	})

	return c.setNoticeGroupConfig(ctx, teamID, configDos...)
}

func (c *configImpl) setNoticeGroupConfig(ctx context.Context, teamID uint32, configs ...*do.NoticeGroupConfig) error {
	configDos := slices.ToMapWithValue(configs, func(v *do.NoticeGroupConfig) (string, any) { return v.UniqueKey(), v })
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
