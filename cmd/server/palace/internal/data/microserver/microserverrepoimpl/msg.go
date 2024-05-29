package microserverrepoimpl

import (
	"context"
	"encoding/json"

	hookapi "github.com/aide-cloud/moon/api/rabbit/hook"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data/microserver"
	"github.com/go-kratos/kratos/v2/log"
)

func NewMsgRepository(cli *microserver.RabbitConn) repository.Msg {
	return &msgRepositoryImpl{cli: cli}
}

type msgRepositoryImpl struct {
	cli *microserver.RabbitConn
}

func (m *msgRepositoryImpl) Send(ctx context.Context, msg *bo.Message) error {
	dataBytes, _ := json.Marshal(msg.Data)
	sendMsg, err := m.cli.SendMsg(ctx, &hookapi.SendMsgRequest{
		JsonData: string(dataBytes),
		Route:    "test",
	})
	if err != nil {
		return err
	}
	log.Infow("sendMsg", sendMsg)
	return nil
}
