package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
)

type Server interface {
	RegisterServer(ctx context.Context, req *bo.ServerRegisterReq) error
	DeregisterServer(ctx context.Context, req *bo.ServerRegisterReq) error
}
