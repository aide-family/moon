package microrepository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
)

// SendAlert 发送告警事件接口
type SendAlert interface {
	// Send 发送告警事件
	Send(context.Context, *bo.SendMsg)
}
