package manager

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/rabbit"
	"github.com/aide-family/moon/pkg/rabbit/metrics"
	"github.com/aide-family/moon/pkg/runtime/cache"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2/klogr"
)

var _ rabbit.ProcessorProvider = &Manager{}

// Options 是 Manager 的配置项
type Options struct {
	// MaxConcurrentWorkers 是最大并发 worker 数量，默认为1
	MaxConcurrentWorkers int

	// Log 日志记录器
	Log *logr.Logger

	// Backoff 消息再次加入时，用来决定消息重新入队的时间
	Backoff *rabbit.Backoff

	// MaxRetries 表示数据最多尝试的入队次数，超过这个次数，数据将被丢弃
	MaxRetries int
}

// Manager 是一个消息队列的 Manager
type Manager struct {
	name                 string
	MaxConcurrentWorkers int
	Started              bool
	Queue                rabbit.MessageQueue
	ctx                  context.Context
	Log                  *logr.Logger
	receivers            map[string]rabbit.Receiver
	rg                   cache.Cache
	fm                   *FilterManager
	am                   *AggregatorManager
	tm                   *TemplaterManager
	sm                   *SenderManager
}

// New 创建一个 Manager
func New(name string, opts *Options) (*Manager, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("invalid manager name")
	}
	if opts.MaxConcurrentWorkers <= 0 {
		opts.MaxConcurrentWorkers = 1
	}

	if opts.Log == nil {
		log := klogr.New()
		opts.Log = &log
	}
	log := opts.Log.WithName(name)
	return &Manager{
		name:                 name,
		MaxConcurrentWorkers: opts.MaxConcurrentWorkers,
		Started:              false,
		Queue:                rabbit.NewQueue(opts.MaxRetries, opts.Backoff),
		Log:                  &log,
		receivers:            make(map[string]rabbit.Receiver),
	}, nil
}

const (
	labelError       = "error"
	labelRequeue     = "requeue"
	labelFilter      = "filter"
	labelAggregation = "aggregation"
	labelTemplate    = "template"
	labelSend        = "send"
)

// Name 返回 Manager 的名称
func (m *Manager) Name() string {
	return m.name
}

func (m *Manager) initMetrics() {
	metrics.ActiveWorkers.WithLabelValues().Set(0)
	metrics.WorkerErrors.WithLabelValues().Add(0)
	metrics.WorkerCount.WithLabelValues().Set(float64(m.MaxConcurrentWorkers))
	metrics.WorkerTotal.WithLabelValues(labelError).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelRequeue).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelFilter).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelAggregation).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelTemplate).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelSend).Add(0)
}

// RegisterReceivers 注册一个或多个 Receiver
func (m *Manager) RegisterReceivers(receivers ...rabbit.Receiver) {
	for _, receiver := range receivers {
		m.receivers[receiver.Name()] = receiver
	}
}

// Start 启动 Manager
func (m *Manager) Start(ctx context.Context) error {
	if m.Started {
		return errors.New("manager was started more than once")
	}
	m.initMetrics()

	m.ctx = ctx

	wg := &sync.WaitGroup{}
	err := func() error {
		defer HandleCrash()

		for name, receiver := range m.receivers {
			rch, err := receiver.Receive()
			if err != nil {
				m.Log.Error(err, "receiver failed to receive", "name", name)
				return err
			}
			metrics.ReceiverTotal.WithLabelValues(name).Add(0)
			wg.Add(1)
			go func(_ctx context.Context, group *sync.WaitGroup, ch <-chan *api.Message, name string) {
				defer group.Done()
				for {
					select {
					case msg := <-ch:
						m.Queue.Add(msg)
						metrics.ReceiverTotal.WithLabelValues(name).Inc()
					case <-_ctx.Done():
						return
					}
				}
			}(ctx, wg, rch, name)
		}

		// Launch workers to process resources
		m.Log.Info("Starting workers", "worker count", m.MaxConcurrentWorkers)
		wg.Add(m.MaxConcurrentWorkers)
		for i := 0; i < m.MaxConcurrentWorkers; i++ {
			go func() {
				defer wg.Done()
				// processNextWorkItem 会从队列中拿出需要处理的对应，使用队列强制保证了即使启动多个worker也不会消费同一条数据
				for m.processNextWorkItem(ctx) {
				}
			}()
		}
		m.Started = true
		return nil
	}()
	if err != nil {
		return err
	}
	<-ctx.Done()
	m.Log.Info("Shutdown signal received, waiting for all workers to finish")
	m.Queue.ShutDown()
	wg.Wait()
	m.Log.Info("All workers finished")
	return err
}

func (m *Manager) processNextWorkItem(ctx context.Context) bool {
	obj, ok := m.Queue.Get()
	if ok {
		// 当从消息队列中获取消息失败时，表明发生了意料之外的错误，
		// 此时，当前worker会退出。
		return false
	}

	// 处理完成每条消息后，我们应该告知队列这条数据已经处理完成。
	// 以便于其从消息队列中移除
	defer m.Queue.Done(obj)

	metrics.ActiveWorkers.WithLabelValues().Add(1)
	defer metrics.ActiveWorkers.WithLabelValues().Add(-1)
	m.worker(ctx, obj)
	return true
}

func (m *Manager) worker(ctx context.Context, info *rabbit.QueueInfo) {
	workerStartTS := time.Now()
	defer func() {
		metrics.WorkerTime.WithLabelValues().Observe(time.Since(workerStartTS).Seconds())
	}()

	log := m.Log.WithValues("message", info.Key)

	// 根据队列消息构建消息快照，用于处理消息，构建失败则重新加入队列
	// 对于构建失败的消息，应该将其重新加入队列，使其在达到退避时间后重新入队
	// 当然，不是每一次失败，消息都能够重新入队，对于失败过多的消息，应该将其抛弃
	snapshot := rabbit.NewMessageSnapshot(ctx, info.Message)
	err := snapshot.CompleteMessageSnapshot(ctx, m, m)
	if err != nil {
		log.Error(err, "build snapshot failed")
		metrics.WorkerErrors.WithLabelValues().Inc()
		metrics.WorkerTotal.WithLabelValues(labelError).Inc()
		if m.Queue.TryAgain(info) {
			metrics.WorkerTotal.WithLabelValues(labelRequeue).Inc()
		}
		return
	}

	// 对于快照构建成功的消息，将会取出 processor 进行工作。
	// 此时，processor 如果发生错误或者失败，消息将不在重新入队。
	// 如果 processor 是发生错误，表明消息与 processor 不匹配或者 processor 构建不对，
	// 应该检查消息是否正确或修复 processor，以期望其正确运行，而非将消息重新入队进行消费。
	for _, processor := range snapshot.Processors {
		m.processor(snapshot.Context, log, snapshot.Message, processor)
	}

}

func (m *Manager) processor(ctx context.Context, log logr.Logger, message *api.Message, processor *rabbit.Processor) {

	// 对需要处理的消息进行过滤，通过交给 Aggregator 进行聚合
	metrics.WorkerTotal.WithLabelValues(labelFilter).Inc()
	if processor.Filter.Allow(message) {
		var err error

		// 对于过滤通过的消息进行聚合,聚合完成则返回聚合后的消息，交给 Templater 进行模版解析
		metrics.WorkerTotal.WithLabelValues(labelAggregation).Inc()
		newMessage, err := processor.Aggregator.Group(message)
		if err != nil {
			log.Error(err, "group message failed", "aggregator", processor.Aggregator.Name(), "labels", message.Labels)
		}
		if newMessage != nil {
			// 对于聚合完成的消息，则使用模版进行解析，解析成功则进行交给 Sender 进行发送
			buf := &bytes.Buffer{}
			metrics.WorkerTotal.WithLabelValues(labelTemplate).Inc()
			err = processor.Templater.Parse(newMessage.Content, buf)
			if err != nil {
				log.Error(err, "template parsing error", "templater", processor.Templater.Name(), "labels", newMessage.Labels)
				return
			}
			// 对与模版解析成功的，使用 Sender 进行发送
			metrics.WorkerTotal.WithLabelValues(labelSend).Inc()
			err = processor.Sender.Send(ctx, buf.Bytes())
			if err != nil {
				log.Error(err, "send error", "sender", processor.Sender.Name(), "labels", message.Labels)
				return
			}
			log.V(5).Info("send successful", "sender", processor.Sender.Name(), "labels", message.Labels)
		} else {
			log.V(5).Info("waiting message group finish", "aggregator", processor.Aggregator.Name(), "labels", message.Labels)
		}
	} else {
	}
}

func (m *Manager) RuleGroup(ctx context.Context, name string) (*rabbit.RuleGroup, error) {
	rule := &rabbit.RuleGroup{}
	err := m.rg.Get(ctx, name, rule)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (m *Manager) Filter(ctx context.Context, ruleName string) (rabbit.Filter, error) {
	return m.fm.Filter(ctx, ruleName)
}

func (m *Manager) Aggregator(ctx context.Context, ruleName string) (rabbit.Aggregator, error) {
	return m.am.Aggregator(ctx, ruleName)
}

func (m *Manager) Templater(ctx context.Context, ruleName string) (rabbit.Templater, error) {
	return m.tm.Templater(ctx, ruleName)
}

func (m *Manager) Sender(ctx context.Context, ruleName string) (rabbit.Sender, error) {
	return m.sm.Sender(ctx, ruleName)
}

type FilterManager struct {
	rules     cache.Cache
	processor map[string]rabbit.Filter
}

func (x *FilterManager) Filter(ctx context.Context, ruleName string) (rabbit.Filter, error) {
	rule := &rabbit.MessageFilterRule{}
	err := x.rules.Get(ctx, ruleName, rule)
	if err != nil {
		return nil, err
	}
	processor, ok := x.processor[rule.Use]
	if !ok {
		return nil, fmt.Errorf("filter %s not found", rule.Use)
	}
	return processor.Inject(rabbit.NewFilterRuleBuilder(rule.Rule))
}

type AggregatorManager struct {
	rules     cache.Cache
	processor map[string]rabbit.Aggregator
}

func (x *AggregatorManager) Aggregator(ctx context.Context, ruleName string) (rabbit.Aggregator, error) {
	rule := &rabbit.MessageAggregationRule{}
	err := x.rules.Get(ctx, ruleName, rule)
	if err != nil {
		return nil, err
	}
	processor, ok := x.processor[rule.Use]
	if !ok {
		return nil, fmt.Errorf("aggregator %s not found", rule.Use)
	}
	return processor.Inject(rabbit.NewAggregationRuleBuilder(rule.Rule))
}

type TemplaterManager struct {
	rules     cache.Cache
	processor map[string]rabbit.Templater
}

func (x *TemplaterManager) Templater(ctx context.Context, ruleName string) (rabbit.Templater, error) {
	rule := &rabbit.MessageTemplateRule{}
	err := x.rules.Get(ctx, ruleName, rule)
	if err != nil {
		return nil, err
	}
	processor, ok := x.processor[rule.Use]
	if !ok {
		return nil, fmt.Errorf("templater %s not found", rule.Use)
	}
	return processor.Inject(rabbit.NewTemplateRuleBuilder(rule.Rule))
}

type SenderManager struct {
	rules     cache.Cache
	processor map[string]rabbit.Sender
}

func (x *SenderManager) Sender(ctx context.Context, ruleName string) (rabbit.Sender, error) {
	rule := &rabbit.MessageSendRule{}
	err := x.rules.Get(ctx, ruleName, rule)
	if err != nil {
		return nil, err
	}
	processor, ok := x.processor[rule.Use]
	if !ok {
		return nil, fmt.Errorf("sender %s not found", rule.Use)
	}
	return processor.Inject(rabbit.NewSendRuleBuilder(rule.Rule))
}
