# watch服务做什么

它主要用来监听某一类事件或者信息，然后能根据不同的信息类型完成对应的处理。

那么我们大概需要如下东西

* 消息体
  * 用于维护消息的类型，内容等，如果需要重试，还可以维护消息的重试次数、重试逻辑等
* 消息队列
  * 用于消息的发送接收，也用于消息削峰等，给watch提供一个持续的消息流
* 消息处理器
  * 根据消息类型匹配具体的处理器完成逻辑处理
* 消息存储器
  * 持久化存储消息，避免消息丢失，也可以用于缓存等
* 消息编码器
  *  消息编码器，用于序列化消息，方便存储和消费

## 消息索引

定义不同消息的索引规则， 方便持久化存储和消费

```go
// Indexer 索引器
type Indexer interface {
    Index() string
}
```

## 消息体

```go
// Message watch 消息结构体
type Message struct {
	// 并发锁结构， 避免并发读写问题
    lock sync.Mutex

    // 传输的消息内容， 由用户自定义
    data Indexer

    // 消息类型， 如需要增加新的类型，去vobj包增加
    topic vobj.Topic

    // 注册编码器
    schema Schemer

    // 重试次数
    retry int

    // 最大消息重试次数
    retryMax int

    // 是否已经处理过
    handleCtx context.Context
}
```

## 消息队列

```go
// Queue 消息队列
type Queue interface {
    // Next 获取下一个消息
    Next() (*Message, bool)
    
    // Push 添加消息
    Push(msg *Message) error
    
    // Close 关闭队列
    Close() error
    
    // Len 获取队列长度
    Len() int
    
    // Clear 清空队列
    Clear()
}
```

## 存储器

```go
// Storage 存储器
type Storage interface {
    // Get 获取消息
    Get(index Indexer) *Message

    // Put 放入消息
    Put(msg *Message) error

    // Clear 清空消息
    Clear()

    // Remove 移除消息
    Remove(index Indexer)

    // Close 关闭存储器
    Close() error

    // Len 长度
    Len() int

    // Range 遍历
    //  f返回值为bool类型，如果返回false，则停止range
    Range(f func(index Indexer, msg *Message) bool)
}
```

## 消息处理器

```go
// Handler 消息处理
type Handler interface {
    // Handle 处理消息
    //
    // 	ctx 上下文
    // 	msg 消息
    Handle(ctx context.Context, msg *Message) error
}
```

## 消息编码器

```go
type Schemer interface {
    // Decode 解码
    Decode(in *Message, out any) error

    // Encode 编码
    Encode(in *Message, out any) error
}
```

## watch服务定义

以上完成了基础数据结构定义，现在我们需要吧这些功能组合起来，完成我们的watch功能

```go
type Watcher struct {
    // 停止监听的通道
    stopCh chan struct{}
    // 存储器
    storage Storage
    // 消息队列
    queue Queue
    // 消息处理器
    handler Handler
    // 超时时间
    timeout time.Duration
}


func (w *Watcher) Start(_ context.Context) error {
  go func() {
    defer after.RecoverX()
    for {
      select {
        case <-w.stopCh:
          log.Infow("method", "stop watcher")
          w.clear()
          return
        default:
          if types.IsNil(w.queue) {
            log.Warnw("method", "queue is empty")
            continue
          }
          w.reader()
      }
    }
  }()
  return nil
}

func (w *Watcher) Stop(_ context.Context) error {
  w.stopCh <- struct{}{}
  return nil
}

// clear 清理资源
func (w *Watcher) clear() {
  if !types.IsNil(w.queue) {
    if err := w.queue.Close(); err != nil {
      log.Errorw("method", "close queue error", "error", err)
    }
  }
  
  if !types.IsNil(w.storage) {
    if err := w.storage.Close(); err != nil {
        log.Errorw("method", "close storage error", "error", err)
    }
  }
  
  close(w.stopCh)
  log.Infow("method", "clear resources", "res", "done")
}

// retry 重试
func (w *Watcher) retry(msg *Message) {
  if msg.GetRetry() >= msg.GetRetryMax() {
    // 重试次数超过最大次数不再重试
    return
  }
  // 消息重试次数+1
  msg.RetryInc()
  if err := w.queue.Push(msg); err != nil {
    log.Errorw("method", "push message to queue error", "error", err)
  }
}

func (w *Watcher) reader() {
  msg, ok := w.queue.Next()
  if !ok {
    return
  }
  
  if !types.IsNil(w.handler) {
    // 递交消息给处理器，由处理器决定消息去留， 如果失败，会进入重试逻辑
    ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
    defer cancel()
    if err := w.handler.Handle(ctx, msg); err != nil {
      log.Errorw("method", "handle message error", "error", err)
      w.retry(msg)
      return
    }
  }
  
  if !types.IsNil(w.storage) {
    // 存储消息
    if err := w.storage.Put(msg); err != nil {
      log.Errorw("method", "put message to storage error", "error", err)
      w.retry(msg)
      return
    }
  }
}
```

watch调用start方法后启动watch监听，监听到消息后，会调用handler处理消息，处理完成后，如果失败会根据消息的配置，决定是否需要重试。处理成功会加入存储器，如果失败会进入重试逻辑。

在重试逻辑中，判断消息是否已经达到最大的重试次数， 如果没有达到，则重新入队， 重试次数+1.

watch提供了stop方法，可以由调用方去通过系统信号或者其他方式停止监听服务，watch停止前会先清理依赖的队列、存储等、然后退出监听协程。

watch的实现平实简单，但是需要考虑很多细节，比如消息重试、消息存储、消息队列等，这些细节需要根据具体的业务场景进行设计，比如消息重试的次数、消息存储的存储方式、消息队列的实现方式等。

## 测试

```go
package watch_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
	"github.com/go-kratos/kratos/v2/log"
)

type MyMsg struct {
	Data int
}

func (m *MyMsg) Index() string {
	return fmt.Sprintf("my-msg-%d", m.Data)
}

func msgHandle(ctx context.Context, msg *watch.Message) error {
	log.Debugw("default handler", msg.GetData())

	if err := msg.GetSchema().Encode(msg, msg.GetData()); err != nil {
		log.Errorw("method", "Encode", "err", err)
	}
	if err := msg.GetSchema().Decode(msg, msg.GetData()); err != nil {
		log.Errorw("method", "Decode", "err", err)
	}
	d := msg.GetData().(*MyMsg)
	if d.Data%3 == 0 {
		return errors.New("模拟错误， 检测重试")
	}
	return nil
}

func TestNewWatcher(t *testing.T) {
	defaultQueue := watch.NewDefaultQueue(100)
	defaultStorage := watch.NewDefaultStorage()

	opts := []watch.WatcherOption{
		watch.WithWatcherQueue(defaultQueue),
		watch.WithWatcherStorage(defaultStorage),
		watch.WithWatcherTimeout(3 * time.Second),
		watch.WithWatcherHandler(watch.NewDefaultHandler(
			watch.WithDefaultHandlerTopicHandle(vobj.TopicUnknown, msgHandle),
		)),
	}
	w := watch.NewWatcher(opts...)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w.Start(ctx)

	msgCount := 100
	schema := watch.NewDefaultSchemer()
	msgOpts := []watch.MessageOption{
		watch.WithMessageSchema(schema),
		watch.WithMessageRetryMax(3),
	}
	go func() {
		for msgCount > 0 {
			time.Sleep(1 * time.Second) // 延时1秒发送
			value := msgCount
			msg := watch.NewMessage(&MyMsg{Data: value}, vobj.TopicUnknown, msgOpts...)
			msgCount--
			if err := w.GetQueue().Push(msg); err != nil {
				continue
			}
		}
	}()

	go func() {
		for {
			log.Infow("默认存储的数据长度", w.GetStorage().Len())
			time.Sleep(3 * time.Second)
		}
	}()

	time.Sleep(10 * time.Second)
	w.Stop(context.Background())
}

```

系统实现一些简单的消息队列、存储器等作为默认实现， 可以通过配置或者参数的方式来替换默认实现，比如替换队列实现为kafka、替换存储器实现为redis等。

在handle中，我们假定每3次就会失败， 依次来检测重试逻辑。

```bash
INFO 默认存储的数据长度=0
DEBUG default handler=&{100}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
DEBUG default handler=&{99}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{99}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{99}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{99}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
INFO 默认存储的数据长度=1
DEBUG default handler=&{98}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
DEBUG default handler=&{97}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
DEBUG default handler=&{96}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{96}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{96}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{96}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
INFO 默认存储的数据长度=3
DEBUG default handler=&{95}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
DEBUG default handler=&{94}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
DEBUG default handler=&{93}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{93}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{93}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
DEBUG default handler=&{93}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
ERROR method=handle message error error=模拟错误， 检测重试
INFO 默认存储的数据长度=5
DEBUG default handler=&{92}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
DEBUG default handler=&{91}
ERROR method=Encode err=rpc error: code = Unimplemented desc = encode unimplemented topic: 未知
ERROR method=Decode err=rpc error: code = Unimplemented desc = decode unimplemented topic: 未知
INFO method=stop watcher
INFO method=clear resources res=done
```

