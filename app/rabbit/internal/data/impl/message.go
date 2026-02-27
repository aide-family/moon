package impl

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/conf"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl/query"
	"github.com/aide-family/rabbit/internal/data/impl/state"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
	"github.com/aide-family/rabbit/pkg/message"
)

func NewMessageRepository(
	c *conf.Bootstrap,
	d *data.Data,
	messageLogRepo repository.MessageLog,
) (repository.Message, error) {
	jobCore := c.GetJobCore()
	repo := &messageRepository{
		Data:           d,
		messageLogRepo: messageLogRepo,
		clusters:       safety.NewSyncMap(make(map[string]ClusterSender)),
		messageChan:    make(chan *state.MessageTask, jobCore.GetBufferSize()),
		stopChan:       make(chan struct{}),
		workerTotal:    int(jobCore.GetWorkerTotal()),
		maxRetryCount:  int(jobCore.GetMaxRetry()),
		timeout:        jobCore.GetTimeout().AsDuration(),
		wg:             sync.WaitGroup{},
	}
	state.RegisterMessageTaskProcess(enum.MessageStatus_MessageStatus_UNKNOWN, repo.unknownMessageTaskProcess)
	state.RegisterMessageTaskProcess(enum.MessageStatus_PENDING, repo.pendingMessageTaskProcess)
	state.RegisterMessageTaskProcess(enum.MessageStatus_SENDING, repo.sendingMessageTaskProcess)
	state.RegisterMessageTaskProcess(enum.MessageStatus_SENT, repo.sentMessageTaskProcess)
	state.RegisterMessageTaskProcess(enum.MessageStatus_FAILED, repo.failedMessageTaskProcess)
	state.RegisterMessageTaskProcess(enum.MessageStatus_CANCELLED, repo.cancelledMessageTaskProcess)
	for _, value := range enum.MessageStatus_value {
		status := enum.MessageStatus(value)
		state.RegisterMessageTaskState(status, state.NewMessageTaskState(status, repo.timeout))
	}

	query.SetDefault(d.DB())
	if err := repo.initClusters(c.GetJobClusters()); err != nil {
		return nil, err
	}
	if err := repo.Start(context.Background()); err != nil {
		return nil, err
	}
	d.AppendClose("messageRepo", func() error { return repo.Stop(context.Background()) })
	return repo, nil
}

type messageRepository struct {
	messageLogRepo repository.MessageLog
	stopChan       chan struct{}
	messageChan    chan *state.MessageTask
	wg             sync.WaitGroup
	workerTotal    int
	timeout        time.Duration
	clusters       *safety.SyncMap[string, ClusterSender]
	maxRetryCount  int
	*data.Data
}

func (m *messageRepository) unknownMessageTaskProcess(ctx context.Context, task *state.MessageTask) error {
	changed, err := m.messageLogRepo.UpdateMessageLogStatusIf(ctx, task.MessageUID, enum.MessageStatus_MessageStatus_UNKNOWN, enum.MessageStatus_PENDING)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	task.SetNextStatus(enum.MessageStatus_PENDING)
	return nil
}

func (m *messageRepository) pendingMessageTaskProcess(ctx context.Context, task *state.MessageTask) error {
	changed, err := m.messageLogRepo.UpdateMessageLogStatusSendingIf(ctx, task.MessageUID, enum.MessageStatus_PENDING)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	task.SetNextStatus(enum.MessageStatus_SENDING)
	return nil
}

func (m *messageRepository) sendingMessageTaskProcess(ctx context.Context, task *state.MessageTask) (err error) {
	defer func() {
		if err != nil {
			changed, _err := m.messageLogRepo.UpdateMessageLogLastErrorIf(ctx, task.MessageUID, enum.MessageStatus_SENDING, err.Error())
			if _err != nil {
				err = errors.Join(err, _err)
			}
			if changed {
				task.SetNextStatus(enum.MessageStatus_FAILED)
			}
		}
	}()
	messageLog, err := m.messageLogRepo.GetMessageLogWithLock(ctx, task.MessageUID)
	if err != nil {
		return err
	}
	driver, ok := message.GetDriver(messageLog.MessageType)
	if !ok {
		return merr.ErrorInternalServer("message driver not found, message type %s", messageLog.MessageType)
	}
	msgConfig, err := messageLog.ToMessageConfig()
	if err != nil {
		return merr.ErrorInternalServer("message config convert failed, message type %s, error %v", messageLog.MessageType, err)
	}
	sender, err := driver(msgConfig)
	if err != nil {
		return merr.ErrorInternalServer("message driver create failed, message type %s, error %v", messageLog.MessageType, err)
	}
	msg := message.NewMessage(messageLog.MessageType, []byte(messageLog.Message))
	if err = sender.Send(ctx, msg); err != nil {
		return err
	}
	changed, err := m.messageLogRepo.UpdateMessageLogStatusSuccessIf(ctx, task.MessageUID)
	if err != nil {
		return err
	}
	if changed {
		task.SetNextStatus(enum.MessageStatus_SENT)
	}

	return nil
}

func (m *messageRepository) sentMessageTaskProcess(ctx context.Context, task *state.MessageTask) error {
	task.StopNext()
	return nil
}

func (m *messageRepository) cancelledMessageTaskProcess(ctx context.Context, task *state.MessageTask) error {
	task.StopNext()
	return nil
}

func (m *messageRepository) failedMessageTaskProcess(ctx context.Context, task *state.MessageTask) error {
	if task.IsMaxRetry() {
		task.StopNext()
		return nil
	}

	if err := m.messageLogRepo.MessageLogRetryIncrement(ctx, task.MessageUID); err != nil {
		return err
	}

	task.Retry()
	return nil
}

// AppendMessage implements [repository.Message].
func (m *messageRepository) AppendMessage(ctx context.Context, messageUID snowflake.ID) error {
	task := &state.MessageTask{
		NamespaceUID: contextx.GetNamespace(ctx),
		MessageUID:   messageUID,
	}
	return m.appendMessageToChannel(ctx, task)
}

func (m *messageRepository) appendMessageToChannel(ctx context.Context, task *state.MessageTask) error {
	task.AvailableRetryCount(m.maxRetryCount)
	select {
	case m.messageChan <- task:
		klog.Context(ctx).Debugw("msg", "append message success", "task", task)
		return nil
	default:
		klog.Context(ctx).Debugw("msg", "append message channel full", "task", task)
		if task.IsMaxRetry() {
			klog.Context(ctx).Warnw("msg", "append message retry count reached max", "task", task)
			return nil
		}

		for _, cluster := range m.clusters.Values() {
			if err := cluster.Send(ctx, task.MessageUID); err != nil {
				klog.Warnw("msg", "send message to cluster failed", "error", err, "cluster", cluster)
			}
		}
		klog.Context(ctx).Debugw("msg", "append message retry", "task", task)
		task.Retry()
		return m.appendMessageToChannel(ctx, task)
	}
}

func (m *messageRepository) worker(number int) {
	for {
		select {
		case <-m.stopChan:
			klog.Debugw("msg", "message worker stopped", "workerIndex", number)
			return
		case task := <-m.messageChan:
			m.processMessageTask(task)
		}
	}
}

func (m *messageRepository) processMessageTask(task *state.MessageTask) {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	ctx = contextx.WithNamespace(ctx, task.NamespaceUID)
	messageLog, err := m.messageLogRepo.GetMessageLogWithLock(ctx, task.MessageUID)
	if err != nil {
		klog.Warnw("msg", "get message log failed", "error", err, "messageUID", task.MessageUID)
		return
	}
	messageTaskState, ok := state.GetMessageTaskState(messageLog.Status)
	if !ok {
		klog.Warnw("msg", "message task state not found", "status", messageLog.Status)
		return
	}
	messageTaskState.Process(task)
}

// Start implements [repository.Message].
func (m *messageRepository) Start(ctx context.Context) error {
	klog.Infow("msg", "start message worker", "workerTotal", m.workerTotal)
	for i := 0; i < m.workerTotal; i++ {
		index := i
		m.wg.Go(func() {
			m.worker(index)
		})
	}
	return nil
}

// Stop implements [repository.Message].
func (m *messageRepository) Stop(_ context.Context) error {
	close(m.stopChan)
	m.wg.Wait()
	klog.Infow("msg", "message worker stopped")
	for _, cluster := range m.clusters.Values() {
		if err := cluster.Close(); err != nil {
			klog.Warnw("msg", "close cluster failed", "error", err, "name", cluster.Name())
		}
		klog.Debugw("msg", "close cluster success", "name", cluster.Name())
	}
	klog.Infow("msg", "close clusters success")
	return nil
}

type ClusterSender interface {
	Name() string
	Send(ctx context.Context, messageUID snowflake.ID) error
	Close() error
}

func NewClusterSender(name string, sendFunc func(ctx context.Context, req *apiv1.SendMessageRequest) (*apiv1.SendReply, error), closeFunc func() error) ClusterSender {
	return &clusterSender{
		name:      name,
		sendFunc:  sendFunc,
		closeFunc: closeFunc,
	}
}

type clusterSender struct {
	name      string
	sendFunc  func(ctx context.Context, req *apiv1.SendMessageRequest) (*apiv1.SendReply, error)
	closeFunc func() error
}

func (c *clusterSender) Name() string {
	return c.name
}

func (c *clusterSender) Send(ctx context.Context, messageUID snowflake.ID) error {
	req := &apiv1.SendMessageRequest{
		Uid: messageUID.Int64(),
	}
	_, err := c.sendFunc(ctx, req)
	return err
}

func (c *clusterSender) Close() error {
	return c.closeFunc()
}

func (m *messageRepository) initClusters(c *config.ClusterConfig) error {
	clusterEndpoints, clusterTimeout, clusterName, protocol := strutil.SplitSkipEmpty(c.GetEndpoints(), ","), c.GetTimeout().AsDuration(), c.GetName(), c.GetProtocol().String()
	for _, clusterEndpoint := range clusterEndpoints {
		opts := []connect.InitOption{
			connect.WithDiscovery(m.Registry()),
		}
		initConfig := connect.NewDefaultConfig(clusterName, clusterEndpoint, clusterTimeout, protocol)

		name := strings.Join([]string{clusterName, clusterEndpoint}, ":")
		var clusterSender ClusterSender
		switch protocol {
		case connect.ProtocolHTTP:
			httpClient, err := connect.InitHTTPClient(initConfig, opts...)
			if err != nil {
				klog.Warnw("msg", "create HTTP client failed", "endpoint", clusterEndpoint, "error", err)
				return err
			}

			httpSender := apiv1.NewSenderHTTPClient(httpClient)
			clusterSender = NewClusterSender(name, func(ctx context.Context, req *apiv1.SendMessageRequest) (*apiv1.SendReply, error) {
				return httpSender.SendMessage(ctx, req)
			}, httpClient.Close)
		case connect.ProtocolGRPC:
			grpcClient, err := connect.InitGRPCClient(initConfig, opts...)
			if err != nil {
				klog.Warnw("msg", "create GRPC client failed", "endpoint", clusterEndpoint, "error", err)
				return err
			}
			grpcSender := apiv1.NewSenderClient(grpcClient)
			clusterSender = NewClusterSender(name, func(ctx context.Context, req *apiv1.SendMessageRequest) (*apiv1.SendReply, error) {
				return grpcSender.SendMessage(ctx, req)
			}, grpcClient.Close)
		default:
			klog.Warnw("msg", "unknown protocol", "endpoint", clusterEndpoint, "protocol", protocol)
			return merr.ErrorInternalServer("unknown protocol: %s", protocol)
		}

		m.clusters.Set(name, clusterSender)
	}

	return nil
}
