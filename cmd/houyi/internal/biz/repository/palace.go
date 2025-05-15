package repository

import (
	"context"

	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/api/palace"
)

type ServerRegister interface {
	Register(ctx context.Context, server *common.ServerRegisterRequest) error
}

type Callback interface {
	SyncMetadata(ctx context.Context, req *palace.SyncMetadataRequest) error
}
