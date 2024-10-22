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
	"github.com/go-kratos/kratos/v2/log"
)

func NewSendAlertRepository(data *data.Data, rabbitConn *data.RabbitConn) microrepository.SendAlert {
	return &sendAlertRepositoryImpl{rabbitConn: rabbitConn, data: data}
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
				setOK, err := s.data.GetCacher().SetNX(context.Background(), v.Key(route), v.Fingerprint, 2*time.Hour)
				if err != nil {
					log.Warnf("set cache failed: %v", err)
					continue
				}
				if !setOK {
					continue
				}
				task := &hookapi.SendMsgRequest{JsonData: v.GetAlertItemString(), Route: route}
				s.send(task)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return nil
}

func (s *sendAlertRepositoryImpl) send(task *hookapi.SendMsgRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.rabbitConn.SendMsg(ctx, task); err != nil {
		log.Warnf("send alert failed: %v", err)
	}
}
