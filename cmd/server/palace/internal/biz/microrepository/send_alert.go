package microrepository

import (
	"context"

	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

// SendAlert 发送告警事件接口
type SendAlert interface {
	// Send 发送告警事件
	Send(context.Context, map[string]*alarmmodel.AlarmRaw) error
}
