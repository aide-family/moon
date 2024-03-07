package server

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"prometheus-manager/api"
	"prometheus-manager/api/agent"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/data"
	"prometheus-manager/app/prom_agent/internal/service"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/util/interflow"
)

var _ transport.Server = (*Watch)(nil)

type EventHandler = interflow.Callback

type Watch struct {
	ticker *time.Ticker
	log    *log.Helper

	loadService   *service.LoadService
	d             *data.Data
	interflowConf *conf.Interflow

	groups            *sync.Map
	exitCh            chan struct{}
	eventHandlers     map[consts.TopicType]EventHandler
	interflowInstance interflow.Interflow
}

func NewWatch(
	c *conf.WatchProm,
	interflowConf *conf.Interflow,
	d *data.Data,
	loadService *service.LoadService,
	logger log.Logger,
) (*Watch, error) {
	w := &Watch{
		interflowConf:     interflowConf,
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
		consts.ServerOnlineTopic: func(topic consts.TopicType, key, value []byte) error {
			return w.onlineNotify()
		},
		consts.ServerOfflineTopic: func(topic consts.TopicType, key, value []byte) error {
			return w.onlineNotify()
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

func (w *Watch) loadGroupAllEventHandler(_ consts.TopicType, _, value []byte) error {
	w.log.Info("strategyGroupAllTopic", string(value))
	// 把新规则刷进内存
	groupBytes := value
	var groupDetail *api.GroupSimple
	if err := json.Unmarshal(groupBytes, &groupDetail); err != nil {
		w.log.Warnf("unmarshal groupList error: %s", err.Error())
		return err
	}
	w.log.Info("groupDetail", groupDetail)
	w.groups.Store(groupDetail.GetId(), groupDetail)
	return nil
}

func (w *Watch) removeGroupEventHandler(topic consts.TopicType, key, value []byte) error {
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
				w.log.Info("[Watch] server tick")
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
	return w.onlineNotify()
}

func (w *Watch) evaluate(groupList []*api.GroupSimple) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _ = w.loadService.Evaluate(ctx, &agent.EvaluateRequest{GroupList: groupList})
}

func (w *Watch) Stop(_ context.Context) error {
	w.log.Info("[Watch] server stopping")
	if err := w.offlineNotify(); err != nil {
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

// onlineNotify 上线通知
func (w *Watch) onlineNotify() error {
	w.log.Info("[Watch] server online notify")
	topic := string(consts.AgentOnlineTopic)
	serverUrl := w.interflowConf.GetServer()
	agentUrl := w.interflowConf.GetAgent()
	msg := &interflow.HookMsg{
		Topic: topic,
		Value: []byte(agentUrl),
		Key:   []byte(agentUrl),
	}

	go func() {
		defer after.Recover(w.log)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for w.interflowInstance.Send(ctx, serverUrl, msg) != nil {
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			time.Sleep(10 * time.Second)
		}
		cancel()
	}()
	return nil
}

// offlineNotify 下线通知
func (w *Watch) offlineNotify() error {
	w.log.Info("[Watch] server offline notify")
	topic := string(consts.AgentOfflineTopic)
	w.log.Infow("topic", topic)
	serverUrl := w.interflowConf.GetServer()
	agentUrl := w.interflowConf.GetAgent()
	msg := &interflow.HookMsg{
		Topic: topic,
		Value: nil,
		Key:   []byte(agentUrl),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := w.interflowInstance.Send(ctx, serverUrl, msg)
	if err != nil {
		return err
	}
	// 等待1秒，等kafka消费完消息
	time.Sleep(1 * time.Second)
	return nil
}
