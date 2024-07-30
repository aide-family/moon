package microserverrepoimpl

import (
	"context"
	"encoding/json"

	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/microserver"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

// NewMsgRepository 创建消息操作
func NewMsgRepository(cli *microserver.RabbitConn) repository.Msg {
	return &msgRepositoryImpl{cli: cli}
}

type msgRepositoryImpl struct {
	cli *microserver.RabbitConn
}

// Send 发送消息
func (m *msgRepositoryImpl) Send(ctx context.Context, msg *bo.Message) error {
	dataBytes, _ := json.Marshal(msg.Data)
	sendMsg, err := m.cli.SendMsg(ctx, &hookapi.SendMsgRequest{
		JsonData: string(dataBytes),
		Route:    "test",
	})
	if !types.IsNil(err) {
		return err
	}
	log.Infow("sendMsg", sendMsg)
	return nil
}
