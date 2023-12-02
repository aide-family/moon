package helper

import (
	"github.com/go-kratos/kratos/v2/log"
)

type (
	RecoverCallback func(log *log.Helper, err error)
)

func Recover(logHelper *log.Helper, calls ...RecoverCallback) {
	if err := recover(); err != nil {
		logHelper.Errorf("module: %s, error: %v", err)
		for _, call := range calls {
			call(logHelper, err.(error))
		}
	}
}
