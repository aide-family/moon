package repository

import (
	"context"

	common "github.com/aide-family/moon/pkg/api/common"
)

type ServerRegister interface {
	Register(ctx context.Context, server *common.ServerRegisterRequest) error
}
