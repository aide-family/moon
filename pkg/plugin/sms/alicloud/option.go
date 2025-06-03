package alicloud

import (
	"github.com/go-kratos/kratos/v2/log"
)

func WithLogger(logger log.Logger) Option {
	return func(a *aliCloudImpl) {
		a.helper = log.NewHelper(log.With(logger, "module", "plugin.sms.aliyun"))
	}
}
