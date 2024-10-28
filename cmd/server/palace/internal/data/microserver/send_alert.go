package microserver

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
	"github.com/go-kratos/kratos/v2/log"
)

func NewSendAlertRepository(data *data.Data, rabbitConn *data.RabbitConn) microrepository.SendAlert {
	return &sendAlertRepositoryImpl{rabbitConn: rabbitConn, data: data}
}

type sendAlertRepositoryImpl struct {
	data       *data.Data
	rabbitConn *data.RabbitConn
}

func (s *sendAlertRepositoryImpl) Send(_ context.Context, alertMsg *bo.SendMsg) {
	go func() {
		defer after.RecoverX()
		s.send(alertMsg)
	}()
}

func (s *sendAlertRepositoryImpl) send(task *bo.SendMsg) {
	setOK, err := s.data.GetCacher().SetNX(context.Background(), task.RequestID, "1", 2*time.Hour)
	if err != nil {
		log.Warnf("set cache failed: %v", err)
		return
	}
	if !setOK {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.rabbitConn.SendMsg(ctx, task.SendMsgRequest); err != nil {
		// 删除缓存
		if err := s.data.GetCacher().Delete(context.Background(), task.RequestID); err != nil {
			log.Warnf("send alert failed")
		}
		// 加入消息队列，重试
		s.data.GetAlertQueue().Push(watch.NewMessage(task, vobj.TopicAlertMsg))
	}
}
