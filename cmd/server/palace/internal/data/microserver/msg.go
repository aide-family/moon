package microserver

import (
	"context"

	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/types"
)

// NewMsgRepository 创建消息操作
func NewMsgRepository(cli *data.RabbitConn) repository.Msg {
	return &msgRepositoryImpl{cli: cli}
}

type msgRepositoryImpl struct {
	cli *data.RabbitConn
}

// Send 发送消息
func (m *msgRepositoryImpl) Send(ctx context.Context, msg *bo.Message) error {
	dataBytes, _ := types.Marshal(msg.Data)
	err := m.cli.SendMsg(ctx, &hookapi.SendMsgRequest{
		Json:  string(dataBytes),
		Route: "test",
	})
	if !types.IsNil(err) {
		return err
	}
	return nil
}
