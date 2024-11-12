package biz

import (
	"context"
	"sync"

	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

// NewConfigBiz 创建配置业务
func NewConfigBiz(c *rabbitconf.Bootstrap) *ConfigBiz {
	return &ConfigBiz{
		c: c,
	}
}

// GetConfigData 获取配置数据
func GetConfigData() *Config {
	if types.IsNil(configData) {
		configData = &Config{
			Receivers: new(sync.Map),
			Templates: new(sync.Map),
		}
	}
	return configData
}

var configData = &Config{
	Receivers: new(sync.Map),
	Templates: new(sync.Map),
}

// GetReceivers 获取接收人
func (l *Config) GetReceivers(route string) (*conf.Receiver, bool) {
	if types.IsNil(l) {
		return nil, false
	}
	val, ok := GetConfigData().Receivers.Load(route)
	if ok {
		receivers, ok := val.(*conf.Receiver)
		return receivers, ok
	}
	return nil, false
}

// GetTemplates 获取模板
func (l *Config) GetTemplates(temp string) string {
	if types.IsNil(l) {
		return ""
	}
	val, ok := GetConfigData().Templates.Load(temp)
	if ok {
		template, ok := val.(string)
		if ok {
			return template
		}
	}
	return ""
}

// ConfigBiz 配置业务
type ConfigBiz struct {
	c *rabbitconf.Bootstrap
}

// Config 配置数据
type Config struct {
	Receivers *sync.Map `json:"receivers"`
	Templates *sync.Map `json:"templates"`
}

// Set 设置接收人
func (l *Config) Set(_ context.Context, params *bo.CacheConfigParams) {
	log.Debugw("method", "设置接收人", "params", params)
	for k, v := range params.Receivers {
		r := &conf.Receiver{
			Hooks: types.SliceTo(v.GetHooks(), func(item *conf.ReceiverHook) *conf.ReceiverHook {
				return &conf.ReceiverHook{
					Type:     item.GetType(),
					Webhook:  item.GetWebhook(),
					Content:  item.GetContent(),
					Template: item.GetTemplate(),
					Secret:   item.GetSecret(),
				}
			}),
			Phones: types.SliceTo(v.GetPhones(), func(item *conf.ReceiverPhone) *conf.ReceiverPhone {
				return &conf.ReceiverPhone{}
			}),
			Emails: types.SliceTo(v.GetEmails(), func(item *conf.ReceiverEmail) *conf.ReceiverEmail {
				return &conf.ReceiverEmail{
					To:          item.GetTo(),
					Subject:     item.GetSubject(),
					Content:     item.GetContent(),
					Template:    item.GetTemplate(),
					Cc:          item.GetCc(),
					AttachUrl:   item.GetAttachUrl(),
					ContentType: item.GetContentType(),
				}
			}),
		}
		if !types.IsNil(v.GetEmailConfig()) {
			r.EmailConfig = &conf.EmailConfig{
				User: v.GetEmailConfig().GetUser(),
				Pass: v.GetEmailConfig().GetPass(),
				Host: v.GetEmailConfig().GetHost(),
				Port: v.GetEmailConfig().GetPort(),
			}
		}
		l.Receivers.Store(k, r)
	}
	for k, v := range params.Templates {
		l.Templates.Store(k, v)
	}
}

// CacheConfig 缓存配置
func (b *ConfigBiz) CacheConfig(ctx context.Context, params *bo.CacheConfigParams) {
	configData.Set(ctx, params)
}

// LoadConfig 加载配置
func (b *ConfigBiz) LoadConfig(ctx context.Context) {
	defer log.Debug("加载配置完成")
	// 加载配置文件配置进内存
	yamlConfig := &bo.CacheConfigParams{
		Receivers: b.c.GetReceivers(),
		Templates: b.c.GetTemplates(),
	}

	configData.Set(ctx, yamlConfig)
}
