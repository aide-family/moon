// Package state is the implementation package for the message task state.
package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
)

type MessageTask struct {
	NamespaceUID snowflake.ID `json:"namespace_uid"`
	MessageUID   snowflake.ID `json:"message_uid"`

	mu                  sync.Mutex
	nextStatus          enum.MessageStatus
	isNext              bool
	availableRetryCount int
	err                 error
}

func (m *MessageTask) AvailableRetryCount(count int) *MessageTask {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.availableRetryCount = count
	return m
}

func (m *MessageTask) Retry() *MessageTask {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.availableRetryCount--
	m.isNext = true
	m.nextStatus = enum.MessageStatus_PENDING
	return m
}

func (m *MessageTask) IsMaxRetry() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.availableRetryCount <= 0
}

func (m *MessageTask) NextStatus() enum.MessageStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nextStatus
}

func (m *MessageTask) IsNext() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isNext
}

func (m *MessageTask) SetNextStatus(nextStatus enum.MessageStatus) *MessageTask {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextStatus = nextStatus
	m.isNext = true
	return m
}

func (m *MessageTask) StopNext() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isNext = false
}

func (m *MessageTask) SetError(err error) *MessageTask {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
	m.isNext = true
	m.nextStatus = enum.MessageStatus_FAILED
	return m
}

func (m *MessageTask) Error() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.err
}

type ProcessFunc func(ctx context.Context, task *MessageTask) error

/*
MessageTaskState state transition rule matrix (reference design document)

	+-------------------+----------+------------+-------------+----------+---------+
	| current status    | Start    | SendSuccess| SendFailure | Cancel   | Retry   |
	+-------------------+----------+------------+-------------+----------+---------+
	| PENDING           | SENDING  | -          | -           | CANCELLED| -       |
	| SENDING           | -        | SENT       | FAILED      | CANCELLED| -       |
	| FAILED            | -        | -          | -           | CANCELLED| PENDING |
	| SENT              | terminated state (reject all events)                     |
	| CANCELLED         | terminated state (reject all events)                     |
	| UNKNOWN           | only allowed Start → PENDING (initialization)            |
	+-------------------+----------+------------+-------------+----------+---------+
*/
type MessageTaskState struct {
	processFunc ProcessFunc
	status      enum.MessageStatus
	timeout     time.Duration
}

func NewMessageTaskState(status enum.MessageStatus, timeout time.Duration) MessageTaskState {
	processFunc, ok := GetMessageTaskProcess(status)
	if !ok {
		panic(fmt.Sprintf("message status %s state process func not found", status))
	}
	return MessageTaskState{
		processFunc: processFunc,
		status:      status,
		timeout:     timeout,
	}
}

func (m *MessageTaskState) Process(task *MessageTask) {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()
	ctx = contextx.WithNamespace(ctx, task.NamespaceUID)
	if err := m.processFunc(ctx, task); err != nil {
		klog.Errorw("msg", "process message task failed", "error", err, "task", task, "status", m.status)
		return
	}
	if task.IsNext() {
		if nextState, ok := GetMessageTaskState(task.NextStatus()); ok {
			nextState.Process(task)
		}
	}
}
