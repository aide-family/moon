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
	w := watch.NewWatcher("test_watch", opts...)

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
