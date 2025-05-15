package repository

import (
	"context"

	"github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/api/palace"
)

type ServerRegister interface {
	Register(ctx context.Context, server *common.ServerRegisterRequest) error
}

type Callback interface {
	SyncMetadata(ctx context.Context, req *palace.SyncMetadataRequest) error
}
