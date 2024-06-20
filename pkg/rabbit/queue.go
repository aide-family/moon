package rabbit

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/aide-family/moon/api/rabbit/rule"

	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// MessageQueue 消息队列接口，用来处理信息
//
//	队列中，至多允许一个相同 key 的数据存在。
//	加入时，如果数据未被消费，则会更新数据，如果是延时加入，则以第一次加入的为准。
//	如果已经被消费，则会重新创建一个key，然后将数据存储起来。
type MessageQueue interface {
	// Add 用来将一条数据加入到队列中
	Add(item *rule.Message)

	// TryAgain 尝试将数据加入再次加入到队列
	//
	//  尝试加入队列时，会通过即将加入的 item 的去计算再次加入的时间，如果超过限制，则会被抛弃
	TryAgain(item *QueueInfo) bool

	// Get 用来顺序获取消息队列中的一条消息数据
	//
	//  如果队列中没有数据，则会阻塞，直到有数据入队,队列关闭则返回false
	Get() (*QueueInfo, bool)

	// Done 用来销毁队列中的一条数据
	//
	// 当 Done 被调用时，表示这条数据已经处理完成，不需要重新入队，它将被完全移除，
	// 正常情况下，我们每次处理完成数据，都应该调用这个方法，以保证队列中没有溢出的数据产生。
	Done(item *QueueInfo)

	// ShutDown 用来关闭队列
	//
	// 当 ShutDown 被调用时，队列将直接关闭，无论队列中的数据是否被消费完成
	ShutDown()
	// ShuttingDown 用于判断队列是否关闭
	ShuttingDown() bool
	// ShutDownWithDrain 用来关闭队列
	//
	// 当 ShutDownWithDrain 被调用时，队列将在数据被消费完成后关闭
	ShutDownWithDrain()
	Len() int
}

var _ MessageQueue = &PriorityQueue{}

type PriorityQueue struct {
	// 退避管理器，消息再次加入时，用来决定消息重新入队的时间
	backoff *Backoff
	//queue 用来存储消息的索引
	queue workqueue.DelayingInterface
	// dirty 用来存储消息
	dirty *dirty
	// runner 用来存储正在处理的消息数据
	runner set
	// maxRetries 表示数据最多尝试的入队次数，超过这个次数，数据将被丢弃
	maxRetries int
}

func NewQueue(maxRetries int, backoff *Backoff) *PriorityQueue {
	if backoff == nil {
		backoff = DefaultBackOff
	}
	if maxRetries <= 0 {
		maxRetries = DefaultMaxReties
	}
	return &PriorityQueue{
		backoff:    backoff,
		queue:      workqueue.NewDelayingQueue(),
		dirty:      &dirty{},
		runner:     set{v: map[interface{}]empty{}},
		maxRetries: maxRetries,
	}
}

// Add 往队列里面添加一条消息数据
//
//	加入对象是 *Message
//	case 1：对象断言失败，则添加失败
//	case 2：调用 tryAdd 尝试添加数据到队列
func (p *PriorityQueue) Add(msg *rule.Message) {
	p.tryAdd(msg)
}

// tryAdd 尝试往队列中添加数据
func (p *PriorityQueue) tryAdd(msg *rule.Message) {
	info := InitQueuedMessageInfo(msg)

	p.addMessageInfo(info, 0)
}

// TryAgain 用来再次往队列里面添加数据
func (p *PriorityQueue) TryAgain(info *QueueInfo) bool {
	duration := info.NextAttempt(p.backoff)
	// 尝试次数过多，丢弃该key
	if info.Attempts > p.maxRetries {
		return false
	}
	p.addMessageInfo(info, duration)
	return true
}

// AddMessageInfo 往队列里面添加一条消息数据
// case 1：对象已经被添加，则更新对象
// case 2：对象未被添加，则往队列索引里新增key，往dirty里存储数据
func (p *PriorityQueue) addMessageInfo(info *QueueInfo, duration time.Duration) {
	if !p.dirty.has(info.Key) {
		if duration == 0 {
			klog.V(2).Infof("add message %s", info.Key)
			p.queue.Add(info.Key)
		} else {
			klog.V(2).Infof("add message %s after %.2f seconds", info.Key, duration.Seconds())
			p.queue.AddAfter(info.Key, duration)
		}
	}
	p.dirty.update(info)
}

// Get 获取下一个需要处理的消息
func (p *PriorityQueue) Get() (*QueueInfo, bool) {
	for {
		// 从索引队列中获取消息的索引，获取索引失败则表示消息队列关闭
		data, shutdown := p.queue.Get()
		if shutdown {
			return nil, false
		}

		p.queue.Done(data)

		key, ok := data.(string)
		if !ok {
			panic(fmt.Sprintf("Invalid queue key: %+v", data))
		}

		info, ok := p.dirty.loadAndDelete(key)
		if !ok {
			panic(fmt.Sprintf("Invalid queue info, key:%s", key))
		}

		// 检查该消息是否正在处理
		// case 1: 是就将消息重新入队，继续循环
		// case 2: 否就记录key，并返回数据
		if !p.runner.add(key) {
			p.TryAgain(info)
			continue
		} else {
			klog.V(2).Infof("pop message %s from queue", key)
			return info, true
		}
	}
}

func (p *PriorityQueue) Done(item *QueueInfo) {
	p.runner.delete(item.Key)
	p.queue.Done(item.Key)
}

func (p *PriorityQueue) Len() int {
	return p.queue.Len()
}

func (p *PriorityQueue) ShutDown() {
	p.queue.ShutDown()
}

func (p *PriorityQueue) ShuttingDown() bool {
	return p.queue.ShuttingDown()
}

func (p *PriorityQueue) ShutDownWithDrain() {
	p.queue.ShutDownWithDrain()
}

type dirty struct {
	sync.Map
}

func (d *dirty) has(key string) bool {
	_, ok := d.get(key)
	return ok
}

func (d *dirty) loadAndDelete(key string) (*QueueInfo, bool) {
	value, ok := d.LoadAndDelete(key)
	if ok {
		return value.(*QueueInfo), ok
	}
	return nil, false
}

func (d *dirty) get(key string) (*QueueInfo, bool) {
	value, ok := d.Load(key)
	if ok {
		return value.(*QueueInfo), ok
	}
	return nil, false
}

func (d *dirty) delete(key string) {
	d.Delete(key)
}

func (d *dirty) update(info *QueueInfo) {
	value, ok := d.Load(info.Key)
	if ok {
		old := value.(*QueueInfo)
		old.UpdateMessage(info.Message)
		d.Store(info.Key, old)
	} else {
		d.Store(info.Key, info)
	}
}

func (d *dirty) add(info *QueueInfo) {
	d.Store(info.Key, info)
}

type empty struct{}
type set struct {
	lock sync.RWMutex
	v    map[interface{}]empty
}

func (s *set) has(item interface{}) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, exists := s.v[item]
	return exists
}

func (s *set) add(item interface{}) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, exists := s.v[item]; exists {
		return false
	}
	s.v[item] = empty{}
	return true
}

func (s *set) delete(item interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.v, item)
}

const (
	// DefaultMaxBackoffDuration = DefaultInitialBackoffDuration * math.Pow(DefaultBackoffFactor, DefaultBackoffTimes)
	// 默认最大退避时间，即需要处理的消息最长将在多久之后加入消息队列，
	// 退避时间=退避时间*退避因子^退避次数
	DefaultMaxBackoffDuration     time.Duration = 5 * time.Minute
	DefaultInitialBackoffDuration               = 10 * time.Second
	DefaultBackoffTimes           int           = 5
	DefaultBackoffFactor                        = 2
	DefaultBackoffJitter          float64       = 0
	DefaultMaxReties              int           = 5
)

var (
	DefaultBackOff = &Backoff{
		Duration: DefaultInitialBackoffDuration,
		Factor:   DefaultBackoffFactor,
		Jitter:   DefaultBackoffJitter,
		Steps:    DefaultBackoffTimes,
		Cap:      DefaultMaxBackoffDuration,
	}
)

// Backoff struct copy from "k8s.io/apimachinery/pkg/util/wait"
type Backoff struct {
	// The initial duration.
	Duration time.Duration
	// Duration is multiplied by factor each iteration, if factor is not zero
	// and the limits imposed by Steps and Cap have not been reached.
	// Should not be negative.
	// The jitter does not contribute to the updates to the duration parameter.
	Factor float64
	// The sleep at each iteration is the duration plus an additional
	// amount chosen uniformly at random from the interval between
	// zero and `jitter*duration`.
	Jitter float64
	// The remaining number of iterations in which the duration
	// parameter may change (but progress can be stopped earlier by
	// hitting the cap). If not positive, the duration is not
	// changed. Used for exponential backoff in combination with
	// Factor and Cap.
	Steps int
	// A limit on revised values of the duration parameter. If a
	// multiplication by the factor parameter would make the duration
	// exceed the cap then the duration is set to the cap and the
	// steps parameter is set to zero.
	Cap time.Duration
}

func (b *Backoff) Next(attempts int) time.Duration {
	// 退避次数小于0 返回默认退避时间
	if b.Steps < 1 {
		if b.Jitter > 0 {
			return Jitter(b.Duration, b.Jitter)
		}
		return b.Duration
	}

	duration := b.Duration

	if b.Factor != 0 {
		// 判断尝试次数是否大于等于最大退避次数
		if b.Steps <= attempts {
			attempts = b.Steps
		}
		// 计算影响因子
		factor := math.Pow(b.Factor, float64(attempts))
		duration = time.Duration(float64(b.Duration) * factor)
		if b.Cap > 0 && duration > b.Cap {
			duration = b.Cap
		}
	}

	if b.Jitter > 0 {
		duration = Jitter(duration, b.Jitter)
	}

	return duration
}

func Jitter(duration time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
	return wait
}

// QueueInfo 是一个消息包装器，其中包含与进程队列中的 Message 状态相关的附加信息，例如添加到队列时的时间戳。
type QueueInfo struct {
	// 消息的Key
	Key string
	// Message 原始消息
	Message *rule.Message
	// 消息添加到队列中的时间。
	// 随着每次加入变更
	Timestamp time.Time
	// 成功处理前的尝试次数
	// 它用于记录尝试次数指标和退避管理
	Attempts int
	// InitialAttemptTimestamp 是消息首次加入队列的时间。
	// 在成功处理之前，该消息可能会被多次添加回队列。
	// 初始化后不应更新。
	InitialAttemptTimestamp time.Time
}

// InitQueuedMessageInfo 初始化需要加入到消息队列的信息
func InitQueuedMessageInfo(message *rule.Message) *QueueInfo {
	first := time.Now()
	return &QueueInfo{
		Key:                     message.Id,
		Message:                 message,
		Timestamp:               first,
		Attempts:                0,
		InitialAttemptTimestamp: first,
	}
}

func (p *QueueInfo) UpdateMessage(message *rule.Message) {
	if message != nil {
		p.Message = message
	}
}

// NextAttempt 返回下一次尝试的时间
func (p *QueueInfo) NextAttempt(backoff *Backoff) time.Duration {
	p.Attempts++
	duration := backoff.Next(p.Attempts)
	p.Timestamp = time.Now().Add(duration)
	return duration
}
