package service

import (
	"context"
	"time"

	pushapi "github.com/aide-family/moon/api/rabbit/push"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
)

type ConfigService struct {
	pushapi.UnimplementedConfigServer

	configBiz *biz.ConfigBiz
}

func NewConfigService(configBiz *biz.ConfigBiz) *ConfigService {
	return &ConfigService{
		configBiz: configBiz,
	}
}

func (s *ConfigService) NotifyObject(ctx context.Context, req *pushapi.NotifyObjectRequest) (*pushapi.NotifyObjectReply, error) {
	if err := s.configBiz.CacheConfig(ctx, &bo.CacheConfigParams{
		Receivers: req.GetReceivers(),
		Templates: req.GetTemplates(),
	}); !types.IsNil(err) {
		return nil, err
	}
	return &pushapi.NotifyObjectReply{
		Msg:  "ok",
		Code: 0,
		Time: types.NewTime(time.Now()).String(),
	}, nil
}

// LoadNotifyObject 加载配置
func (s *ConfigService) LoadNotifyObject(ctx context.Context) error {
	return s.configBiz.LoadConfig(ctx)
}
