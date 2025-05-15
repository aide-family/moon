package repository

import (
	"context"

	common "github.com/moon-monitor/moon/pkg/api/common"
)

type ServerRegister interface {
	Register(ctx context.Context, server *common.ServerRegisterRequest) error
}
