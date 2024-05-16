package after

import (
	"github.com/go-kratos/kratos/v2/log"
)

type (
	RecoverCallback func(err error)
)

func Recover(logHelper *log.Helper, calls ...RecoverCallback) {
	if err := recover(); err != nil {
		logHelper.Errorf("panic error: %v", err)
		for _, call := range calls {
			call(err.(error))
		}
	}
}

func RecoverX(calls ...RecoverCallback) {
	if err := recover(); err != nil {
		log.Errorw("type", "panic", "err", err)
		for _, call := range calls {
			call(err.(error))
		}
	}
}
