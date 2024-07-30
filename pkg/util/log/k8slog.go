package log

import (
	"github.com/go-kratos/kratos/v2/log"
	"k8s.io/klog/v2"
)

var _ log.Logger = (*k8sLogger)(nil)

// NewK8sLogger 创建k8s日志
func NewK8sLogger(opts ...K8sLoggerOption) log.Logger {
	k := &k8sLogger{}
	for _, o := range opts {
		o(k)
	}
	return k
}

type (
	k8sLogger struct {
		log klog.Logger
	}

	// K8sLoggerOption k8s日志配置
	K8sLoggerOption func(k *k8sLogger)
)

// Log k8s日志实现kratos接口
func (k *k8sLogger) Log(level log.Level, keyvals ...interface{}) error {
	//TODO implement me
	panic("implement me")
}

// WithK8sLoggerLog 配置k8s日志
func WithK8sLoggerLog(logger klog.Logger) K8sLoggerOption {
	return func(k *k8sLogger) {
		k.log = logger
	}
}
