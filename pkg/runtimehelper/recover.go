package runtimehelper

import (
	"github.com/go-kratos/kratos/v2/log"
)

func Recover(module string) {
	if err := recover(); err != nil {
		log.Errorf("module: %s, error: %v", module, err)
	}
}
