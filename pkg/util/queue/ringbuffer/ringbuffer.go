package ringbuffer

import (
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type RingBuffer[T any] struct {
	data     []T
	capacity int

	mutex sync.Mutex
	index int

	ticker    *time.Ticker
	interval  time.Duration
	onTrigger func([]T)
	closed    chan struct{}
	helper    *log.Helper
}

// New create a new ring buffer
func New[T any](capacity int, interval time.Duration, logger log.Logger) *RingBuffer[T] {
	rb := &RingBuffer[T]{
		data:      make([]T, capacity),
		capacity:  capacity,
		interval:  interval,
		onTrigger: func([]T) {},
		closed:    make(chan struct{}),
		helper:    log.NewHelper(log.With(logger, "module", "ringbuffer")),
	}
	rb.startTicker()
	return rb
}

// RegisterOnTrigger register the on trigger function
func (rb *RingBuffer[T]) RegisterOnTrigger(onTrigger func([]T)) {
	rb.onTrigger = func(items []T) {
		defer func() {
			if err := recover(); err != nil {
				rb.helper.Errorw("method", "onTrigger", "err", err)
			}
		}()
		onTrigger(items)
	}
}

// Add add an item to the ring buffer
func (rb *RingBuffer[T]) Add(item T) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.data[rb.index] = item
	rb.index++

	if rb.index == rb.capacity {
		rb.flushLocked()
	}
}

// flushLocked flush the ring buffer when the lock is held
func (rb *RingBuffer[T]) flushLocked() {
	if rb.index == 0 {
		return
	}
	items := make([]T, rb.index)
	copy(items, rb.data[:rb.index])
	rb.index = 0

	go rb.onTrigger(items)
}

// Stop stop the ticker
func (rb *RingBuffer[T]) Stop() {
	close(rb.closed)
	if rb.ticker != nil {
		rb.ticker.Stop()
	}
	rb.data = nil
	rb.onTrigger = nil
	rb.closed = nil
	rb.ticker = nil
	rb.interval = 0
	rb.helper.Info("ringbuffer stopped")
}

// startTicker start the ticker
func (rb *RingBuffer[T]) startTicker() {
	if rb.interval <= 0 {
		rb.helper.Warnf("interval is 0, set to 10 seconds")
		rb.interval = 10 * time.Second
	}
	rb.ticker = time.NewTicker(rb.interval)
	go func() {
		for {
			select {
			case <-rb.ticker.C:
				rb.mutex.Lock()
				rb.flushLocked()
				rb.mutex.Unlock()
			case <-rb.closed:
				return
			}
		}
	}()
}
