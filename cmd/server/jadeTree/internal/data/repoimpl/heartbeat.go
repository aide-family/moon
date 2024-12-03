package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/data/microserver"
)

// NewHeartbeatRepository 初始化心跳存储层
func NewHeartbeatRepository(palaceConn *microserver.PalaceConn) repository.Heartbeat {
	return &heartbeatRepositoryImpl{
		palaceConn: palaceConn,
	}
}

type heartbeatRepositoryImpl struct {
	palaceConn *microserver.PalaceConn
}

func (h *heartbeatRepositoryImpl) Heartbeat(ctx context.Context, in *api.HeartbeatRequest) error {
	_, err := h.palaceConn.Heartbeat(ctx, in)

	return err
}
