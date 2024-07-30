package after

import (
	"github.com/go-kratos/kratos/v2/log"
)

type (
	// RecoverCallback 恢复回调
	RecoverCallback func(err error)
)

// Recover 恢复
func Recover(logHelper *log.Helper, calls ...RecoverCallback) {
	if err := recover(); err != nil {
		logHelper.Errorf("panic error: %v", err)
		for _, call := range calls {
			call(err.(error))
		}
	}
}

// RecoverX 恢复, 默认使用log.Errorw
func RecoverX(calls ...RecoverCallback) {
	if err := recover(); err != nil {
		log.Errorw("type", "panic", "err", err)
		for _, call := range calls {
			call(err.(error))
		}
	}
}
