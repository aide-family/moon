package hook

import (
	"context"
	"sync"
	"time"

	"github.com/aide-family/moon/pkg/after"
	"github.com/aide-family/moon/pkg/helper/consts"
	"github.com/aide-family/moon/pkg/httpx"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
)

var _ interflow.AgentInterflow = (*hookHttpInterflow)(nil)
var _ interflow.ServerInterflow = (*hookHttpInterflow)(nil)

type (
	hookHttpInterflow struct {
		log     *log.Helper
		handles map[consts.TopicType]interflow.Callback
		server  HttpServerConfig
		agent   HttpServerConfig
		lock    sync.RWMutex
		closeCh chan struct{}
	}
)

func (l *hookHttpInterflow) SendAgent(ctx context.Context, to string, msg *interflow.HookMsg) error {
	_, err := httpx.NewHttpX().POSTWithContext(ctx, to, msg.Bytes())
	return err
}

func (l *hookHttpInterflow) ServerOnlineNotify(agentUrls []string) error {
	eg := new(errgroup.Group)
	topic := string(consts.ServerOnlineTopic)
	msg := &interflow.HookMsg{
		Topic: topic,
		Value: nil,
	}
	for _, agentUrl := range agentUrls {
		eg.Go(func() error {
			defer after.Recover(l.log)
			ctx, cancel := context.WithTimeout(context.Background(), interflow.Timeout)
			defer cancel()
			err := l.SendAgent(ctx, agentUrl, msg)
			if err != nil {
				l.log.Warnw("send server online notify error", err)
			}
			return nil
		})
	}
	return eg.Wait()
}

func (l *hookHttpInterflow) ServerOfflineNotify(agentUrls []string) error {
	eg := new(errgroup.Group)
	topic := string(consts.ServerOfflineTopic)
	msg := &interflow.HookMsg{
		Topic: topic,
		Value: nil,
	}
	for _, agentUrl := range agentUrls {
		eg.Go(func() error {
			defer after.Recover(l.log)
			ctx, cancel := context.WithTimeout(context.Background(), interflow.Timeout)
			defer cancel()
			return l.SendAgent(ctx, agentUrl, msg)
		})
	}
	return eg.Wait()
}

func (l *hookHttpInterflow) OnlineNotify() error {
	topic := string(consts.AgentOnlineTopic)

	msg := &interflow.HookMsg{
		Topic: topic,
		Value: []byte(l.agent.GetUrl()),
	}

	go func() {
		defer after.Recover(l.log)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), interflow.Timeout)
			err := l.Send(ctx, msg)
			cancel()
			if err == nil {
				break
			}
			l.log.Warnw("send online notify error", err)
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

func (l *hookHttpInterflow) OfflineNotify() error {
	topic := string(consts.AgentOfflineTopic)
	l.log.Infow("topic", topic)
	msg := &interflow.HookMsg{
		Topic: topic,
		Value: []byte(l.agent.GetUrl()),
	}
	count := 1
	for {
		if count > 3 {
			break
		}
		ctx, cancel := context.WithTimeout(context.Background(), interflow.Timeout)
		if err := l.Send(ctx, msg); err != nil {
			cancel()
			l.log.Warnw("send offline notify error", err)
			count++
			// 等待1秒
			time.Sleep(1 * time.Second)
			continue
		}
		cancel()
		break
	}

	return nil
}

func (l *hookHttpInterflow) Close() error {
	close(l.closeCh)
	return nil
}

func (l *hookHttpInterflow) Send(ctx context.Context, msg *interflow.HookMsg) (err error) {
	_, err = httpx.NewHttpX().POSTWithContext(ctx, l.server.GetUrl(), msg.Bytes())
	return err
}

func (l *hookHttpInterflow) Receive() error {
	receiveCh := interflow.GetReceiveInterflowCh()
	go func() {
		defer after.Recover(l.log)
		for {
			select {
			case msg := <-receiveCh:
				if handle, ok := l.handles[consts.TopicType(msg.Topic)]; ok {
					if err := handle(consts.TopicType(msg.Topic), msg.Value); err != nil {
						l.log.Warnw("err", err, "topic", msg.Topic, "value", string(msg.Value))
					}
				}
			case <-l.closeCh:
				l.log.Info("hookHttpInterflow closed")
				return
			}
		}
	}()
	return nil
}

func (l *hookHttpInterflow) SetHandles(handles map[consts.TopicType]interflow.Callback) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.handles = handles
	return nil
}

func NewAgentHookHttpInterflow(c HttpConfig, logger log.Logger) (interflow.AgentInterflow, error) {
	return &hookHttpInterflow{
		log:     log.NewHelper(log.With(logger, "module", "interflow.hook.http")),
		server:  c.GetServer(),
		agent:   c.GetAgent(),
		closeCh: make(chan struct{}),
	}, nil
}

func NewServerHookHttpInterflow(logger log.Logger) interflow.ServerInterflow {
	return &hookHttpInterflow{
		log:     log.NewHelper(log.With(logger, "module", "interflow.hook.http")),
		closeCh: make(chan struct{}),
	}
}
