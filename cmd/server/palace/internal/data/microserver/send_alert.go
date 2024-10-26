package microserver

import (
	"context"
	"strings"
	"time"

	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
	"github.com/go-kratos/kratos/v2/log"
)

var _ watch.Indexer = (*SendMsg)(nil)

type SendMsg struct {
	*hookapi.SendMsgRequest
}

func (s *SendMsg) Index() string {
	return s.RequestID
}

func NewSendAlertRepository(data *data.Data, rabbitConn *data.RabbitConn) microrepository.SendAlert {
	s := &sendAlertRepositoryImpl{rabbitConn: rabbitConn, data: data}
	go func() {
		defer after.RecoverX()
		msgCh := s.data.GetAlertPersistenceMsgQueue().Next()
		for {
			select {
			case msg, ok := <-msgCh:
				if !ok {
					break
				}
				sendMsg, ok := msg.GetData().(*SendMsg)
				if !ok {
					break
				}
				s.send(sendMsg.SendMsgRequest)
			}
		}
	}()
	return s
}

type sendAlertRepositoryImpl struct {
	data       *data.Data
	rabbitConn *data.RabbitConn
}

func (s *sendAlertRepositoryImpl) Send(_ context.Context, alerts []*bo.AlertItemRawParams, rowMap map[string]*alarmmodel.AlarmRaw) error {
	go func() {
		defer after.RecoverX()
		for _, v := range alerts {
			row, ok := rowMap[v.Fingerprint]
			if !ok {
				continue
			}
			routes := strings.Split(row.Receiver, ",")
			if len(routes) == 0 {
				continue
			}
			for _, route := range routes {
				key := v.NoticeKey(route)
				task := &hookapi.SendMsgRequest{
					Json:      v.GetAlertItemString(),
					Route:     route,
					RequestID: key,
				}
				s.send(task)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return nil
}

func (s *sendAlertRepositoryImpl) send(task *hookapi.SendMsgRequest) {
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
	if err := s.rabbitConn.SendMsg(ctx, task); err != nil {
		// 删除缓存
		if err := s.data.GetCacher().Delete(context.Background(), task.RequestID); err != nil {
			log.Warnf("send alert failed")
		}
		// 加入消息队列，重试
		s.data.GetAlertPersistenceMsgQueue().Push(watch.NewMessage(&SendMsg{task}, vobj.TopicAlertMsg))
	}
}
