package manager

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aide-family/moon/pkg/rabbit"
	"github.com/aide-family/moon/pkg/rabbit/metrics"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2/klogr"
	"sync"
	"time"
)

var _ rabbit.ConfigGetter = &Manager{}

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

type Manager struct {
	name                 string
	MaxConcurrentWorkers int
	Started              bool
	Queue                rabbit.MessageQueue
	ctx                  context.Context
	Log                  *logr.Logger
	tm                   *TemplateManger
	rm                   *SuppressionRuleManager
	receivers            map[string]rabbit.Receiver
	senders              map[string]rabbit.Sender
}

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
		senders:              make(map[string]rabbit.Sender),
	}, nil
}

const (
	labelError         = "error"
	labelRequeue       = "requeue"
	labelPropose       = "propose"
	labelProposeFail   = "propose_fail"
	labelProposePass   = "propose_pass"
	labelProposeCancel = "propose_cancel"
	labelProposeFinish = "propose_finish"
)

func (m *Manager) Name() string {
	return m.name
}

func (m *Manager) initMetrics() {
	metrics.ActiveWorkers.WithLabelValues().Set(0)
	metrics.WorkerErrors.WithLabelValues().Add(0)
	metrics.WorkerCount.WithLabelValues().Set(float64(m.MaxConcurrentWorkers))
	metrics.WorkerTotal.WithLabelValues(labelError).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelRequeue).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelPropose).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelProposeFail).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelProposePass).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelProposeCancel).Add(0)
	metrics.WorkerTotal.WithLabelValues(labelProposeFinish).Add(0)
}

func (m *Manager) RegisterReceivers(receivers ...rabbit.Receiver) {
	for _, receiver := range receivers {
		m.receivers[receiver.Name()] = receiver
	}
}

func (m *Manager) RegisterSenders(senders ...rabbit.Sender) {
	for _, sender := range senders {
		m.senders[sender.Name()] = sender
	}
}

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
			go func(_ctx context.Context, group *sync.WaitGroup, ch <-chan *rabbit.Message, name string) {
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
	obj, ok := m.Queue.Next()
	if ok {
		// 当从消息队列中获取消息失败时，表明发生了意料之外的错误，
		// 此时，当前worker会退出。
		// TODO: 获取消息失败不代表队列关闭，此处功能耦合，需要进行优化
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
	err := snapshot.CompleteMessageSnapshot(ctx, m)
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

func (m *Manager) processor(ctx context.Context, log logr.Logger, message *rabbit.Message, processor *rabbit.Processor) {
	// 在每个 processor 中，首先应该对消息发起提案，进行表决，
	// 提案表决通过则进行下一步处理，反之则抛弃
	metrics.WorkerTotal.WithLabelValues(labelPropose).Inc()
	if processor.Suppressor.Propose(processor.Index) {
		metrics.WorkerTotal.WithLabelValues(labelProposePass).Inc()

		var err error
		defer func() {
			if err != nil {
				processor.Suppressor.Cancel(processor.Index)
				metrics.WorkerTotal.WithLabelValues(labelProposeCancel).Inc()
			} else {
				processor.Suppressor.Finish(processor.Index)
				metrics.WorkerTotal.WithLabelValues(labelProposeFinish).Inc()
			}
		}()

		// 对于通提案的消息，则使用模版进行解析，解析成功则进行下一步，否则放弃提案
		buf := &bytes.Buffer{}
		err = processor.Templater.Parse(message.Content, buf)
		if err != nil {
			log.Error(err, "template parsing error", "index", processor.Index)
			return
		}
		// 对与模版解析成功的，使用 Sender 进行发送，发送成功则提案完成，否则放弃提案
		err = processor.Sender.Send(ctx, buf.Bytes(), processor.Secret)
		if err != nil {
			log.Error(err, "send error", "index", processor.Index)
			return
		}
		log.V(5).Info("send successful", "index", processor.Index)
	} else {
		metrics.WorkerTotal.WithLabelValues(labelProposeFail).Inc()
		log.V(4).Info("propose fail", "index", processor.Index)
	}
}

func (m *Manager) GetSecret(context context.Context, id int64) ([]byte, error) {
	return m.tm.GetSecret(context, id)
}

func (m *Manager) GetTemplater(context context.Context, id int64) (rabbit.Templater, error) {
	return m.tm.Get(context, id)
}

func (m *Manager) GetSuppressorByTemplate(context context.Context, id int64) (rabbit.Suppressor, error) {
	suppressorID, err := m.tm.GetSuppressorID(context, id)
	if err != nil {
		return nil, err
	}
	return m.rm.Get(context, suppressorID)
}

func (m *Manager) GetSenderByTemplate(context context.Context, id int64) (rabbit.Sender, error) {
	senderName, err := m.tm.GetSenderName(context, id)
	if err != nil {
		return nil, err
	}
	return m.senders[senderName], nil
}
