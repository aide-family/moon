package ringbuffer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/util/queue/ringbuffer"
)

func TestRingBuffer(t *testing.T) {
	rb, err := ringbuffer.New[string](10, 5, 3*time.Second, func(items []string) {
		fmt.Println("Flushed:", items)
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for i := 0; i < 12; i++ {
		rb.Add(fmt.Sprintf("item-%d", i))
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)
	rb.Stop()
}
