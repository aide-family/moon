package microrepository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

// SendAlert 发送告警事件接口
type SendAlert interface {
	// Send 发送告警事件
	Send(context.Context, []*bo.AlertItemRawParams, map[string]*alarmmodel.AlarmRaw) error
}
