package microserver

import (
	"context"
	"strings"
	"time"

	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/go-kratos/kratos/v2/log"
)

func NewSendAlertRepository(data *data.RabbitConn) microrepository.SendAlert {
	return &sendAlertRepositoryImpl{rabbitConn: data}
}

type sendAlertRepositoryImpl struct {
	rabbitConn *data.RabbitConn
}

func (s *sendAlertRepositoryImpl) Send(_ context.Context, m map[string]*alarmmodel.AlarmRaw) error {
	tasks := make(chan *hookapi.SendMsgRequest, 100)
	go func() {
		defer after.RecoverX()
		for _, v := range m {
			routes := strings.Split(v.Receiver, ",")
			if len(routes) == 0 {
				continue
			}
			for _, route := range routes {
				tasks <- &hookapi.SendMsgRequest{JsonData: v.RawInfo, Route: route}
			}
		}
		for len(tasks) > 0 {
		}
		close(tasks)
	}()
	select {
	case task, ok := <-tasks:
		if !ok {
			break
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.rabbitConn.SendMsg(ctx, task); err != nil {
			log.Warnf("send alert failed: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
