package server

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/agent"
	"github.com/aide-family/moon/app/prom_agent/internal/conf"
	"github.com/aide-family/moon/app/prom_agent/internal/data"
	"github.com/aide-family/moon/app/prom_agent/internal/service"
	"github.com/aide-family/moon/pkg/after"
	"github.com/aide-family/moon/pkg/helper/consts"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*Watch)(nil)

const timeout = 10 * time.Second

type EventHandler = interflow.Callback

type Watch struct {
	ticker *time.Ticker
	log    *log.Helper

	loadService *service.LoadService
	d           *data.Data

	groups            *sync.Map
	exitCh            chan struct{}
	eventHandlers     map[consts.TopicType]EventHandler
	interflowInstance interflow.Interflow
}

func NewWatch(
	c *conf.WatchProm,
	d *data.Data,
	loadService *service.LoadService,
	logger log.Logger,
) (*Watch, error) {
	w := &Watch{
		exitCh:            make(chan struct{}, 1),
		ticker:            time.NewTicker(c.GetInterval().AsDuration()),
		log:               log.NewHelper(log.With(logger, "module", "server.watch")),
		loadService:       loadService,
		groups:            new(sync.Map),
		eventHandlers:     make(map[consts.TopicType]EventHandler),
		interflowInstance: d.Interflow(),
	}
	w.eventHandlers = map[consts.TopicType]EventHandler{
		consts.StrategyGroupAllTopic: w.loadGroupAllEventHandler,
		consts.RemoveGroupTopic:      w.removeGroupEventHandler,
		consts.ServerOnlineTopic: func(topic consts.TopicType, value []byte) error {
			w.log.Debugw("server online", "", "value", string(value), "topic", topic.String())
			return w.interflowInstance.OnlineNotify()
		},
		consts.ServerOfflineTopic: func(topic consts.TopicType, value []byte) error {
			w.log.Debugw("server offline", "", "value", string(value), "topic", topic.String())
			return w.interflowInstance.OnlineNotify()
		},
	}

	if err := w.interflowInstance.SetHandles(w.eventHandlers); err != nil {
		return nil, err
	}

	if err := w.interflowInstance.Receive(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watch) loadGroupAllEventHandler(_ consts.TopicType, value []byte) error {
	w.log.Info("strategyGroupAllTopic", string(value))
	// 把新规则刷进内存
	groupBytes := value
	var groupDetail *api.GroupSimple
	if err := json.Unmarshal(groupBytes, &groupDetail); err != nil {
		w.log.Warnf("unmarshal groupList error: %s", err.Error())
		return err
	}
	w.log.Debugw("groupDetail", groupDetail)
	w.groups.Store(groupDetail.GetId(), groupDetail)
	return nil
}

func (w *Watch) removeGroupEventHandler(topic consts.TopicType, value []byte) error {
	w.log.Info("removeGroupTopic", string(value))
	groupId, err := strconv.ParseUint(string(value), 10, 64)
	if err != nil {
		w.log.Warnf("parse groupId error: %s", value)
		return nil
	}
	if groupId > 0 {
		w.groups.Delete(uint32(groupId))
	}
	return nil
}

func (w *Watch) Start(_ context.Context) error {
	go func() {
		defer after.Recover(w.log, func(err error) {
			w.log.Errorf("recover error: %s", err.Error())
		})
		for {
			select {
			case <-w.exitCh:
				w.shutdown()
				return
			case <-w.ticker.C:
				//w.log.Debug("[Watch] server tick")
				groupList := make([]*api.GroupSimple, 0)
				w.groups.Range(func(key, value any) bool {
					if group, ok := value.(*api.GroupSimple); ok && group != nil {
						groupList = append(groupList, group)
					}
					return true
				})
				w.evaluate(groupList)
			}
		}
	}()
	w.log.Info("[Watch] server started")
	return w.interflowInstance.OnlineNotify()
}

func (w *Watch) evaluate(groupList []*api.GroupSimple) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, _ = w.loadService.Evaluate(ctx, &agent.EvaluateRequest{GroupList: groupList})
}

func (w *Watch) Stop(_ context.Context) error {
	w.log.Info("[Watch] server stopping")
	if err := w.interflowInstance.OnlineNotify(); err != nil {
		return err
	}
	close(w.exitCh)
	return nil
}

func (w *Watch) shutdown() {
	w.groups = nil
	w.ticker.Stop()
	w.log.Info("[Watch] server stopped")
}
