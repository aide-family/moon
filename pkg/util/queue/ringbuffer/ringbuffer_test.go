package ringbuffer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/util/queue/ringbuffer"
)

func TestRingBuffer(t *testing.T) {
	rb := ringbuffer.New[string](5, 0*time.Second, log.DefaultLogger)
	rb.RegisterOnTrigger(func(items []string) {
		log.Infof("%s Flushed: %v", time.Now().Format("15:04:05"), items)
	})

	for i := 0; i < 12; i++ {
		rb.Add(fmt.Sprintf("item-%d", i))
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)
	rb.Stop()
}
