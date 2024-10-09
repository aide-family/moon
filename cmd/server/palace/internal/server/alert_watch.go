package server

import (
	"context"
	"time"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	alarmHistoryWatchJobTimeout = 10 * time.Second
)

var _ transport.Server = (*AlertWatch)(nil)

type (
	// AlertWatch 报警历史记录
	AlertWatch struct {
		data         *data.Data
		timeout      time.Duration
		stopCh       chan struct{}
		alertService *service.AlertService
	}
)

func newAlertWatch(c *palaceconf.Bootstrap, data *data.Data, alertService *service.AlertService) *AlertWatch {
	return &AlertWatch{
		data:         data,
		timeout:      alarmHistoryWatchJobTimeout,
		stopCh:       make(chan struct{}),
		alertService: alertService,
	}
}

func (s *AlertWatch) Stop(_ context.Context) error {
	defer log.Infof("[AlarmHistoryWatch] server stopped")
	s.stopCh <- struct{}{}
	return nil
}

// Start 启动告警历史任务执行器
func (s *AlertWatch) Start(_ context.Context) error {
	if types.IsNil(s.data) {
		return merr.ErrorNotificationSystemError("data is nil")
	}
	if types.IsNil(s.data) {
		return merr.ErrorNotificationSystemError("data is nil")
	}
	if types.IsNil(s.data.GetAlartHistoryQueue()) {
		return merr.ErrorNotificationSystemError("history queue is nil")
	}

	go func() {
		defer after.RecoverX()
		for {
			select {
			case <-s.stopCh:
				return
			case msg, ok := <-s.data.GetAlartHistoryQueue().Next():
				if !ok || !msg.GetTopic().IsAlert() {
					continue
				}
				params := msg.GetData().(*bo.CreateAlarmHookRawParams)
				if err := s.alertService.CreateAlarmInfo(context.Background(), params); err != nil {
					log.Errorw("save Alert queue err", err)
				}
			}
		}
	}()
	log.Info("[AlarmHistoryWatch] server started")
	return nil
}
